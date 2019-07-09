/*
Copyright 2017 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"context"
	"strconv"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/kapacitor"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"

	"github.com/ghodss/yaml"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kubeapi "k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func runAlertsWatcher(ctx context.Context, kubernetesClient *kubernetes.Client, retryC chan<- func() error) error {
	kapacitorClient, err := kapacitor.NewClient()
	if err != nil {
		return trace.Wrap(err)
	}

	alertLabel, err := kubernetes.MatchLabel(constants.MonitoringLabel, constants.MonitoringUpdateAlert)
	if err != nil {
		return trace.Wrap(err)
	}

	targetLabel, err := kubernetes.MatchLabel(constants.MonitoringLabel, constants.MonitoringUpdateAlertTarget)
	if err != nil {
		return trace.Wrap(err)
	}

	smtpLabel, err := kubernetes.MatchLabel(constants.MonitoringLabel, constants.MonitoringUpdateSMTP)
	if err != nil {
		return trace.Wrap(err)
	}

	alertCh := make(chan kubernetes.ConfigMapUpdate)
	alertTargetCh := make(chan kubernetes.ConfigMapUpdate)
	configmaps := []kubernetes.ConfigMap{
		{alertLabel, alertCh},
		{targetLabel, alertTargetCh},
	}
	smtpCh := make(chan kubernetes.SecretUpdate)

	go kubernetesClient.WatchConfigMaps(ctx, configmaps...)
	go kubernetesClient.WatchSecrets(ctx, kubernetes.Secret{smtpLabel, smtpCh})
	receiverLoop(ctx, kubernetesClient.Clientset, kapacitorClient,
		alertCh, alertTargetCh, smtpCh, retryC)

	return nil
}

func receiverLoop(ctx context.Context, kubeClient *kubeapi.Clientset, kClient *kapacitor.Client,
	alertCh, alertTargetCh <-chan kubernetes.ConfigMapUpdate,
	smtpCh <-chan kubernetes.SecretUpdate, retryC chan<- func() error) {
	for {
		select {
		case update := <-alertCh:
			log := logrus.WithField("configmap", update.ResourceUpdate.Meta())
			spec := []byte(update.Data[constants.ResourceSpecKey])
			if isSpecEmpty(spec) {
				log.Error("empty configuration")
				continue
			}
			switch update.EventType {
			case watch.Added, watch.Modified:
				handler := func() error {
					return createAlert(kClient, spec, log)
				}
				err := handler()
				if err == nil {
					// Success - no need to retry
					break
				}
				log.WithError(err).Warnf("failed to create alert from spec %s.", spec)
				select {
				case retryC <- handler:
				// Queue handler on retry list
				case <-ctx.Done():
				}
			}
		case update := <-smtpCh:
			log := logrus.WithField("secret", update.ResourceUpdate.Meta())
			spec := update.Data[constants.ResourceSpecKey]
			if isSpecEmpty(spec) {
				log.Error("empty configuration")
				continue
			}
			client := kubeClient.CoreV1().Secrets(constants.MonitoringNamespace)
			switch update.EventType {
			case watch.Added, watch.Modified:
				handler := func() error {
					return updateSMTPConfig(client, kClient, spec, log)
				}
				err := handler()
				if err == nil {
					// Success - no need to retry
					break
				}
				log.WithError(err).Warnf("failed to update SMTP configuration from spec %s", spec)
				select {
				case retryC <- handler:
				// Queue handler on retry list
				case <-ctx.Done():
				}
			}
		case update := <-alertTargetCh:
			log := logrus.WithField("configmap", update.ResourceUpdate.Meta())
			spec := []byte(update.Data[constants.ResourceSpecKey])
			if isSpecEmpty(spec) {
				log.Error("empty configuration")
				continue
			}

			client := kubeClient.CoreV1().ConfigMaps(constants.MonitoringNamespace)
			switch update.EventType {
			case watch.Added, watch.Modified:
				handler := func() error {
					return updateAlertTarget(client, kClient, spec, log)
				}
				err := handler()
				if err == nil {
					// Success - no need to retry
					break
				}
				log.WithError(err).Warnf("failed to update alert target from spec %s", spec)
				select {
				case retryC <- handler:
				// Queue handler on retry list
				case <-ctx.Done():
				}
			case watch.Deleted:
				handler := func() error {
					return deleteAlertTarget(kClient, log)
				}
				err := handler()
				if err == nil {
					// Success - no need to retry
					break
				}
				log.WithError(err).Warn("failed to delete alert target")
				select {
				case retryC <- handler:
				// Queue handler on retry list
				case <-ctx.Done():
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func isSpecEmpty(spec []byte) bool {
	return len(bytes.TrimSpace(spec)) == 0
}

func createAlert(client *kapacitor.Client, spec []byte, log *log.Entry) error {
	if len(bytes.TrimSpace(spec)) == 0 {
		return trace.NotFound("empty configuration")
	}

	var alert alert
	err := yaml.Unmarshal(spec, &alert)
	if err != nil {
		return trace.Wrap(err, "failed to unmarshal %s", spec)
	}

	err = client.CreateAlert(alert.Name, alert.Spec.Formula)
	if err != nil {
		return trace.Wrap(err, "failed to create task")
	}
	return nil
}

func updateSMTPConfig(client corev1.SecretInterface, kClient *kapacitor.Client, spec []byte, log *logrus.Entry) error {
	log.Debugf("update SMTP config from spec %s", spec)
	var config smtpConfig
	err := yaml.Unmarshal(spec, &config)
	if err != nil {
		return trace.Wrap(err, "failed to unmarshal %s", spec)
	}

	portS := strconv.FormatInt(int64(config.Spec.Port), 10)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KapacitorSMTPSecret,
			Namespace: constants.MonitoringNamespace,
		},
		Data: map[string][]byte{
			"host": []byte(config.Spec.Host),
			"port": []byte(portS),
			"user": []byte(config.Spec.Username),
			"pass": []byte(config.Spec.Password),
		},
	}

	_, err = client.Update(secret)
	if err != nil {
		return trace.Wrap(err)
	}

	err = kClient.UpdateSMTPConfig(config.Spec.Host, config.Spec.Port, config.Spec.Username,
		config.Spec.Password)
	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func updateAlertTarget(client corev1.ConfigMapInterface, kClient *kapacitor.Client, spec []byte, log *logrus.Entry) error {
	log.Debugf("update alert target from spec %s", spec)
	var target alertTarget
	err := yaml.Unmarshal(spec, &target)
	if err != nil {
		return trace.Wrap(err, "failed to unmarshal %s", spec)
	}

	config := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KapacitorAlertTargetConfigMap,
			Namespace: constants.MonitoringNamespace,
		},
		Data: map[string]string{
			"from": constants.KapacitorAlertFrom,
			"to":   target.Spec.Email,
		},
	}

	_, err = client.Update(config)
	if err != nil {
		return trace.Wrap(err)
	}

	err = kClient.UpdateAlertTarget(target.Spec.Email)
	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func deleteAlertTarget(client *kapacitor.Client, log *logrus.Entry) error {
	log.Debug("delete alert target")
	return trace.Wrap(client.DeleteAlertTarget())
}

// alert defines the monitoring alert resource
type alert struct {
	Metadata `json:"metadata" yaml:"metadata"`
	// Spec defines the alert
	Spec alertSpec `json:"spec" yaml:"spec"`
}

// smtpConfig defines the cluster SMTP configuration resource
type smtpConfig struct {
	Metadata `json:"metadata" yaml:"metadata"`
	// Spec defines the SMTP configuration
	Spec smtpConfigSpec `json:"spec" yaml:"spec"`
}

// alertTarget defines the monitoring alert target resource
type alertTarget struct {
	Metadata `json:"metadata" yaml:"metadata"`
	// Spec defines the alert target
	Spec alertTargetSpec `json:"spec" yaml:"spec"`
}

// Metadata defines the common resource metadata
type Metadata struct {
	// Name is the name of the resource
	Name string `json:"name" yaml:"name"`
}

// alertSpec defines a monitoring alert
type alertSpec struct {
	// Formula specifies the Kapacitor formula
	Formula string `json:"formula" yaml:"formula"`
}

// smtpConfigSpec defines a SMTP configuration
type smtpConfigSpec struct {
	// Host specifies the SMTP service host
	Host string `json:"host" yaml:"host"`
	// Port specifies the SMTP service port
	Port int `json:"port" yaml:"port"`
	// Username specifies the name of the user to connect
	Username string `json:"username" yaml:"username"`
	// Password specifies the password to connect
	Password string `json:"password" yaml:"password"`
}

// alertTargetSpec defines a monitoring alert target
type alertTargetSpec struct {
	// Email specifies the recipient's email
	Email string `json:"email" yaml:"email"`
}

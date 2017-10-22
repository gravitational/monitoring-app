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

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/kapacitor"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"

	"github.com/ghodss/yaml"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeapi "k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func runAlertsWatcher(kubernetesClient *kubernetes.Client) error {
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

	alertCh := make(chan map[string]string)
	alertTargetCh := make(chan map[string]string)
	configmaps := []kubernetes.ConfigMap{
		{alertLabel, alertCh},
		{targetLabel, alertTargetCh},
	}
	smtpCh := make(chan map[string][]byte)

	go kubernetesClient.WatchConfigMaps(context.TODO(), configmaps...)
	go kubernetesClient.WatchSecrets(context.TODO(), kubernetes.Secret{smtpLabel, smtpCh})
	receiverLoop(context.TODO(), kubernetesClient.Clientset, kapacitorClient,
		alertCh, alertTargetCh, smtpCh)

	return nil
}

func receiverLoop(ctx context.Context, kubeClient *kubeapi.Clientset, kClient *kapacitor.Client,
	alertCh, alertTargetCh <-chan map[string]string, smtpCh <-chan map[string][]byte) {
	for {
		select {
		case update := <-alertCh:
			spec := []byte(update[constants.ResourceSpecKey])
			if err := createAlert(kClient, spec); err != nil {
				log.Warnf("failed to create alert: %v", trace.DebugReport(err))
			}
		case update := <-smtpCh:
			spec := update[constants.ResourceSpecKey]
			client := kubeClient.Secrets(api.NamespaceSystem)
			if err := updateSMTPConfig(client, kClient, spec); err != nil {
				log.Warnf("failed to update SMTP configuration: %v", trace.DebugReport(err))
			}
		case update := <-alertTargetCh:
			spec := []byte(update[constants.ResourceSpecKey])
			client := kubeClient.ConfigMaps(api.NamespaceSystem)
			if err := updateAlertTarget(client, kClient, spec); err != nil {
				log.Warnf("failed to update alert target: %v", trace.DebugReport(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func createAlert(client *kapacitor.Client, spec []byte) error {
	if len(bytes.TrimSpace(spec)) == 0 {
		return trace.NotFound("empty configurtion")
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

func updateSMTPConfig(client corev1.SecretInterface, kClient *kapacitor.Client, spec []byte) error {
	log.Debugf("update SMTP config from spec %s", spec)
	if len(bytes.TrimSpace(spec)) == 0 {
		return trace.NotFound("empty configurtion")
	}

	var config smtpConfig
	err := yaml.Unmarshal(spec, &config)
	if err != nil {
		return trace.Wrap(err, "failed to unmarshal %s", spec)
	}

	portS := strconv.FormatInt(int64(config.Spec.Port), 10)
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KapacitorSMTPSecret,
			Namespace: api.NamespaceSystem,
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

func updateAlertTarget(client corev1.ConfigMapInterface, kClient *kapacitor.Client, spec []byte) error {
	log.Debugf("update alert target from spec %s", spec)
	if len(bytes.TrimSpace(spec)) == 0 {
		return trace.NotFound("empty configurtion")
	}

	var target alertTarget
	err := yaml.Unmarshal(spec, &target)
	if err != nil {
		return trace.Wrap(err, "failed to unmarshal %s", spec)
	}

	config := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KapacitorAlertTargetConfigMap,
			Namespace: api.NamespaceSystem,
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

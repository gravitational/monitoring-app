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
	"time"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"
	"github.com/gravitational/monitoring-app/watcher/lib/resources"

	"github.com/ghodss/yaml"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
	kubeapi "k8s.io/client-go/kubernetes"
)

func runAlertsWatcher(ctx context.Context, kubernetesClient *kubernetes.Client) error {
	monitoringClient, err := kubernetes.NewMonitoringClient()
	if err != nil {
		return trace.Wrap(err)
	}

	rClient, err := resources.New(resources.ClientConfig{
		KubernetesClient: kubernetesClient.Clientset,
		MonitoringClient: monitoringClient,
		Namespace:        constants.MonitoringNamespace,
	})
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
	receiverLoop(ctx, kubernetesClient.Clientset, rClient,
		alertCh, alertTargetCh, smtpCh)

	return nil
}

func receiverLoop(ctx context.Context, kubeClient *kubeapi.Clientset, rClient resources.Resources,
	alertCh, alertTargetCh <-chan kubernetes.ConfigMapUpdate, smtpCh <-chan kubernetes.SecretUpdate) {
	for {
		select {
		case update := <-alertCh:
			log := log.WithField("configmap", update.ResourceUpdate.Meta())
			spec := []byte(update.Data[constants.ResourceSpecKey])
			switch update.EventType {
			case watch.Added, watch.Modified:
				if err := createAlert(rClient, spec, log); err != nil {
					log.Warnf("Failed to create alert from spec %s: %v.", spec, trace.DebugReport(err))
				}
			case watch.Deleted:
				if err := deleteAlert(rClient, spec, log); err != nil {
					log.Warnf("Failed to delete alert from spec %s: %v.", spec, trace.DebugReport(err))
				}
			}
		case update := <-smtpCh:
			log := log.WithField("secret", update.ResourceUpdate.Meta())
			spec := update.Data[constants.ResourceSpecKey]
			switch update.EventType {
			case watch.Added, watch.Modified:
				if err := updateSMTPConfig(rClient, spec, log); err != nil {
					log.Warnf("Failed to update SMTP configuration from spec %s: %v.", spec, trace.DebugReport(err))
				}
			case watch.Deleted:
				if err := deleteSMTPConfig(rClient, log); err != nil {
					log.Warnf("Failed to delete SMTP configuration: %v.", trace.DebugReport(err))
				}
			}
		case update := <-alertTargetCh:
			log := log.WithField("configmap", update.ResourceUpdate.Meta())
			spec := []byte(update.Data[constants.ResourceSpecKey])
			switch update.EventType {
			case watch.Added, watch.Modified:
				if err := updateAlertTarget(rClient, spec, log); err != nil {
					log.Warnf("Failed to update alert target from spec %s: %v.", spec, trace.DebugReport(err))
				}
			case watch.Deleted:
				if err := deleteAlertTarget(rClient, log); err != nil {
					log.Warnf("Failed to delete alert target: %v.", trace.DebugReport(err))
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func createAlert(client resources.Resources, spec []byte, log *log.Entry) error {
	log.Debugf("Creating alert from spec %s.", spec)

	if len(bytes.TrimSpace(spec)) == 0 {
		return trace.NotFound("empty configuration")
	}

	var alert alert
	err := yaml.Unmarshal(spec, &alert)
	if err != nil {
		return trace.Wrap(err, "failed to unmarshal %s", spec)
	}

	err = client.UpsertAlert(resources.Alert{
		CRDName:     alert.Name,
		AlertName:   alert.Spec.AlertName,
		GroupName:   alert.Spec.GroupName,
		Formula:     alert.Spec.Formula,
		Delay:       alert.Spec.Delay,
		Labels:      alert.Spec.Labels,
		Annotations: alert.Spec.Annotations,
	})
	if err != nil {
		return trace.Wrap(err, "failed to create task")
	}
	return nil
}

func deleteAlert(client resources.Resources, spec []byte, log *log.Entry) error {
	log.Debugf("Deleting alert from spec %s.", spec)

	var alert alert
	if err := yaml.Unmarshal(spec, &alert); err != nil {
		return trace.Wrap(err)
	}

	return client.DeleteAlert(alert.Name)
}

func updateSMTPConfig(client resources.Resources, spec []byte, log *log.Entry) error {
	log.Debugf("Updating SMTP config from spec: %s.", spec)
	if len(bytes.TrimSpace(spec)) == 0 {
		return trace.NotFound("empty configuration")
	}

	var config smtpConfig
	err := yaml.Unmarshal(spec, &config)
	if err != nil {
		return trace.Wrap(err, "failed to unmarshal %s", spec)
	}

	err = client.UpsertSMTPConfig(resources.SMTPConfig{
		Host:     config.Spec.Host,
		Port:     config.Spec.Port,
		Username: config.Spec.Username,
		Password: config.Spec.Password,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func deleteSMTPConfig(client resources.Resources, log *log.Entry) error {
	log.Debug("Deleting SMTP config.")
	return client.DeleteSMTPConfig()
}

func updateAlertTarget(client resources.Resources, spec []byte, log *log.Entry) error {
	log.Debugf("Updating alert target from spec: %s.", spec)
	if len(bytes.TrimSpace(spec)) == 0 {
		return trace.NotFound("empty configuration")
	}

	var target alertTarget
	err := yaml.Unmarshal(spec, &target)
	if err != nil {
		return trace.Wrap(err, "failed to unmarshal %s", spec)
	}

	err = client.UpsertAlertTarget(resources.AlertTarget{
		Email: target.Spec.Email,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func deleteAlertTarget(client resources.Resources, log *log.Entry) error {
	log.Debug("Deleting alert target.")
	return client.DeleteAlertTarget()
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
	// GroupName is the alerting rule group name.
	GroupName string `json:"group_name" yaml:"group_name"`
	// AlertName is the alerting rule name.
	AlertName string `json:"alert_name" yaml:"alert_name"`
	// Formula specifies the alert formula
	Formula string `json:"formula" yaml:"formula"`
	// Delay is an optional delay before alert triggers
	Delay time.Duration `json:"duration" yaml:"duration"`
	// Labels is the alerting rule labels.
	Labels map[string]string `json:"labels"`
	// Annotations is the alerting rule annotations.
	Annotations map[string]string `json:"annotations"`
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

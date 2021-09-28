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

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"

	"github.com/cenkalti/backoff"
	"github.com/gravitational/trace"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	monitoringv1typed "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	runtime.Must(monitoringv1.AddToScheme(scheme.Scheme))
}

// Client is the Kubernetes API client
type Client struct {
	*kubernetes.Clientset
}

// NewClient returns a new Kubernetes API client
func NewClient(kubeconfig string) (*Client, error) {
	var config *rest.Config
	var err error
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, trace.Wrap(err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &Client{Clientset: client}, nil
}

// NewMonitoringClient returns a new in-cluster Prometheus CRD API client.
func NewMonitoringClient(kubeconfig string) (*monitoring.Clientset, error) {
	var config *rest.Config
	var err error
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, trace.Wrap(err)
	}
	client, err := monitoring.NewForConfig(config)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return client, nil
}

// Prometheuses returns Prometheus CRD client in monitoring namespace.
func Prometheuses(kubeconfig string) (monitoringv1typed.PrometheusInterface, error) {
	client, err := NewMonitoringClient(kubeconfig)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return client.MonitoringV1().Prometheuses(constants.MonitoringNamespace), nil
}

// Alertmanagers returns Alertmanager CRD client in monitoring namespace.
func Alertmanagers(kubeconfig string) (monitoringv1typed.AlertmanagerInterface, error) {
	client, err := NewMonitoringClient(kubeconfig)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return client.MonitoringV1().Alertmanagers(constants.MonitoringNamespace), nil
}

// WatchConfigMaps watches Kubernetes API for ConfigMaps using specified configs to match
// and send updates
func (c *Client) WatchConfigMaps(ctx context.Context, configs ...ConfigMap) {
	for _, config := range configs {
		go func(config ConfigMap) {
			retry(ctx, func() error {
				err := watchConfigMap(ctx, c.CoreV1().ConfigMaps(constants.MonitoringNamespace), config)
				return trace.Wrap(err)
			})
		}(config)
	}
}

// WatchSecrets watches Kubernetes API for Secrets using specified configs to match
// and send updates
func (c *Client) WatchSecrets(ctx context.Context, configs ...Secret) {
	for _, config := range configs {
		go func(config Secret) {
			retry(ctx, func() error {
				err := watchSecret(ctx, c.CoreV1().Secrets(constants.MonitoringNamespace), config)
				return trace.Wrap(err)
			})
		}(config)
	}
}

// Label represents a Kubernetes label which is used
// as a search target for ConfigMaps
type Label struct {
	Key   string
	Value string
}

// MatchLabel matches a resource with the specified label
func MatchLabel(key, value string) (labels.Selector, error) {
	req, err := labels.NewRequirement(key, selection.In, []string{value})
	if err != nil {
		return nil, trace.Wrap(err, "failed to build a requirement from %v=%q", key, value)
	}
	selector := labels.NewSelector()
	selector = selector.Add(*req)
	return selector, nil
}

// ConfigMap describes matching and sending updates for ConfigMaps.
// If Match matches a resource, RecvCh channel receives
// the data from the matched resource
type ConfigMap struct {
	// Selector specifies the selector for this ConfigMap
	Selector labels.Selector
	// RecvCh specifies the channel that receives updates on the matched resource
	RecvCh chan ConfigMapUpdate
}

// Secret describes matching and sending updates for Secrets.
// If Match matches a resource, RecvCh channel receives
// the data from the matched resource
type Secret struct {
	// Selector specifies the selector for this Secret
	Selector labels.Selector
	// RecvCh specifies the channel that receives updates on the matched resource
	RecvCh chan SecretUpdate
}

// ConfigMapUpdate describes a ConfigMap update
type ConfigMapUpdate struct {
	// ResourceUpdate describes the common resource update metadata
	ResourceUpdate
	// Data descrines the update data payload
	Data map[string]string
}

// SecretUpdate describes a Secret update
type SecretUpdate struct {
	// ResourceUpdate describes the common resource update metadata
	ResourceUpdate
	// Data descrines the update data payload
	Data map[string][]byte
}

// Meta formats the metadata for readability
func (r ResourceUpdate) Meta() string {
	return fmt.Sprintf("%v(%v/%v)", r.Kind, r.Namespace, r.Name)
}

// String formats this update for readability
func (r ResourceUpdate) String() string {
	return fmt.Sprintf("%v(%v)", r.EventType, r.Meta())
}

// ResourceUpdate describes an update for a resource
type ResourceUpdate struct {
	// EventType specifies the type of event
	watch.EventType
	// TypeMeta references the resource type metadata
	metav1.TypeMeta
	// ObjectMeta references the resource metadata
	metav1.ObjectMeta
}

func watchConfigMap(ctx context.Context, client corev1.ConfigMapInterface, config ConfigMap) error {
	watcher, err := client.Watch(ctx, metav1.ListOptions{LabelSelector: config.Selector.String()})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watcher.Stop()

	log := log.WithFields(log.Fields{"watch": "configmap", "label": config.Selector.String()})
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return trace.Retry(nil, "configmap watcher closed")
			}

			switch configMap := event.Object.(type) {
			case *v1.ConfigMap:
				log.Infof("Detected event %v for configmap %v.", event.Type, configMap.Name)
				config.RecvCh <- ConfigMapUpdate{
					ResourceUpdate{event.Type, configMap.TypeMeta, configMap.ObjectMeta},
					configMap.Data,
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func watchSecret(ctx context.Context, client corev1.SecretInterface, config Secret) error {
	watcher, err := client.Watch(ctx, metav1.ListOptions{LabelSelector: config.Selector.String()})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watcher.Stop()

	log := log.WithFields(log.Fields{"watch": "secret", "label": config.Selector.String()})
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return trace.Retry(nil, "secret watcher closed")
			}

			switch secret := event.Object.(type) {
			case *v1.Secret:
				log.Infof("Detected event %v for secret %v.", event.Type, secret.Name)
				config.RecvCh <- SecretUpdate{
					ResourceUpdate{event.Type, secret.TypeMeta, secret.ObjectMeta},
					secret.Data,
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func retry(ctx context.Context, fn func() error) (err error) {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 0
	err = backoff.RetryNotify(
		func() error {
			return trace.Wrap(fn())
		},
		backoff.WithContext(b, ctx),
		func(err error, d time.Duration) {
			log.Debugf("retrying: %v (time %v)", trace.DebugReport(err), d)
		})
	return trace.Wrap(err)
}

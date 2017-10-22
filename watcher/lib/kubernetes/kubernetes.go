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
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

// Client is the Kubernetes API client
type Client struct {
	*kubernetes.Clientset
}

// NewClient returns a new Kubernetes API client
func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &Client{Clientset: client}, nil
}

// WatchConfigMaps watches Kubernetes API for ConfigMaps using specified configs to match
// and send updates
func (c *Client) WatchConfigMaps(ctx context.Context, configs ...ConfigMap) {
	for _, config := range configs {
		go func(config ConfigMap) {
			retry(ctx, func() error {
				err := watchConfigMap(ctx, c.CoreV1().ConfigMaps(api.NamespaceSystem), config)
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
				err := watchSecret(ctx, c.CoreV1().Secrets(api.NamespaceSystem), config)
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
	RecvCh chan map[string]string
}

// Secret describes matching and sending updates for Secrets.
// If Match matches a resource, RecvCh channel receives
// the data from the matched resource
type Secret struct {
	// Selector specifies the selector for this Secret
	Selector labels.Selector
	// RecvCh specifies the channel that receives updates on the matched resource
	RecvCh chan map[string][]byte
}

func watchConfigMap(ctx context.Context, client corev1.ConfigMapInterface, config ConfigMap) error {
	watcher, err := client.Watch(metav1.ListOptions{LabelSelector: config.Selector.String()})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watcher.Stop()

	log := log.WithFields(log.Fields{"watch": "configmap", "label": config.Selector.String()})
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				log.Debug("watcher closed")
				return nil
			}

			if event.Type != watch.Added {
				log.Debugf("ignoring event: %v", event)
				continue
			}

			switch configMap := event.Object.(type) {
			case *v1.ConfigMap:
				log.Infof("detected %q", configMap.Name)
				config.RecvCh <- configMap.Data
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func watchSecret(ctx context.Context, client corev1.SecretInterface, config Secret) error {
	watcher, err := client.Watch(metav1.ListOptions{LabelSelector: config.Selector.String()})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watcher.Stop()

	log := log.WithFields(log.Fields{"watch": "secret", "label": config.Selector.String()})
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				log.Debug("watcher closed")
				return nil
			}

			if event.Type != watch.Added {
				log.Debugf("ignoring event: %v", event)
				continue
			}

			switch secret := event.Object.(type) {
			case *v1.Secret:
				log.Infof("detected %q", secret.Name)
				config.RecvCh <- secret.Data
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

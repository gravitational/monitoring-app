package lib

import (
	"context"
	"strings"
	"time"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

// KubernetesLabel represents a Kubernetes label which is used
// as a search target for ConfigMaps
type KubernetesLabel struct {
	Key   string
	Value string
}

// KubernetesClient is the Kubernetes API client
type KubernetesClient struct {
	*kubernetes.Clientset
}

// NewKubernetesClient returns a new Kubernetes API client
func NewKubernetesClient() (*KubernetesClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &KubernetesClient{Clientset: client}, nil
}

// WatchConfigMaps watches Kubernetes API for ConfigMaps using specified configs to match
// and send updates
func (c *KubernetesClient) WatchConfigMaps(ctx context.Context, configs ...ConfigMap) {
	for {
		select {
		case <-time.After(time.Second):
			err := c.restartConfigMapWatch(ctx, configs...)
			if err != nil {
				log.Errorf(trace.DebugReport(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

// WatchSecrets watches Kubernetes API for Secrets using specified configs to match
// and send updates
func (c *KubernetesClient) WatchSecrets(ctx context.Context, configs ...Secret) {
	for {
		select {
		case <-time.After(time.Second):
			err := c.restartSecretWatch(ctx, configs...)
			if err != nil {
				log.Errorf(trace.DebugReport(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *KubernetesClient) restartConfigMapWatch(ctx context.Context, configs ...ConfigMap) error {
	log.Info("restarting watch")

	watcher, err := c.ConfigMaps(api.NamespaceSystem).Watch(metav1.ListOptions{})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				log.Warn("watcher channel closed")
				return nil
			}

			if event.Type != watch.Added {
				log.Debugf("ignoring event: %v", event)
				continue
			}

			configMap := event.Object.(*v1.ConfigMap)
			for _, config := range configs {
				if config.Match(configMap.ObjectMeta) {
					log.Infof("detected configmap %q", configMap.Name)
					config.RecvCh <- configMap.Data
				}
			}

		case <-ctx.Done():
			log.Debug("stopping watcher")
			return nil
		}
	}
}

func (c *KubernetesClient) restartSecretWatch(ctx context.Context, configs ...Secret) error {
	log.Info("restarting watch")

	watcher, err := c.Secrets(api.NamespaceSystem).Watch(metav1.ListOptions{})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				log.Warn("watcher channel closed")
				return nil
			}

			if event.Type != watch.Added {
				log.Debugf("ignoring event: %v", event)
				continue
			}

			secret := event.Object.(*v1.Secret)
			for _, config := range configs {
				if config.Match(secret.ObjectMeta) {
					log.Infof("detected secret %q", secret.Name)
					config.RecvCh <- secret.Data
				}
			}

		case <-ctx.Done():
			log.Debug("stopping watcher")
			return nil
		}
	}
}

// MatchName matches a resource using a strict name comparison
func MatchName(name string) ResourceMatchFunc {
	return func(resource metav1.ObjectMeta) bool {
		return resource.Name == name
	}
}

// MatchPrefix matches a resource with the specified name prefix
func MatchPrefix(prefix string) ResourceMatchFunc {
	return func(resource metav1.ObjectMeta) bool {
		if strings.HasPrefix(resource.Name, prefix) {
			return true
		}
		return false
	}
}

// MatchLabel matches a resource with the specified label
func MatchLabel(label KubernetesLabel) ResourceMatchFunc {
	return func(resource metav1.ObjectMeta) bool {
		for k, v := range resource.Labels {
			if k == label.Key && v == label.Value {
				return true
			}
		}
		return false
	}
}

// ResourceMatchFunc defines a function that matches resources.
type ResourceMatchFunc func(metav1.ObjectMeta) bool

// ConfigMap describes matching and sending updates for ConfigMaps.
// If Match matches a resource, RecvCh channel receives
// the data from the matched resource
type ConfigMap struct {
	// Match matches a resource to receive updates about
	Match ResourceMatchFunc
	// RecvCh specifies the channel that receives updates on the matched resource
	RecvCh chan map[string]string
}

// Secret describes matching and sending updates for Secrets.
// If Match matches a resource, RecvCh channel receives
// the data from the matched resource
type Secret struct {
	// Match matches a resource to receive updates about
	Match ResourceMatchFunc
	// RecvCh specifies the channel that receives updates on the matched resource
	RecvCh chan map[string][]byte
}

package lib

import (
	"context"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/gravitational/trace"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/watch"
	"k8s.io/client-go/1.4/rest"
)

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

// WatchConfigMaps watches Kubernetes API for configmaps with the specified name prefix in the system
// namespace and submits them to the provided channel
func (c *KubernetesClient) WatchConfigMaps(ctx context.Context, prefix string, ch chan<- string) {
	for {
		select {
		case <-time.After(time.Second):
			err := c.restartWatch(ctx, prefix, ch)
			if err != nil {
				log.Errorf(trace.DebugReport(err))
			}
		case <-ctx.Done():
			close(ch)
			return
		}
	}
}

func (c *KubernetesClient) restartWatch(ctx context.Context, prefix string, ch chan<- string) error {
	log.Infof("restarting watch")

	watcher, err := c.ConfigMaps("kube-system").Watch(api.ListOptions{})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				log.Warningf("watcher channel closed: %v", event)
				return nil
			}

			if event.Type != watch.Added {
				log.Infof("ignoring event: %v", event)
				continue
			}

			configMap := event.Object.(*v1.ConfigMap)
			if !strings.HasPrefix(configMap.Name, prefix) {
				log.Infof("ignoring configmap: %v", configMap.Name)
				continue
			}

			log.Infof("detected configmap: %v", configMap.Name)
			for _, v := range configMap.Data {
				ch <- v
			}
		case <-ctx.Done():
			log.Infof("stopping watcher")
			return nil
		}
	}
}

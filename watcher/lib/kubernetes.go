package lib

import (
	"context"
	"strings"

	log "github.com/Sirupsen/logrus"
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

// WatchDashboards watches Kubernetes API for configmaps with dashboards in system namespace and
// submits them to the returned channel
func (c *KubernetesClient) WatchDashboards(ctx context.Context) (chan string, error) {
	watcher, err := c.ConfigMaps("kube-system").Watch(api.ListOptions{})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	ch := make(chan string)
	go watchDashboards(ctx, watcher, ch)

	return ch, nil
}

func watchDashboards(ctx context.Context, watcher watch.Interface, ch chan<- string) {
	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type != watch.Added {
				log.Infof("ignoring event: %v", event.Type)
				continue
			}

			configMap := event.Object.(*v1.ConfigMap)
			if !strings.HasPrefix(configMap.Name, DashboardPrefix) {
				log.Infof("ignoring configmap: %v", configMap.Name)
				continue
			}

			log.Infof("detected dashboard configmap: %v", configMap.Name)
			for _, v := range configMap.Data {
				ch <- v
			}
		case <-ctx.Done():
			log.Infof("stopping watcher")
			watcher.Stop()
			return
		}
	}
}

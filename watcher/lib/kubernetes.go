package lib

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/trace"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/watch"
	"k8s.io/client-go/1.4/rest"
)

type KubernetesClient struct {
	*kubernetes.Clientset
}

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

func (c *KubernetesClient) WatchDashboards(ch chan string) error {
	watcher, err := c.ConfigMaps("kube-system").Watch(api.ListOptions{})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watcher.Stop()

	log.Infof("start watching dashboards")
	for event := range watcher.ResultChan() {
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
	}

	return nil
}

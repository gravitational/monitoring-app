/*
Copyright 2020 Gravitational, Inc.

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
	"context"
	"time"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

type autoscaleConfig struct {
	// nodes is the Nodes API client.
	nodes v1.NodeInterface
	// alertmanagers is the Alertmanagers CRD API client.
	alertmanagers monitoringv1.AlertmanagerInterface
	// prometheuses is the Prometheuses CRD API client.
	prometheuses monitoringv1.PrometheusInterface
	// interval is the reconciliation interval.
	interval time.Duration
}

func (c *autoscaleConfig) checkAndSetDefaults() error {
	if c.nodes == nil {
		return trace.BadParameter("missing Nodes client")
	}
	if c.alertmanagers == nil {
		return trace.BadParameter("missing Alertmanagers client")
	}
	if c.prometheuses == nil {
		return trace.BadParameter("missing Prometheuses client")
	}
	if c.interval == 0 {
		c.interval = time.Minute
	}
	return nil
}

func runAutoscale(ctx context.Context, config autoscaleConfig) error {
	err := config.checkAndSetDefaults()
	if err != nil {
		return trace.Wrap(err)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	log.Info("Starting autoscaler.")

	for {
		select {
		case <-ticker.C:
			nodes, err := getMasterNodes(config.nodes)
			if err != nil {
				log.WithError(err).Error("Failed to query nodes.")
				continue
			}
			err = reconcileAlertmanager(config.alertmanagers, nodes)
			if err != nil {
				log.WithError(err).Error("Failed to reconcile Alertmanager.")
				continue
			}
			err = reconcilePrometheus(config.prometheuses, nodes)
			if err != nil {
				log.WithError(err).Error("Failed to reconcile Prometheus.")
				continue
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// getMasterNodes returns a list of Kubernetes master nodes.
func getMasterNodes(nodes v1.NodeInterface) ([]corev1.Node, error) {
	masterLabel, err := kubernetes.MatchLabel(constants.NodeRoleLabel, constants.MasterLabel)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	nodeList, err := nodes.List(metav1.ListOptions{
		LabelSelector: masterLabel.String(),
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return nodeList.Items, nil
}

// reconcileAlertmanager adjusts the number of Alertmanager replicas according
// to the provided node list.
func reconcileAlertmanager(alertmanagers monitoringv1.AlertmanagerInterface, nodes []corev1.Node) error {
	alertmanager, err := alertmanagers.Get(constants.AlertmanagerName, metav1.GetOptions{})
	if err != nil {
		return trace.Wrap(err)
	}

	replicas := int32V(alertmanager.Spec.Replicas)

	if len(nodes) > 1 {
		if replicas != 2 {
			log.Infof("Alertmanager has %v replicas, scaling to 2.", replicas)
			alertmanager.Spec.Replicas = int32P(2)
			if _, err := alertmanagers.Update(alertmanager); err != nil {
				return trace.Wrap(err)
			}
		} else {
			log.Debugf("Alertmanager has %v replicas.", replicas)
		}
	} else {
		if replicas != 1 {
			log.Infof("Alertmanager has %v replicas, scaling to 1.", replicas)
			alertmanager.Spec.Replicas = int32P(1)
			if _, err := alertmanagers.Update(alertmanager); err != nil {
				return trace.Wrap(err)
			}
		} else {
			log.Debugf("Alertmanager has %v replicas.", replicas)
		}
	}

	return nil
}

// reconcilePrometheus adjusts the number of Prometheus replicas according to
// the provided node list.
func reconcilePrometheus(prometheuses monitoringv1.PrometheusInterface, nodes []corev1.Node) error {
	prometheus, err := prometheuses.Get(constants.PrometheusName, metav1.GetOptions{})
	if err != nil {
		return trace.Wrap(err)
	}

	replicas := int32V(prometheus.Spec.Replicas)

	if len(nodes) > 1 {
		if replicas != 2 {
			log.Infof("Prometheus has %v replicas, scaling to 2.", replicas)
			prometheus.Spec.Replicas = int32P(2)
			if _, err := prometheuses.Update(prometheus); err != nil {
				return trace.Wrap(err)
			}
		} else {
			log.Debugf("Prometheus has %v replicas.", replicas)
		}
	} else {
		if replicas != 1 {
			log.Infof("Prometheus has %v replicas, scaling to 1.", replicas)
			prometheus.Spec.Replicas = int32P(1)
			if _, err := prometheuses.Update(prometheus); err != nil {
				return trace.Wrap(err)
			}
		} else {
			log.Debugf("Prometheus has %v replicas.", replicas)
		}
	}

	return nil
}

func int32P(v int32) *int32 {
	return &v
}

func int32V(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

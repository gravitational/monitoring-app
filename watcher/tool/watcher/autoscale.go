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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
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

	ticker := time.NewTicker(config.interval)
	defer ticker.Stop()

	log.Info("Starting autoscaler.")

	for {
		select {
		case <-ticker.C:
			nodes, err := getMasterNodes(ctx, config.nodes)
			if err != nil {
				log.WithError(err).Error("Failed to query nodes.")
				continue
			}
			err = reconcileAlertmanager(ctx, config.alertmanagers, nodes)
			if err != nil {
				log.WithError(err).Error("Failed to reconcile Alertmanager.")
			}
			err = reconcilePrometheus(ctx, config.prometheuses, nodes)
			if err != nil {
				log.WithError(err).Error("Failed to reconcile Prometheus.")
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// getMasterNodes returns a list of Kubernetes master nodes.
func getMasterNodes(ctx context.Context, nodes v1.NodeInterface) ([]corev1.Node, error) {
	masterLabel, err := kubernetes.MatchLabel(constants.NodeRoleLabel, constants.MasterLabel)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	nodeList, err := nodes.List(ctx, metav1.ListOptions{
		LabelSelector: masterLabel.String(),
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return nodeList.Items, nil
}

// reconcileAlertmanager adjusts the number of Alertmanager replicas according
// to the provided node list.
func reconcileAlertmanager(ctx context.Context, alertmanagers monitoringv1.AlertmanagerInterface, nodes []corev1.Node) error {
	alertmanager, err := alertmanagers.Get(ctx, constants.AlertmanagerName, metav1.GetOptions{})
	if err != nil {
		return trace.Wrap(err)
	}

	replicas := int32V(alertmanager.Spec.Replicas)

	if len(nodes) > 1 {
		if replicas != 2 {
			alertmanager.Spec.Replicas = int32P(2)
		}
	} else {
		if replicas != 1 {
			alertmanager.Spec.Replicas = int32P(1)
		}
	}

	if int32V(alertmanager.Spec.Replicas) == replicas {
		log.Debugf("Alertmanager has %v replicas.", replicas)
		return nil
	}

	log.Infof("Alertmanager has %v replicas, scaling to %v.", replicas, int32V(alertmanager.Spec.Replicas))
	if _, err := alertmanagers.Update(ctx, alertmanager, metav1.UpdateOptions{}); err != nil {
		return trace.Wrap(err)
	}

	return nil
}

// reconcilePrometheus adjusts the number of Prometheus replicas according to
// the provided node list.
func reconcilePrometheus(ctx context.Context, prometheuses monitoringv1.PrometheusInterface, nodes []corev1.Node) error {
	prometheus, err := prometheuses.Get(ctx, constants.PrometheusName, metav1.GetOptions{})
	if err != nil {
		return trace.Wrap(err)
	}

	replicas := int32V(prometheus.Spec.Replicas)

	if len(nodes) > 1 {
		if replicas != 2 {
			prometheus.Spec.Replicas = int32P(2)
		}
	} else {
		if replicas != 1 {
			prometheus.Spec.Replicas = int32P(1)
		}
	}

	if int32V(prometheus.Spec.Replicas) == replicas {
		log.Debugf("Prometheus has %v replicas.", replicas)
		return nil
	}

	log.Infof("Prometheus has %v replicas, scaling to %v.", replicas, int32V(prometheus.Spec.Replicas))
	if _, err := prometheuses.Update(ctx, prometheus, metav1.UpdateOptions{}); err != nil {
		return trace.Wrap(err)
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

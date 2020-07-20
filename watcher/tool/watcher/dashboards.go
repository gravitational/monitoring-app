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
	"context"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/grafana"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"
	"github.com/gravitational/monitoring-app/watcher/lib/utils"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
)

func runDashboardsWatcher(ctx context.Context, kubernetesClient *kubernetes.Client, retryC chan<- func() error) error {
	grafanaClient, err := grafana.NewClient()
	if err != nil {
		return trace.Wrap(err)
	}

	err = utils.WaitForAPI(ctx, grafanaClient)
	if err != nil {
		return trace.Wrap(err)
	}

	label, err := kubernetes.MatchLabel(constants.MonitoringLabel, constants.MonitoringUpdateDashboard)
	if err != nil {
		return trace.Wrap(err)
	}

	ch := make(chan kubernetes.ConfigMapUpdate)
	go kubernetesClient.WatchConfigMaps(ctx, kubernetes.ConfigMap{label, ch})
	receiveAndCreateDashboards(ctx, grafanaClient, ch, retryC)
	return nil
}

// receiveAndCreateDashboards listens on the provided channel that receives new dashboards data and creates
// them in Grafana using the provided client
func receiveAndCreateDashboards(ctx context.Context, client *grafana.Client,
	ch <-chan kubernetes.ConfigMapUpdate, retryC chan<- func() error) {
	for {
		select {
		case update := <-ch:
			switch update.EventType {
			case watch.Added, watch.Modified:
				log := log.WithField("configmap", update.ResourceUpdate.Meta())
				for _, dashboard := range update.Data {
					dashboard := dashboard
					handler := func() error {
						return client.CreateDashboard(dashboard)
					}
					err := handler()
					if err == nil {
						// Success - no need to retry
						break
					}
					log.WithError(err).Warnf("failed to create dashboard %v", dashboard)
					select {
					case retryC <- handler:
					// Queue handler on retry list
					case <-ctx.Done():
					}
				}
			case watch.Deleted:
				log := log.WithField("configmap", update.ResourceUpdate.Meta())
				for _, dashboard := range update.Data {
					dashboard := dashboard
					handler := func() error {
						return client.DeleteDashboard(dashboard)
					}
					err := handler()
					if err == nil {
						// Success - no need to retry
						break
					}
					log.WithError(err).Warnf("failed to delete dashboard %v", dashboard)
					select {
					case retryC <- handler:
					// Queue handler on retry list
					case <-ctx.Done():
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

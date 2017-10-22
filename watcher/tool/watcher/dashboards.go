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
)

func runDashboardsWatcher(kubernetesClient *kubernetes.Client) error {
	grafanaClient, err := grafana.NewClient()
	if err != nil {
		return trace.Wrap(err)
	}

	err = utils.WaitForAPI(context.TODO(), grafanaClient)
	if err != nil {
		return trace.Wrap(err)
	}

	label, err := kubernetes.MatchLabel(constants.MonitoringLabel, constants.MonitoringUpdateDashboard)
	if err != nil {
		return trace.Wrap(err)
	}

	ch := make(chan map[string]string)
	go kubernetesClient.WatchConfigMaps(context.TODO(), kubernetes.ConfigMap{label, ch})
	receiveAndCreateDashboards(context.TODO(), grafanaClient, ch)
	return nil
}

// receiveAndCreateDashboards listens on the provided channel that receives new dashboards data and creates
// them in Grafana using the provided client
func receiveAndCreateDashboards(ctx context.Context, client *grafana.Client, ch <-chan map[string]string) {
	for {
		select {
		case data := <-ch:
			for _, v := range data {
				err := client.CreateDashboard(v)

				if err != nil {
					log.Errorf("failed to create dashboard: %v", trace.DebugReport(err))
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

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

	"github.com/gravitational/monitoring-app/watcher/lib"
	"github.com/gravitational/monitoring-app/watcher/lib/kapacitor"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

func runAlertsWatcher() error {
	kubernetesClient, err := lib.NewKubernetesClient()
	if err != nil {
		return trace.Wrap(err)
	}

	kapacitorClient, err := kapacitor.NewClient()
	if err != nil {
		return trace.Wrap(err)
	}

	ch := make(chan map[string]string)
	alertLabel := &lib.KubernetesLabel{
		Key:   AlertsLabelKey,
		Value: AlertsLabelValue,
	}
	go kubernetesClient.WatchConfigMaps(context.TODO(), "", alertLabel, ch)
	receiveAndCreateAlerts(context.TODO(), kapacitorClient, ch)

	return nil
}

func receiveAndCreateAlerts(ctx context.Context, client *kapacitor.Client, ch <-chan map[string]string) {
	for {
		select {
		case data, ok := <-ch:
			if !ok {
				log.Warn("alerts channel closed")
				return
			}

			for k, v := range data {
				err := client.CreateAlert(k, v)
				if err != nil {
					log.Errorf("failed to create alert task: %v", trace.DebugReport(err))
				}
			}
		case <-ctx.Done():
			log.Debugln("stopping")
			return
		}
	}
}

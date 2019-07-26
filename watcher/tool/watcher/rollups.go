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
	"encoding/json"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/influxdb"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"
	"github.com/gravitational/monitoring-app/watcher/lib/utils"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
)

func runRollupsWatcher(watcherCfg watcherConfig, influxDBConfig influxdb.Config) error {
	influxDBClient, err := influxdb.NewClient(influxDBConfig)
	if err != nil {
		return trace.Wrap(err)
	}

	err = utils.WaitForAPI(watcherCfg.ctx, influxDBClient)
	if err != nil {
		return trace.Wrap(err)
	}

	err = influxDBClient.Setup(influxDBConfig)
	if err != nil {
		return trace.Wrap(err)
	}

	label, err := kubernetes.MatchLabel(constants.MonitoringLabel, constants.MonitoringUpdateRollup)
	if err != nil {
		return trace.Wrap(err)
	}

	ch := make(chan kubernetes.ConfigMapUpdate)
	go watcherCfg.kubernetesClient.WatchConfigMaps(context.TODO(), kubernetes.ConfigMap{label, ch})
	receiveAndManageRollups(watcherCfg, influxDBClient, ch)
	return nil
}

// receiveAndManageRollups listens on the provided channel that receives new rollups data and creates,
// updates or deletes them in/from InfluxDB using the provided client
func receiveAndManageRollups(config watcherConfig, client *influxdb.Client, ch <-chan kubernetes.ConfigMapUpdate) {
	for {
		select {
		case update := <-ch:
			log := log.WithField("configmap", update.ResourceUpdate.Meta())
			for _, v := range update.Data {
				var rollups []influxdb.Rollup
				err := json.Unmarshal([]byte(v), &rollups)
				if err != nil {
					log.Errorf("failed to unmarshal %s: %v", update.Data, trace.DebugReport(err))
					continue
				}

				for _, rollup := range rollups {
					if err := rollup.Check(); err != nil {
						// Fail immediately on precondition violation
						log.WithError(err).Warn("Failed to validate rollup.")
						continue
					}

					switch update.EventType {
					case watch.Added:
						handler := func() error {
							return client.CreateRollup(rollup)
						}
						err := handler()
						if err == nil {
							// Success - no need to retry
							break
						}
						log.WithError(err).Warnf("Failed to create rollup %v", rollup)
						select {
						case config.retryC <- handler:
						// Queue handler on retry list
						case <-config.ctx.Done():
						}
					case watch.Deleted:
						handler := func() error {
							return client.DeleteRollup(rollup)
						}
						err := handler()
						if err == nil {
							// Success - no need to retry
							break
						}
						log.WithError(err).Warnf("Failed to delete rollup %v", rollup)
						select {
						case config.retryC <- handler:
						// Queue handler on retry list
						case <-config.ctx.Done():
						}
					case watch.Modified:
						handler := func() error {
							return client.UpdateRollup(rollup)
						}
						err := handler()
						if err == nil {
							// Success - no need to retry
							break
						}
						log.WithError(err).Warnf("Failed to update rollup %v", rollup)
						select {
						case config.retryC <- handler:
						// Queue handler on retry list
						case <-config.ctx.Done():
						}
					}
				}
			}
		case <-config.ctx.Done():
			return
		}
	}
}

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
	"time"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/influxdb"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"
	"github.com/gravitational/monitoring-app/watcher/lib/utils"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
)

func runRollupsWatcher(kubernetesClient *kubernetes.Client) error {
	influxDBClient, err := influxdb.NewClient()
	if err != nil {
		return trace.Wrap(err)
	}

	err = utils.WaitForAPI(context.TODO(), influxDBClient)
	if err != nil {
		return trace.Wrap(err)
	}

	err = influxDBClient.Setup()
	if err != nil {
		return trace.Wrap(err)
	}

	label, err := kubernetes.MatchLabel(constants.MonitoringLabel, constants.MonitoringUpdateRollup)
	if err != nil {
		return trace.Wrap(err)
	}

	chK := make(chan kubernetes.ConfigMapUpdate)
	chR := make(chan func() error)
	go kubernetesClient.WatchConfigMaps(context.TODO(), kubernetes.ConfigMap{label, chK})
	receiveAndManageRollups(context.TODO(), influxDBClient, chK, chR)
	return nil
}

// receiveAndManageRollups listens on the provided channel that receives new rollups data and creates,
// updates or deletes them in/from InfluxDB using the provided client
func receiveAndManageRollups(ctx context.Context, client *influxdb.Client, ch <-chan kubernetes.ConfigMapUpdate, retryC chan<- func() error) {
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
					switch update.EventType {
					case watch.Added:
						if err := rollup.Check(); err != nil {
							// Fail immediately on precondition violation
							log.WithError(err).Warn("Failed to create rollup.")
							continue
						}
						handler := func() error {
							err := client.CreateRollup(rollup)
							return err
						}
						if err := handler(); err == nil {
							// Success - no need to retry
							break
						}
						log.WithError(err).Warnf("Failed to create rollup %v", rollup)
						select {
						case retryC <- handler:
						// Queue handler on retry list
						case <-ctx.Done():
							return
						}
					case watch.Deleted:
						err := retry(ctx, func() error {
							err := client.DeleteRollup(rollup)
							return err
						})
						if err != nil {
							log.Errorf("failed to delete rollup %v: %v", rollup, trace.DebugReport(err))
						}
					case watch.Modified:
						err := retry(ctx, func() error {
							err := client.UpdateRollup(rollup)
							return err
						})
						if err != nil {
							log.Errorf("failed to alter rollup %v: %v", rollup, trace.DebugReport(err))
						}
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func retryLoop(ctx context.Context, retryC <-chan func() error) {
	var handlers []func() error
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for {
		select {
		case handler := <-retryC:
			handlers = append(handlers)
		case <-timer.C:
			for i, handler := range handlers {
				if err := handler(); err != nil {
					log.WithError(err).Warn("Failed to complete handler.")
					continue
				}
				// Remove handler
				handlers = append(handlers[:i], handlers[i+1:]...)
			}
		case <-ctx.Done():
			return
		}
	}
}

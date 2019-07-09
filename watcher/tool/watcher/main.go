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
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/influxdb"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"
	"github.com/sirupsen/logrus"

	"github.com/gravitational/trace"
	"github.com/gravitational/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	mode           string
	debug          bool
	influxDBConfig influxdb.Config

	envs = map[string]string{
		"INFLUXDB_ADMIN_USERNAME":    "influxdb-admin-username",
		"INFLUXDB_ADMIN_PASSWORD":    "influxdb-admin-password",
		"INFLUXDB_GRAFANA_USERNAME":  "influxdb-grafana-username",
		"INFLUXDB_GRAFANA_PASSWORD":  "influxdb-grafana-password",
		"INFLUXDB_TELEGRAF_USERNAME": "influxdb-telegraf-username",
		"INFLUXDB_TELEGRAF_PASSWORD": "influxdb-telegraf-password",
	}

	rootCmd = &cobra.Command{
		Use:   "watcher",
		Short: "Utility to manage InfluxDB/Grafana/Alerts",
		RunE:  root,
	}
)

func init() {
	level := logrus.InfoLevel

	rootCmd.PersistentFlags().StringVar(&mode, "mode", "", fmt.Sprintf("Watcher mode: %v", constants.AllModes))
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBAdminUser, "influxdb-admin-username", constants.InfluxDBAdminUser, "InfluxDB administrator username")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBAdminPassword, "influxdb-admin-password", constants.InfluxDBAdminUser, "InfluxDB administrator password")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBGrafanaUser, "influxdb-grafana-username", constants.InfluxDBGrafanaUser, "InfluxDB grafana username")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBGrafanaPassword, "influxdb-grafana-password", constants.InfluxDBGrafanaUser, "InfluxDB grafana password")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBTelegrafUser, "influxdb-telegraf-username", constants.InfluxDBTelegrafUser, "InfluxDB telegraf username")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBTelegrafPassword, "influxdb-telegraf-password", constants.InfluxDBTelegrafUser, "InfluxDB telegraf password")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debugging mode")

	trace.SetDebug(debug)
	if debug {
		level = logrus.DebugLevel
	}
	logrus.SetFormatter(&trace.TextFormatter{})
	logrus.SetLevel(level)

	bindFlagEnv(rootCmd.PersistentFlags())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(trace.DebugReport(err))
		os.Exit(255)
	}
}

func root(ccmd *cobra.Command, args []string) error {
	var mode string
	flag.StringVar(&mode, "mode", "", fmt.Sprintf("watcher mode: %v", constants.AllModes))
	ver := flag.Bool("version", false, "print version")
	flag.Parse()

	if *ver {
		version.Print()
		return nil
	}

	client, err := kubernetes.NewClient()
	if err != nil {
		return trace.Wrap(err)
	}

	ctx := context.TODO()
	retryC := runRetryLoop(ctx)

	switch mode {
	case constants.ModeDashboards:
		err = runDashboardsWatcher(ctx, client, retryC)
	case constants.ModeRollups:
		err = runRollupsWatcher(ctx, influxDBConfig, client, retryC)
	case constants.ModeAlerts:
		err = runAlertsWatcher(ctx, client, retryC)
	default:
		return trace.Errorf("ERROR: unknown mode %q\n", mode)
	}

	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

// bindFlagEnv binds environment variables to command flags
func bindFlagEnv(flagSet *pflag.FlagSet) {
	for env, flag := range envs {
		cmdFlag := flagSet.Lookup(flag)
		if cmdFlag != nil {
			if value := os.Getenv(env); value != "" {
				cmdFlag.Value.Set(value)
			}
		}
	}
}

func runRetryLoop(ctx context.Context) chan<- func() error {
	retryC := make(chan func() error)
	go func() {
		var handlers []func() error
		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()
		for {
			select {
			case handler := <-retryC:
				handlers = append(handlers, handler)
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
	}()
	return retryC
}

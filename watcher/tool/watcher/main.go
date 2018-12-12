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
	"fmt"
	"os"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/influxdb"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"
	"github.com/sirupsen/logrus"

	"github.com/gravitational/trace"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var (
	log            = logrus.New()
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
	log.SetLevel(logrus.InfoLevel)

	rootCmd.PersistentFlags().StringVar(&mode, "mode", "", fmt.Sprintf("Watcher mode: %v", constants.AllModes))
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBAdminUser, "influxdb-admin-username", constants.InfluxDBAdminUser, "InfluxDB administrator username")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBAdminPassword, "influxdb-admin-password", constants.InfluxDBAdminUser, "InfluxDB administrator password")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBGrafanaUser, "influxdb-grafana-username", constants.InfluxDBGrafanaUser, "InfluxDB grafana username")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBGrafanaPassword, "influxdb-grafana-password", constants.InfluxDBGrafanaUser, "InfluxDB grafana password")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBTelegrafUser, "influxdb-telegraf-username", constants.InfluxDBTelegrafUser, "InfluxDB telegraf username")
	rootCmd.PersistentFlags().StringVar(&influxDBConfig.InfluxDBTelegrafPassword, "influxdb-telegraf-password", constants.InfluxDBTelegrafUser, "InfluxDB telegraf password")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debugging mode")

	if debug {
		log.Level = logrus.DebugLevel
	}

	bindFlagEnv(rootCmd.PersistentFlags())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(trace.DebugReport(err))
		os.Exit(255)
	}
}

func root(ccmd *cobra.Command, args []string) error {
	client, err := kubernetes.NewClient()
	if err != nil {
		return trace.Wrap(err)
	}

	switch mode {
	case constants.ModeDashboards:
		err = runDashboardsWatcher(client)
	case constants.ModeRollups:
		err = runRollupsWatcher(client, influxDBConfig)
	case constants.ModeAlerts:
		err = runAlertsWatcher(client)
	default:
		return trace.Errorf("ERROR: unknown mode %q\n", mode)
	}

	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

// bindFlagEnv binds environment variables to command flags
func bindFlagEnv(flagSet *flag.FlagSet) {
	for env, flag := range envs {
		cmdFlag := flagSet.Lookup(flag)
		if cmdFlag != nil {
			if value := os.Getenv(env); value != "" {
				cmdFlag.Value.Set(value)
			}
		}
	}
}

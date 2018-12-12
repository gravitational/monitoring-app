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
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	mode                  string
	influxdbAdminUsername string
	influxdbAdminPassword string

	envs = map[string]string{
		"influxdb-admin-username": "INFLUXDB_ADMIN_USERNAME",
		"influxdb-admin-password": "INFLUXDB_ADMIN_PASSWORD",
	}
)

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "", fmt.Sprintf("Watcher mode: %v", constants.AllModes))
	flag.StringVar(&influxdbAdminUsername, "influxdb-admin-username", constants.InfluxDBAdminUser, "InfluxDB administrator username")
	flag.StringVar(&influxdbAdminPassword, "influxdb-admin-password", constants.InfluxDBAdminUser, "InfluxDB administrator password")
	flag.Parse()

	bindFlagEnv(&influxdbAdminUsername)
	bindFlagEnv(&influxdbAdminPassword)

	client, err := kubernetes.NewClient()
	if err != nil {
		exitWithError(err)
	}

	switch mode {
	case constants.ModeDashboards:
		err = runDashboardsWatcher(client)
	case constants.ModeRollups:
		err = runRollupsWatcher(client)
	case constants.ModeAlerts:
		err = runAlertsWatcher(client)
	default:
		fmt.Printf("ERROR: unknown mode %q\n", mode)
		os.Exit(255)
	}

	if err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	log.Error(trace.DebugReport(err))
	fmt.Printf("ERROR: %v\n", err.Error())
	os.Exit(255)
}

// bindFlagEnv binds environment variables to command flags
func bindFlagEnv(flag *string) {
	if val, ok := envs[*flag]; ok {
		if value := os.Getenv(val); value != "" {
			*flag = value
		}
	}
}

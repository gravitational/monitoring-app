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

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

var (
	mode, kubeconfig string
	debug            bool
)

func main() {
	flag.StringVar(&mode, "mode", "", fmt.Sprintf("watcher mode: %v", constants.AllModes))
	flag.StringVar(&kubeconfig, "kubeconfig", "", "optional kubeconfig path")
	flag.BoolVar(&debug, "debug", false, "turn on debug logging")
	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	err := run()
	if err != nil {
		exitWithError(err)
	}
}

func run() error {
	client, err := kubernetes.NewClient(kubeconfig)
	if err != nil {
		return trace.Wrap(err)
	}

	switch mode {
	case constants.ModeDashboards:
		err := runDashboardsWatcher(client)
		if err != nil {
			return trace.Wrap(err)
		}

	case constants.ModeAlerts:
		err := runAlertsWatcher(context.Background(), client, kubeconfig)
		if err != nil {
			return trace.Wrap(err)
		}

	case constants.ModeAutoscale:
		alertmanagers, err := kubernetes.Alertmanagers(kubeconfig)
		if err != nil {
			return trace.Wrap(err)
		}
		prometheuses, err := kubernetes.Prometheuses(kubeconfig)
		if err != nil {
			return trace.Wrap(err)
		}
		err = runAutoscale(context.Background(), autoscaleConfig{
			nodes:         client.CoreV1().Nodes(),
			alertmanagers: alertmanagers,
			prometheuses:  prometheuses,
		})
		if err != nil {
			return trace.Wrap(err)
		}

	default:
		return trace.BadParameter("unknown mode %q", mode)
	}

	return nil
}

func exitWithError(err error) {
	log.Error(trace.DebugReport(err))
	fmt.Printf("ERROR: %v\n", err.Error())
	os.Exit(255)
}

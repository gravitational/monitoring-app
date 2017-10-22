package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gravitational/monitoring-app/watcher/lib/constants"
	"github.com/gravitational/monitoring-app/watcher/lib/kubernetes"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "", fmt.Sprintf("watcher mode: %v", constants.AllModes))
	flag.Parse()

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

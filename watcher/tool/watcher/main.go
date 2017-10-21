package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gravitational/monitoring-app/watcher/lib"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "", fmt.Sprintf("watcher mode: %v", lib.AllModes))
	flag.Parse()

	if !lib.OneOf(mode, lib.AllModes) {
		fmt.Printf("invalid mode")
		os.Exit(255)
	}

	var err error
	switch mode {
	case lib.ModeDashboards:
		err = runDashboardsWatcher()
	case lib.ModeRollups:
		err = runRollupsWatcher()
	case lib.ModeAlerts:
		err = runAlertsWatcher()
	}

	if err != nil {
		log.Error(trace.DebugReport(err))
		fmt.Printf("ERROR: %v\n", err.Error())
		os.Exit(255)
	}
}

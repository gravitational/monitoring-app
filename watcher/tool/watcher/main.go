package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/monitoring-app/watcher/lib"
	"github.com/gravitational/trace"
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
	if mode == lib.ModeDashboards {
		err = runDashboardsWatcher()
	} else {
		err = runRollupsWatcher()
	}

	if err != nil {
		log.Error(trace.DebugReport(err))
		fmt.Printf("ERROR: %v\n", err.Error())
		os.Exit(255)
	}
}

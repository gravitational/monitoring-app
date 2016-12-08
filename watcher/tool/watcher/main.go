package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/monitoring-app/watcher/lib"
	"github.com/gravitational/trace"
)

func main() {
	err := run()
	if err != nil {
		log.Error(trace.DebugReport(err))
		fmt.Printf("ERROR: %v\n", err.Error())
		os.Exit(255)
	}
}

func run() error {
	kubernetesClient, err := lib.NewKubernetesClient()
	if err != nil {
		return trace.Wrap(err)
	}

	grafanaClient, err := lib.NewGrafanaClient()
	if err != nil {
		return trace.Wrap(err)
	}

	ch := make(chan string)

	go func() {
		kubernetesClient.WatchDashboards(ch)
	}()

	for data := range ch {
		err = grafanaClient.CreateDashboard(data)
		if err != nil {
			log.Errorf("failed to create dashboard: %v", trace.DebugReport(err))
		}
	}

	return nil
}

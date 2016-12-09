package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	// grafana might be still starting up, wait for it
	err = waitForGrafana(context.TODO(), grafanaClient)
	if err != nil {
		return trace.Wrap(err)
	}

	ch, err := kubernetesClient.WatchDashboards(context.TODO())
	if err != nil {
		return trace.Wrap(err)
	}

	receiveAndCreateDashboards(context.TODO(), grafanaClient, ch)
	return nil
}

// waitForGrafana spins until Grafana API can be reached successfully or the provided context is cancelled
func waitForGrafana(ctx context.Context, grafanaClient *lib.GrafanaClient) error {
	for {
		select {
		case <-time.After(2 * time.Second):
			err := grafanaClient.Health()
			if err != nil {
				log.Infof("cannot reach grafana: %v", trace.DebugReport(err))
			} else {
				return nil
			}
		case <-ctx.Done():
			return trace.Errorf("failed to reach grafana")
		}
	}
}

// receiveAndCreateDashboards listens on the provided channel that receives new dashboards data and creates
// them in Grafana using the provided client
func receiveAndCreateDashboards(ctx context.Context, grafanaClient *lib.GrafanaClient, ch chan string) {
	for {
		select {
		case data := <-ch:
			err := grafanaClient.CreateDashboard(data)
			if err != nil {
				log.Errorf("failed to create dashboard: %v", trace.DebugReport(err))
			}
		case <-ctx.Done():
			log.Infof("stopping")
			return
		}
	}
}

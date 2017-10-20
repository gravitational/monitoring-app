package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/gravitational/monitoring-app/watcher/lib"
	"github.com/gravitational/trace"
)

func runDashboardsWatcher() error {
	kubernetesClient, err := lib.NewKubernetesClient()
	if err != nil {
		return trace.Wrap(err)
	}

	grafanaClient, err := lib.NewGrafanaClient()
	if err != nil {
		return trace.Wrap(err)
	}

	err = lib.WaitForAPI(context.TODO(), grafanaClient)
	if err != nil {
		return trace.Wrap(err)
	}

	ch := make(chan map[string]string)
	go kubernetesClient.WatchConfigMaps(context.TODO(), lib.DashboardPrefix, "", ch)
	receiveAndCreateDashboards(context.TODO(), grafanaClient, ch)
	return nil
}

// receiveAndCreateDashboards listens on the provided channel that receives new dashboards data and creates
// them in Grafana using the provided client
func receiveAndCreateDashboards(ctx context.Context, client *lib.GrafanaClient, ch <-chan map[string]string) {
	for {
		select {
		case data, ok := <-ch:
			if !ok {
				log.Warningf("dashboards channel closed")
				return
			}

			for _, v := range data {
				err := client.CreateDashboard(v)

				if err != nil {
					log.Errorf("failed to create dashboard: %v", trace.DebugReport(err))
				}
			}
		case <-ctx.Done():
			log.Infof("stopping")
			return
		}
	}
}

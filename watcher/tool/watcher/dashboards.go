package main

import (
	"context"

	"github.com/gravitational/monitoring-app/watcher/lib"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

func runDashboardsWatcher(kubernetesClient *lib.KubernetesClient) error {
	grafanaClient, err := lib.NewGrafanaClient()
	if err != nil {
		return trace.Wrap(err)
	}

	err = lib.WaitForAPI(context.TODO(), grafanaClient)
	if err != nil {
		return trace.Wrap(err)
	}

	ch := make(chan map[string]string)
	go kubernetesClient.WatchConfigMaps(context.TODO(),
		lib.ConfigMap{lib.MatchPrefix(lib.DashboardPrefix), ch})
	receiveAndCreateDashboards(context.TODO(), grafanaClient, ch)
	return nil
}

// receiveAndCreateDashboards listens on the provided channel that receives new dashboards data and creates
// them in Grafana using the provided client
func receiveAndCreateDashboards(ctx context.Context, client *lib.GrafanaClient, ch <-chan map[string]string) {
	for {
		select {
		case data := <-ch:
			for _, v := range data {
				err := client.CreateDashboard(v)

				if err != nil {
					log.Errorf("failed to create dashboard: %v", trace.DebugReport(err))
				}
			}
		case <-ctx.Done():
			log.Debug("stopping")
			return
		}
	}
}

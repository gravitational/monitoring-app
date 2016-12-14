package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gravitational/monitoring-app/watcher/lib"
	"github.com/gravitational/trace"
)

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "",
		fmt.Sprintf("watcher mode: %v or %v", lib.ModeDashboards, lib.ModeRollups))
	flag.Parse()

	if !lib.OneOf(mode, []string{lib.ModeDashboards, lib.ModeRollups}) {
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

	ch, err := kubernetesClient.WatchConfigMaps(context.TODO(), lib.DashboardPrefix)
	if err != nil {
		return trace.Wrap(err)
	}

	receiveAndCreateDashboards(context.TODO(), grafanaClient, ch)
	return nil
}

func runRollupsWatcher() error {
	kubernetesClient, err := lib.NewKubernetesClient()
	if err != nil {
		return trace.Wrap(err)
	}

	influxdbClient, err := lib.NewInfluxdbClient()
	if err != nil {
		return trace.Wrap(err)
	}

	err = lib.WaitForAPI(context.TODO(), influxdbClient)
	if err != nil {
		return trace.Wrap(err)
	}

	ch, err := kubernetesClient.WatchConfigMaps(context.TODO(), lib.RollupsPrefix)
	if err != nil {
		return trace.Wrap(err)
	}

	receiveAndCreateRollups(context.TODO(), influxdbClient, ch)
	return nil
}

// receiveAndCreateDashboards listens on the provided channel that receives new dashboards data and creates
// them in Grafana using the provided client
func receiveAndCreateDashboards(ctx context.Context, client *lib.GrafanaClient, ch <-chan string) {
	for {
		select {
		case data, ok := <-ch:
			if !ok {
				log.Warningf("dashboards channel closed")
				return
			}

			err := client.CreateDashboard(data)
			if err != nil {
				log.Errorf("failed to create dashboard: %v", trace.DebugReport(err))
			}
		case <-ctx.Done():
			log.Infof("stopping")
			return
		}
	}
}

// receiveAndCreateRollups listens on the provided channel that receives new rollups data and creates
// them in Influxdb using the provided client
func receiveAndCreateRollups(ctx context.Context, client *lib.InfluxdbClient, ch <-chan string) {
	for {
		select {
		case data, ok := <-ch:
			if !ok {
				log.Warningf("rollups channel closed")
				return
			}

			var rollups []lib.Rollup
			err := json.Unmarshal([]byte(data), &rollups)
			if err != nil {
				log.Errorf("failed to unmarshal: %v %v", data, trace.DebugReport(err))
				continue
			}

			for _, rollup := range rollups {
				err := client.CreateRollup(rollup)
				if err != nil {
					log.Errorf("failed to create rollup: %v %v", rollup, trace.DebugReport(err))
				}
			}
		case <-ctx.Done():
			log.Infof("stopping")
			return
		}
	}
}

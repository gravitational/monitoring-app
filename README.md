# Gravity Cluster Monitoring

This gravity app provides an InfluxDB, Heapster + Grafana based monitoring system.

## Overview

As alluded to, there are 3 main components in the monitoring system: InfluxDB, Heapster and Grafana

### InfluxDB

InfluxDB is the main data store for current + future monitoring time series data. It provides the service `influxdb.kube-system.svc.cluster.local`.

### Heapster

Heapster monitors Kubernetes components and reports statistics and information to InfluxDB about nodes and pods.

### Grafana

Grafana is the dashboard system that provides visualization information on all the information stored in InfluxDB. It is exposed as the service `grafana.kube-system.svc.cluster.local`. The default username and password is `admin`/`admin`.

Grafana comes with several pre-built dashboards to monitor pods and nodes. Dashboards can be created in Grafana's interface and saved to JSON. When placed in the [images/grafana/dashboards](images/grafana/dashboards) folder, they can be included automatically with this app.

### Production

This app is automatically included with any `k8s-*` app and anything inheriting from it.

## Development

The `Makefile` has development targets for quick iteration

## Future work

 - [ ] Better InfluxDB persistence, availability work
 - [ ] More default Grafana Dashboards

# Gravity Cluster Monitoring

This gravity app provides an InfluxDB, Heapster + Grafana based monitoring system. Also provided is a Kapacitor deployment, for alerting.

## Overview

As alluded to, there are 4 main components in the monitoring system: InfluxDB, Heapster, Grafana and Kapacitor.

### InfluxDB

InfluxDB is the main data store for current + future monitoring time series data. It provides the service `influxdb.kube-system.svc.cluster.local`.

### Heapster

Heapster monitors Kubernetes components and reports statistics and information to InfluxDB about nodes and pods.

### Grafana

Grafana is the dashboard system that provides visualization information on all the information stored in InfluxDB. It is exposed as the service `grafana.kube-system.svc.cluster.local`. Grafana credentials are generated during initial installation and placed into a Secret `grafana` in `kube-system` namespace.

### Kapacitor

Kapacitor is the alerting system, that streams data from InfluxDB and sends alerts as configured by the end user. It exposes the service `kapacitor.kube-system.svc.cluster.local`.

## Grafana integration

This app is shipped with two pre-configured Grafana dashboards providing machine- and pod-level overview of the installed site. Grafana UI is integrated with Gravity control panel. To view dashboards once the site is up and running, navigate to that site's `Monitoring` page.

Grafana is configured with anonymous mode which allows anyone logged into Gravity OpsCenter (or site) to use it.

In development (sites installed with virsh locally) the anonymous mode has full Admin permissions that allows creating new and modifying existing dashboards which is convenient for development.

In production the anonymous mode is read-only that allows only viewing of existing dashboards.

## Pluggable dashboards

Other applications can ship their own Grafana dashboards by using ConfigMaps. A custom dashboard ConfigMap should be assigned a `monitoring`
label with value `dashboard` and created in the `kube-system` namespace for it to be recognized and loaded at startup:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: mydashboard
  namespace: kube-system
  labels:
    monitoring: dashboard
data:
  mydashboard: |
    { ... dashboard JSON ... }
```


An example of a dashboard ConfigMap can be seen in the monitoring-app's own resources.

A dashboard ConfigMap may contain multiple keys with dashboards, key names are not relevant. Dashboard JSON can be obtained from Grafana by building a dashboard and then exporting it (or viewing its raw JSON representation).

## Metrics collection

All default metrics collected by heapster go into `k8s` database in InfluxDB. All other applications that collect metrics should submit them into the same `k8s` database in order for proper retention policies to be enforced.

## Retention policies

The app comes with 3 pre-configured retention policies:

* default / 24h
* medium / 4w
* long / 52w

The `default` retention policy is supposed to store high-precision metrics (for example, all default metrics collected by heapster with 10s interval). The `default` policy is "default" for k8s database which means that metrics that do not specify retention policy explicitly go in there.

The other two policies - `medium` and `long` are supposed to store metric rollups and should not be used directly.

Durations for each of the retention policies can be configured via Gravity control panel.

## Rollups

Metric rollups are meant to provide access to historical data for longer time period but at lower resolution.

Monitoring app allows to configure two "types" of rollups for any collected metric.

* "medium" rollup aggregates (or filters) data over 5-minute interval and goes into "medium" retention policy
* "long" rollup aggregates (or filters) data over 1-hour interval and goes into "long" retention policy

This app comes with rollups pre-configured for some of the metrics collected by default. Applications that collect their own metrics can configure their own rollups as well, via ConfigMaps.

A custom rollup ConfigMap should be assigned a `monitoring` label with value `rollup` and created in the `kube-system` namespace
so it is recognized and loaded at startup:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: rollups-myrollups
  namespace: kube-system
  labels:
    monitoring: rollup
data:
  rollups: |
    [
      {
        "retention": "medium",
        "measurement": "cpu/usage_rate",
        "name": "cpu/usage_rate/medium",
        "functions": [
          {
            "function": "max",
            "field": "value",
            "alias": "value_max"
          },
          {
            "function": "mean",
            "field": "value",
            "alias": "value_mean"
          }
        ]
      }
    ]
```

Each rollup is a JSON object with the following fields:

* `retention` - name of the retention policy (and hence the aggregation interval) for this rollup, can be `medium` or `long`
* `measurement` - name of the metric for the rollup (i.e. which metric is being "rolled up")
* `name` - name of the resulting "rolled up" metric
* `functions` - list of rollup functions to apply to metric `measurement`
* `function` - function name, can be `mean`, `median`, `sum`, `max`, `min` or `percentile_XXX` where `0 <= XXX <= 100`
* `field` - name of the field to apply rollup function to (e.g. "value")
* `alias` - new name for the rolled up field (e.g. "value_max")

## Kapacitor integration

Kapacitor provides alerting for default and user-defined alerts.

### Basic configuration

To configure Kapacitor to send email alerts, change the values of the Kubernetes Secret `smtp-configuration`. To configure the 'to' and 'from' email addresses for sending alerts, update the `alerting-addresses` ConfigMap. You will need to reload any running Kapacitor pods after making these changes for them to take effect.

### Custom and default alerts

Alerts (written in [TICKscript](https://docs.influxdata.com/kapacitor/v1.2/tick/)) are automatically detected, loaded and enabled. They are read from the Kubernetes ConfigMap named `kapacitor-alerts`. To create new alerts, add your alert scripts as new key/values to that ConfigMap.

## Future work

 - [ ] Better InfluxDB persistence, availability work
 - [ ] More default Grafana Dashboards


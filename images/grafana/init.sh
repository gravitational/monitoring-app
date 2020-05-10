#!/usr/bin/env bash

set -o nounset
set -o errexit

mkdir -p /etc/grafana/provisioning/{dashboards,datasources}
cat /tmp/datasources/influxdb.yaml | envsubst > /etc/grafana/provisioning/datasources/influxdb.yaml
chown -R $SERVICE_UID /etc/grafana

#!/bin/bash

set -o errexit
set -o nounset

echo "Substitute env variables in influxdb datasource"
mkdir -p /etc/grafana/provisioning/datasources
cat /etc/grafana/datasources/influxdb.yaml | envsubst > /etc/grafana/provisioning/datasources/influxdb.yaml

echo "Starting Grafana"
exec /usr/share/grafana/bin/grafana-server --homepath=/usr/share/grafana --config=/etc/grafana/cfg/grafana.ini

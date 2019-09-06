#!/bin/bash

set -m
echo "Starting Grafana"
cat /etc/grafana/provisioning-noenv/datasources/influxdb.yaml | envsubst > /etc/grafana/provisioning/datasources/influxdb.yaml && \
exec /usr/share/grafana/bin/grafana-server --config=/etc/grafana/cfg/grafana.ini

#!/bin/bash

set -m
echo "Starting Grafana"
mkdir -p /etc/grafana/provisioning/datasources && \
/usr/bin/envsubst /etc/grafana/provisioning-noenv/datasources/influxdb.yaml > /tmp/conf && mv /tmp/conf /etc/grafana/provisioning/datasources/influxdb.yaml && \
exec /usr/share/grafana/bin/grafana-server --homepath=/usr/share/grafana --config=/etc/grafana/cfg/grafana.ini

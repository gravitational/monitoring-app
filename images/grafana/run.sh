#!/bin/bash

set -m
echo "Starting Grafana"
/usr/bin/envsubst /etc/grafana/provisioning-noenv/datasources/influxdb.yaml > /tmp/conf && mv /tmp/conf /etc/grafana/provisioning/datasources/influxdb.yaml && \
exec /usr/share/grafana/bin/grafana-server --config=/etc/grafana/cfg/grafana.ini

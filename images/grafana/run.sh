#!/bin/bash

set -m
echo "Starting Grafana"
exec /usr/share/grafana/bin/grafana-server --config=/etc/grafana/cfg/grafana.ini

#!/bin/bash

set -m
echo "Starting Grafana"
exec /usr/share/grafana/bin/grafana-server --homepath=/usr/share/grafana --config=/etc/grafana/cfg/grafana.ini

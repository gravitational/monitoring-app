#!/bin/bash

set -m
echo "Starting Grafana"
exec /usr/sbin/grafana-server --homepath=/usr/share/grafana --config=/etc/grafana/cfg/grafana.ini

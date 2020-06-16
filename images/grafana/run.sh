#!/bin/bash

set -m
echo "Starting Grafana"
/usr/share/grafana/bin/grafana-server --config=/etc/grafana/cfg/grafana.ini

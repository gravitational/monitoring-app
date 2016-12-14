#!/bin/bash
set -m

echo "Starting InfluxDB in the background"
exec /influxdb/influxd -config /etc/influxdb.toml &

echo "Waiting for InfluxDB to come up..."
until $(curl --fail --output /dev/null --silent http://localhost:8086/ping); do
    printf "." && sleep 2
done

echo "Configuring retention policies"
curl --silent http://localhost:8086/query --data-urlencode "q=create database k8s with duration 24h"
curl --silent http://localhost:8086/query --data-urlencode "q=create retention policy medium on k8s duration 4w replication 1"
curl --silent http://localhost:8086/query --data-urlencode "q=create retention policy long on k8s duration 52w replication 1"

echo "Bringing InfluxDB back to the foreground"
fg

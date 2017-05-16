#!/bin/bash

URL=${URL:-http://localhost:9092}
ALERTDIR=${ALERTDIR:-/opt/alerts}

while true; do
    for alert in $ALERTDIR/*.tick; do
        filename=$(basename "$alert")
        alertname="${filename%.*}"
        if ! kapacitor -url $URL list tasks | grep -q $alertname ; then
            echo "alert $alertname doesn't exist, creating"
            kapacitor -url $URL define $alertname -type stream -dbrp k8s.default -tick $alert
            kapacitor -url $URL enable $alertname
        fi
    done
    sleep 5
done

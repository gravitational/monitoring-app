#!/bin/sh
if [ "$DEVMODE" = "true" ]; then
    /opt/bin/kubectl create -f /var/lib/gravity/resources/grafana-ini-dev.yaml
else
    /opt/bin/kubectl create -f /var/lib/gravity/resources/grafana-ini-prod.yaml
fi

/opt/bin/kubectl create -f /var/lib/gravity/resources/resources.yaml

#!/bin/sh
/opt/bin/kubectl create secret generic grafana \
                 --namespace=kube-system \
                 --from-literal=username=admin \
                 --from-literal=password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

if [ "$DEVMODE" = "true" ]; then
    /opt/bin/kubectl create -f /var/lib/gravity/resources/grafana-ini-dev.yaml
else
    /opt/bin/kubectl create -f /var/lib/gravity/resources/grafana-ini-prod.yaml
fi

/opt/bin/kubectl create -f /var/lib/gravity/resources/resources.yaml

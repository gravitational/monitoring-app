#!/bin/sh
set -e

echo "Assuming changeset from the environment: $RIG_CHANGESET"
# note that rig does not take explicit changeset ID
# taking it from the environment variables
if [ $1 = "update" ]; then
    echo "Checking: $RIG_CHANGESET"
    if rig status $RIG_CHANGESET --retry-attempts=1 --retry-period=1s; then exit 0; fi

    echo "Starting update, changeset: $RIG_CHANGESET"
    rig cs delete --force -c cs/$RIG_CHANGESET

    echo "Deleting old replication controller 'heapster'"
    rig delete rc/heapster --resource-namespace=kube-system --force

    echo "Deleting old replication controller 'influxdb'"
    rig delete rc/influxdb --resource-namespace=kube-system --force

    echo "Deleting old replication controller 'grafana'"
    rig delete rc/grafana --resource-namespace=kube-system --force

    echo "Deleting old configmap 'grafana'"
    rig delete configmaps/grafana --resource-namespace=kube-system --force

    echo "Deleting old secret 'grafana'"
    rig delete secrets/grafana --resource-namespace=kube-system --force

    echo "Creating new secret 'grafana'"
    password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
    sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana-creds.yaml
    rig upsert -f /var/lib/gravity/resources/grafana-creds.yaml --debug

    echo "Creating new configmap 'grafana'"
    if [ "$DEVMODE" = "true" ]; then
        rig upsert -f /var/lib/gravity/resources/grafana-ini-dev.yaml --debug
    else
        rig upsert -f /var/lib/gravity/resources/grafana-ini-prod.yaml --debug
    fi

    echo "Creating or updating resources"
    rig upsert -f /var/lib/gravity/resources/resources.yaml --debug
    rig upsert -f /var/lib/gravity/resources/alerts.yaml --debug

    echo "Checking status"
    rig status $RIG_CHANGESET --retry-attempts=120 --retry-period=1s --debug

    echo "Freezing"
    rig freeze
elif [ $1 = "rollback" ]; then
    echo "Reverting changeset $RIG_CHANGESET"
    rig revert
else
    echo "Missing argument, should be either 'update' or 'rollback'"
fi

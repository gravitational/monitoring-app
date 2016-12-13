#!/bin/sh
set -e

echo "Asuming changeset from the envrionment: $RIG_CHANGESET"
# note that rig does not take explicit changeset ID
# taking it from the environment variables
if [ $1 = "update" ]; then
    echo "Starting update, changeset: $RIG_CHANGESET"
    rig cs delete --force -c cs/$RIG_CHANGESET
    echo "Deleting old replication controller rc/heapster"
    rig delete rc/heapster --resource-namespace=kube-system --force
    echo "Deleting old replication controller rc/influxdb"
    rig delete rc/influxdb --resource-namespace=kube-system --force
    echo "Deleting old replication controller rc/grafana"
    rig delete rc/grafana --resource-namespace=kube-system --force
    echo "Creating or updating resources"
    rig upsert -f /var/lib/gravity/resources/resources.yaml --debug
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

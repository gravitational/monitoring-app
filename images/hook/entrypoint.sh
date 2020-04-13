#!/bin/sh
set -xe

echo "---> Assuming changeset from the environment: $RIG_CHANGESET"
# note that rig does not take explicit changeset ID
# taking it from the environment variables
if [ $1 = "update" ]; then
    echo "---> Checking: $RIG_CHANGESET"
    if rig status $RIG_CHANGESET --retry-attempts=1 --retry-period=1s; then exit 0; fi

    echo "---> Starting update, changeset: $RIG_CHANGESET"
    rig cs delete --force -c cs/$RIG_CHANGESET

    echo "---> Deleting old 'heapster' resources"
    rig delete deployments/heapster --resource-namespace=monitoring --force

    echo "---> Deleting old 'influxdb' resources"
    rig delete deployments/influxdb --resource-namespace=monitoring --force

    echo "---> Deleting old 'grafana' resources"
    rig delete deployments/grafana --resource-namespace=monitoring --force

    echo "---> Deleting old 'telegraf' resources"
    rig delete deployments/telegraf --resource-namespace=monitoring --force
    rig delete daemonsets/telegraf-node --resource-namespace=monitoring --force

    echo "---> Deleting old deployment 'kapacitor'"
    rig delete deployments/kapacitor --resource-namespace=monitoring --force

    echo "---> Deleting old secrets"
    for secret in grafana grafana-influxdb-creds smtp-configuration
    do
        rig delete secrets/$secret --resource-namespace=monitoring --force
    done

    echo "---> Deleting old configmaps"
    for configmap in influxdb grafana-cfg grafana grafana-dashboards-cfg grafana-dashboards grafana-datasources kapacitor-alerts rollups-default alerting-addresses
    do
        rig delete configmaps/$configmap --resource-namespace=monitoring --force
    done

    echo "---> Creating monitoring namespace"
    rig upsert -f /var/lib/gravity/resources/namespace.yaml --debug

    for file in /var/lib/gravity/resources/crds/*
    do
        rig upsert -f $file --debug
    done

    # Generate password for Grafana administrator
    password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
    sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana.yaml

    echo "---> Creating or updating resources"
    for name in security grafana watcher
    do
        rig upsert -f /var/lib/gravity/resources/${name}.yaml --debug
    done

    for file in /var/lib/gravity/resources/prometheus/*
    do
        rig upsert -f $file --debug
    done

    for file in /var/lib/gravity/resources/nethealth/*
    do
        rig upsert -f $file --debug
    done

    echo "---> Checking status"
    rig status $RIG_CHANGESET --retry-attempts=120 --retry-period=1s --debug

    echo "---> Freezing"
    rig freeze
elif [ $1 = "rollback" ]; then
    echo "---> Reverting changeset $RIG_CHANGESET"
    rig revert
    rig cs delete --force -c cs/$RIG_CHANGESET
else
    echo "---> Missing argument, should be either 'update' or 'rollback'"
fi

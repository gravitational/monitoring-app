#!/bin/sh
set -xe

echo "---> Assuming changeset from the environment: $RIG_CHANGESET"
# note that rig does not take explicit changeset ID
# taking it from the environment variables
if [ $1 = "update" ]; then
    echo "---> Deleting old 'heapster' resources"
    rig delete deployments/heapster --resource-namespace=monitoring --force

    echo "---> Deleting old 'influxdb' resources"
    rig delete deployments/influxdb --resource-namespace=monitoring --force

    echo "---> Deleting old 'telegraf' resources"
    rig delete deployments/telegraf --resource-namespace=monitoring --force
    rig delete daemonsets/telegraf-node --resource-namespace=monitoring --force
    rig delete daemonsets/telegraf-node-worker --resource-namespace=monitoring --force
    rig delete daemonsets/telegraf-node-master --resource-namespace=monitoring --force

    echo "---> Deleting old deployment 'kapacitor'"
    rig delete deployments/kapacitor --resource-namespace=monitoring --force

    echo "---> Deleting old configmaps"
    for configmap in influxdb grafana kapacitor-alerts rollups-default grafana-dashboard-k8s-cluster-rsrc-use grafana-dashboard-k8s-resources-cluster \
      grafana-dashboard-k8s-resources-namespace grafana-dashboard-k8s-resources-pod grafana-dashboard-k8s-resources-workload grafana-dashboard-k8s-resources-workloads-namespace \
      grafana-dashboard-nodes grafana-dashboard-pods grafana-dashboard-nethealth
    do
        rig delete configmaps/$configmap --resource-namespace=monitoring --force
    done

    echo "---> Creating monitoring namespace"
    rig upsert -f /var/lib/gravity/resources/namespace.yaml --debug

    for file in /var/lib/gravity/resources/kube-prometheus-setup/*
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

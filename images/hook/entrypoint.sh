#!/bin/sh
set -e

echo "---> Assuming changeset from the environment: $RIG_CHANGESET"
# note that rig does not take explicit changeset ID
# taking it from the environment variables
if [ $1 = "update" ]; then
    if ! /opt/bin/kubectl get namespaces monitoring > /dev/null 2>&1
    then
        /opt/bin/kubectl create namespace monitoring
    fi

    for namespace in kube-system monitoring
    do
        echo "---> Checking: $RIG_CHANGESET"
        if rig status $RIG_CHANGESET --retry-attempts=1 --retry-period=1s; then exit 0; fi

        echo "---> Starting update, changeset: $RIG_CHANGESET"
        rig cs delete --force -c cs/$RIG_CHANGESET

        echo "---> Deleting old 'heapster' resources"
        rig delete deployments/heapster --resource-namespace=$namespace --force

        echo "---> Deleting old 'influxdb' resources"
        rig delete deployments/influxdb --resource-namespace=$namespace --force

        echo "---> Deleting old 'grafana' resources"
        rig delete deployments/grafana --resource-namespace=$namespace --force

        echo "---> Deleting old 'telegraf' resources"
        rig delete deployments/telegraf --resource-namespace=$namespace --force
        rig delete daemonsets/telegraf-node --resource-namespace=$namespace --force

        echo "---> Deleting old deployment 'kapacitor'"
        rig delete deployments/kapacitor --resource-namespace=$namespace --force

        echo "---> Deleting old secrets"
        rig delete secrets/grafana --resource-namespace=$namespace --force
        rig delete secrets/grafana-influxdb-creds --resource-namespace=$namespace --force

        echo "---> Deleting old configmaps"
        for cfm in influxdb grafana-cfg grafana grafana-dashboards-cfg grafana-dashboards grafana-datasources
        do
            rig delete configmaps/$cfm --resource-namespace=$namespace --force
        done
    done

    echo "---> Creating new 'grafana' password"
    password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
    sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana.yaml

    echo "---> Creating or updating resources"
    for filename in security smtp influxdb grafana heapster kapacitor telegraf alerts
    do
        rig upsert -f /var/lib/gravity/resources/${filename}.yaml --debug
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

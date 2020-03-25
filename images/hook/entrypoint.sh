#!/usr/bin/env bash
set -e

echo "---> Assuming changeset from the environment: $RIG_CHANGESET"
# note that rig does not take explicit changeset ID
# taking it from the environment variables
if [ $1 = "update" ]; then
    echo "---> Checking: $RIG_CHANGESET"
    if rig status $RIG_CHANGESET --retry-attempts=1 --retry-period=1s; then exit 0; fi
    echo "---> Starting update, changeset: $RIG_CHANGESET"
    rig cs delete --force -c cs/$RIG_CHANGESET

    echo "---> Creating monitoring namespace"
    /opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml

    for namespace in kube-system monitoring
    do
        echo "---> Deleting resources in $namespace namespace"
        echo "---> Deleting old 'heapster' resources"
        rig delete deployments/heapster --resource-namespace=$namespace --force

        echo "---> Deleting old 'influxdb' resources"
        # Patch influxdb deployment but keep the pod scheduled to the same node after the update
        if kubectl --namespace=$namespace get deployment influxdb --ignore-not-found=false 2>/dev/null; then
            NODE_NAME=$(kubectl --namespace=$namespace get pod -l app=monitoring,component=influxdb -o go-template --template='{{(index .items 0).spec.nodeName}}')
        fi

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
        rig delete secrets/telegraf-influxdb-creds --resource-namespace=$namespace --force
        rig delete secrets/heapster-influxdb-creds --resource-namespace=$namespace --force

        echo "---> Deleting old configmaps"
        for cfm in influxdb grafana-cfg grafana grafana-dashboards-cfg grafana-dashboards grafana-datasources kapacitor-alerts rollups-default
        do
            rig delete configmaps/$cfm --resource-namespace=$namespace --force
        done
    done

    echo  "---> Moving smtp-cofiguration secret to monitoring namespace"
    if ! /opt/bin/kubectl --namespace=monitoring get secret smtp-configuration > /dev/null 2>&1; then
        if /opt/bin/kubectl --namespace=kube-system get secret smtp-configuration > /dev/null 2>&1; then
            /opt/bin/kubectl --namespace=kube-system get secret smtp-configuration --export=true -o json | \
                jq '.metadata.namespace = "monitoring"' > /tmp/resource.json
            rig upsert -f /tmp/resource.json --debug
            rig delete secrets/smtp-configuration --resource-namespace=kube-system --force
        fi
    fi

    echo  "---> Moving alerting-addresses configmap to monitoring namespace"
    if ! /opt/bin/kubectl --namespace=monitoring get configmap alerting-addresses > /dev/null 2>&1; then
        if /opt/bin/kubectl --namespace=kube-system get configmap alerting-addresses > /dev/null 2>&1; then
            /opt/bin/kubectl --namespace=kube-system get configmap alerting-addresses --export=true -o json | \
                jq '.metadata.namespace = "monitoring"' > /tmp/resource.json
            rig upsert -f /tmp/resource.json --debug
            rig delete configmaps/alerting-addresses --resource-namespace=kube-system --force
        fi
    fi

    # create influxdb secret in case it does not exist because upgrading from old gravity version
    if ! /opt/bin/kubectl --namespace=monitoring get secret influxdb > /dev/null 2>&1;
    then
        # backwards-compatible password
        sed -i s/b1lIV3gyVDlmQVd3SzdsZTRrZDY=/cm9vdA==/g /var/lib/gravity/resources/influxdb-secret.yaml
        rig upsert -f /var/lib/gravity/resources/influxdb-secret.yaml --debug
    fi
    
    # Generate password for Grafana administrator
    password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
    sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/secrets.yaml

    # Generate password for InfluxDB grafana user
    password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ')
    sed -i s/cGRGY29ma2NHYllIekRZMUdadmg=/$(echo -n $password | /opt/bin/base64)/g /var/lib/gravity/resources/secrets.yaml

    # Generate password for InfluxDB telegraf user
    password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
    sed -i s/bllMNU1sdEREeHFSTlFxMEVsZkY=/$password/g /var/lib/gravity/resources/secrets.yaml

    # Generate password for InfluxDB heapster user
    password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
    sed -i s/MTIxMzQyNDMyZHdkY2RldmdyZWc=/$password/g /var/lib/gravity/resources/secrets.yaml

    echo "---> Creating or updating resources"
    for filename in security secrets smtp influxdb grafana heapster kapacitor telegraf rollups alerts
    do
        rig upsert -f /var/lib/gravity/resources/${filename}.yaml --debug
    done

    TMPFILE="$(mktemp)"
    cat >$TMPFILE<<EOF
spec:
  template:
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 1
            preference:
              matchExpressions:
              - key: kubernetes.io/hostname
                operator: In
                values:
                - $NODE_NAME
EOF
    kubectl --namespace=monitoring patch deployment influxdb -p "$(cat $TMPFILE)"
    rm $TMPFILE

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

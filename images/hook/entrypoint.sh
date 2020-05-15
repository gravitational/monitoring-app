#!/usr/bin/env bash
set -e

echo "---> Assuming changeset from the environment: $RIG_CHANGESET"
# note that rig does not take explicit changeset ID
# taking it from the environment variables
if [ $1 = "update" ]; then
    echo "---> Creating monitoring namespace"
    /opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml

    echo "---> Getting node name for influxdb pod"
    # Patch influxdb deployment but keep the pod scheduled to the same node after the update
    if kubectl --namespace=monitoring get pod | grep -q influxdb; then
        NODE_NAME=$(kubectl --namespace=monitoring get pod -l app=monitoring,component=influxdb -o go-template --template='{{(index .items 0).spec.nodeName}}')
    fi

    # migrate old influxdb data if it exist after upgrade
    if [ -n "$NODE_NAME" ]; then
        sed "s/NODE_SELECTOR/$NODE_NAME/" resources/influxdb-data-migration.yaml | /opt/bin/kubectl create -f -
        /opt/bin/kubectl --namespace=monitoring wait --for=condition=complete --timeout=2m job/influxdb-data-migration
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
    for filename in security secrets smtp influxdb grafana heapster kapacitor telegraf alerts
    do
        rig upsert -f /var/lib/gravity/resources/${filename}.yaml --debug
    done

    if [ -n "$NODE_NAME" ]; then
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
        kubectl --namespace=monitoring delete pod -l app=monitoring,component=influxdb
        rm $TMPFILE
    fi

    echo "---> Checking status of updated resources"
    rig status $RIG_CHANGESET --retry-attempts=120 --retry-period=1s --debug

    echo "---> Updating rollups"
    rig upsert -f /var/lib/gravity/resources/rollups.yaml --debug

    for file in /var/lib/gravity/resources/nethealth/nethealth.yaml
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

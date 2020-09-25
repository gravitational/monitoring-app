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

    echo "---> Removing last applied configuration from resources"
    for deployment in grafana kube-state-metrics prometheus-adapter prometheus-operator watcher
    do
      # || true is needed in case resources does not have last applied configuration
      kubectl --namespace=monitoring patch deployments.apps $deployment --type=json -p='[{"op": "replace", "path": "/metadata/managedFields", "value": [{}]}]' || true
      kubectl --namespace=monitoring patch deployments.apps $deployment --type=json -p='[{"op": "remove", "path": "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"}]' || true
    done
    for daemonset in nethealth node-exporter
    do
      # || true is needed in case resources does not have last applied configuration
      kubectl --namespace=monitoring patch daemonsets.apps $daemonset --type=json -p='[{"op": "replace", "path": "/metadata/managedFields", "value": [{}]}]' || true
      kubectl --namespace=monitoring patch daemonsets.apps $daemonset --type=json -p='[{"op": "remove", "path": "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"}]' || true
    done
    for configmap in adapter-config grafana-cfg prometheus-k8s-rulefiles-0 grafana-dashboard-k8s-cluster-rsrc-use grafana-dashboard-k8s-resources-cluster \
      grafana-dashboard-k8s-resources-namespace grafana-dashboard-k8s-resources-pod grafana-dashboard-k8s-resources-workload grafana-dashboard-k8s-resources-workloads-namespace \
      grafana-dashboard-nodes grafana-dashboard-pods grafana-dashboard-nethealth grafana-dashboards
    do
      # || true is needed in case resources does not have last applied configuration
      kubectl --namespace=monitoring patch configmaps $configmap --type=json -p='[{"op": "replace", "path": "/metadata/managedFields", "value": [{}]}]' || true
      kubectl --namespace=monitoring patch configmaps $configmap --type=json -p='[{"op": "remove", "path": "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"}]' || true
    done

    echo "---> Deleting old configmaps"
    for configmap in influxdb grafana kapacitor-alerts rollups-default grafana-dashboard-k8s-cluster-rsrc-use grafana-dashboard-k8s-resources-cluster \
      grafana-dashboard-k8s-resources-namespace grafana-dashboard-k8s-resources-pod grafana-dashboard-k8s-resources-workload grafana-dashboard-k8s-resources-workloads-namespace \
      grafana-dashboard-nodes grafana-dashboard-pods grafana-dashboard-nethealth
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

    if [ $(kubectl get nodes -lgravitational.io/k8s-role=master --output=go-template --template="{{len .items}}") -gt 1 ]
    then
	kubectl --namespace monitoring patch prometheuses.monitoring.coreos.com k8s --type=json -p='[{"op": "replace", "path": "/spec/replicas", "value": 2}]'
	kubectl --namespace monitoring patch alertmanagers.monitoring.coreos.com main --type=json -p='[{"op": "replace", "path": "/spec/replicas", "value": 2}]'
    fi
    # check for readiness of prometheus pod
    kubectl --namespace monitoring wait --for=condition=ready pod prometheus-k8s-0


    # Remove unused nethealth objects
    # Todo: can be removed when upgrades from gravity 7.0 are no longer supported.
    /opt/bin/kubectl delete -n monitoring servicemonitor/nethealth || true
    /opt/bin/kubectl delete -n monitoring prometheusrule/prometheus-nethealth-rules || true
elif [ $1 = "rollback" ]; then
    echo "---> Reverting changeset $RIG_CHANGESET"
    rig revert
    rig cs delete --force -c cs/$RIG_CHANGESET
else
    echo "---> Missing argument, should be either 'update' or 'rollback'"
fi

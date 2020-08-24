#!/bin/sh

/opt/bin/kubectl create -f /var/lib/gravity/resources/namespace.yaml

for file in /var/lib/gravity/resources/crds/*
do
    /opt/bin/kubectl create -f -
done

# Wait until setup of CRDs are done
until /opt/bin/kubectl get servicemonitors --all-namespaces ; do echo $(date) ": Waiting for CRDs to setup"; sleep 5; done

# Generate password for Grafana administrator
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana.yaml

for name in security grafana watcher
do
    /opt/bin/kubectl create -f /var/lib/gravity/resources/${name}.yaml
done

/opt/bin/kubectl create -f /var/lib/gravity/resources/prometheus/
/opt/bin/kubectl create -f /var/lib/gravity/resources/nethealth/

if [ $(/opt/bin/kubectl get nodes -lgravitational.io/k8s-role=master --output=go-template --template="{{len .items}}") -gt 1 ]
then
    /opt/bin/kubectl --namespace monitoring patch prometheuses.monitoring.coreos.com k8s --type=json -p='[{"op": "replace", "path": "/spec/replicas", "value": 2}]'
    /opt/bin/kubectl --namespace monitoring patch alertmanagers.monitoring.coreos.com main --type=json -p='[{"op": "replace", "path": "/spec/replicas", "value": 2}]'
fi

# check for readiness of prometheus pod
/opt/bin/kubectl --namespace monitoring wait --for=condition=ready pod prometheus-k8s-0

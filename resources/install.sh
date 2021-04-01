#!/bin/sh
set -x

/opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml

for file in /var/lib/gravity/resources/kube-prometheus-setup/*
do
    /opt/bin/kubectl create -f $file
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

sed -i "s/runAsUser: -1/runAsUser: $GRAVITY_SERVICE_USER/" /var/lib/gravity/resources/prometheus/prometheus-prometheus.yaml
/opt/bin/kubectl create -f /var/lib/gravity/resources/prometheus/
/opt/bin/kubectl create -f /var/lib/gravity/resources/nethealth/

if [ $(/opt/bin/kubectl get nodes -lgravitational.io/k8s-role=master --output=go-template --template="{{len .items}}") -gt 1 ]
then
    /opt/bin/kubectl --namespace monitoring patch prometheuses.monitoring.coreos.com k8s --type=json -p='[{"op": "replace", "path": "/spec/replicas", "value": 2}]'
    /opt/bin/kubectl --namespace monitoring patch alertmanagers.monitoring.coreos.com main --type=json -p='[{"op": "replace", "path": "/spec/replicas", "value": 2}]'
fi

# check for readiness of prometheus pod
timeout 5m sh -c "while ! /opt/bin/kubectl get pod prometheus-k8s-0; do sleep 10; done"
/opt/bin/kubectl --namespace monitoring wait --for=condition=ready pod prometheus-k8s-0

#!/bin/sh
set -eux

for name in namespace priority-class
do
    /opt/bin/kubectl apply -f /var/lib/gravity/resources/${name}.yaml
done

# Generate password for Grafana administrator
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)

/opt/bin/helm3 install gravity --namespace monitoring /var/lib/gravity/resources/charts/kube-prometheus-stack
sed -i "s/runAsUser: -1/runAsUser: $GRAVITY_SERVICE_USER/" /var/lib/gravity/resources/prometheus/prometheus-prometheus.yaml
/opt/bin/kubectl create -f /var/lib/gravity/resources/prometheus/
/opt/bin/kubectl create -f /var/lib/gravity/resources/nethealth/

if [ $(/opt/bin/kubectl get nodes -lgravitational.io/k8s-role=master --output=go-template --template="{{len .items}}") -gt 1 ]
then
    /opt/bin/kubectl --namespace monitoring patch prometheuses.monitoring.coreos.com k8s --type=json -p='[{"op": "replace", "path": "/spec/replicas", "value": 2}]'
    /opt/bin/kubectl --namespace monitoring patch alertmanagers.monitoring.coreos.com main --type=json -p='[{"op": "replace", "path": "/spec/replicas", "value": 2}]'
fi

# check for readiness of prometheus pod
timeout 5m sh -c "while ! /opt/bin/kubectl --namespace=monitoring get pod prometheus-k8s-0; do sleep 10; done"
/opt/bin/kubectl --namespace monitoring wait --for=condition=ready pod prometheus-k8s-0

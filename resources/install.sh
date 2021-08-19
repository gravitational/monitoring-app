#!/bin/sh
set -eux

for name in namespace priority-class
do
    /opt/bin/kubectl apply -f /var/lib/gravity/resources/${name}.yaml
done

# Generate password for Grafana administrator
# password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)

/opt/bin/helm3 install nethealth --namespace monitoring /var/lib/gravity/resources/charts/nethealth

# check for readiness of prometheus pod
# timeout 5m sh -c "while ! /opt/bin/kubectl --namespace=monitoring get pod prometheus-k8s-0; do sleep 10; done"
# /opt/bin/kubectl --namespace monitoring wait --for=condition=ready pod prometheus-k8s-0

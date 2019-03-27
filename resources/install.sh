#!/bin/sh

/opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml
for name in security smtp grafana metrics-server alerts kube-state-metrics
do
    /opt/bin/kubectl apply -f /var/lib/gravity/resources/${name}.yaml
done

/opt/bin/kubectl apply -f /var/lib/gravity/resources/prometheus/

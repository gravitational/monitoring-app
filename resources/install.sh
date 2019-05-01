#!/bin/sh

/opt/bin/kubectl create -f /var/lib/gravity/resources/namespace.yaml

for file in /var/lib/gravity/resources/crds/*
do
    head -n -6 $file | /opt/bin/kubectl apply -f -
done

for name in security smtp grafana
do
    /opt/bin/kubectl create -f /var/lib/gravity/resources/${name}.yaml
done

/opt/bin/kubectl create -f /var/lib/gravity/resources/prometheus/

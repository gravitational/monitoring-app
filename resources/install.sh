#!/bin/sh

/opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml

for file in /var/lib/gravity/resources/crds/*
do
    head -n -6 $file | /opt/bin/kubectl apply -f -
done

# Generate password for Grafana administrator
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana.yaml

for name in security grafana watcher
do
    /opt/bin/kubectl apply -f /var/lib/gravity/resources/${name}.yaml
done

/opt/bin/kubectl apply -f /var/lib/gravity/resources/prometheus/
/opt/bin/kubectl apply -f /var/lib/gravity/resources/nethealth/

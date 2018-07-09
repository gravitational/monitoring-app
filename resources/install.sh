#!/bin/sh
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana-creds.yaml
/opt/bin/kubectl apply -f /var/lib/gravity/resources/influxdb-conf.yaml
# do not recreate secret if it exists
if ! kubectl --namespace=kube-system get secret grafana > /dev/null 2>&1
then
    /opt/bin/kubectl apply -f /var/lib/gravity/resources/grafana-creds.yaml
fi
/opt/bin/kubectl apply -f /var/lib/gravity/resources/grafana-cfg.yaml
/opt/bin/kubectl apply -f /var/lib/gravity/resources/resources.yaml
/opt/bin/kubectl apply -f /var/lib/gravity/resources/alerts.yaml

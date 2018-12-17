#!/bin/sh
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana.yaml

/opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml
for name in security smtp influxdb grafana metrics-server kapacitor telegraf alerts
do
    if [ -d $name ]; then
        /opt/bin/kubectl create -f /var/lib/gravity/resources/${name}/
    else
        /opt/bin/kubectl create -f /var/lib/gravity/resources/${name}.yaml
    fi
done

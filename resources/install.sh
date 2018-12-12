#!/bin/sh
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana.yaml

# Generate another password for InfluxDB administrator
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 16 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/influxdb.yaml

/opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml
for filename in security smtp influxdb grafana heapster kapacitor telegraf alerts
do
    /opt/bin/kubectl create -f /var/lib/gravity/resources/${filename}.yaml
done

#!/bin/sh
# Generate password for Grafana administrator
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/secrets.yaml

# Generate password for InfluxDB administrator
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/b1lIV3gyVDlmQVd3SzdsZTRrZDY=/$password/g /var/lib/gravity/resources/secrets.yaml

# Generate password for InfluxDB grafana user
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGRGY29ma2NHYllIekRZMUdadmg=/$password/g /var/lib/gravity/resources/secrets.yaml

# Generate password for InfluxDB telegraf user
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/bllMNU1sdEREeHFSTlFxMEVsZkY=/$password/g /var/lib/gravity/resources/secrets.yaml

/opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml
for filename in security secrets smtp influxdb grafana heapster kapacitor telegraf alerts
do
    /opt/bin/kubectl create -f /var/lib/gravity/resources/${filename}.yaml
done

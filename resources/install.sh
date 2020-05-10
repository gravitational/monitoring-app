#!/usr/bin/env bash
# Generate password for Grafana administrator
set -o nounset
set -o errexit

password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/secrets.yaml

# Generate password for InfluxDB administrator
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/b1lIV3gyVDlmQVd3SzdsZTRrZDY=/$password/g /var/lib/gravity/resources/influxdb-secret.yaml

# Generate password for InfluxDB grafana user
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ')
sed -i s/cGRGY29ma2NHYllIekRZMUdadmg=/$(echo -n $password | /opt/bin/base64)/g /var/lib/gravity/resources/secrets.yaml

# Generate password for InfluxDB telegraf user
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/bllMNU1sdEREeHFSTlFxMEVsZkY=/$password/g /var/lib/gravity/resources/secrets.yaml

# Generate password for InfluxDB heapster user
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/MTIxMzQyNDMyZHdkY2RldmdyZWc=/$password/g /var/lib/gravity/resources/secrets.yaml

/opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml
cat <<EOF | /opt/bin/kubectl create -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: gravity-service-user
  namespace: monitoring
data:
  service-uid: "$GRAVITY_SERVICE_USER"
EOF

for filename in security secrets influxdb-secret smtp influxdb grafana heapster kapacitor rollups telegraf alerts
do
    /opt/bin/kubectl create -f /var/lib/gravity/resources/${filename}.yaml
done

/opt/bin/kubectl apply -f /var/lib/gravity/resources/nethealth/

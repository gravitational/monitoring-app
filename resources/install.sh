#!/bin/sh
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana-creds.yaml
/opt/bin/kubectl create -f /var/lib/gravity/resources/grafana-creds.yaml

if [ "$DEVMODE" = "true" ]; then
    /opt/bin/kubectl create -f /var/lib/gravity/resources/grafana-ini-dev.yaml
else
    /opt/bin/kubectl create -f /var/lib/gravity/resources/grafana-ini-prod.yaml
fi

/opt/bin/kubectl create -f /var/lib/gravity/resources/resources.yaml

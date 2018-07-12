#!/bin/sh
password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)
sed -i s/cGFzc3dvcmQtZ29lcy1oZXJlCg==/$password/g /var/lib/gravity/resources/grafana.yaml

/opt/bin/kubectl apply -f /var/lib/gravity/resources/namespace.yaml
for filename in security smtp influxdb grafana heapster kapacitor telegraf alerts kube-state-metrics
do
    /opt/bin/kubectl create -f /var/lib/gravity/resources/${filename}.yaml
done

if [ $(/opt/bin/kubectl get nodes -l gravitational.io/k8s-role=master -o name | wc -l) -ge 3 ]
then
    /opt/bin/kubectl --namespace=kube-system scale --replicas=3 deployment kube-state-metrics.yaml
fi

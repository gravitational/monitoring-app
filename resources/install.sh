#!/bin/sh
/opt/bin/kubectl create secret generic grafana \
                 --namespace=kube-system \
                 --from-literal=username=admin \
                 --from-literal=password=$(date | md5sum | cut -d ' ' -f1)
/opt/bin/kubectl apply \
                 --filename=/var/lib/gravity/resources/resources.yaml

#!/bin/sh
/opt/bin/kubectl create secret generic grafana \
                 --namespace=kube-system \
                 --from-literal=username=admin \
                 --from-literal=password=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
/opt/bin/kubectl apply \
                 --filename=/var/lib/gravity/resources/resources.yaml

#!/bin/sh

kubectl delete --namespace kube-system \
   job/monitoring-bootstrap \
   rc/heapster \
   svc/heapster \
   rc/influxdb \
   svc/influxdb \
   rc/grafana \
   svc/grafana

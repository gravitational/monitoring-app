#!/bin/sh
/opt/bin/kubectl delete -f /var/lib/gravity/resources/nethealth/
/opt/bin/kubectl delete -f /var/lib/gravity/resources/prometheus/
/opt/bin/kubectl delete -f /var/lib/gravity/resources/security.yaml
/opt/bin/kubectl delete -f /var/lib/gravity/resources/grafana.yaml
/opt/bin/kubectl delete -f /var/lib/gravity/resources/watcher.yaml
/opt/bin/kubectl delete -f /var/lib/gravity/resources/crds/*
/opt/bin/kubectl delete -f /var/lib/gravity/resources/namespace.yaml

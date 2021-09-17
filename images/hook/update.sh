#!/usr/bin/env bash
set -eux

echo "---> Assuming changeset from the environment: $RIG_CHANGESET"
echo "---> Removing last applied configuration from resources"
for deployment in autoscaler grafana kube-state-metrics prometheus-adapter prometheus-operator watcher
do
  # || true is needed in case resource does not have last applied configuration
  kubectl --namespace=monitoring patch deployments.apps $deployment --type=json -p='[{"op": "replace", "path": "/metadata/managedFields", "value": [{}]}]' || true
  kubectl --namespace=monitoring patch deployments.apps $deployment --type=json -p='[{"op": "remove", "path": "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"}]' || true
done
for daemonset in nethealth node-exporter
do
  # || true is needed in case resource does not have last applied configuration
  kubectl --namespace=monitoring patch daemonsets.apps $daemonset --type=json -p='[{"op": "replace", "path": "/metadata/managedFields", "value": [{}]}]' || true
  kubectl --namespace=monitoring patch daemonsets.apps $daemonset --type=json -p='[{"op": "remove", "path": "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"}]' || true
done
for configmap in adapter-config grafana-cfg prometheus-k8s-rulefiles-0 grafana-dashboards
do
  # || true is needed in case resource does not have last applied configuration
  kubectl --namespace=monitoring patch configmaps $configmap --type=json -p='[{"op": "replace", "path": "/metadata/managedFields", "value": [{}]}]' || true
  kubectl --namespace=monitoring patch configmaps $configmap --type=json -p='[{"op": "remove", "path": "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"}]' || true
done

echo "---> Deleting pre-helm chart resources to avoid conflicts"
for daemonset in nethealth node-exporter
do
    rig delete daemonsets/$daemonset --resource-namespace=monitoring --force
done
for deployment in autoscaler grafana kube-state-metrics prometheus-adapter prometheus-operator watcher
do
    rig delete deployments/$deployment --resource-namespace=monitoring --force
done
for statefulset in alertmanager-main prometheus-k8s
do
    rig delete statefulsets/$statefulset --resource-namespace=monitoring --force
done
for configmap in adapter-config grafana-cfg prometheus-k8s-rulefiles-0 grafana-dashboards
do
    rig delete configmaps/$configmap --resource-namespace=monitoring --force
done
for secret in alertmanager-main grafana grafana-datasources prometheus-additional-scrape-configs \
  prometheus-k8s prometheus-k8s-tls-assets
do
    rig delete secrets/$secret --resource-namespace=monitoring --force
done
for service in alertmanager-main alertmanager-operated grafana kube-state-metrics \
  nethealth node-exporter prometheus-adapter prometheus-k8s prometheus-operated prometheus-operator
do
    rig delete services/$service --resource-namespace=monitoring --force
done
for serviceaccount in alertmanager-main kube-state-metrics monitoring monitoring-autoscaler monitoring-updater \
  nethealth node-exporter prometheus-adapter prometheus-k8s prometheus-operator
do
    rig delete serviceaccounts/$serviceaccount --resource-namespace=monitoring --force
done
rig delete alertmanagers/main --resource-namespace=monitoring --force
rig delete prometheuses/k8s --resource-namespace=monitoring --force
rig delete psp/nethealth --resource-namespace=monitoring --force
for promrule in gravity-k8s-rules prometheus-k8s-rules prometheus-nethealth-rules
do
    rig delete prometheusrules/$promrule --resource-namespace=monitoring --force
done
for servicemonitor in alertmanager coredns grafana kube-apiserver kube-state-metrics kubelet node-exporter prometheus prometheus-operator nethealth
do
    rig delete servicemonitors/$servicemonitor --resource-namespace=monitoring --force
done
for role in alertmanager-main kube-state-metrics monitoring monitoring:autoscaler monitoring:updater nethealth prometheus-k8s prometheus-k8s-config
do
    rig delete roles/"$role" --resource-namespace=monitoring --force
done
for rolebinding in alertmanager-main kube-state-metrics monitoring monitoring:autoscaler monitoring:updater nethealth prometheus-k8s prometheus-k8s-config
do
    rig delete rolebindings/"$rolebinding" --resource-namespace=monitoring --force
done
for clusterrole in kube-state-metrics monitoring:autoscaler monitoring:metrics nethealth node-exporter prometheus-adapter prometheus-k8s prometheus-operator
do
    rig delete clusterroles/"$clusterrole" --force
done
for clusterrolebinding in kube-state-metrics monitoring:autoscaler monitoring:metrics nethealth node-exporter prometheus-adapter prometheus-k8s prometheus-operator
do
    rig delete clusterrolebindings/"$clusterrolebinding" --force
done
for crd in alertmanagers.monitoring.coreos.com podmonitors.monitoring.coreos.com prometheuses.monitoring.coreos.com \
  prometheusrules.monitoring.coreos.com servicemonitors.monitoring.coreos.com
do
    rig delete crds/"$crd" --force
done

echo "---> Checking status"
rig status "$RIG_CHANGESET" --retry-attempts=120 --retry-period=1s --debug

echo "---> Freezing"
rig freeze

echo "---> Creating monitoring namespace"
kubectl apply -f /var/lib/gravity/resources/namespace.yaml

# Generate password for Grafana administrator
password=$(tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w 32 | head -n 1 | tr -d '\n ' | /opt/bin/base64)

/opt/bin/helm3 upgrade --install nethealth --namespace monitoring /var/lib/gravity/resources/charts/nethealth
/opt/bin/helm3 upgrade --install monitoring --namespace monitoring /var/lib/gravity/resources/charts/kube-prometheus-stack -f /var/lib/gravity/resources/custom-values.yaml \
    --set grafana.adminPassword="${password}" --set alertmanager.alertmanagerSpec.securityContext.runAsUser="$GRAVITY_SERVICE_USER" --set prometheus.prometheusSpec.securityContext.runAsUser="$GRAVITY_SERVICE_USER"
/opt/bin/helm3 upgrade --install watcher --namespace monitoring /var/lib/gravity/resources/charts/watcher -f /var/lib/gravity/resources/custom-values-watcher.yaml

# check for readiness of prometheus pod
timeout 5m sh -c "while ! kubectl --namespace=monitoring get pod prometheus-monitoring-kube-prometheus-prometheus-0; do sleep 10; done"
kubectl --namespace monitoring wait --for=condition=ready pod prometheus-monitoring-kube-prometheus-prometheus-0

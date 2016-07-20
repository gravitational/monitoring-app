apiVersion: v1
kind: SystemApplication
metadata:
  name: monitoring-app
  namespace: kube-system
  repository: gravitational.io
  resourceVersion: VERSION
hooks:
  install:
    spec:
      apiVersion: batch/v1
      kind: Job
      metadata:
        name: monitoring-app-install
      spec:
        template:
          metadata:
            name: monitoring-app-install
          spec:
            restartPolicy: OnFailure
            containers:
              - name: hook
                image: quay.io/gravitational/debian-tall:0.0.1
                command: ["/usr/local/bin/kubectl", "apply", "-f", "/var/lib/gravity/resources/resources.yaml"]
  uninstall:
    spec:
      apiVersion: batch/v1
      kind: Job
      metadata:
        name: monitoring-app-uninstall
      spec:
        template:
          metadata:
            name: monitoring-app-uninstall
          spec:
            restartPolicy: OnFailure
            containers:
              - name: hook
                image: quay.io/gravitational/debian-tall:0.0.1
                command: ["/usr/local/bin/kubectl", "delete", "-f", "/var/lib/gravity/resources/resources.yaml"]

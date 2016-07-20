apiVersion: v1
kind: ReplicationController
metadata:
  labels:
    k8s-app: heapster
    name: heapster
  name: heapster
  namespace: kube-system
spec:
  replicas: 1
  selector:
    k8s-app: heapster
  template:
    metadata:
      labels:
        k8s-app: heapster
    spec:
      containers:
      - name: heapster
        image: quay.io/gravitational/monitoring-heapster:VERSION
        imagePullPolicy: Always
        command:
        - /heapster
        - --source=kubernetes
        - --sink=influxdb:http://influxdb:8086
---
apiVersion: v1
kind: Service
metadata:
  labels:
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: Heapster
  name: heapster
  namespace: kube-system
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 8082
    nodePort: 30082
  selector:
    k8s-app: heapster
---
apiVersion: v1
kind: ReplicationController
metadata:
  labels:
    name: influxdb
  name: influxdb
  namespace: kube-system
spec:
  replicas: 1
  selector:
    name: influxdb
  template:
    metadata:
      labels:
        name: influxdb
    spec:
      containers:
      - name: influxdb
        image: quay.io/gravitational/monitoring-influxdb:VERSION
        imagePullPolicy: Always
        volumeMounts:
        - mountPath: /data
          name: influxdb-storage
      volumes:
      - name: influxdb-storage
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  labels: null
  name: influxdb
  namespace: kube-system
spec:
  type: NodePort
  ports:
  - name: http
    port: 8083
  - name: api
    port: 8086
  selector:
    name: influxdb
---
apiVersion: v1
kind: ReplicationController
metadata:
  labels:
    k8s-app: grafana
    name: grafana
  name: grafana
  namespace: kube-system
spec:
  replicas: 1
  selector:
    app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: quay.io/gravitational/monitoring-grafana:VERSION
        imagePullPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  labels:
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: Grafana
  name: grafana
  namespace: kube-system
spec:
  type: NodePort
  ports:
  - name: grafana
    port: 3000
  selector:
    app: grafana

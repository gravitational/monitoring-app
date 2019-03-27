FROM quay.io/gravitational/debian-venti:go1.10.3-stretch as builder
ARG KUBE_STATE_METRICS_VERSION
ENV PKG k8s.io/kube-state-metrics
WORKDIR $GOPATH
RUN go get k8s.io/kube-state-metrics && \
    cd $GOPATH/src/k8s.io/kube-state-metrics && \
    git checkout v$KUBE_STATE_METRICS_VERSION && \
    GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w -X ${PKG}/version.Release=${KUBE_STATE_METRICS_VERSION} -X ${PKG}/version.Commit=${Commit}" -o kube-state-metrics

FROM quay.io/gravitational/debian-tall:stretch
COPY --from=builder /gopath/src/k8s.io/kube-state-metrics/kube-state-metrics /usr/local/bin/
ENTRYPOINT ["/usr/bin/dumb-init", "/usr/local/bin/kube-state-metrics", "--port=8080", "--telemetry-port=8081"]

EXPOSE 8080 8081

ARG GOLANG_VERSION=1.10
FROM golang:${GOLANG_VERSION} as builder

ARG ARCH=amd64
ARG METRICS_SERVER_VERSION=0.3.1

WORKDIR /go/src/github.com/kubernetes-incubator/metrics-server/
RUN git clone https://github.com/kubernetes-incubator/metrics-server.git . && \
    git checkout v${METRICS_SERVER_VERSION} && \
    make


FROM quay.io/gravitational/debian-tall:stretch
ARG ARCH=amd64
COPY --from=builder /go/src/github.com/kubernetes-incubator/metrics-server/_output/${ARCH}/metrics-server /usr/bin/metrics-server

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/bin/metrics-server"]

FROM quay.io/gravitational/debian-grande:stretch AS builder
ARG TELEGRAF_VERSION

ENV DEBIAN_FRONTEND=noninteractive \
    TERM=xterm

RUN apt-get update && \
    apt-get install --yes --no-install-recommends curl tar && \
    curl -sSL https://dl.influxdata.com/telegraf/releases/telegraf-${TELEGRAF_VERSION}_linux_amd64.tar.gz -o /telegraf.tar.gz && \
    tar xzf /telegraf.tar.gz --strip-components=2

FROM quay.io/gravitational/debian-tall:stretch
COPY --from=builder /usr/bin/telegraf /usr/bin/telegraf
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/bin/telegraf", "--config", "/etc/telegraf/telegraf.conf"]

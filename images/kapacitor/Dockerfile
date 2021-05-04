FROM quay.io/gravitational/debian-grande:stretch as downloader
ARG KAPACITOR_VERSION

RUN set -ex && apt-get update && \
    apt-get install -y curl tar && \
    curl -sSL https://dl.influxdata.com/kapacitor/releases/kapacitor-${KAPACITOR_VERSION}_linux_amd64.tar.gz -o /kapacitor.tar.gz && \
    tar xzf /kapacitor.tar.gz --strip-components=2

FROM quay.io/gravitational/debian-tall:stretch
COPY --from=downloader /usr/bin/kapacitor /usr/bin/kapacitord /usr/bin/tickfmt /usr/bin/

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN set -ex && test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

EXPOSE 9092

VOLUME /var/lib/kapacitor

COPY kapacitor.conf /etc/kapacitor/kapacitor.conf
COPY entrypoint.sh /entrypoint.sh
COPY loadalerts.sh /loadalerts.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/entrypoint.sh"]

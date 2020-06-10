ARG GRAFANA_VERSION
FROM grafana/grafana:${GRAFANA_VERSION}

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

COPY rootfs/ /
COPY run.sh /run.sh

EXPOSE 3000
ENTRYPOINT /run.sh

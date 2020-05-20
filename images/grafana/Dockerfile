FROM debian:stretch-slim

ARG GRAFANA_VERSION
ARG GRAFANA_TGZ="https://dl.grafana.com/oss/release/grafana-${GRAFANA_VERSION}.linux-amd64.tar.gz"

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN set -ex && test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

RUN set -ex && apt-get update && apt-get install -qq -y tar && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

ADD ${GRAFANA_TGZ} /tmp/grafana.tar.gz

RUN set -ex && mkdir /tmp/grafana && tar xfvz /tmp/grafana.tar.gz --strip-components=1 -C /tmp/grafana

FROM debian:stretch-slim

ARG GF_UID="472"
ARG GF_GID="472"

ENV PATH=/usr/share/grafana/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    GF_PATHS_CONFIG="/etc/grafana/grafana.ini" \
    GF_PATHS_DATA="/var/lib/grafana" \
    GF_PATHS_HOME="/usr/share/grafana" \
    GF_PATHS_LOGS="/var/log/grafana" \
    GF_PATHS_PLUGINS="/var/lib/grafana/plugins" \
    GF_PATHS_PROVISIONING="/etc/grafana/provisioning"

WORKDIR $GF_PATHS_HOME

RUN set -ex && apt-get update && apt-get install -qq -y libfontconfig ca-certificates gettext-base && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

COPY --from=0 /tmp/grafana "$GF_PATHS_HOME"

RUN set -ex && groupadd -r -g $GF_GID grafana && \
    useradd -r -u $GF_UID -g grafana grafana && \
    mkdir -p "$GF_PATHS_PROVISIONING/datasources" \
             "$GF_PATHS_PROVISIONING/dashboards" \
             "$GF_PATHS_LOGS" \
             "$GF_PATHS_PLUGINS" \
             "$GF_PATHS_DATA" && \
    chown -R grafana:grafana "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS" "$GF_PATHS_PROVISIONING" && \
    chmod 777 "$GF_PATHS_DATA" "$GF_PATHS_LOGS" "$GF_PATHS_PLUGINS"

EXPOSE 3000

COPY ./run.sh /run.sh

USER grafana
ENTRYPOINT [ "/run.sh" ]

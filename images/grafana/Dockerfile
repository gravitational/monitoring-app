ARG GRAFANA_VERSION
FROM grafana/grafana:${GRAFANA_VERSION}

RUN apt-get update && \
    apt-get install gettext --no-install-recommends && \
    apt-get clean && \
    rm -rf \
        /var/lib/apt/lists/* \
        ~/.bashrc \
        /usr/share/doc/ \
        /usr/share/doc-base/ \
        /usr/share/man/ \
        /tmp/*

COPY run.sh /run.sh

EXPOSE 3000
ENTRYPOINT /run.sh

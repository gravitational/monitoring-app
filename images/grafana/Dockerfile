ARG GRAFANA_VERSION
FROM grafana/grafana:${GRAFANA_VERSION}

RUN apt-get update && \
    apt-get install -qq -y gettext-base && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

COPY run.sh /run.sh

EXPOSE 3000
ENTRYPOINT /run.sh

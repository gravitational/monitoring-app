ARG GRAFANA_VERSION
FROM grafana/grafana:${GRAFANA_VERSION}

RUN apt-get update && \
    apt-get install -y curl

COPY run.sh /run.sh

EXPOSE 3000
ENTRYPOINT /run.sh

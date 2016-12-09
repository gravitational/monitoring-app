FROM grafana/grafana:3.0.4

RUN apt-get update && \
    apt-get install -y curl

COPY run.sh /run.sh

EXPOSE 3000
ENTRYPOINT /run.sh

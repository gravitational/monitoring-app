FROM quay.io/gravitational/debian-grande:0.0.1

RUN apt-get update && apt-get install -y curl

ENV INFLUXDB_HOME /influxdb
ENV PATH $INFLUXDB_HOME:$PATH

# admin, http, udp, cluster, graphite, opentsdb, collectd
EXPOSE 8083 8086 8086/udp 8088 2003 4242 25826

RUN mkdir -p $INFLUXDB_HOME /data /logs

ADD Dockerfile /
ADD build/influx* $INFLUXDB_HOME/
ADD config.toml /etc/influxdb.toml

# volume used for storing database logs and data
VOLUME ["/data"]

# volume used for storing debug logs
VOLUME ["/logs"]

COPY run.sh /run.sh
ENTRYPOINT /run.sh

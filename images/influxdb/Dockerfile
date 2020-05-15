FROM quay.io/gravitational/debian-tall:stretch

ENV INFLUXDB_HOME /influxdb
ENV PATH $INFLUXDB_HOME:$PATH

# admin, http, udp, cluster, graphite, opentsdb, collectd
EXPOSE 8083 8086 8086/udp 8088 2003 4242 25826

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

RUN mkdir -p $INFLUXDB_HOME /data /logs

ADD build/influx* $INFLUXDB_HOME/

# volume used for storing database logs and data
VOLUME ["/data"]

# volume used for storing debug logs
VOLUME ["/logs"]

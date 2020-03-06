FROM quay.io/gravitational/debian-grande:stretch
ARG KAPACITOR_VERSION

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

RUN apt-get update && \
    apt-get install -y curl wget gnupg dirmngr

RUN for key in 05CE15085FC09D18E99EFB22684A14CF2582E0C5; \
    do \
      gpg --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys "$key" || \
      gpg --keyserver hkp://ipv4.pool.sks-keyservers.net --recv-keys "$key" || \
      gpg --keyserver hkp://pgp.mit.edu:80 --recv-keys "$key" ; \
    done

RUN wget -q https://dl.influxdata.com/kapacitor/releases/kapacitor_${KAPACITOR_VERSION}_amd64.deb.asc && \
    wget -q https://dl.influxdata.com/kapacitor/releases/kapacitor_${KAPACITOR_VERSION}_amd64.deb && \
    gpg --batch --verify kapacitor_${KAPACITOR_VERSION}_amd64.deb.asc kapacitor_${KAPACITOR_VERSION}_amd64.deb && \
    dpkg -i kapacitor_${KAPACITOR_VERSION}_amd64.deb && \
    rm -f kapacitor_${KAPACITOR_VERSION}_amd64.deb*
COPY kapacitor.conf /etc/kapacitor/kapacitor.conf

EXPOSE 9092

VOLUME /var/lib/kapacitor

RUN test -f /cleanup.sh && sh /cleanup.sh

COPY entrypoint.sh /entrypoint.sh
COPY loadalerts.sh /loadalerts.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/entrypoint.sh"]

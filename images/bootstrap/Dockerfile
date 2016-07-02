FROM quay.io/gravitational/debian-tall:0.0.1

RUN mkdir -p /usr/local/bin/

ADD ./kubectl /usr/local/bin/

RUN test -f /cleanup.sh && sh /cleanup.sh

CMD ["/usr/local/bin/kubectl", "create", "-f", "/var/lib/gravity/resources/resources.yaml"]

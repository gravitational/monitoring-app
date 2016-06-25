FROM quay.io/gravitational/debian-tall:0.0.1

RUN mkdir -p /usr/local/bin/

ADD ./kubectl /usr/local/bin/
ADD ./resources.yml /usr/local/etc/resources.yml

RUN test -f /cleanup.sh && sh /cleanup.sh

CMD ["/usr/local/bin/kubectl", "create", "-f", "/usr/local/etc/resources.yml"]

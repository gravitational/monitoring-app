FROM quay.io/gravitational/debian-tall:0.0.1

RUN mkdir -p /usr/local/bin/

ADD ./kubectl /usr/local/bin/
ADD ./uninstall.sh /uninstall.sh

RUN test -f /cleanup.sh && sh /cleanup.sh

CMD ["/bin/sh", "/uninstall.sh"]

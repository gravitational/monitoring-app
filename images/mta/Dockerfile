FROM quay.io/gravitational/debian-grande:stretch

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

RUN apt-get update && \
    apt-get install -y exim4-daemon-light && \
    apt-get clean && \
    rm -rf \
        /var/lib/apt/lists/* \
        ~/.bashrc \
        /usr/share/doc/ \
        /usr/share/doc-base/ \
        /usr/share/man/ \
        /tmp/*

COPY entrypoint.sh /usr/local/bin/

RUN chmod a+x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/entrypoint.sh", "exim", "-bdf", "-v", "-q15m"]

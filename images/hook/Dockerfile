FROM quay.io/gravitational/rig:7.1.7

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

ARG CHANGESET
ENV RIG_CHANGESET $CHANGESET

RUN apt-get update && \
    apt-get install --yes --no-install-recommends jq && \
    apt-get clean && \
    rm -rf \
        /var/lib/apt/lists/* \
        ~/.bashrc \
        /usr/share/doc/ \
        /usr/share/doc-base/ \
        /usr/share/man/ \
        /tmp/*

ADD update.sh rollback.sh /
RUN chmod +x /update.sh /rollback.sh

ENTRYPOINT ["dumb-init", "/update.sh"]

FROM quay.io/gravitational/rig:5.5.8

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN set -ex && test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

ARG CHANGESET
ENV RIG_CHANGESET $CHANGESET

RUN set -ex && apt-get update && \
    apt-get install --yes --no-install-recommends jq && \
    apt-get clean && \
    rm -rf \
        /var/lib/apt/lists/* \
        ~/.bashrc \
        /usr/share/doc/ \
        /usr/share/doc-base/ \
        /usr/share/man/ \
        /tmp/*

ADD entrypoint.sh /

ENTRYPOINT ["dumb-init", "/entrypoint.sh"]

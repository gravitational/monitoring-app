FROM quay.io/gravitational/rig:0.0.13

ARG CHANGESET
ENV RIG_CHANGESET $CHANGESET

RUN apt-get update && \
    apt-get --yes --no-install-recommends jq && \
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

FROM quay.io/gravitational/rig:0.0.9-dev

ARG CHANGESET
ENV RIG_CHANGESET $CHANGESET

ADD entrypoint.sh /

ENTRYPOINT ["dumb-init", "/entrypoint.sh"]

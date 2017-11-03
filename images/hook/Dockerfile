FROM quay.io/gravitational/rig:0.0.6

ARG CHANGESET
ENV RIG_CHANGESET $CHANGESET

ADD entrypoint.sh /

ENTRYPOINT ["dumb-init", "/entrypoint.sh"]

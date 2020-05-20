FROM quay.io/gravitational/debian-tall:stretch

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN set -ex && test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

# FIXME: check this
# cAdvisor discovery via external files.
VOLUME /var/run/heapster/hosts

ADD Dockerfile /
ADD build/heapster /heapster

ENTRYPOINT ["/heapster"]

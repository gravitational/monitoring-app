FROM quay.io/gravitational/debian-tall:stretch

# Ensure nsswitch is set so localhost will be resolved locally
# https://github.com/gravitational/gravity/issues/1046
# https://github.com/golang/go/issues/35305
RUN test -e /etc/nsswitch.conf || echo 'hosts: files dns' > /etc/nsswitch.conf

ADD Dockerfile /
ADD build/watcher /
ENTRYPOINT ["/usr/bin/dumb-init", "/watcher"]

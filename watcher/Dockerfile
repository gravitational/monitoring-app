FROM quay.io/gravitational/debian-tall
ADD Dockerfile /
ADD build/watcher /
ENTRYPOINT ["/usr/bin/dumb-init", "/watcher"]

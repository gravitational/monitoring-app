FROM quay.io/gravitational/debian-tall:stretch
ADD Dockerfile /
ADD build/watcher /
ENTRYPOINT ["/usr/bin/dumb-init", "/watcher"]

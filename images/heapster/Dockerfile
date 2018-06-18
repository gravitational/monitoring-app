FROM quay.io/gravitational/debian-tall:stretch

# FIXME: check this
# cAdvisor discovery via external files.
VOLUME /var/run/heapster/hosts

ADD Dockerfile /
ADD build/heapster /heapster

ENTRYPOINT ["/heapster"]

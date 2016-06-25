FROM quay.io/gravitational/debian-tall:0.0.1

# FIXME: check this
# cAdvisor discovery via external files.
VOLUME /var/run/heapster/hosts

ADD Dockerfile /
ADD build/heapster /heapster

ENTRYPOINT ["/heapster"]

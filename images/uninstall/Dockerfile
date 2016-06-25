FROM ubuntu:14.04

RUN apt-get update && apt-get install -y curl

RUN curl -O https://storage.googleapis.com/kubernetes-release/release/v1.2.0/bin/linux/amd64/kubectl && \
    cp kubectl /usr/local/bin && \
    chmod +x /usr/local/bin/kubectl

ADD ./uninstall.sh /uninstall.sh

CMD ["/bin/sh", "/uninstall.sh"]

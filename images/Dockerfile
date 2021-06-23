FROM quay.io/gravitational/debian-venti:go1.9-stretch

ARG UID
ARG GID

RUN getent group  $GID || groupadd builder --gid=$GID -o && \
    getent passwd $UID || useradd builder --uid=$UID --gid=$GID --create-home --shell=/bin/sh;

RUN mkdir -p /.cache && chown -R $UID:$GID /.cache && chown -R $UID:$GID /gopath

USER $UID:$GID

RUN go get github.com/tools/godep

ENV PATH="$PATH:/gopath/bin"

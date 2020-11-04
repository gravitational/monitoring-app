BUILDBOX := quay.io/gravitational/debian-venti:go1.15.3-buster
BUILDDIR := $(abspath build)
CURDIR := $(dir $(CURDIR)/$(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST)))
DSTDIR := /gopath/src/github.com/gravitational/monitoring-app/watcher

NAME = watcher
VERSION ?= 0.0.3


#
# all builds a "watcher" container
#
.PHONY: all
all: container


#
# dev builds the "watcher" binary on host
#
.PHONY: dev
dev: build-inside-container


#
# container builds a docker container with the "watcher" binary
#
.PHONY: container
container: build
	docker build -t $(NAME):$(VERSION) .


#
# build builds the "watcher" binary
#
.PHONY: build
build:
	mkdir -p $(BUILDDIR)
	docker run -i --rm=true \
		-u $$(id -u):$$(id -g) \
		-v $(CURDIR):$(DSTDIR) \
		-e GOCACHE=/tmp/.cache \
		$(BUILDBOX) \
		/bin/bash -c "make -C $(DSTDIR) build-inside-container"


#
# build-inside-container builds the "watcher" binary inside the buildbox
#
.PHONY: build-inside-container
build-inside-container: tool/watcher/*.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-w" -o $(BUILDDIR)/$(NAME) github.com/gravitational/monitoring-app/watcher/tool/watcher


#
# clean deletes the built "watcher" binary
#
.PHONY: clean
clean:
	rm -rf $(BUILDDIR)

BUILDBOX := quay.io/gravitational/debian-venti:go1.7.1-jessie
BUILDDIR := $(abspath build)
CURDIR := $(dir $(CURDIR)/$(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST)))
DSTDIR := /gopath/src/github.com/gravitational/monitoring-app/watcher

NAME = watcher
VER ?= 0.0.2


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
	docker build -t $(NAME):$(VER) .


#
# build builds the "watcher" binary
#
.PHONY: build
build:
	mkdir -p $(BUILDDIR)
	docker run -i --rm=true \
		-v $(CURDIR):$(DSTDIR) \
		$(BUILDBOX) \
		/bin/bash -c "make -C $(DSTDIR) build-inside-container"


#
# build-inside-container builds the "watcher" binary inside the buildbox
#
.PHONY: build-inside-container
build-inside-container:
	go build -ldflags "-w" -o $(BUILDDIR)/$(NAME) github.com/gravitational/monitoring-app/watcher/tool/watcher


#
# clean deletes the built "watcher" binary
#
.PHONY: clean
clean:
	rm -rf $(BUILDDIR)

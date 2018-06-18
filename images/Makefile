.PHONY: all heapster heapster-build heapster-clean influxdb influxdb-build \
	influxdb-clean grafana kapacitor telegraf deploy

GRAFANA_VERSION := 5.1.0
# TODO: heapster is marked as deprecated project and will be removed in k8s 1.13
# https://github.com/kubernetes/heapster/blob/master/docs/deprecation.md
HEAPSTER_VERSION := 1.5.0
INFLUXDB_VERSION := 1.5.2
KAPACITOR_VERSION := 1.5.0
TELEGRAF_VERSION := 1.6.4

PWD=$(shell pwd)
.DEFAULT_GOAL := all

INFLUXDB_TAG  := monitoring-influxdb:$(VERSION)
HEAPSTER_TAG  := monitoring-heapster:$(VERSION)
GRAFANA_TAG   := monitoring-grafana:$(VERSION)
KAPACITOR_TAG := monitoring-kapacitor:$(VERSION)
TELEGRAF_TAG  := monitoring-telegraf:$(VERSION)
HOOK_TAG := monitoring-hook:$(VERSION)

REPO_URL := quay.io/gravitational

.PHONY: hook
hook:
	$(eval CHANGESET = $(shell echo $$VERSION | sed -e 's/[\.]//g'))
	if [ -z "$(CHANGESET)" ]; then \
	  echo "CHANGESET is not set"; exit 1; \
	fi;
	$(MAKE) -e BUILDIMAGE=$(HOOK_TAG) EXTRA_ARGS="--build-arg CHANGESET=monitoring-$(CHANGESET)" TARGETDIR=$@ make-docker-image

heapster-build:
	$(MAKE) -C $(PWD)/heapster -e VER=v$(HEAPSTER_VERSION) TARGET=heapster TARGETDIR=heapster -f $(PWD)/buildbox.mk

heapster-clean:
	$(MAKE) -C $(PWD)/heapster -f Makefile clean

heapster: heapster-build
	$(MAKE) -e BUILDIMAGE=$(HEAPSTER_TAG) TARGETDIR=$@ make-docker-image

influxdb-build:
	$(MAKE) -C $(PWD)/influxdb -e VER=v$(INFLUXDB_VERSION) TARGET=influxd TARGETDIR=influxdb -f $(PWD)/buildbox.mk

influxdb-clean:
	$(MAKE) -C $(PWD)/influxdb clean

influxdb: influxdb-build
	$(MAKE) -e BUILDIMAGE=$(INFLUXDB_TAG) TARGETDIR=$@ make-docker-image

grafana:
	$(MAKE) -e BUILDIMAGE=$(GRAFANA_TAG) EXTRA_ARGS="--build-arg GRAFANA_VERSION=$(GRAFANA_VERSION)" TARGETDIR=$@ make-docker-image

kapacitor:
	$(MAKE) -e BUILDIMAGE=$(KAPACITOR_TAG) EXTRA_ARGS="--build-arg KAPACITOR_VERSION=$(KAPACITOR_VERSION)" TARGETDIR=$@ make-docker-image

telegraf:
	$(MAKE) -e BUILDIMAGE=$(TELEGRAF_TAG) EXTRA_ARGS="--build-arg TELEGRAF_VERSION=$(TELEGRAF_VERSION)" TARGETDIR=$@ make-docker-image

all: heapster influxdb grafana kapacitor telegraf hook

clean: heapster-clean influxdb-clean

.PHONY: deploy
deploy:
	$(foreach ct,$(INFLUXDB_TAG) $(HEAPSTER_TAG) $(GRAFANA_TAG) $(KAPACITOR_TAG), \
		docker tag $(ct) $(REPO_URL)/$(ct) ; \
		docker push $(REPO_URL)/$(ct) ; )

make-docker-image:
	docker build --pull $(EXTRA_ARGS) -t $(BUILDIMAGE) $(TARGETDIR)

.PHONY: all metrics-server metrics-server-build metrics-server-clean influxdb influxdb-build \
	influxdb-clean grafana kapacitor telegraf deploy kube-state-metrics

GRAFANA_VERSION := 5.1.0
METRICS_SERVER_VERSION := 0.3.1
INFLUXDB_VERSION := 1.5.2
KAPACITOR_VERSION := 1.5.0
TELEGRAF_VERSION := 1.9.1
KUBE_STATE_METRICS_VERSION := 1.4.0

PWD=$(shell pwd)
.DEFAULT_GOAL := all

INFLUXDB_TAG  := monitoring-influxdb:$(VERSION)
METRICS_SERVER_TAG  := monitoring-metrics-server:$(VERSION)
GRAFANA_TAG   := monitoring-grafana:$(VERSION)
KAPACITOR_TAG := monitoring-kapacitor:$(VERSION)
TELEGRAF_TAG  := monitoring-telegraf:$(VERSION)
HOOK_TAG := monitoring-hook:$(VERSION)
KUBE_STATE_METRICS_TAG := monitoring-kube-state-metrics:$(VERSION)

REPO_URL := quay.io/gravitational

.PHONY: hook
hook:
	$(eval CHANGESET = $(shell echo $$VERSION | sed -e 's/[\.]//g'))
	if [ -z "$(CHANGESET)" ]; then \
	  echo "CHANGESET is not set"; exit 1; \
	fi;
	$(MAKE) -e BUILDIMAGE=$(HOOK_TAG) EXTRA_ARGS="--build-arg CHANGESET=monitoring-$(CHANGESET)" TARGETDIR=$@ make-docker-image

metrics-server:
	$(MAKE) -e BUILDIMAGE=$(METRICS_SERVER_TAG) TARGETDIR=$@ make-docker-image

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

kube-state-metrics:
	$(MAKE) -e BUILDIMAGE=$(KUBE_STATE_METRICS_TAG) EXTRA_ARGS="--build-arg KUBE_STATE_METRICS_VERSION=$(KUBE_STATE_METRICS_VERSION)" TARGETDIR=$@ make-docker-image

all: metrics-server influxdb grafana kapacitor telegraf hook kube-state-metrics

clean: influxdb-clean

.PHONY: deploy
deploy:
	$(foreach ct,$(INFLUXDB_TAG) $(METRICS_SERVER_TAG) $(GRAFANA_TAG) $(KAPACITOR_TAG), \
		docker tag $(ct) $(REPO_URL)/$(ct) ; \
		docker push $(REPO_URL)/$(ct) ; )

make-docker-image:
	docker build --pull $(EXTRA_ARGS) -t $(BUILDIMAGE) $(TARGETDIR)

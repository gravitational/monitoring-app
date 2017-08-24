.PHONY: all heapster heapster-build heapster-clean influxdb influxdb-build \
	influxdb-clean grafana kapacitor telegraf deploy

PWD=$(shell pwd)
.DEFAULT_GOAL := all

INFLUXDB_TAG  := monitoring-influxdb:1.2.2
HEAPSTER_TAG  := monitoring-heapster:1.4.1
GRAFANA_TAG   := monitoring-grafana:3.0.4
KAPACITOR_TAG := monitoring-kapacitor:1.3
TELEGRAF_TAG  := monitoring-telegraf:1.2.1
HOOK_TAG := monitoring-hook:$(VER)

REPO_URL := quay.io/gravitational

.PHONY: hook
hook:
	$(eval CHANGESET = $(shell echo $$VER | sed -e 's/[\.]//g'))
	if [ -z "$(CHANGESET)" ]; then \
	  echo "CHANGESET is not set"; exit 1; \
	fi;
	docker build --build-arg CHANGESET=monitoring-$(CHANGESET) --pull -t $(HOOK_TAG) hook

heapster-build:
	$(MAKE) -C $(PWD)/heapster -e TARGET=heapster TARGETDIR=heapster -f $(PWD)/buildbox.mk

heapster-clean:
	$(MAKE) -C $(PWD)/heapster -f Makefile clean

heapster: heapster-build
	docker build --pull -t $(HEAPSTER_TAG) $@

influxdb-build:
	$(MAKE) -C $(PWD)/influxdb -e TARGET=influxd TARGETDIR=influxdb -f $(PWD)/buildbox.mk

influxdb-clean:
	$(MAKE) -C $(PWD)/influxdb clean

influxdb: influxdb-build
	docker build --pull -t $(INFLUXDB_TAG) $@

grafana:
	docker build --pull -t $(GRAFANA_TAG) $@

kapacitor:
	docker build --pull -t $(KAPACITOR_TAG) $@

telegraf:
	docker build --pull -t $(TELEGRAF_TAG) $@

all: heapster influxdb grafana kapacitor telegraf hook

clean: heapster-clean influxdb-clean

.PHONY: deploy
deploy:
	$(foreach ct,$(INFLUXDB_TAG) $(HEAPSTER_TAG) $(GRAFANA_TAG) $(KAPACITOR_TAG), \
		docker tag $(ct) $(REPO_URL)/$(ct) ; \
		docker push $(REPO_URL)/$(ct) ; )

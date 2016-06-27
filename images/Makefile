.PHONY: all heapster heapster-build heapster-clean influxdb influxdb-build \
	influxdb-clean grafana bootstrap uninstall deploy

PWD=$(shell pwd)
.DEFAULT_GOAL := all

VER := 0.0.1
BOOTSTRAP_TAG := monitoring-bootstrap:$(VER)
UNINSTALL_TAG := monitoring-uninstall:$(VER)
INFLUXDB_TAG  := monitoring-influxdb:0.13.0
HEAPSTER_TAG  := monitoring-heapster:1.0.2
GRAFANA_TAG   := monitoring-grafana:v2.6.0-2

REPO_URL := quay.io/gravitational

KUBECTL_URL := https://storage.googleapis.com/kubernetes-release/release/v1.2.0/bin/linux/amd64/kubectl

kubectl:
	curl -o kubectl -L $(KUBECTL_URL)
	chmod +x kubectl

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

bootstrap: kubectl
	cp kubectl bootstrap/
	docker build --pull -t $(BOOTSTRAP_TAG) bootstrap

uninstall: kubectl
	cp kubectl uninstall/
	docker build --pull -t $(UNINSTALL_TAG) uninstall

all: heapster influxdb grafana bootstrap uninstall

clean: heapster-clean influxdb-clean

.PHONY: deploy
deploy:
	$(foreach ct,$(INFLUXDB_TAG) $(HEAPSTER_TAG) $(GRAFANA_TAG), \
		docker tag $(ct) $(REPO_URL)/$(ct) ; \
		docker push $(REPO_URL)/$(ct) ; )


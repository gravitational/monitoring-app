.PHONY: all heapster heapster-build heapster-clean influxdb influxdb-build \
	influxdb-clean grafana bootstrap uninstall buildbox

PWD=$(shell pwd)
.DEFAULT_GOAL := all

VER := 0.0.1
BOOTSTRAP_TAG := monitoring-bootstrap:$(VER)
UNINSTALL_TAG := monitoring-uninstall:$(VER)

KUBECTL_URL := https://storage.googleapis.com/kubernetes-release/release/v1.2.0/bin/linux/amd64/kubectl

kubectl:
	curl -o kubectl -L $(KUBECTL_URL)
	chmod +x kubectl

heapster-build:
	$(MAKE) -C $(PWD)/heapster -e TARGET=heapster TARGETDIR=heapster -f $(PWD)/buildbox.mk

heapster-clean:
	$(MAKE) -C $(PWD)/heapster -f Makefile clean

heapster: heapster-build
	$(MAKE) -C $(PWD)/heapster docker

influxdb-build:
	$(MAKE) -C $(PWD)/influxdb -e TARGET=influxd TARGETDIR=influxdb -f $(PWD)/buildbox.mk

influxdb-clean:
	$(MAKE) -C $(PWD)/influxdb clean

influxdb: influxdb-build
	$(MAKE) -C $(PWD)/influxdb docker

grafana:
	$(MAKE) -C $(PWD)/grafana

bootstrap: kubectl
	cp kubectl bootstrap/
	docker build --pull -t $(BOOTSTRAP_TAG) bootstrap

uninstall: kubectl
	cp kubectl uninstall/
	docker build --pull -t $(UNINSTALL_TAG) uninstall

all: heapster influxdb grafana bootstrap uninstall

clean: heapster-clean influxdb-clean

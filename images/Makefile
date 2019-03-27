.PHONY: all metrics-server grafana deploy kube-state-metrics

GRAFANA_VERSION := 6.0.2
METRICS_SERVER_VERSION := 0.3.1
KUBE_STATE_METRICS_VERSION := 1.5.0

PWD=$(shell pwd)
.DEFAULT_GOAL := all

METRICS_SERVER_TAG := monitoring-metrics-server:$(VERSION)
GRAFANA_TAG := monitoring-grafana:$(VERSION)
KUBE_STATE_METRICS_TAG := monitoring-kube-state-metrics:$(VERSION)
HOOK_TAG := monitoring-hook:$(VERSION)

REPO_URL := quay.io/gravitational

.PHONY: hook
hook:
	$(eval CHANGESET = $(shell echo $$VERSION | sed -e 's/[\.]//g'))
	if [ -z "$(CHANGESET)" ]; then \
	  echo "CHANGESET is not set"; exit 1; \
	fi;
	$(MAKE) -e BUILDIMAGE=$(HOOK_TAG) EXTRA_ARGS="--build-arg CHANGESET=monitoring-$(CHANGESET)" TARGETDIR=$@ make-docker-image

grafana:
	$(MAKE) -e BUILDIMAGE=$(GRAFANA_TAG) EXTRA_ARGS="--build-arg GRAFANA_VERSION=$(GRAFANA_VERSION)" TARGETDIR=$@ make-docker-image

metrics-server:
	$(MAKE) -e BUILDIMAGE=$(METRICS_SERVER_TAG) TARGETDIR=$@ make-docker-image

kube-state-metrics:
	$(MAKE) -e BUILDIMAGE=$(KUBE_STATE_METRICS_TAG) EXTRA_ARGS="--build-arg KUBE_STATE_METRICS_VERSION=$(KUBE_STATE_METRICS_VERSION)" TARGETDIR=$@ make-docker-image

all: metrics-server grafana hook kube-state-metrics

.PHONY: deploy
deploy:
	$(foreach ct,$(GRAFANA_TAG) $(METRICS_SERVER_TAG) $(KUBE_STATE_METRICS_TAG), \
		docker tag $(ct) $(REPO_URL)/$(ct) ; \
		docker push $(REPO_URL)/$(ct) ; )

make-docker-image:
	docker build --pull --rm $(EXTRA_ARGS) -t $(BUILDIMAGE) $(TARGETDIR)

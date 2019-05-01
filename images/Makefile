.PHONY: all grafana deploy mta hook

GRAFANA_VERSION := 6.0.2

PWD=$(shell pwd)
.DEFAULT_GOAL := all

GRAFANA_TAG := monitoring-grafana:$(VERSION)
HOOK_TAG := monitoring-hook:$(VERSION)
MTA_TAG := monitoring-mta:$(MTA_IMAGE_VERSION)

REPO_URL := quay.io/gravitational

.PHONY: hook
hook:
	$(eval CHANGESET = $(shell echo $$VERSION | sed -e 's/[\.]//g'))
	if [ -z "$(CHANGESET)" ]; then \
	  echo "CHANGESET is not set"; exit 1; \
	fi;
	$(MAKE) -e BUILDIMAGE=$(HOOK_TAG) EXTRA_ARGS="--build-arg CHANGESET=monitoring-$(CHANGESET)" TARGETDIR=$@ make-docker-image

.PHONY: grafana
grafana:
	$(MAKE) -e BUILDIMAGE=$(GRAFANA_TAG) EXTRA_ARGS="--build-arg GRAFANA_VERSION=$(GRAFANA_VERSION)" TARGETDIR=$@ make-docker-image

.PHONY: mta
mta:
	$(MAKE) -e BUILDIMAGE=$(MTA_TAG) TARGETDIR=$@ make-docker-image

.PHONY: all
all: grafana hook mta

.PHONY: deploy
deploy:
	$(foreach ct,$(GRAFANA_TAG),$(MTA_TAG) \
		docker tag $(ct) $(REPO_URL)/$(ct) ; \
		docker push $(REPO_URL)/$(ct) ; )

make-docker-image:
	docker build --pull --rm $(EXTRA_ARGS) -t $(BUILDIMAGE) $(TARGETDIR)

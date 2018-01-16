export VERSION ?= $(shell git describe --long --tags --always|awk -F'[.-]' '{print $$1 "." $$2 "." $$4}')
REPOSITORY := gravitational.io
NAME := monitoring-app
OPS_URL ?= https://opscenter.localhost.localdomain:33009
OUT ?= $(NAME).tar.gz
GRAVITY ?= gravity
export

EXTRA_GRAVITY_OPTIONS ?=

IMPORT_IMAGE_FLAGS := --set-image=monitoring-influxdb:$(VERSION) \
	--set-image=monitoring-heapster:$(VERSION) \
	--set-image=monitoring-grafana:$(VERSION) \
	--set-image=monitoring-kapacitor:$(VERSION) \
	--set-image=monitoring-telegraf:$(VERSION) \
	--set-image=monitoring-hook:$(VERSION) \
	--set-image=watcher:$(VERSION)

IMPORT_OPTIONS := --vendor \
	--insecure \
	--glob=**/*.yaml \
	--exclude=".git" \
	--exclude="images" \
	--exclude="Makefile" \
	--exclude=".gitignore" \
	--registry-url=apiserver:5000 \
	--ops-url=$(OPS_URL) \
	--repository=$(REPOSITORY) \
	--name=$(NAME) \
	--version=$(VERSION) \
	$(IMPORT_IMAGE_FLAGS)

.PHONY: package
package:
	$(MAKE) -C watcher
	$(MAKE) -C images all

.PHONY: deploy
deploy:
	$(MAKE) -C images deploy

.PHONY:
what-version:
	@echo $(VERSION)

.PHONY: hook
hook:
	$(MAKE) -C images hook

.PHONY: import
import: package
	-$(GRAVITY) app delete --ops-url=$(OPS_URL) $(REPOSITORY)/$(NAME):$(VERSION) \
		--force --insecure $(EXTRA_GRAVITY_OPTIONS)
	$(GRAVITY) app import $(IMPORT_OPTIONS) $(EXTRA_GRAVITY_OPTIONS) .

.PHONY: clean
clean:
	$(MAKE) -C watcher clean
	$(MAKE) -C images clean

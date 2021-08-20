export VERSION ?= $(shell git describe --tags)
REPOSITORY := gravitational.io
NAME := monitoring-app
OPS_URL ?= https://opscenter.localhost.localdomain:32009
OUT ?= $(NAME).tar.gz
GRAVITY ?= gravity
export

EXTRA_GRAVITY_OPTIONS ?=

MTA_IMAGE_VERSION := 1.0.0

IMPORT_IMAGE_FLAGS := --set-image=monitoring-grafana:$(VERSION) \
	--set-image=monitoring-hook:$(VERSION) \
	--set-image=monitoring-mta:$(MTA_IMAGE_VERSION) \
	--set-image=watcher:$(VERSION)

IMPORT_OPTIONS := --vendor \
	--insecure \
	--glob=**/*.yaml \
	--exclude=".git" \
	--exclude="images" \
	--exclude="Makefile" \
	--exclude=".gitignore" \
	--registry-url=leader.telekube.local:5000 \
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
get-version:
	@echo $(VERSION)

.PHONY: hook
hook:
	$(MAKE) -C images hook

.PHONY: import
import: package
	echo "image.tag: $(VERSION)" > resources/custom-values-watcher.yaml
	-$(GRAVITY) app delete \
		--ops-url=$(OPS_URL) \
		$(REPOSITORY)/$(NAME):$(VERSION) \
		--force --insecure $(EXTRA_GRAVITY_OPTIONS)
	$(GRAVITY) app import \
		$(IMPORT_OPTIONS) \
		$(EXTRA_GRAVITY_OPTIONS) \
		--include=resources --include=registry .

.PHONY: tarball
tarball: import
	$(GRAVITY) package export \
		--ops-url=$(OPS_URL) \
		--insecure \
		$(EXTRA_GRAVITY_OPTIONS) \
		$(REPOSITORY)/$(NAME):$(VERSION) $(NAME)-$(VERSION).tar.gz

.PHONY: clean
clean:
	$(MAKE) -C watcher clean

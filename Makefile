VER := 0.0.7
REPOSITORY := gravitational.io
NAME := monitoring-app
OPS_URL ?= https://opscenter.localhost.localdomain:33009
OUT ?= $(NAME).tar.gz
GRAVITY ?= gravity

.PHONY: package
package:
	$(MAKE) -C images all

.PHONY: deploy
deploy:
	$(MAKE) -C images deploy

.PHONY:
what-version:
	@echo $(VER)

.PHONY: import
import: package
	-$(GRAVITY) app delete --ops-url=$(OPS_URL) $(REPOSITORY)/$(NAME):$(VER) \
		--force --insecure
	$(GRAVITY) app import --vendor --glob=**/*.yaml --registry-url=apiserver:5000 \
		--ops-url=$(OPS_URL) --repository=$(REPOSITORY) --name=$(NAME) \
		--version=$(VER) --insecure .

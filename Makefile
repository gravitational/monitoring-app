export VERSION := 0.0.3

.PHONY: package
package:
	$(MAKE) -C images all
	$(MAKE) -C resources all

.PHONY: deploy
deploy:
	$(MAKE) -C images deploy


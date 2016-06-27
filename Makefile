.PHONY: package
package:
	$(MAKE) -C images all

.PHONY: deploy
deploy:
	$(MAKE) -C images deploy


ifndef VERSION
$(error 'VERSION not specified')
endif

.PHONY: all
all:
	for f in *.yaml.tpl ; do \
		sed -e 's/VERSION/$(VERSION)/' $$f > $$(basename $$f .tpl) ; \
	done

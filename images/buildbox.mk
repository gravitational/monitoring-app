# Generic GO builder makefile -- uses sub projects to build any GO-related packages
# Conventions used:
# TARGET names the directory where target is to be built as well as the resulting binary
# If the directory name does not equal that of the binary, override the directory with TARGETDIR
# With REPOBUILD defined, the docker image is store in docker repository - otherwise, the image
# is written to <target:version>.tar file.
#
ifndef TARGETDIR
ASSETS=$(PWD)/$(TARGET)
else
ASSETS=$(PWD)/$(TARGETDIR)
endif
override BUILDDIR=$(ASSETS)/build

.PHONY: all prepare buildbox

# Configuration by convention: use TARGET as a directory name
BINARIES=$(BUILDDIR)/$(TARGET)

BBOX := quay.io/gravitational/debian-venti:latest

all: prepare $(BINARIES)

$(BINARIES): buildbox $(ASSETS)/Makefile
	@echo "\n---> BuildBox for $(TARGET):\n"
	docker run --rm=true \
		--volume=$(ASSETS):/assets \
		--volume=$(BUILDDIR):/targetdir \
		--env="BUILD_UID=$(shell id -u)" \
		--env="TARGETDIR=/targetdir" \
		$(BBOX) \
		make -f /assets/Makefile

prepare:
	mkdir -p $(BUILDDIR)

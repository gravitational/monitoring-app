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

# Configuration by convention: use TARGET as a directory name
BINARIES=$(BUILDDIR)/$(TARGET)

# BBOX_IID stores the docker image hash of the buildbox
BBOX_IID=$(BUILDDIR)/.bbox.iid

.PHONY: all
all: $(BINARIES)

$(BINARIES): $(BBOX_IID) $(ASSETS)/Makefile | $(BUILDDIR)
	@echo "\n---> BuildBox for $(TARGET):\n"
	docker run --rm=true \
		--user $$(id -u):$$(id -g) \
		--volume=$(ASSETS):/assets \
		--volume=$(BUILDDIR):/targetdir \
		--env="TARGETDIR=/targetdir" \
		--env="VER=$(VER)" \
		$$(cat $(BBOX_IID)) \
		make -f /assets/Makefile

$(BBOX_IID): $(PWD)/Dockerfile $(PWD)/buildbox.mk | $(BUILDDIR)
	docker build $(PWD) --build-arg UID=$$(id -u) --build-arg GID=$$(id -g) --iidfile $(BBOX_IID)

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

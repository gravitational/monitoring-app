.PHONY: all clean docker

REPODIR:=$(GOPATH)/src/github.com/influxdata/influxdb
VER=v0.13.0
OUT=/targetdir/influxd

# FIXME: define build attributes and build with this:
# CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-X main.version $(version) -X main.branch $(branch) -X main.commit $(commit) -X main.buildTime $(date)" -o influxd ./cmd/influxd
version=v0.13.0
branch=master
commit=
date=

all: $(OUT)

$(OUT):
	@echo "\n---> Building $@\n"
	mkdir -p $(GOPATH)/src/github.com/influxdata
	cd $(GOPATH)/src/github.com/influxdata && git clone https://github.com/influxdata/influxdb -b $(VER)
	cd $(REPODIR) && python build.py --release --platform=linux --arch=static_amd64 --branch=$(version)
	mv $(REPODIR)/build/influxd /targetdir/

clean:
	sudo rm build/influxd


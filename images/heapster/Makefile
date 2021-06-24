.PHONY: all clean docker

VER ?= v1.4.1
REPODIR:=$(GOPATH)/src/k8s.io/heapster
PREFIX=gravity
GOFLAGS=-a -installsuffix cgo --ldflags '-w'
OUT=/targetdir/heapster

all: $(OUT)

$(OUT):
	@echo "\n---> Building $@\n"
	mkdir -p $(GOPATH)/src/k8s.io
	cd $(GOPATH)/src/k8s.io && git clone https://github.com/kubernetes/heapster -b $(VER)
	cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build $(GOFLAGS) ./...
	cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o heapster $(GOFLAGS) k8s.io/heapster/metrics && mv $(REPODIR)/heapster /targetdir/
	cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o eventer $(GOFLAGS) k8s.io/heapster/events && mv $(REPODIR)/eventer /targetdir/

clean:
	rm -f build/heapster
	rm -f build/eventer

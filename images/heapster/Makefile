.PHONY: all clean docker

REPODIR:=$(GOPATH)/src/k8s.io/heapster
VER=v1.0.2
PREFIX=gravity
GOFLAGS=-a -installsuffix cgo --ldflags '-w'
OUT=/targetdir/heapster

all: $(OUT)

$(OUT):
	@echo "\n---> Building $@\n"
	go get github.com/tools/godep
	mkdir -p $(GOPATH)/src/k8s.io
	cd $(GOPATH)/src/k8s.io && git clone https://github.com/kubernetes/heapster -b $(VER)
	cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build $(GOFLAGS) ./...
	cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o heapster $(GOFLAGS) k8s.io/heapster/metrics && mv $(REPODIR)/heapster /targetdir/
	cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o eventer $(GOFLAGS) k8s.io/heapster/events && mv $(REPODIR)/eventer /targetdir/

clean:
	sudo rm build/heapster
	sudo rm build/eventer

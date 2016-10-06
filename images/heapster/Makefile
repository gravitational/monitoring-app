.PHONY: all clean docker

REPODIR:=$(GOPATH)/src/k8s.io/heapster
VER=v1.0.2
PREFIX=gravity
GOFLAGS=-a -installsuffix cgo --ldflags '-w'
OUT=/targetdir/heapster
BUILD_UID?=1000

all: $(OUT)

$(OUT):
	useradd --uid $$BUILD_UID builder
	chown -R builder $(GOPATH)
	@echo "\n---> Building $@\n"
	sudo -E -u builder HOME=~builder PATH="$$PATH" bash -c "go get github.com/tools/godep"
	sudo -E -u builder HOME=~builder PATH="$$PATH" bash -c "mkdir -p $(GOPATH)/src/k8s.io"
	sudo -E -u builder HOME=~builder PATH="$$PATH" bash -c "cd $(GOPATH)/src/k8s.io && git clone https://github.com/kubernetes/heapster -b $(VER)"
	sudo -E -u builder HOME=~builder PATH="$$PATH:$(GOPATH)/bin" bash -c "cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build $(GOFLAGS) ./..."
	sudo -E -u builder HOME=~builder PATH="$$PATH:$(GOPATH)/bin" bash -c "cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o heapster $(GOFLAGS) k8s.io/heapster/metrics && mv $(REPODIR)/heapster /targetdir/"
	sudo -E -u builder HOME=~builder PATH="$$PATH:$(GOPATH)/bin" bash -c "cd $(REPODIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -o eventer $(GOFLAGS) k8s.io/heapster/events && mv $(REPODIR)/eventer /targetdir/"

clean:
	rm build/heapster
	rm build/eventer

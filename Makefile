export GOBIN := $(PWD)/.bin
export PATH := $(GOBIN):$(PATH)

PLATFORM := $(shell uname)
ARCH     := $(shell uname -m)

## Tools
GORELEASER_VERSION := 1.23.0
GORELEASER         := $(GOBIN)/goreleaser

$(GOBIN):
	mkdir -p $(GOBIN)

$(GORELEASER): $(GOBIN)
ifeq ($(PLATFORM),Linux)
	curl -L --fail -o $(GOBIN)/goreleaser.tar.gz https://github.com/goreleaser/goreleaser/releases/download/v$(GORELEASER_VERSION)/goreleaser_Linux_$(ARCH).tar.gz
else
	curl -L --fail -o $(GOBIN)/goreleaser.tar.gz https://github.com/goreleaser/goreleaser/releases/download/v$(GORELEASER_VERSION)/goreleaser_Darwin_all.tar.gz
endif
	cd $(GOBIN) && tar -xvf $(GOBIN)/goreleaser.tar.gz goreleaser
	rm -rf $(GOBIN)/goreleaser.tar.gz
	touch $(GOBIN)/goreleaser

snapshot: $(GORELEASER)
	$(GORELEASER) release --clean --snapshot

clean:
	rm -rf dist ui/*

bindata:
	GO111MODULE=off go get -u -v github.com/go-bindata/go-bindata/...

json-rpc2:
	GO111MODULE=off go get -u -v github.com/reddec/jsonrpc2/cmd/...

ui/src:
	cd ui && git reset --hard && git pull origin master && git lfs pull

ui/dist: ui/src
	cd ui && npm install . && npx quasar build

update_ui:
	git submodule init
	git submodule foreach --recursive git reset --hard
	git submodule update --init --recursive
	cd ui && git pull origin master && git lfs pull && npm install . && npx @quasar/cli build

regen: json-rpc2
	go generate api/handlers/*.go

embed_ui: bindata update_ui
	cd assets && $(shell go env GOPATH)/bin/go-bindata -pkg assets -prefix ../ui/dist/spa -fs ../ui/dist/spa/...

test:
	go test -v ./...
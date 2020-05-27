clean:
	rm -rf dist ui/*

bindata:
	GO111MODULE=off go get -u -v github.com/go-bindata/go-bindata/...

ui/src:
	cd ui && git reset --hard && git pull origin master && git lfs pull

ui/dist: ui/src
	cd ui && npm install . && npx quasar build

update_ui:
	cd ui && git reset --hard && git pull origin master && git lfs pull && npx quasar build

embed_ui: bindata update_ui
	cd assets && $(shell go env GOPATH)/bin/go-bindata -pkg assets -prefix ../ui/dist/spa -fs ../ui/dist/spa/...
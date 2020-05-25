clean:
	rm -rf dist ui/dist

bindata:
	GO111MODULE=off go get -u -v github.com/go-bindata/go-bindata/...

ui/dist:
	cd ui && npm install . && npx quasar build


embed_ui: bindata ui/dist
	cd assets && $(shell go env GOPATH)/bin/go-bindata -pkg assets -prefix ../ui/dist/spa -fs ../ui/dist/spa/...
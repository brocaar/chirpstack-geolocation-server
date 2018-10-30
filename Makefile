.PHONY: build api test test-race clean dist snapshot requirements docker-test
VERSION := $(shell git describe --always |sed -e "s/^v//")

build:
	@echo "Compiling source"
	mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/lora-geo-server cmd/lora-geo-server/main.go

api:
	@echo "Generating API code from .proto files"
	go generate internal/test/test.go

test:
	@echo "Running tests"
	go test -v ./...

test-race:
	@echo "Running test with race detector enabled"
	go test -v -race ./...

clean:
	@echo "Cleaning up workspace"
	rm -rf build
	rm -rf dist
	rm -rf docs/public

dist:
	goreleaser
	mkdir -p dist/upload/tar
	mkdir -p dist/upload/deb
	mv dist/*.tar.gz dist/upload/tar
	mv dist/*.deb dist/upload/deb

snapshot:
	@echo "Building snapshot binaries"
	goreleaser --snapshot

dev-requirements:
	@echo "Installing requirements"
	go get -u golang.org/x/tools/cmd/stringer
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/goreleaser/goreleaser
	go get -u github.com/goreleaser/nfpm

requirements:
	dep ensure -v

docker-test:
	@echo "Running tests inside docker container"
	docker-compose run --rm geoserver test

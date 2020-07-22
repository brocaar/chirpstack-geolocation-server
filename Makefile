.PHONY: build api test test-race clean dist snapshot requirements docker-test
VERSION := $(shell git describe --always |sed -e "s/^v//")

build:
	@echo "Compiling source"
	mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/chirpstack-geolocation-server cmd/chirpstack-geolocation-server/main.go

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

dist:
	goreleaser
	mkdir -p dist/upload/tar
	mkdir -p dist/upload/deb
	mkdir -p dist/upload/rpm
	mv dist/*.tar.gz dist/upload/tar
	mv dist/*.deb dist/upload/deb
	mv dist/*.rpm dist/upload/rpm

snapshot:
	@echo "Building snapshot binaries"
	goreleaser --snapshot

dev-requirements:
	@echo "Installing requirements"
	go install golang.org/x/tools/cmd/stringer
	go install github.com/golang/protobuf/protoc-gen-go
	go install github.com/goreleaser/goreleaser
	go install github.com/goreleaser/nfpm

docker-test:
	@echo "Running tests inside docker container"
	docker-compose run --rm geolocationserver test

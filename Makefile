GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")
VETPACKAGES ?= $(shell $(GO) list ./... | grep -v /examples/)
CLI_VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
VER ?= $(shell cat VERSION)

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

vet:
	$(GO) vet $(VETPACKAGES)

build:
	$(GO) build -o bin/yomo -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go

build-arm:
	GOARCH=arm64 GOOS=linux $(GO) build -o bin/yomo-${VER}-aarch64-linux -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go

build-linux:
	GOOS=linux $(GO) build -o bin/yomo-${VER}-x86_64-linux -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go

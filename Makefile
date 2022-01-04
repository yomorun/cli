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

build-release:
	GOARCH=arm64 GOOS=darwin $(GO) build -o bin/yomo-${VER}-arm64-Darwin -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go
	GOARCH=amd64 GOOS=darwin $(GO) build -o bin/yomo-${VER}-x86_64-Darwin -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go
	GOARCH=arm64 GOOS=linux $(GO) build -o bin/yomo-${VER}-arm64-Linux -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go
	GOARCH=amd64 GOOS=linux $(GO) build -o bin/yomo-${VER}-x86_64-Linux -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go

archive-release: build-release
	tar -czf bin/yomo-${VER}-arm64-Darwin.tar.gz bin/yomo-${VER}-arm64-Darwin
	tar -czf bin/yomo-${VER}-x86_64-Darwin.tar.gz bin/yomo-${VER}-x86_64-Darwin
	tar -czf bin/yomo-${VER}-arm64-Linux.tar.gz bin/yomo-${VER}-arm64-Linux
	tar -czf bin/yomo-${VER}-x86_64-Linux.tar.gz bin/yomo-${VER}-x86_64-Linux

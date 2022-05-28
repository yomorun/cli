GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")
VETPACKAGES ?= $(shell $(GO) list ./... | grep -v /examples/)
CLI_VERSION ?= $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
GO_LDFLAGS ?= -X $(shell $(GO) list -m)/cmd.Version=$(CLI_VERSION)
VER ?= $(shell cat VERSION)

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

vet:
	$(GO) vet $(VETPACKAGES)

build:
	$(GO) build -o bin/yomo -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go

archive-release:
	rm -rf bin/yomo
	GOARCH=arm64 GOOS=darwin $(GO) build -o bin/yomo -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go
	tar -C ./bin -czf bin/yomo-${VER}-arm64-Darwin.tar.gz yomo
	rm -rf bin/yomo
	GOARCH=amd64 GOOS=darwin $(GO) build -o bin/yomo -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go
	tar -C ./bin -czf bin/yomo-${VER}-x86_64-Darwin.tar.gz yomo
	rm -rf bin/yomo
	GOARCH=arm64 GOOS=linux $(GO) build -o bin/yomo -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go
	tar -C ./bin -czf bin/yomo-${VER}-arm64-Linux.tar.gz yomo
	rm -rf bin/yomo
	GOARCH=amd64 GOOS=linux $(GO) build -o bin/yomo -ldflags "-s -w ${GO_LDFLAGS}" ./yomo/main.go
	tar -C ./bin -czf bin/yomo-${VER}-x86_64-Linux.tar.gz yomo
	rm -rf bin/yomo
	cp ./bina.json ./bin/.

tar-release: build-release
	tar -C ./bin -czf bin/yomo-${VER}-arm64-Darwin.tar.gz yomo
	tar -C ./bin -czf bin/yomo-${VER}-x86_64-Darwin.tar.gz yomo
	tar -C ./bin -czf bin/yomo-${VER}-arm64-Linux.tar.gz yomo
	tar -C ./bin -czf bin/yomo-${VER}-x86_64-Linux.tar.gz yomo

build-w-sym:
	GOARCH=amd64 GOOS=linux $(GO) build -o bin/yomo -ldflags "${GO_LDFLAGS}" -gcflags=-l ./yomo/main.go


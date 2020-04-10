SHELL = /bin/bash
GO-VERSION = 1.14
GO-VER = go$(GO-VERSION)

SRC = $(shell find . -name "*.go" | grep -v "_test\." )
VERSION=0.1.0
ifeq ($(USE_GO_CONTAINERS),)
GO=go
GOIMPORTS=goimports
else
UID:=$(shell id -u)
DOCKER_OPTS=--rm -u $(UID) -v $(HOME):$(HOME) -e HOME -e USER=$(USER) -e USERNAME=$(USER) -w $(PWD)
GO=docker run $(DOCKER_OPTS) -e GOARCH -e GOOS golang:$(GO-VERSION) go
GOTOOLS=docker run $(DOCKER_OPTS) jare/go-tools
GOIMPORTS=$(GOTOOLS) goimports
HAS_GO_IMPORTS=true
endif

EXE_NAME=sqlfailover

.DEFAULT_TARGET: build

.PHONY: deps-go-binary
deps-go-binary:
	echo "Expect: $(GO-VER)" && \
		echo "Actual: $$($(GO) version)" && \
	 	$(GO) version | grep $(GO-VER) > /dev/null

HAS_GO_IMPORTS := $(shell command -v goimports;)

deps-goimports: deps-go-binary
ifndef HAS_GO_IMPORTS
	go get -u golang.org/x/tools/cmd/goimports
endif

.PHONY: build
build: ../../build/$(EXE_NAME)_$(VERSION)_linux_amd64.zip ../../build/$(EXE_NAME)_$(VERSION)_darwin_amd64.zip

../../build/$(EXE_NAME)_$(VERSION)_linux_amd64.zip: build/linux/$(EXE_NAME)
	zip -j -r $@ $<

../../build/$(EXE_NAME)_$(VERSION)_darwin_amd64.zip: build/darwin/$(EXE_NAME)
	zip -j -r $@ $<

build/linux/$(EXE_NAME): $(SRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GO) build -o $@

build/darwin/$(EXE_NAME): $(SRC)
	GOARCH=amd64 GOOS=darwin $(GO) build -o $@

.PHONY: lint
lint: deps-goimports
	git ls-files | grep '.go$$' | xargs $(GOIMPORTS) -l -w	

.PHONY: test
test: lint

.PHONY: clean
clean:
	rm -rf build

.PHONY: rebuild
rebuild: clean build

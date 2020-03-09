INSTALL_TO := /usr/local/bin

SRC = $(shell find . -name "*.go" | grep -v "_test\." )
VERSION=0.2.1
GO=go

.DEFAULT_TARGET: build

.PHONY: build
build: build/terraform-provider-sqlserver_$(VERSION)_linux_amd64.zip build/terraform-provider-sqlserver_$(VERSION)_darwin_amd64.zip

build/terraform-provider-sqlserver_$(VERSION)_linux_amd64.zip: build/linux/terraform-provider-sqlserver
	zip -j -r $@ $<

build/terraform-provider-sqlserver_$(VERSION)_darwin_amd64.zip: build/darwin/terraform-provider-sqlserver
	zip -j -r $@ $<

build/linux/terraform-provider-sqlserver: $(SRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GO) build -o $@

build/darwin/terraform-provider-sqlserver: $(SRC)
	GOARCH=amd64 GOOS=darwin $(GO) build -o $@

.PHONY: clean
clean:
	rm -rf build

.PHONY: rebuild
rebuild: clean build

.PHONY: tf_init
tf_init:
	terraform init

.PHONY: test
test: build tf_init
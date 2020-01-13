SHELL = /bin/bash
GO-VER = go1.13

OSFAMILY=$(shell uname)
ifeq ($(OSFAMILY),Darwin)
OSFAMILY=darwin
else
OSFAMILY=linux
endif

ifeq ($(USE_GO_CONTAINERS),)
GO=go
GOIMPORTS=goimports
else
UID:=$(shell id -u)
DOCKER_OPTS=--rm -u $(UID) -v $(HOME):$(HOME) -e HOME -e USER=$(USER) -e USERNAME=$(USER) -w $(PWD)
GO=docker run $(DOCKER_OPTS) -e GOARCH -e GOOS golang go
GOTOOLS=docker run $(DOCKER_OPTS) jare/go-tools
GOIMPORTS=$(GOTOOLS) goimports
HAS_GO_IMPORTS=true
endif

.PHONY: deps-go-binary
deps-go-binary:
	echo "Expect: $(GO-VER)" && \
		echo "Actual: $$($(GO) version)" && \
	 	go version | grep $(GO-VER) > /dev/null

HAS_GO_IMPORTS := $(shell command -v goimports;)

deps-goimports: deps-go-binary
ifndef HAS_GO_IMPORTS
	go get -u golang.org/x/tools/cmd/goimports
endif

.PHONY: test
test: deps-go-binary
	$(GO) test -v ./... -tags=service_broker

./build/gcp-service-broker.linux:
	GOARCH=amd64 GOOS=linux $(GO) build -o ./build/gcp-service-broker.linux

./build/gcp-service-broker.darwin:
	GOARCH=amd64 GOOS=darwin $(GO) build -o ./build/gcp-service-broker.darwin

.PHONY: build
build: deps-go-binary ./build/gcp-service-broker.linux ./build/gcp-service-broker.darwin
	./build/gcp-service-broker.$(OSFAMILY) generate tile > ./tile.yml
	./build/gcp-service-broker.$(OSFAMILY) generate manifest > ./manifest.yml
	./build/gcp-service-broker.$(OSFAMILY) generate customization > docs/customization.md
	./build/gcp-service-broker.$(OSFAMILY) generate use --destination-dir="docs/"
	./build/gcp-service-broker.$(OSFAMILY) generate use > docs/use.md

.PHONY: clean
clean: deps-go-binary
	$(GO) clean --modcache
	rm -rf ./build

.PHONY: lint
lint: deps-goimports
	git ls-files | grep '.go$$' | xargs $(GOIMPORTS) -l -w

.PHONY: root-service-account-json
root-service-account-json:
ifndef ROOT_SERVICE_ACCOUNT_JSON
	$(error variable ROOT_SERVICE_ACCOUNT_JSON not defined)
endif

.PHONY: security-user-name
security-user-name:
ifndef SECURITY_USER_NAME
	$(error variable SECURITY_USER_NAME not defined)
endif

.PHONY: security-user-password
security-user-password:
ifndef SECURITY_USER_PASSWORD
	$(error variable SECURITY_USER_PASSWORD not defined)
endif

.PHONY: db-host
db-host:
ifndef DB_HOST
	$(error variable DB_HOST not defined)
endif

.PHONY: db-username
db-username:
ifndef DB_USERNAME
	$(error variable DB_USERNAME not defined)
endif

.PHONY: db-password
db-password:
ifndef DB_PASSWORD
	$(error variable DB_PASSWORD not defined)
endif

.PHONY: check-env-vars
check-env-vars: root-service-account-json security-user-name security-user-password db-host db-username db-password

.PHONY: run
run: check-env-vars ./build/gcp-service-broker.$(OSFAMILY)
	./build/gcp-service-broker.$(OSFAMILY) serve
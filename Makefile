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

SRC = $(shell find . -name "*.go" | grep -v "_test\." )

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

.PHONY: test-units
test-units: deps-go-binary
	$(GO) test -v ./... -tags=service_broker

.PHONY: test-acceptance security-user-name security-user-password
test-acceptance: ./build/cloud-service-broker.$(OSFAMILY)
	./build/cloud-service-broker.$(OSFAMILY) client run-examples

./build/cloud-service-broker.linux: $(SRC)
	GOARCH=amd64 GOOS=linux $(GO) build -o ./build/cloud-service-broker.linux

./build/cloud-service-broker.darwin: $(SRC)
	GOARCH=amd64 GOOS=darwin $(GO) build -o ./build/cloud-service-broker.darwin

.PHONY: build
build: deps-go-binary ./build/cloud-service-broker.linux ./build/cloud-service-broker.darwin

.PHONY: package
package: ./build/cloud-service-broker.$(OSFAMILY) ./tile.yml ./manifest.yml docs/customization.md

./tile.yml:
	./build/cloud-service-broker.$(OSFAMILY) generate tile > ./tile.yml

./manifest.yml:
	./build/cloud-service-broker.$(OSFAMILY) generate manifest > ./manifest.yml

docs/customization.md:
	./build/cloud-service-broker.$(OSFAMILY) generate customization > docs/customization.md

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

.PHONY: run-broker
run-broker: check-env-vars ./build/cloud-service-broker.$(OSFAMILY) vmware-brokers/google-services-1.0.0.brokerpak
	./build/cloud-service-broker.$(OSFAMILY) serve

vmware-brokers/google-services-1.0.0.brokerpak: ./build/cloud-service-broker.$(OSFAMILY) ./vmware-brokers/*.yml
	cd ./vmware-brokers && ../build/cloud-service-broker.$(OSFAMILY) pak build

.PHONY: push-broker
push-broker: check-env-vars ./build/cloud-service-broker.$(OSFAMILY) vmware-brokers/google-services-1.0.0.brokerpak
	./scripts/push-broker.sh
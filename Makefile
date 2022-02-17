SHELL = /bin/bash
GO-VERSION = 1.17
GO-VER = go$(GO-VERSION)

OSFAMILY=$(shell uname)
ifeq ($(OSFAMILY),Darwin)
OSFAMILY=darwin
else
OSFAMILY=linux
endif

ifeq ($(USE_GO_CONTAINERS),)
GO=go
GOFMT=gofmt
else
UID:=$(shell id -u)
DOCKER_OPTS=--rm -u $(UID) -v $(HOME):$(HOME) -e HOME -e USER=$(USER) -e USERNAME=$(USER) -w $(PWD)
GO=docker run $(DOCKER_OPTS) -e GOARCH -e GOOS -e CGO_ENABLED golang:$(GO-VERSION) go
GOFMT=docker run $(DOCKER_OPTS) -e GOARCH -e GOOS -e CGO_ENABLED golang:$(GO-VERSION) gofmt
endif

SRC = $(shell find . -name "*.go" | grep -v "_test\." )

VERSION := $(or $(VERSION), dev)

LDFLAGS="-X github.com/cloudfoundry/cloud-service-broker/utils.Version=$(VERSION)"

PKG="github.com/cloudfoundry/cloud-service-broker"

.PHONY: deps-go-binary
deps-go-binary:
ifeq ($(SKIP_GO_VERSION_CHECK),)
	echo "Expect: $(GO-VER)" && \
		echo "Actual: $$($(GO) version)" && \
	 	$(GO) version | grep $(GO-VER) > /dev/null
endif

###### Help ###################################################################

.DEFAULT_GOAL = help

.PHONY: help

help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Test ###################################################################

.PHONY: test
test: download lint test-units test-integration ## run lint and unit tests

.PHONY: test-units
test-units: deps-go-binary ## run unit tests
	$(GO) test $(PKG)/brokerapi/... $(PKG)/cmd/... $(PKG)/db_service/... $(PKG)/internal/... $(PKG)/pkg/... $(PKG)/utils/... -tags=service_broker

.PHONY: test-integration
test-integration: deps-go-binary ## run integration tests
	$(GO) run github.com/onsi/ginkgo/v2/ginkgo -p integrationtest/...

###### Build ##################################################################

./build/cloud-service-broker.linux: $(SRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GO) build -o ./build/cloud-service-broker.linux -ldflags ${LDFLAGS}

./build/cloud-service-broker.darwin: $(SRC)
	GOARCH=amd64 GOOS=darwin $(GO) build -o ./build/cloud-service-broker.darwin -ldflags ${LDFLAGS}

.PHONY: build
build: deps-go-binary ./build/cloud-service-broker.linux ./build/cloud-service-broker.darwin ## build binary

.PHONY: generate
generate: ## generate test fakes
	${GO} generate ./...
	make format

.PHONY: download
download: ## download go module dependencies
	${GO} mod download

###### Clean ##################################################################

.PHONY: clean
clean: deps-go-binary ## clean up from previous builds
	-$(GO) clean --modcache
	-rm -rf ./build

###### Lint ###################################################################

.PHONY: lint
lint: checkformat checkimports vet staticcheck ## lint the source

checkformat: ## Checks that the code is formatted correctly
	@@if [ -n "$$(${GOFMT} -s -e -l -d .)" ]; then       \
		echo "gofmt check failed: run 'make format'"; \
		exit 1;                                       \
	fi

checkimports: ## Checks that imports are formatted correctly
	@@if [ -n "$$(${GO} run golang.org/x/tools/cmd/goimports -l -d .)" ]; then \
		echo "goimports check failed: run 'make format'";                      \
		exit 1;                                                                \
	fi

vet: ## Runs go vet
	${GO} vet ./...

staticcheck: ## Runs staticcheck
	${GO} run honnef.co/go/tools/cmd/staticcheck ./...

###### Format #################################################################

.PHONY: format
format: ## format the source
	${GOFMT} -s -e -l -w .
	${GO} run golang.org/x/tools/cmd/goimports -l -w .

###### Image ##################################################################

.PHONY: build-image
build-image: Dockerfile ## build a Docker image
	docker build --tag csb .

###### Env Var Checks #########################################################

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

###### End ####################################################################

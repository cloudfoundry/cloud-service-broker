SHELL = /bin/bash
GO-VERSION = 1.21.3
GO-VER = go$(GO-VERSION)

PAK_CACHE=$(PWD)/.pak-cache

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

ifeq ($(VERSION),)
	TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
	VERSION := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
	COMMIT := $(shell git rev-parse --short HEAD)
	ifneq ($(COMMIT), $(TAG_COMMIT))
		VERSION := $(VERSION)-$(COMMIT)
	endif
	ifneq ($(shell git status --porcelain),)
		VERSION := $(VERSION)-dirty
	endif
endif

LDFLAGS="-X github.com/cloudfoundry/cloud-service-broker/utils.Version=$(VERSION)"

PKG="github.com/cloudfoundry/cloud-service-broker"

.PHONY: deps-go-binary
deps-go-binary:
ifeq ($(SKIP_GO_VERSION_CHECK),)
	@@if [ "$$($(GO) version | awk '{print $$3}')" != "${GO-VER}" ]; then \
		echo "Go version does not match: expected: ${GO-VER}, got $$($(GO) version | awk '{print $$3}')"; \
		exit 1; \
	fi
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
	$(GO) test $(PKG)/brokerapi/... $(PKG)/cmd/... $(PKG)/dbservice/... $(PKG)/internal/... $(PKG)/pkg/... $(PKG)/utils/... -tags=service_broker

.PHONY: test-integration
test-integration: deps-go-binary .pak-cache ## run integration tests
	PAK_BUILD_CACHE_PATH=$(PAK_CACHE) $(GO) run github.com/onsi/ginkgo/v2/ginkgo -p integrationtest/...

.pak-cache:
	mkdir -p $(PAK_CACHE)

.PHONY: test-units-coverage
test-units-coverage: ## test-units coverage score
	$(GO) list ./... | grep -v fake > /tmp/csb-non-fake.txt
	paste -sd "," /tmp/csb-non-fake.txt > /tmp/csb-coverage-pkgs.txt
	$(GO) test -coverpkg=`cat /tmp/csb-coverage-pkgs.txt` -coverprofile=/tmp/csb-coverage.out `$(GO) list ./... | grep -v integrationtest`
	$(GO) tool cover -func /tmp/csb-coverage.out | grep total

###### Build ##################################################################

./build/cloud-service-broker.linux: $(SRC)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GO) build -o ./build/cloud-service-broker.linux -ldflags ${LDFLAGS}

./build/cloud-service-broker.darwin: $(SRC)
	GOARCH=amd64 GOOS=darwin $(GO) build -o ./build/cloud-service-broker.darwin -ldflags ${LDFLAGS}

.PHONY: build
build: deps-go-binary ./build/cloud-service-broker.linux ./build/cloud-service-broker.darwin ## build binary

.PHONY: install
install: deps-go-binary ## install as /usr/local/bin/csb
	$(GO) build -o csb -ldflags ${LDFLAGS}
	mv csb /usr/local/bin/csb

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
	-rm -rf /tmp/csb-non-fake.txt
	-rm -rf /tmp/csb-coverage.out
	-rm -rf /tmp/csb-coverage-pkgs.txt

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
	${GO} list ./... | grep -v 'fakes$$' | xargs ${GO} run honnef.co/go/tools/cmd/staticcheck

###### Format #################################################################

.PHONY: format
format: ## format the source
	${GOFMT} -s -e -l -w .
	${GO} run golang.org/x/tools/cmd/goimports -l -w .

###### Image ##################################################################

.PHONY: build-image
build-image: Dockerfile ## build a Docker image
	docker build --tag csb . --build-arg CSB_VERSION=$(VERSION)

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

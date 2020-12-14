
IAAS=azure
DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak #--network=host
CSB=cfplatformeng/csb

.PHONY: build
build: $(IAAS)-services-*.brokerpak 

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*.tf ./build/psqlcmd_*.zip ./build/sqlfailover_*.zip
	docker run $(DOCKER_OPTS) $(CSB) pak build

SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)

.PHONY: run
run: build  arm-subscription-id arm-tenant-id arm-client-id arm-client-secret
	docker run $(DOCKER_OPTS) \
	-p 8080:8080 \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e ARM_SUBSCRIPTION_ID \
    -e ARM_TENANT_ID \
	-e ARM_CLIENT_ID \
	-e ARM_CLIENT_SECRET \
	-e "DB_TYPE=sqlite3" \
	-e "DB_PATH=/tmp/csb-db" \
	-e GSB_PROVISION_DEFAULTS \
	$(CSB) serve

.PHONY: docs
docs: build brokerpak-user-docs.md

brokerpak-user-docs.md: *.yml
	docker run $(DOCKER_OPTS) \
	$(CSB) pak docs /brokerpak/$(shell ls *.brokerpak) > $@

.PHONY: run-examples
run-examples: build
	docker run $(DOCKER_OPTS) \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e "GSB_API_HOSTNAME=host.docker.internal" \
	-e USER \
	$(CSB) pak run-examples /brokerpak/$(shell ls *.brokerpak)

.PHONY: info
info: build
	docker run $(DOCKER_OPTS) \
	$(CSB) pak info /brokerpak/$(shell ls *.brokerpak)

.PHONY: validate
validate: build
	docker run $(DOCKER_OPTS) \
	$(CSB) pak validate /brokerpak/$(shell ls *.brokerpak)

# fetching bits for cf push broker
cloud-service-broker:
	wget $(shell curl -sL https://api.github.com/repos/pivotal/cloud-service-broker/releases/latest | jq -r '.assets[] | select(.name == "cloud-service-broker") | .browser_download_url')
	chmod +x ./cloud-service-broker

APP_NAME := $(or $(APP_NAME), cloud-service-broker-azure)
DB_TLS := $(or $(DB_TLS), skip-verify)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), '{"resource_group": "broker-cf-test"}')

.PHONY: push-broker
push-broker: cloud-service-broker build arm-subscription-id arm-tenant-id arm-client-id arm-client-secret
	MANIFEST=cf-manifest.yml APP_NAME=$(APP_NAME) DB_TLS=$(DB_TLS) GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)' ../scripts/push-broker.sh

.PHONY: clean
clean:
	- rm $(IAAS)-services-*.brokerpak
	- rm ./cloud-service-broker
	- rm ./brokerpak-user-docs.md

./build/psqlcmd_*.zip: ../tools/psqlcmd/*.go
	cd ../tools/psqlcmd; USE_GO_CONTAINERS=1 $(MAKE) build

./build/sqlfailover_*.zip: ../tools/sqlfailover/*.go
	cd ../tools/sqlfailover; USE_GO_CONTAINERS=1 $(MAKE) build

.PHONY: arm-subscription-id
arm-subscription-id:
ifndef ARM_SUBSCRIPTION_ID
	$(error variable ARM_SUBSCRIPTION_ID not defined)
endif

.PHONY: arm-tenant-id
arm-tenant-id:
ifndef ARM_TENANT_ID
	$(error variable ARM_TENANT_ID not defined)
endif

.PHONY: arm-client-id
arm-client-id:
ifndef ARM_CLIENT_ID
	$(error variable ARM_CLIENT_ID not defined)
endif

.PHONY: arm-client-secret
arm-client-secret:
ifndef ARM_CLIENT_SECRET
	$(error variable ARM_CLIENT_SECRET not defined)
endif
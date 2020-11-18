
IAAS=aws
DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak #--network=host
CSB=cfplatformeng/csb

.PHONY: build
build: $(IAAS)-services-*.brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*.tf terraform/*.tf 
	docker run $(DOCKER_OPTS) $(CSB) pak build

clean:
	- rm $(IAAS)-services-*.brokerpak

SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)

.PHONY: run
run: build aws_access_key_id aws_secret_access_key
	docker run $(DOCKER_OPTS) \
	-p 8080:8080 \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e AWS_ACCESS_KEY_ID \
    -e AWS_SECRET_ACCESS_KEY \
	-e "DB_TYPE=sqlite3" \
	-e "DB_PATH=/tmp/csb-db" \
	$(CSB) serve

.PHONY: docs
docs: build
	docker run $(DOCKER_OPTS) \
	$(CSB) pak docs /brokerpak/$(shell ls *.brokerpak)

.PHONY: run-examples
run-examples: build
	docker run $(DOCKER_OPTS) \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e "GSB_API_HOSTNAME=host.docker.internal" \
	-e USER \
	$(CSB) pak run-examples /brokerpak/$(shell ls *.brokerpak)

.PHONY: aws_access_key_id
aws_access_key_id:
ifndef AWS_ACCESS_KEY_ID
	$(error variable AWS_ACCESS_KEY_ID not defined)
endif

.PHONY: aws_secret_access_key
aws_secret_access_key:
ifndef AWS_SECRET_ACCESS_KEY
	$(error variable AWS_SECRET_ACCESS_KEY not defined)
endif

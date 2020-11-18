
IAAS=aws
DOCKER_OPTS=--rm -v $(PWD):/brokerpak -w /brokerpak
CSB=csb

.PHONY: build
build: $(IAAS)-services-*.brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*.tf terraform/*.tf 
	docker run $(DOCKER_OPTS) $(CSB) pak build

clean:
	- rm $(IAAS)-services-*.brokerpak

.PHONY: run aws_access_key_id aws_secret_access_key
run: build
	GSB_BROKERPAK_BUILTIN_PATH=/brokerpak \
	DB_TYPE=sqlite3 \
	DB_PATH=/tmp/csb-db \
	docker run $(DOCKER_OPTS) \
	-p 8080:8080 \
	-e SECURITY_USER_NAME \
	-e SECURITY_USER_PASSWORD \
	-e AWS_ACCESS_KEY_ID \
    -e AWS_SECRET_ACCESS_KEY \
	-e DB_TYPE \
	-e DB_PATH \
	-e GSB_BROKERPAK_BUILTIN_PATH \
	$(CSB) serve

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

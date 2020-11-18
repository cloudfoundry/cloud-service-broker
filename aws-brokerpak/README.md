# AWS Brokerpak

A brokerpak for the [Cloud Service Broker](https://github.com/pivotal/cloud-service-broker) that provides support for AWS services.

## Development Requirements

* [Docker](https://docs.docker.com/get-docker/) - tooling provided as docker images
* make - covers development lifecycle steps

A docker container for the cloud service broker binary is available at *cfplatformeng/csb*

## AWS account information

To provision services, the brokerpak currently requires AWS access key id and secret. The brokerpak expects them in environment variables:

* AWS_ACCESS_KEY_ID
* AWS_SECRET_ACCESS_KEY

## Development Tools

A Makefile supports the full development lifecycle for the brokerpak.

- `make` will build the brokerpak
- `make run` runs the brokerpak locally
- `make docs` will generate markdown documentation from brokerpak
- `make run-examples` will run example provision, bind, unbind, deprovision against broker started with `make run`


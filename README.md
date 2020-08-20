[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Cloud Service Broker

This is a service broker built to be used with [Cloud Foundry](https://docs.cloudfoundry.org/services/overview.html) and Kubernetes. It adheres to the [Open Service Broker API v2.13 (https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md).

Cloud Service Broker is an extension of the [GCP Service Broker](https://github.com/GoogleCloudPlatform/gcp-service-broker). As long as your target public cloud has a [Terraform provider] (https://www.terraform.io/docs/providers/index.html), services can be provisioned via a common interface using standard `cf` CLI commands.

Some of the benefits over traditional, IaaS-provided, service brokers include: 
- **Easily extensible and maintainable** Less talking to far flung teams, more getting work done. 
- **One common broker for all brokered services.** Cloud Service Broker decouples the service broker functionality from the catalog of services that it exposes.
- **Credhub integration out-of-the-box** CredHub encrypts and manages all the secrets associated with your usage of cloud services.
- **Community** When you expose a service via a [Brokerpak](https://github.com/pivotal/cloud-service-broker/blob/master/docs/brokerpak-intro.md), you can make it available to everyone who uses CSB.
- **Easy to migrate existing services using [TF Import](https://www.terraform.io/docs/import/index.html)** *Coming Soon...*

## Public Roadmap
For a list of currently supported services (and broker features), see our up-to-date roadmap on Trello here: https://trello.com/b/6wgNQZLB/cloud-service-broker-public-product-roadmap

## Installation

This service broker can be installed as a CF application. See the instructions for:

- [AWS Installation](./docs/aws-installation.md)
- [Azure Installation](./docs/azure-installation.md) 
- [GCP Installation](./docs/gcp-installation.md) 

## Usage

**For operators**: see [docs/configuration.md](./docs/configuration.md) for details about configuring the service broker.

**For developers**: see [docs/use.md](./docs/use.md) for service options and details.

You can get documentation specific to your install from the `/docs` endpoint of your deployment.

## Commands

The service broker can be run as both a server (the service broker) and as a general purpose command line utility.
It supports the following sub-commands:

 * `client` - A CLI client for the service broker.
 * `config` - Show and merge configuration options together.
 * `generate` - Generate documentation and tiles.
 * `help` - Help about any command.
 * `serve` - Start the service broker.

## Development

`make` is used to orchestrate most development tasks. 
`golang` 1.14 is required to build the broker. If you don't have `golang` installed, it is possible to use a `docker` image to build and unit test the broker. If the environment variable `USE_GO_CONTAINERS` exists, `make` will use `docker` versions of the tools so you don't need to have them installed locally. 

There are make targets for most common dev tasks. 

| command | action |
|---------|--------|
`make build` | builds broker into `./build`
`make test-units` | runs unit tests
`make run-broker-gcp` | builds broker and broker pak and starts broker for gcp
`make run-broker-azure` | builds broker and broker pak and starts broker for azure
`make run-broker-aws` | builds broker and broker pak and starts broker for aws
`make test-acceptance` | runs broker [client run-examples](./TESTING.md) tests
`make clean` | removes binaries and built broker paks
`make push-broker-gcp` | will push and register the broker in PAS for GCP
`make push-broker-azure` | will push and register the broker in PAS for Azure

## Bug Reports, Feature Requests, Documentation Requests & Support

[File a GitHub issue](https://github.com/pivotal/cloud-service-broker/issues) for bug reports and documentation or feature requests. Please use the provided templates.  

## Contributing

See [the contributing file](https://github.com/pivotal/cloud-service-broker/blob/master/CONTRIBUTING.md) for more information. 

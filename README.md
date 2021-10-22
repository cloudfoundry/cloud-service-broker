[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![test](https://github.com/cloudfoundry-incubator/cloud-service-broker/workflows/test/badge.svg?branch=master)](https://github.com/cloudfoundry-incubator/cloud-service-broker/actions?query=workflow%3Atest+branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudfoundry-incubator/cloud-service-broker)](https://goreportcard.com/report/github.com/cloudfoundry-incubator/cloud-service-broker)

# Cloud Service Broker

This is a service broker built to be used with [Cloud Foundry](https://docs.cloudfoundry.org/services/overview.html) and Kubernetes. It adheres to the [Open Service Broker API v2.13](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md).

Cloud Service Broker is an extension of the [GCP Service Broker](https://github.com/GoogleCloudPlatform/gcp-service-broker) and uses [Brokerpaks](https://github.com/cloudfoundry-incubator/cloud-service-broker/blob/master/docs/brokerpak-intro.md) to expose services. As long as your target public cloud has a [Terraform provider](https://www.terraform.io/docs/providers/index.html), services can be provisioned via a common interface using standard `cf` CLI commands.

Some of the benefits over traditional, IaaS-provided, service brokers include: 
- **Easily extensible and maintainable** Less talking to far flung teams, more getting work done. 
- **One common broker for all brokered services.** Cloud Service Broker decouples the service broker functionality from the catalog of services that it exposes.
- **Credhub integration out-of-the-box** CredHub encrypts and manages all the secrets associated with your usage of cloud services.
- **Community** When you expose a service via a [Brokerpak](https://github.com/cloudfoundry-incubator/cloud-service-broker/blob/master/docs/brokerpak-intro.md), you can make it available to everyone who uses CSB.
- **Easy to migrate existing services using [TF Import](https://www.terraform.io/docs/import/index.html)** We call this "Migration-less" Migration, more on this **coming soon!** 

## Architecture
![Architecture Diagram](https://lh6.googleusercontent.com/GoNJx-4dQ51pEY6mCLkus1peKhZJbDMj4JHpdu83stfQrbcsjd45ypBPzpspfWAPPYrc63BREaawwRHS4Ht4U7m2yWAHItwaIgfuwUtn_KxfF96s6Jby7BRIliZ6BZz1HL-KhaI)



## Public Roadmap
The previously mentioned Trello roadmap has been archived in favor of a new Roadmap, which will live in Github, is TBD. 

For updates on this roadmap, please reach out on the #cloudservicebroker channel in the [Cloud Foundry Slack](https://slack.cloudfoundry.org/)! 

## Installation

This service broker can be installed as a CF application. See the instructions for:

* [General configuration](./docs/configuration.md)
* [AWS](https://github.com/cloudfoundry-incubator/csb-brokerpak-aws/blob/main/docs/aws-installation.md)
* [Azure](https://github.com/cloudfoundry-incubator/csb-brokerpak-azure/blob/main/docs/azure-installation.md)
  * [Azure configuration examples](https://github.com/cloudfoundry-incubator/csb-brokerpak-azure/blob/main/docs/azure-example-configs.md)
* [GCP](https://github.com/cloudfoundry-incubator/csb-brokerpak-gcp/blob/main/docs/gcp-installation.md)


## CSB-Provided Brokerpaks 

To examine, submit issues or pull requests to the Brokerpaks which have been created for the major public clouds (AWS, Azure, GCP) see the repos below:

* [AWS Brokerpak](https://github.com/cloudfoundry-incubator/csb-brokerpak-aws)
* [Azure Brokerpak](https://github.com/cloudfoundry-incubator/csb-brokerpak-azure)
* [GCP Brokerpak](https://github.com/cloudfoundry-incubator/csb-brokerpak-gcp)

## Usage

**For operators**: see [docs/configuration.md](./docs/configuration.md) for details about configuring the service broker.

**For developers**: see [docs/](./docs) ReadMe for service options and details.

You can get documentation specific to your install from the `/docs` endpoint of your deployment.


## Commands

The service broker can be run as both a server (the service broker) and as a general purpose command line utility.
It supports the following sub-commands:

 * `client` - A CLI client for the service broker.
 * `config` - Show and merge configuration options together.
 * `help` - Help about any command.
 * `serve` - Start the service broker.

## Development

`make` is used to orchestrate most development tasks. 
`golang` 1.16 is required to build the broker. If you don't have `golang` installed, it is possible to use a `docker` image to build and unit test the broker. If the environment variable `USE_GO_CONTAINERS` exists, `make` will use `docker` versions of the tools so you don't need to have them installed locally. 

There are make targets for most common dev tasks. 

| command | action |
|---------|--------|
`make build` | builds broker into `./build`
`make test-units` | runs unit tests
`make clean` | removes binaries and built broker paks

## Bug Reports, Feature Requests, Documentation Requests & Support

[File a GitHub issue](https://github.com/cloudfoundry-incubator/cloud-service-broker/issues) for bug reports and documentation or feature requests. Please use the provided templates.  

## Contributing
We are always looking for folks to contribute Brokerpaks! 

See [Brokerpak Dissection](https://github.com/cloudfoundry-incubator/cloud-service-broker/blob/master/docs/brokerpak-dissection.md) for more information on how to build one yourself.

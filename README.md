[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[![Go Report Card](https://goreportcard.com/badge/github.com/pivotal/cloud-service-broker)](https://goreportcard.com/report/github.com/pivotal/cloud-service-broker)

# Cloud Service Broker

This is a service broker built to be used with [Cloud Foundry](https://docs.cloudfoundry.org/services/overview.html) and Kubernetes. It adheres to the [Open Service Broker API v2.13](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md).

Cloud Service Broker is an extension of the [GCP Service Broker](https://github.com/GoogleCloudPlatform/gcp-service-broker) and uses [Brokerpaks](https://github.com/pivotal/cloud-service-broker/blob/master/docs/brokerpak-intro.md) to expose services. As long as your target public cloud has a [Terraform provider](https://www.terraform.io/docs/providers/index.html), services can be provisioned via a common interface using standard `cf` CLI commands.

Some of the benefits over traditional, IaaS-provided, service brokers include: 
- **Easily extensible and maintainable** Less talking to far flung teams, more getting work done. 
- **One common broker for all brokered services.** Cloud Service Broker decouples the service broker functionality from the catalog of services that it exposes.
- **Credhub integration out-of-the-box** CredHub encrypts and manages all the secrets associated with your usage of cloud services.
- **Community** When you expose a service via a [Brokerpak](https://github.com/pivotal/cloud-service-broker/blob/master/docs/brokerpak-intro.md), you can make it available to everyone who uses CSB.
- **Easy to migrate existing services using [TF Import](https://www.terraform.io/docs/import/index.html)** We call this "Migration-less" Migration, more on this **coming soon!** 

## Architecture
![Architecture Diagram](https://lh6.googleusercontent.com/GoNJx-4dQ51pEY6mCLkus1peKhZJbDMj4JHpdu83stfQrbcsjd45ypBPzpspfWAPPYrc63BREaawwRHS4Ht4U7m2yWAHItwaIgfuwUtn_KxfF96s6Jby7BRIliZ6BZz1HL-KhaI)



## Public Roadmap
For a list of currently "Core Broker" (IaaS agnostic) features, see our up-to-date roadmap on Trello here: https://trello.com/b/m873oYyJ/csb-core-broker-public-roadmap

Azure Roadmap: https://trello.com/b/IJKM3bcG/csb-azure-public-roadmap

AWS Roadmap: https://trello.com/b/eBe25hzx/csb-aws-public-roadmap

GCP Roadmap: https://trello.com/b/MNL1QzrQ/csb-gcp-public-roadmap

## Installation

This service broker can be installed as a CF application. See the instructions for:

- [AWS Installation](./docs/aws-installation.md)
- [Azure Installation](./docs/azure-installation.md) 
- [GCP Installation](./docs/gcp-installation.md) 


The service broker currently provides access to the below services. **Where it exists, we have linked to the documentation for each service.** 

*Note: You can also use CSB with your own custom Brokerpaks. See our [Brokerpak Developer Guide](./docs/brokerpak-development.md) for more information*

| [GCP](./docs/gcp-installation.md) | [Azure]((./docs/azure-installation.md)) | [AWS]((./docs/aws-installation.md)) |	
|-----|-------| ----|	
|[GCP Cloud SQL (MySQL)](./docs/mysql-plans-and-config.md)|[Azure SQL](./docs/mssql-plans-and-config.md)|[Amazon RDS for MySQL](./docs/mysql-plans-and-config.md)|	
|[GCP Cloud SQL (PostgreSQL)](./docs/postgresql-plans-and-config.md)|[Azure SQL DB](./docs/mssql-db-plans-and-config.md)|[Amazon ElastiCache for Redis](./docs/redis-plans-and-config.md)|	
|GCP Memorystore for Redis|[Azure SQL Failover Group](./docs/mssql-fog-plans-and-config.md)|[Amazon RDS for PostgreSQL](./docs/postgresql-plans-and-config.md)|	
|GCP BigQuery|[Azure SQL Failover Group Failover Runner](./docs/azure-fog-failover-runner.md)|[Amazon S3 Bucket](./docs/s3-bucket-plans-and-config.md)|	
|GCP Spanner|[Azure SQL DB Failover Groups](./docs/mssql-db-fog-config.md)||	
|GCP Cloud Storage|[Azure SQL Server](./docs/mssql-server-plans-and-config.md)||	
|GCP Dataproc|[MySQL](docs/mysql-plans-and-config.md)||	
|[Google Stack Driver Trace](./docs/stack-driver-trace.md)|[Azure Database for PostgreSQL](./docs/postgresql-plans-and-config.md)||	
||[Azure Storage Account](./docs/azure-storage-account-plans-and-config.md)||
||[Azure Redis](./docs/redis-plans-and-config.md)||
||[Azure Eventhubs](./docs/azure-event-hubs.md)||
||[MongoDB for CosmosDB](./docs/mongo-plans-and-config.md)||
||Azure CosmosDB||

## Usage

**For operators**: see [docs/configuration.md](./docs/configuration.md) for details about configuring the service broker.

**For developers**: see [docs/](./docs) ReadMe for service options and details.

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
We are always looking for folks to contribute Brokerpaks! 

See [Brokerpak Dissection](https://github.com/pivotal/cloud-service-broker/blob/master/docs/brokerpak-dissection.md) for more information on how to build one yourself.

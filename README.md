[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Open Service Broker for Cloud Platform (GCP, Azure, AWS)

This is a service broker built to be used with [Cloud Foundry](https://docs.cloudfoundry.org/services/overview.html) and Kubernetes.
It adheres to the [Open Service Broker API v2.13](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md).

Service brokers provide a consistent way to create resources and accounts that can access those resources across a variety of different services.

The service broker uses [Terraform](https://www.terraform.io/) to provision services.

The service broker provides support for:

| GCP | Azure | AWS |
|-----|-------| ----|
|[GCP Cloud SQL (MySQL)](https://cloud.google.com/sql/)|[Azure Database for MySQL](https://azure.microsoft.com/en-us/services/mysql/?&ef_id=EAIaIQobChMImtPm8_DK5wIVgf5kCh1lEAqOEAAYASABEgIwjfD_BwE:G:s&OCID=AID2000128_SEM_VfuRONbO&MarinID=VfuRONbO_307794721357_azure%20mysql_e_c_Qml9BhwJ_46775457259_kwd-310296951725&lnkd=Google_Azure_Brand&gclid=EAIaIQobChMImtPm8_DK5wIVgf5kCh1lEAqOEAAYASABEgIwjfD_BwE)|[Amazon RDS for MySQL](https://aws.amazon.com/rds/mysql/)|
|[GCP Memorystore for Redis](https://cloud.google.com/memorystore/docs/redis/)|[Azure Cache for Redis](https://azure.microsoft.com/en-us/services/cache/?&ef_id=EAIaIQobChMIzc-t2vHK5wIVsh-tBh3Z8wteEAAYASAAEgJ0cvD_BwE:G:s&OCID=AID2000128_SEM_SeUFPHct&MarinID=SeUFPHct_287547165334_azure%20redis_e_c__46775456859_kwd-310342681850&lnkd=Google_Azure_Brand&gclid=EAIaIQobChMIzc-t2vHK5wIVsh-tBh3Z8wteEAAYASAAEgJ0cvD_BwE)|
||[MongoDB](https://docs.microsoft.com/en-us/azure/cosmos-db/mongodb-introduction) for [CosmosDB](https://azure.microsoft.com/en-us/services/cosmos-db/)||

## Installation

This application can be installed as either a PCF Ops Man Tile _or_ deployed as a PCF application.
See the [installation instructions](https://github.com/pivotal/cloud-service-broker/blob/master/docs/installation.md) for a more detailed walkthrough.

## Upgrading

If you're upgrading, check the [upgrade guide](https://github.com/pivotal/cloud-service-broker/blob/master/docs/upgrading.md).

## Usage

For operators: see [docs/customization.md](https://github.com/pivotal/cloud-service-broker/blob/master/docs/customization.md) for details about configuring the service broker.

For developers: see [docs/use.md](https://github.com/pivotal/cloud-service-broker/blob/master/docs/use.md) for information about creating and binding specific GCP services with the broker.


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
`golang` 1.13 is required to build the broker. If you don't have `golang` installed, it is possible to use a `docker` image to build and unit test the broker. If the environment variable `USE_GO_CONTAINERS` exists, `make` will use `docker` versions of the tools so you don't need to have them installed locally. 

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

## Support

[File a GitHub issue](https://github.com/pivotal/cloud-service-broker/issues) for functional issues or feature requests.

## Contributing

See [the contributing file](https://github.com/pivotal/cloud-service-broker/blob/master/CONTRIBUTING.md) for more information.

This is not an officially supported VMware product.

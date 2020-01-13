[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Open Service Broker for Cloud Platform (GCP, Azure, AWS)

This is a service broker built to be used with [Cloud Foundry](https://docs.cloudfoundry.org/services/overview.html) and Kubernetes.
It adheres to the [Open Service Broker API v2.13](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md).

Service brokers provide a consistent way to create resources and accounts that can access those resources across a variety of different services.

The service broker uses [Terraform](https://www.terraform.io/) to provision services.

The service broker provides support for:

* [GCP BigQuery](https://cloud.google.com/bigquery/)
* [GCP Bigtable](https://cloud.google.com/bigtable/)
* [GCP Cloud SQL](https://cloud.google.com/sql/)
* [GCP Cloud Storage](https://cloud.google.com/storage/)
* [GCP Dataflow](https://cloud.google.com/dataflow/) (preview)
* [GCP Dataproc](https://cloud.google.com/dataproc/docs/overview) (preview)
* [GCP Datastore](https://cloud.google.com/datastore/)
* [GCP Dialogflow](https://cloud.google.com/dialogflow-enterprise/) (preview)
* [GCP Firestore](https://cloud.google.com/firestore/) (preview)
* [GCP Memorystore for Redis](https://cloud.google.com/memorystore/docs/redis/) (preview)
* [GCP ML APIs](https://cloud.google.com/ml/)
* [GCP PubSub](https://cloud.google.com/pubsub/)
* [GCP Spanner](https://cloud.google.com/spanner/)
* [GCP Stackdriver Debugger](https://cloud.google.com/debugger/)
* [GCP Stackdriver Monitoring](https://cloud.google.com/monitoring/) (preview)
* [GCP Stackdriver Trace](https://cloud.google.com/trace/)
* [GCP Stackdriver Profiler](https://cloud.google.com/profiler/)

## Installation

This application can be installed as either a PCF Ops Man Tile _or_ deployed as a PCF application.
See the [installation instructions](https://github.com/pivotal/cloud-service-broker/blob/master/docs/installation.md) for a more detailed walkthrough.

## Upgrading

If you're upgrading, check the [upgrade guide](https://github.com/pivotal/cloud-service-broker/blob/master/docs/upgrading.md).

## Usage

For operators: see [docs/customization.md](https://github.com/pivotal/cloud-service-broker/blob/master/docs/customization.md) for details about configuring the service broker.

For developers: see [docs/use.md](https://github.com/pivotal/cloud-service-broker/blob/master/docs/use.md) for information about creating and binding specific GCP services with the broker.
Complete Spring Boot sample applications which use services can be found in the [service-broker-samples repository](https://github.com/GoogleCloudPlatform/service-broker-samples).

You can get documentation specific to your install from the `/docs` endpoint of your deployment.

## Commands

The service broker can be run as both a server (the service broker) and as a general purpose command line utility.
It supports the following sub-commands:

 * `client` - A CLI client for the service broker.
 * `config` - Show and merge configuration options together.
 * `generate` - Generate documentation and tiles.
 * `help` - Help about any command.
 * `serve` - Start the service broker.

## Testing

Pull requests are unit-tested with Travis. You can run the same tests Travis does using `go test ./...`.

Integration tests are run on a private [Concourse](https://concourse-ci.org/) pipeline for all changes to the `master` branch.
You can set up your own pipeline using the sources in the `ci` directory if you like.

## Support

[File a GitHub issue](https://github.com/pivotal/cloud-service-broker/issues) for functional issues or feature requests.


## Contributing

See [the contributing file](https://github.com/pivotal/cloud-service-broker/blob/master/CONTRIBUTING.md) for more information.

This is not an officially supported Google product.

# Brokerpak Development

So you want to build a brokerpak from scratch? As long as you can write some terraform to control whatever resource you're interested in controlling, and the lifecycle can be mapped into the [OSBAPI API](https://www.openservicebrokerapi.org/), you should be able to write a brokerpak to fulfill your needs.

Hopefully this document helps you along the way.

## Requirements

### Talent
* familiarity with [yaml](https://yaml.org/spec/1.2/spec.html) - yaml glues all the pieces together.
* knowledge of [terraform](https://www.terraform.io/docs/index.html) - terraform scripts do the heavy lifting of managing resource lifecycle.

### Tools
* [golang 1.14](https://golang.org/dl/) - the core broker is written in golang. You probably won't need to delve into the source code, but currently you'll need to build the broker locally to use it to generate and test your brokerpak.
* make - much of the development lifecycle is automated in a makefile.
* [docker](https://docs.docker.com/get-docker/) - some of the toolchain can currently be run in docker (the broker can be built with golang 1.14 docker image instead of having to install golang if you'd like) and eventually release versions of the core broker will be made available as a docker image so the requirement to build the broker locally will go away.

## References
* [Brokerpak Introduction](./brokerpak-intro.md)
* [Brokerpak Specification](./brokerpak-specification.md)

## Prerequisites

> The broker makefile currently supports OSX and Linux environments.

### Fetch the repo
```bash
git clone https://github.com/pivotal/cloud-service-broker
```

### Build the broker
#### Option 1 - with docker
```bash
cd cloud-service-broker
USE_GO_CONTAINERS=1 make build
```

#### Option 2 - with golang 1.14 installed locally
```bash
cd cloud-service-broker
make build
```

After building the broker there should be two executables in the `./build` directory:

```bash
$ ls ./build
cloud-service-broker.darwin cloud-service-broker.linux
```

*cloud-service-broker.darwin* is compiled for OSX, *cloud-service-broker.linux* is compiled for Linux. 

> The rest of this document assumes development on OSX and will always reference *./build/cloud-service-broker.darwin*, if you're developing on Lunix, you should use *./build/cloud-service-broker.linux*

## A New Brokerpak

Now you should be ready to create a new brokerpak.

### manifest.yml

The root of a brokerpak is a *manifest.yml* file. It is a description of the tools and service yml files that will get built into the brokerpak.

#### Create a Bokerpak Directory

```bash
mkdir my-brokerpak
cd my-brokerpak
```

#### The Manifest

Create a file named *manifest.yml* with the following contents:

```yaml
packversion: 1
name: azure-services
version: 0.1.0
metadata:
  author: VMware
platforms:
- os: linux
  arch: amd64
- os: darwin
  arch: amd64
terraform_binaries:
- name: terraform
  version: 0.12.26
  source: https://github.com/hashicorp/terraform/archive/v0.12.26.zip 
env_config_mapping:
  ARM_SUBSCRIPTION_ID: azure.subscription_id
  ARM_TENANT_ID: azure.tenant_id
  ARM_CLIENT_ID: azure.client_id
  ARM_CLIENT_SECRET: azure.client_secret
service_definitions:
  
```
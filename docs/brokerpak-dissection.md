# Dissecting a Brokerpak

So you want to build a brokerpak from scratch? As long as you can write some terraform to control whatever resource you're interested in controlling, and the lifecycle can be mapped into the [OSBAPI API](https://www.openservicebrokerapi.org/), you should be able to write a brokerpak to fulfill your needs.

Hopefully this document helps familiarize you enough with layout and details of the brokerpak specification to help you along the way.

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

## The Azure Brokerpak

Have a look at the Azure brokerpak, starting with *manifest.yml*

```bash
cd azure-brokerpak
cat manifest.yml
```

The output should resemble:

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
- name: terraform-provider-azurerm
  version: 2.20.0
  source: https://github.com/terraform-providers/terraform-provider-azurerm/archive/v2.20.0.zip
- name: terraform-provider-random
  version: 2.2.1
  source: https://releases.hashicorp.com/terraform-provider-random/2.2.1/terraform-provider-random_2.2.1_linux_amd64.zip
- name: terraform-provider-mysql
  version: 1.9.0
  source: https://releases.hashicorp.com/terraform-provider-mysql/1.9.0/terraform-provider-mysql_1.9.0_linux_amd64.zip 
- name: terraform-provider-null
  version: 2.1.2
  source: https://releases.hashicorp.com/terraform-provider-null/2.1.2/terraform-provider-null_2.1.2_linux_amd64.zip
- name: psqlcmd
  version: 0.1.0
  source: https://packages.microsoft.com/config/rhel/7/packages-microsoft-prod.rpm
  url_template: ../build/${name}_${version}_${os}_${arch}.zip
- name: sqlfailover
  version: 0.1.0
  source: https://packages.microsoft.com/config/rhel/7/packages-microsoft-prod.rpm
  url_template: ../build/${name}_${version}_${os}_${arch}.zip  
- name: terraform-provider-postgresql
  version: 1.5.0
  source: https://github.com/terraform-providers/terraform-provider-postgresql/archive/v1.5.0.zip
env_config_mapping:
  ARM_SUBSCRIPTION_ID: azure.subscription_id
  ARM_TENANT_ID: azure.tenant_id
  ARM_CLIENT_ID: azure.client_id
  ARM_CLIENT_SECRET: azure.client_secret
service_definitions:
- azure-redis.yml
- azure-mysql.yml
- azure-mssql.yml
- azure-mssql-failover.yml
- azure-mongodb.yml
- azure-eventhubs.yml
- azure-mssql-db.yml
- azure-mssql-server.yml
- azure-mssql-db-failover.yml
- azure-mssql-fog-run-failover.yml
- azure-resource-group.yml
- azure-postgres.yml
- azure-storage-account.yml
- azure-cosmosdb-sql.yml
- azure-mssql-db-subsume.yml
- azure-mssql-db-masb-subsume.yml
```

Let's break it down.

### Header

```yaml
packversion: 1
name: azure-services
version: 0.1.0
metadata:
  author: VMware
```

Metadata about the brokerpak.

| Field | Value |
|-------|-------|
| packversion | should always be 1 |
| name | name of brokerpak |
| version | version of brokerpak |
| metadata | a map of metadata to add to broker |

Besides *packversion* (which should always be 1,) these values are left to the brokerpak author to describe the brokerpak.

### Platforms

```yaml
platforms:
- os: linux
  arch: amd64
- os: darwin
  arch: amd64
```

Describes which platforms the brokerpak should support. Typically *os: linux* is the minimum required as when `cf push`ing the broker into CloudFoundry. For local development on OSX, adding *os: darwin* allows the broker to run locally.

### Binaries

```yaml
terraform_binaries:
- name: terraform
  version: 0.12.26
  source: https://github.com/hashicorp/terraform/archive/v0.12.26.zip  
- name: terraform-provider-azurerm
  version: 2.20.0
  source: https://github.com/terraform-providers/terraform-provider-azurerm/archive/v2.20.0.zip
- name: terraform-provider-random
  version: 2.2.1
  source: https://releases.hashicorp.com/terraform-provider-random/2.2.1/terraform-provider-random_2.2.1_linux_amd64.zip
- name: terraform-provider-mysql
  version: 1.9.0
  source: https://releases.hashicorp.com/terraform-provider-mysql/1.9.0/terraform-provider-mysql_1.9.0_linux_amd64.zip 
- name: terraform-provider-null
  version: 2.1.2
  source: https://releases.hashicorp.com/terraform-provider-null/2.1.2/terraform-provider-null_2.1.2_linux_amd64.zip
- name: psqlcmd
  version: 0.1.0
  source: https://packages.microsoft.com/config/rhel/7/packages-microsoft-prod.rpm
  url_template: ../build/${name}_${version}_${os}_${arch}.zip
- name: sqlfailover
  version: 0.1.0
  source: https://packages.microsoft.com/config/rhel/7/packages-microsoft-prod.rpm
  url_template: ../build/${name}_${version}_${os}_${arch}.zip  
- name: terraform-provider-postgresql
  version: 1.5.0
  source: https://github.com/terraform-providers/terraform-provider-postgresql/archive/v1.5.0.zip
  ```

This section defines all the binaries and terraform providers that will be bundled into the brokerpak. 
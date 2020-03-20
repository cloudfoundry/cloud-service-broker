# Installing the broker on Azure

The broker service and the Azure brokerpak can be pushed and registered on a foundation running on Azure.

## Requirements

### CloudFoundry running on Azure.
The Azure brokerpak services are provisioned with firewall rules that only allow internal connectivity. This allows `cf pushed` applications access, while denying any public access.

### Azure Service Credentials
The services need to be provisioned in the same Azure account that the foundation is running in. To do this, the broker needs the following credentials to manage resources within that account:
- `ARM_SUBSCRIPTION_ID`
- `ARM_TENANT_ID`
- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`

### MySQL Database for Broker State
The broker keeps service instance and binding information in a MySQL database. 

#### Binding a MySQL Database
If there is an existing broker in the foundation that can provision a MySQL instance ([MASB](https://github.com/Azure/meta-azure-service-broker) is one option,) use `cf create-service` to create a new MySQL instance. Then use `cf bind-service` to bind that instance to the service broker.

#### Manually Provisioning a MySQL Database
If a MySQL instance needs to be manually provisioned, it must be accessible to applications running within the foundation so that the `cf push`ed broker can access it. The following configuration parameters will be needed:
- `DB_HOST`
- `DB_USERNAME`
- `DB_PASSWORD`

## Step By Step The Quick Way
Maybe the simplest way is to create the state database with MASB and then use the Makefile to deploy and register the broker.

### Requirements
The following tools are needed on your workstation:
- [go 1.13](https://golang.org/dl/)
- make
- [cf cli](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html)

The [MASB](https://github.com/Azure/meta-azure-service-broker) service broker must be installed in your Cloud Foundry foundation.

### Assumptions
The `cf` CLI has been used to authenticate with a foundation (`cf api` and `cf login`,) and an org and space have been targeted (`cf target`)

### Clone the Repo
The following commands will clone the service broker repository and cd into the resulting directory.
```bash
git clone https://github.com/pivotal/"${APP_NAME}".git
cd "${APP_NAME}"
```
### Set Required Environment Variables
Collect the Azure service credentials for your account and set them as environment variables:
```bash
export ARM_SUBSCRIPTION_ID=your subscription id
export ARM_TENANT_ID=your tenant id
export ARM_CLIENT_ID=your client id
export ARM_CLIENT_SECRET=your client secret
```
Generate username and password for the broker - Cloud Foundry will use these credentials to authenticate API calls to the service broker.
```bash
export SECURITY_USER_NAME=someusername
export SECURITY_USER_PASSWORD=somepassword
```
### Create a MySQL instance with MASB
The following command will create a basic MySQL database instance named `csb-sql`
```bash
cf create-service azure-mysqldb basic1 csb-sql
```
### Use the Makefile to Deploy the Broker
There is a make target that will build the broker and brokerpak and deploy to and register with Cloud Foundry as a space scoped broker. This will be local and private to the org and space your `cf` CLI is targeting.
```bash
make push-broker-azure
```
Once this completes, the output from `cf marketplace` should include:
```
azure-mongodb                small, medium, large                        The Cosmos DB service implements wire protocols for MongoDB.  Azure Cosmos DB is Microsoft's globally distributed, multi-model database service for mission-critical application
azure-mssql                  small, medium, large, extra-large           Azure SQL Database is a fully managed service for the Azure Platform
azure-mssql-db               small, medium, large, extra-large           Manage Azure SQL Databases on pre-provisioned database servers
azure-mssql-failover-group   small, medium, large                        Manages auto failover group for managed Azure SQL on the Azure Platform
azure-mssql-server           standard                                    Azure SQL Server (no database attached)
azure-mysql                  small, medium,                              Mysql is a fully managed service for the Azure Platform
azure-redis                  small, medium, large, ha-small, ha-medium,  Redis is a fully managed service for the Azure Platform
```
## Step By Step Slightly Harder Way

Requirements and assumptions are the same as above. Follow instructions for the first two steps above ([Clone the Repo](#Clone-the-Repo) and [Set Required Environment Variables](Set-Required-Environment-Variables))

### Create a MySQL Database
Its an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access. Set the following environment variables with information about that MySQL instance:
```bash
export DB_HOST=mysql server host
export DB_USERNAME=mysql server username
export DB_PASSWORD=mysql server password
```

### Build the Broker and Brokerpak
Use the makefile to build the broker executable and brokerpak.
```bash
make build-azure-brokerpak
```
### Pushing the Broker
All the steps to push and register the broker:
```
APP_NAME=cloud-service-broker

cf push --no-start

cf set-env "${APP_NAME}" SECURITY_USER_PASSWORD "${SECURITY_USER_PASSWORD}"
cf set-env "${APP_NAME}" SECURITY_USER_NAME "${SECURITY_USER_NAME}"

cf set-env "${APP_NAME}" ARM_SUBSCRIPTION_ID "${ARM_SUBSCRIPTION_ID}"
cf set-env "${APP_NAME}" ARM_TENANT_ID "${ARM_TENANT_ID}"
cf set-env "${APP_NAME}" ARM_CLIENT_ID "${ARM_CLIENT_ID}"
cf set-env "${APP_NAME}" ARM_CLIENT_SECRET "${ARM_CLIENT_SECRET}"

cf set-env "${APP_NAME}" DB_HOST "${DB_HOST}"
cf set-env "${APP_NAME}" DB_USERNAME "${DB_USERNAME}"
cf set-env "${APP_NAME}" DB_PASSWORD "${DB_PASSWORD}"

cf set-env "${APP_NAME}" GSB_BROKERPAK_BUILTIN_PATH ./azure-brokerpak

cf start "${APP_NAME}"

BROKER_NAME=csb-$USER

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
```
Once these steps are complete, the output from `cf marketplace` should resemble the same as above.
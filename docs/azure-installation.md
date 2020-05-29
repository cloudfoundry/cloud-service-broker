# Installing the broker on Azure

The broker service and the Azure brokerpak can be pushed and registered on a foundation running on Azure.

Documentation for broker configuration can be found [here](./configuration.md).

## Requirements

### CloudFoundry running on Azure.
The Azure brokerpak services are provisioned with firewall rules that only allow internal connectivity. This allows `cf push`ed applications access, while denying any public access.

### Azure Service Credentials
The services need to be provisioned in the same Azure account that the foundation is running in. To do this, the broker needs the following service principal credentials to manage resources within that account:
- tenant id
- subscription id
- client id
- client secret

#### Service Principal Roles and Required Providers
The subscription will require registered providers for each of the services that will be deployed.

> If the service principal being used has the `Contributor` role, provider registration should be automatic and the following can just be used for reference. 

> If the service principal being used does not have rights for automatic provider registration, the broker should be configured to disable this feature.
> Make sure the following is part of the `provision.defaults` part of the config file:
> ```yaml
> provision: 
>   defaults: '{
>     "skip_provider_registration": true
>   }' 

You can list the providers in the subscription, and make sure that the namespace is registered. For example, if you want to enable Service Bus service, `Microsoft.ServiceBus` should be registered. If the specific provider is not registered, you need to run `azure provider register <PROVIDER-NAME>` to register it.

```
$ azure provider list
info:    Executing command provider list
+ Getting ARM registered providers
data:    Namespace                  Registered
data:    -------------------------  -------------
data:    Microsoft.Batch            Registered
data:    Microsoft.Cache            Registered
data:    Microsoft.Compute          Registered
data:    Microsoft.DocumentDB       Registered
data:    Microsoft.EventHub         Registered
data:    microsoft.insights         Registered
data:    Microsoft.KeyVault         Registered
data:    Microsoft.MySql            Registered
data:    Microsoft.Network          Registering
data:    Microsoft.ServiceBus       Registered
data:    Microsoft.Sql              Registered
data:    Microsoft.Storage          Registered
data:    Microsoft.ApiManagement    NotRegistered
data:    Microsoft.Authorization    Registered
data:    Microsoft.ClassicCompute   NotRegistered
data:    Microsoft.ClassicNetwork   NotRegistered
data:    Microsoft.ClassicStorage   NotRegistered
data:    Microsoft.Devices          NotRegistered
data:    Microsoft.Features         Registered
data:    Microsoft.HDInsight        NotRegistered
data:    Microsoft.Resources        Registered
data:    Microsoft.Scheduler        Registered
data:    Microsoft.ServiceFabric    NotRegistered
data:    Microsoft.StreamAnalytics  NotRegistered
data:    Microsoft.Web              NotRegistered
info:    provider list command OK
```

##### Services and their required providers
| Service | Namespace |
|---------|-----------|
| redis   | `Microsoft.Cache` |
| mysql   | `Microsoft.DBforMySQL` |
| mssql   | `Microsoft.Sql` |
| mongodb | `Microsoft.DocumentDB` |
| eventhubs | `Microsoft.EventHub` |
| postgresql | `Microsoft.DBforPostgreSQL` |
| storage | `Microsoft.Storage` |
| cosmosdb | `Microsoft.DocumentDB` |

### MySQL Database for Broker State
The broker keeps service instance and binding information in a MySQL database. 

#### Binding a MySQL Database
If there is an existing broker in the foundation that can provision a MySQL instance ([MASB](https://github.com/Azure/meta-azure-service-broker) is one option,) use `cf create-service` to create a new MySQL instance. Then use `cf bind-service` to bind that instance to the service broker.

#### Manually Provisioning a MySQL Database
If a MySQL instance needs to be manually provisioned, it must be accessible to applications running within the foundation so that the `cf push`ed broker can access it. The following configuration parameters will be needed:
- `DB_HOST`
- `DB_USERNAME`
- `DB_PASSWORD`

It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`).

## Step By Step From a Pre-build Release with a Bound MySQL Instance

Fetch a pre-built broker and brokerpak and bind it to a `cf create-service` managed MySQL.

### Requirements

The following tools are needed on your workstation:
- [cf cli](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html)

### Assumptions

The `cf` CLI has been used to authenticate with a foundation (`cf api` and `cf login`,) and an org and space have been targeted (`cf target`)

### Fetch A Broker and Azure Brokerpak

Download a release from https://github.com/pivotal/cloud-service-broker/releases. Find the latest release matching the name pattern `sb-0.1.0-rc.XXX-azure-0.0.1-rc.YY`. This will have a broker and brokerpak that have been tested together. Follow the hyperlink into that release and download `cloud-servic-broker` and `azure-services-0.1.0-rc.YY.brokerpak` into the same directory on your workstation.

### Create a MySQL instance with MASB
The following command will create a basic MySQL database instance named `csb-sql`
```bash
cf create-service azure-mysqldb basic1 csb-sql
```
### Build Config File
To avoid putting any sensitive information in environment variables, a config file can be used.

Create a file named `config.yml` in the same directory the broker and brokerpak have been downloaded to. Its contents should be:

```yaml
azure:
  subscription_id: your subscription id
  tenant_id: your tenant id
  client_id: your client id
  client_secret: your client secret

api:
  user: someusername
  password: somepassword
```

### Push and Register the Broker

Push the broker as a binary application:

```bash
SECURITY_USER_NAME=someusername
SECURITY_USER_PASSWORD=somepassword
APP_NAME=cloud-service-broker

chmod +x cloud-service-broker
cf push "${APP_NAME}" -c './cloud-service-broker serve --config config.yml' -b binary_buildpack --random-route --no-start
```

Bind the MySQL database and start the service broker:
```bash
cf bind-service cloud-service-broker csb-sql
cf start "${APP_NAME}"
```
Register the service broker:
```bash
BROKER_NAME=csb-$USER

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
```
Once this completes, the output from `cf marketplace` should include:
```
csb-azure-mongodb                small, medium, large                        The Cosmos DB service implements wire protocols for MongoDB.  Azure Cosmos DB is Microsoft's globally distributed, multi-model database service for mission-critical application
csb-azure-mssql                  small, medium, large, extra-large           Azure SQL Database is a fully managed service for the Azure Platform
csb-azure-mssql-db               small, medium, large, extra-large           Manage Azure SQL Databases on pre-provisioned database servers
csb-azure-mssql-failover-group   small, medium, large                        Manages auto failover group for managed Azure SQL on the Azure Platform
csb-azure-mssql-server           standard                                    Azure SQL Server (no database attached)
csb-azure-mysql                  small, medium,                              Mysql is a fully managed service for the Azure Platform
csb-azure-redis                  small, medium, large, ha-small, ha-medium,  Redis is a fully managed service for the Azure Platform
```

## Step By Step From a Pre-built Release with a Manually Provisioned MySQL Instance

Fetch a pre-built broker and brokerpak and configure with a manually provisioned MySQL instance.

Requirements and assumptions are the same as above. Follow instructions above to [fetch the broker and brokerpak](#Fetch-A-Broker-and-Azure-Brokerpak)

### Create a MySQL Database
Its an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access. The database connection values (hostname, user name and password) will be needed in the next step. It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`).

### Build Config File
To avoid putting any sensitive information in environment variables, a config file can be used.

Create a file named `config.yml` in the same directory the broker and brokerpak have been downloaded to. Its contents should be:

```yaml
azure:
  subscription_id: your subscription id
  tenant_id: your tenant id
  client_id: your client id
  client_secret: your client secret

db:
  host: your mysql host
  password: your mysql password
  user: your mysql username
api:
  user: someusername
  password: somepassword
```

### Push and Register the Broker

Push the broker as a binary application:

```bash
SECURITY_USER_NAME=someusername
SECURITY_USER_PASSWORD=somepassword
APP_NAME=cloud-service-broker

chmod +x cloud-service-broker
cf push "${APP_NAME}" -c './cloud-service-broker serve --config config.yml' -b binary_buildpack --random-route
```

Register the service broker:
```bash
BROKER_NAME=csb-$USER

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
```

Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Step By Step From Source with Bound MySQL
Grab the source code, build and deploy.

### Requirements

The following tools are needed on your workstation:
- [go 1.14](https://golang.org/dl/)
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
Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Step By Step Slightly Harder Way

Requirements and assumptions are the same as above. Follow instructions for the first two steps above ([Clone the Repo](#Clone-the-Repo) and [Set Required Environment Variables](Set-Required-Environment-Variables))

### Create a MySQL Database
Its an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access. It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`). Set the following environment variables with information about that MySQL instance:
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
```bash
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

## Uninstalling the Broker
First, make sure there are all service instances created with `cf create-service` have been destroyed with `cf delete-service` otherwise removing the broker will fail.

### Unregister the Broker
```bash
cf delete-service-broker csb-$USER
```

### Uninstall the Broker
```bash
cf delete cloud-service-broker
```



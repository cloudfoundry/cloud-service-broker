# Creating an Azure MySQL DB for Service Broker State

The Cloud Service Broker (CSB) requires a MySQL database to keep its internal state. Here are the `az` cli steps to help create a database instance and get the parameters required to configure the CSB to use the native instances.

## Requirements

* [az cli](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest)
* `az login` has been executed to authenticate against Azure account

## Script

There is a script that can automate this [here](https://github.com/pivotal/cloud-service-broker/blob/master/scripts/azure-create-mysql-db.sh)

```bash
azure-create-mysql-db.sh <name> <resource group> <location>
```

You will pick a name for database server and the Azure resource group and Azure location the database should be created in.

The final output will be the credentials needed to configure the broker to use the database instance:

```bash
Server Details
FQDN: <name>.mysql.database.azure.com
Admin Username: ifOuuVydAjJNzYHF@<name>
Admin Password: 9Kagdsl8VWhw1eQpp8WMVQplnOp156Ly
Database Name: csb-db
```

If you're `cf push`ing the broker, these values should be used for the config file values:
* db.host
* db.user
* db.password
* db.name

or the environment variables:
* DB_HOST
* DB_USERNAME
* DB_PASSWORD
* DB_NAME

If you're deploying the broker as a tile through OpsMan, these values should be used for the following fields on the *Cloud Service Broker for Microsoft Azure -> Service Broker Config* tab:
* Database host
* Database username
* Database password
* Database name



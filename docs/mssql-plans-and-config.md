# Azure SQL Plans and Config

## Applies to service *csb-azure-mssql*

These are the default plans and configuration options for Azure SQL on Azure (not currently supported on GCP or AWS.)

## Plans

| Plan       | CPUs | Storage Size |
|------------|------|--------------|
| small       | 2    | 50GB         |
| medium      | 8    | 200GB        |
| large       | 32   | 500GB        |
| extra-large | 64   | 1TB          |

## Configuration Options

The following plan parameters can be configured.

| Option Name | Values              | Default |
|-------------|---------------------|---------|
| max_storage_gb  |            | 50      |
| cores       | 1-64, multiple of 2 | 2       |


## Provision Parameters

The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql ... -c '{...}'`)

| Parameter Name | Type | Description | Default  |
|----------------|------|-------------|----------|
| instance_name  |string| service instance name | csb-azsql-*instance_id* |
| resource_group |string| resource group for instance | rg-*instance_name* |
| db_name        |string| database name | csb-db |
| location       |string| Azure region to deploy service instance | westus |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| sku_name | string | Azure sku (typically, tier [GP_S,GP,BC,HS] + family [Gen4,Gen5] + cores, e.g. GP_S_Gen4_1, GP_Gen5_8, see [vCore](https://docs.microsoft.com/en-us/azure/azure-sql/database/resource-limits-vcore-single-databases) and [DTU](https://docs.microsoft.com/en-us/azure/azure-sql/database/resource-limits-dtu-single-databases)) Will be computed from cores if empty. `az sql db list-editions -l <location> -o table` will show all valid values. | |
| authorized_network | string | The Azure subnet ID (long form) that the instance is connected to via a service endpoint. The subnet must have the `Microsoft.sql` service enabled. | |
| skip_provider_registration | boolean | `true` to skip automatic Azure provider registration, set if service principal being used does not have rights to register providers | `false` |

Note: Currently Azure SQL is not available in all regions. The enum in the YML lists all the valid regions as of 2/12/2020

### Azure Notes

Azure SQL instances are [vCore model](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-service-tiers-vcore?tabs=azure-portal) and Gen5 hardware generation 
unless overridden by `sku_name` parameter.

CPU/memory size mapped into [Azure sku's](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-vcore-resource-limits-single-databases) as follows:  

#### Core to sku mapping

| Cores | Sku |
|-------|-----|
| 1  | GP_Gen5_1 |
| 2  | GP_Gen5_2 |
| 4  | GP_Gen5_4 |
| 8  | GP_Gen5_8  |
| 16 | GP_Gen5_16 |
| 32 | GP_Gen5_32 |
| 80 | GP_Gen5_80 |

> Note in order for `cf update-service -p <new plan>` to work, the sku's must be the same family (GP_S, GP, or HS.) Otherwise Azure will refuse the update request.

## Binding Credentials

The binding credentials for Azure SQL have the following shape:

```json
{
    "name" : "database name",
    "hostname" : "database server host",
    "port" : "database server port",
    "username" : "authentication user name",
    "password" : "authentication password",
    "uri" : "database connection URI",
    "jdbcUrl" : "jdbc format connection URI",
    "sqldbName" : "database name",
    "sqlServerName" : "server name",
    "sqlServerFullyQualifiedDomainName" : "server fqdn",
    "databaseLogin" : "authentication user name",
    "databaseLoginPassword" : "authentication password"
}
```




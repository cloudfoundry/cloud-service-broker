# Azure SQL Plans and Config

## Applies to service *csb-azure-mssql*

These are the default plans and configuration options for Azure SQL on Azure (not currently supported on GCP or AWS.)

## Plans

| Plan       | CPUs | Storage Size |
|------------|------|--------------|
|small       | 2    | 50GB         |
|medium      | 8    | 200GB        |
|large       | 24   | 500GB        |
|extra-large | 40   | 1GB          |

## Configuration Options

The following plan parameters can be configured.

| Option Name | Values              | Default |
|-------------|---------------------|---------|
| pricing_tier| GP_S, GP, HS, BC    |         |
| storage_gb  | 5 - 4096            | 50      |
| cores       | 1-64, multiple of 2 | 2       |

## Provision Parameters

The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql ... -c '{...}'`)

| Parameter Name | Type | Description | Default  |
|----------------|------|-------------|----------|
| instance_name  |string| service instance name | vsb-azsql-*instance_id* |
| resource_group |string| resource group for instance | rg-*instance_name* |
| db_name        |string| database name | vsb-db |
| location       |string| Azure region to deploy service instance | westus |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `arm.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `arm.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `arm.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `arm.client_secret` |

Note: Currently Azure SQL is not available in all regions. The enum in the YML lists all the valid regions as of 2/12/2020

### Azure Notes

Azure SQL instances are [vCore model](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-service-tiers-vcore?tabs=azure-portal) and Gen5 hardware generation.

CPU/memory size mapped into [Azure sku's](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-vcore-resource-limits-single-databases) as follows:

| Plan        | Sku        | Storage | vCores |
|-------------|------------|---------|--------|
| small       | GP_S_Gen5_2 | 50GB   | 2      |
| medium      | GP_Gen5_8  | 200GB   | 8      |
| large       | HS_Gen5_24 | 500GB   | 24     |
| extra-large | BC_Gen5_40 | 1TB     | 40     |

Pricing tiers map:

| Pricing Tier | Description |
|-|-|
| GP_S | General Purpose tier - Serverless pricing model |
| GP   | General Purpose tier - Provisioned pricing model |
| HS   | Hyperscale tier |
| BC   | Business Critical tier |

Each of the Pricing Tiers in Azure has a [min and max](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-vcore-resource-limits-single-databases) cores:

| Pricing Tier | Max vCores |
|--------------|------------|
| GP_S         | 16         |
| GP           | 80         |
| HS           | 80         |
| BC           | 80         |


## Binding Credentials

The binding credentials for Azure SQL have the following shape:

```
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




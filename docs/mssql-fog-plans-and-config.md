# Microsoft SQL Failover Group Plans and Config
## Applies to service *csb-azure-mssql-failover-group*
These are the default plans and configuration options for Azure SQL Failover Group on Azure (not currently supported on GCP or AWS.)

## Plans

| Plan       | CPUs | Storage Size |
|------------|------|--------------|
|small       | 2    | 50GB         |
|medium      | 8    | 200GB        |
|large       | 40   | 1TB          |

## Configuration Options

The following options may be configured.

| Option Name | Values              | Default |
|-------------|---------------------|---------|
| pricing_tier| GP_S, GP, BC        |         |
| storage_gb  | 5 - 4096            | 50      |
| cores       | 1-64, multiple of 2 | 2       |

### Azure Notes

Except as noted below, configuration is generally the same as [Azure SQL](./mssql-plans-and-config.md)

#### Azure specific config parameters

| Parameter | Type | Description |Default |
|-----------|--------|------------|--------|
| instance_name  |string| service instance name | csb-azsql-*instance_id* |
| resource_group |string| resource group for instance | rg-*instance_name* |
| location  |string|Azure location to deploy service instance | westus |
| failover_location |string|Azure location for failover instance | [default regional pair]([failover_region](https://docs.microsoft.com/en-us/azure/best-practices-availability-paired-regions#azure-regional-pairs))|
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |

Note: Currently Azure SQL is not available in all regions. The enum in the YML lists all the valid regions as of 2/12/2020

## Binding Credentials

The binding credentials for Azure SQL Failover Group have the following shape:

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


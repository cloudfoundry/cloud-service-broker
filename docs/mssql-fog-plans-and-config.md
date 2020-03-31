# Microsoft SQL Failover Group Plans and Config
## Applies to service *csb-azure-mssql-failover-group*
These are the default plans and configuration options for Azure SQL Failover Group on Azure (not supported on GCP or AWS.)

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

| Parameter | Value | Default |
|-----------|--------|--------|
| location  | Azure location to deploy service instance | westus |
| failover_location | Azure location for failover instance | [default regional pair]([failover_region](https://docs.microsoft.com/en-us/azure/best-practices-availability-paired-regions#azure-regional-pairs))|

Note: Currently Azure SQL is not available in all regions. The enum in the YML lists all the valid regions as of 2/12/2020

## Binding Credentials

The binding credentials for Azure SQL Failover Group have the following shape:

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


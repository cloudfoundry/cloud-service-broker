# Microsoft Azure SQL Server

## Applies to service *csb-azure-mssql-server*

*csb-azure-mssql-server* manages stand alone Azure SQL server instances. No databases are created or managed.

## Plans

The only plan is *standard*.

## Provision Parameters
 
 The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql-server standard ... -c '{...}'`)

| Parameter Name | Type | Description | Default |
|----------------|------|-------------|---------|
| instance_name | string | instance name for server | vsb-azsql-svr-*instance_id* |
| resource_group | string | resource group for the server | rg-*instance_name* |
| admin_username | string | admin username for server | randomly generated string |
| admin_password | string | admin password for server | randomly generated string |
| region | string | Azure location to create server | westus |

## Binding Credentials

The binding credentials for Azure SQL Failover Group have the following shape:

```
{
    "hostname" : "database server host",
    "port" : "database server port",
    "username" : "authentication user name",
    "password" : "authentication password",
    "sqlServerName" : "server name",
    "sqldbResourceGroup" : "resource group for server",
    "sqlServerFullyQualifiedDomainName" : "server fqdn",
    "databaseLogin" : "authentication user name",
    "databaseLoginPassword" : "authentication password"
}
```
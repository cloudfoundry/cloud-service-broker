# Microsoft Azure SQL Server

## Applies to service *csb-azure-mssql-server*

*csb-azure-mssql-server* manages stand alone Azure SQL server instances on Azure (not currently supported on GCP or AWS.) No databases are created or managed.

## Plans

The only plan is *standard*.

## Provision Parameters
 
 The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql-server standard ... -c '{...}'`)

| Parameter Name | Type | Description | Default |
|----------------|------|-------------|---------|
| instance_name | string | instance name for server | csb-azsql-svr-*instance_id* |
| resource_group | string | resource group for the server | rg-*instance_name* |
| admin_username | string | admin username for server | randomly generated string |
| admin_password | string | admin password for server | randomly generated string |
| location | string | Azure location to create server | westus |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| authorized_network | string | The Azure subnet ID (long form) that the instance is connected to via a service endpoint. The subnet must have the `Microsoft.sql` service enabled. | |
| skip_provider_registration | boolean | `true` to skip automatic Azure provider registration, set if service principal being used does not have rights to register providers | `false` |

## Binding Credentials

The binding credentials for Azure SQL Failover Group have the following shape:

```json
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
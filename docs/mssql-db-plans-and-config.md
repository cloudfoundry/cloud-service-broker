# Microsoft Azure SQL DB Config

## Applies to service *csb-azure-mssql-db*

*csb-azure-mssql-db* manages Azure SQL databases on pre-configured database servers on Azure (not currently supported on GCP or AWS.)

## Plans

| Plan       | CPUs | Storage Size |
|------------|------|--------------|
|small       | 2    | 50GB         |
|medium      | 8    | 200GB        |
|large       | 24   | 500GB        |
|extra-large | 40   | 1TB          |

## Plan Configuration Parameters

The following plan parameters can be configured.

| Parameter Name | Values              | Default |
|-------------|---------------------|---------|
| pricing_tier| GP_S, GP, BC, HS    |         |
| storage_gb  | 5 - 4096            | 50      |
| cores       | 1-64, multiple of 2 | 2       |

## Provision Parameters

The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql-db ... -c '{...}'`)

| Parameter Name | Type | Description | Default |
|----------------|------|-------------|---------|
| db_name | string | database name | csb-fog-db-*instance_id* |
| server  | string | server name from *server_credentials* on which to create the database | |
| server_credentials | JSON | list of server credentials on which databases can be created, *server* must match one of *name*. Format: `{ "name": { "server_name":"...", "server_resource_group":"...", "admin_username":"...", "admin_password":"..."}, ...}`
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |

## Configuring Global Defaults

An operator will likely configure *server_credentials* for developers to use.

See [configuration documentation](./configuration.md) and [Azure installation documentation](azure-installation.md) for reference.

To globally configure *server_credential*, include the following in the configuration file for the broker:

```yaml
service:
  csb-azure-mssql-db:
    provision:
      defaults: '{ 
        "server1": { 
          "admin_username":"...", 
          "admin_password":"...", 
          "server_name":"...", 
          "server_resource_group":..."
        },
        "server2": {
          "admin_username":"...",
          ...
        }
      }' 
```

A developer could create a new failover group database on *server1* like this:
```bash
cf create-service csb-azure-mssql-db medium medium-sql -c '{"server":"server1"}'
```

## Binding Credentials

The binding credentials for Azure SQL DB have the following shape:

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

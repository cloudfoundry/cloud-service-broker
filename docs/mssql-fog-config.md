# Microsoft Azure SQL Failover Group Config

## Applies to service *csb-azure-mssql-db-failover-group*

*csb-azure-mssql-db-failover-group* manages Azure SQL failover group databases on pre-configured database server pairs on Azure (not currently supported on GCP or AWS.)

## Plans

| Plan       | CPUs | Storage Size |
|------------|------|--------------|
|small       | 2    | 50GB         |
|medium      | 8    | 200GB        |
|large       | 40   | 1TB          |

## Plan Configuration Parameters

The following parameters may be configured.

| Parameter Name | Values           | Default |
|-------------|---------------------|---------|
| pricing_tier| GP_S, GP, BC        |         |
| storage_gb  | 5 - 4096            | 50      |
| cores       | 1-64, multiple of 2 | 2       |

## Provision Parameters

The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql-db-failover-group ... -c '{...}'`)

| Parameter Name | Type | Description | Default |
|----------------|------|-------------|---------|
| instance_name | string | instance name for failover group | vsb-azsql-fog-*instance_id* |
| db_name | string | database name | vsb-fog-db-*instance_id* |
| server_pair | string | server pair from *server_credential_pairs* on which to create failover DB | |
| server_credential_pairs | JSON | list of server pairs on which failover DB's can be created, *server_pair* must match one of *name*. Format: `{ "name": { "admin_username":"...", "admin_password":"...", "primary":{"server_name":"...", "resource_group":..."}, "secondary":{"server_name":"...", "resource_group":..."}, ...}`| |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |

## Configuring Global Defaults

An operator will likely configure *server_credential_pairs* for developers to use.

See [configuration documentation](./configuration.md) and [Azure installation documentation](azure-installation.md) for reference.

To globally configure *server_credential_pairs*, include the following in the configuration file for the broker:

```yaml
service:
  csb-azure-mssql-db-failover-group:
    provision:
      defaults: '{ 
        "pair1": { 
          "admin_username":"...", 
          "admin_password":"...", 
          "primary": {
            "server_name":"...", 
            "resource_group":..."
          }, 
          "secondary": {
            "server_name":"...", 
            "resource_group":..."
          },
        "pair2": {
          "admin_username":"...",
          ...
        }
      }' 
```

A developer could create a new failover group database on *pair1* like this:
```bash
cf create-service csb-azure-mssql-db-failover-group medium medium-fog -c '{"server_pair":"pair1"}'
```

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

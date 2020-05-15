# Microsoft Azure SQL Failover Group Config

## Applies to service *csb-azure-mssql-db-failover-group*

*csb-azure-mssql-db-failover-group* manages Azure SQL failover group databases on pre-configured database server pairs on Azure (not currently supported on GCP or AWS.)

## Plans

| Plan       | CPUs | Storage Size |
|------------|------|--------------|
|small       | 2    | 50GB         |
|medium      | 8    | 200GB        |
|large       | 32   | 500GB          |

## Plan Configuration Parameters

The following parameters may be configured.

| Parameter Name | Values           | Default |
|-------------|---------------------|---------|
| max_storage_gb  |             | 50      |
| cores       | 1-64, multiple of 2 | 2       |

## Provision Parameters

The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql-db-failover-group ... -c '{...}'`)

| Parameter Name | Type | Description | Default |
|----------------|------|-------------|---------|
| instance_name | string | instance name for failover group | csb-azsql-fog-*instance_id* |
| db_name | string | database name | csb-fog-db-*instance_id* |
| server_pair | string | server pair from *server_credential_pairs* on which to create failover DB | |
| server_credential_pairs | JSON | list of server pairs on which failover DB's can be created, *server_pair* must match one of *name*. Format: `{ "name": { "admin_username":"...", "admin_password":"...", "primary":{"server_name":"...", "resource_group":..."}, "secondary":{"server_name":"...", "resource_group":..."}, ...}`| |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| sku_name | string | Azure sku (typically, tier [GP_S,GP,BC,HS] + family [Gen4,Gen5] + cores, e.g. GP_S_Gen4_1, GP_Gen5_8, see https://docs.microsoft.com/en-us/azure/mysql/concepts-pricing-tiers) Will be computed from cores if empty. `az sql db list-editions -l <location> -o table` will show all valid values. | |

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

### Azure Notes

Azure SQL instances are [vCore model](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-service-tiers-vcore?tabs=azure-portal) and Gen5 hardware generation 
unless overridden by `sku_name` parameter.

CPU/memory size mapped into [Azure sku's](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-vcore-resource-limits-single-databases) as follows:  

#### Core to sku mapping

| Cores | Sku |
|-------|-----|
| 1  | P_S_Gen5_1 |
| 2  | P_S_Gen5_2 |
| 4  | P_Gen5_4   |
| 8  | P_Gen5_8   |
| 16 | GP_Gen5_16 |
| 32 | BC_Gen5_32 |
| 64 | BC_Gen5_64 |


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

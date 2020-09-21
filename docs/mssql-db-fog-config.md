# Microsoft Azure SQL Failover Group Config

## Applies to service *csb-azure-mssql-db-failover-group*

*csb-azure-mssql-db-failover-group* manages Azure SQL failover group databases on pre-configured database server pairs on Azure (not currently supported on GCP or AWS.)

## Plans

| Plan       | CPUs | Storage Size |
|------------|------|--------------|
|small       | 2    | 50GB         |
|medium      | 8    | 200GB        |
|large       | 32   | 500GB        |
|existing    | n/a  | n/a          |

The `existing` plan connects to an existing failover group DB to allow applications (typically in a second foundation) to bind to the database.

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
| read_write_endpoint_failover_policy | string | Read/Write failover policy - `Automatic` or `Manual` | `Automatic` |
| failover_grace_minutes | number | grace period in minutes before failover with data loss is attempted | 60 |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| sku_name | string | Azure sku (typically, tier [GP_S,GP,BC,HS] + family [Gen4,Gen5] + cores, e.g. GP_S_Gen4_1, GP_Gen5_8, see [vCore](https://docs.microsoft.com/en-us/azure/azure-sql/database/resource-limits-vcore-single-databases) and [DTU](https://docs.microsoft.com/en-us/azure/azure-sql/database/resource-limits-dtu-single-databases)) Will be computed from cores if empty. `az sql db list-editions -l <location> -o table` will show all valid values. | |
| skip_provider_registration | boolean | `true` to skip automatic Azure provider registration, set if service principal being used does not have rights to register providers | `false` |

## Configuring Global Defaults

An operator will likely configure *server_credential_pairs* for developers to use.

See [configuration documentation](./configuration.md) and [Azure installation documentation](azure-installation.md) for reference.

To globally configure *server_credential_pairs*, include the following in the configuration file for the broker:

```yaml
service:
  csb-azure-mssql-db-failover-group:
    provision:
      defaults: '{
        "server_credential_pairs": { 
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
        }
      }' 
```

A developer could create a new failover group database on *pair1* like this:
```bash
cf create-service csb-azure-mssql-db-failover-group medium medium-fog -c '{"server_pair":"pair1"}'
```

To allow multiple foundations to connect to a single database (an example could be foundations in primary and secondary fail over locations) the same server credential pairs would be configured in each foundation.

A developer could create a failover group DB in one foundation:
```bash
cf create-service csb-azure-mssql-db-failover-group medium medium-fog -c '{"server_pair":"pair1", "instance_name":"fog-instance","db_name":"db"}'
```

And then connect to that db in the second foundation:
```bash
cf create-service csb-azure-mssql-db-failover-group existing medium-fog -c '{"server_pair":"pair1","instance_name": "csb-failover-group-test", "db_name":"test-db"}'
```

### Azure Notes

Azure SQL instances are [vCore model](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-service-tiers-vcore?tabs=azure-portal) and Gen5 hardware generation 
unless overridden by `sku_name` parameter.

CPU/memory size mapped into [Azure vcore sku's](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-vcore-resource-limits-single-databases) as follows:  

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

## Notes on the Secondary DB created by FOG
Since the secondary (failover) database is implicitly created when the failover group is created, the broker doesn't currently gain any control over that db instance. This will manifest itself when a `cf update` or `cf delete-service` is executed on the fail over group. Changes to SKU, etc won't get propagated to the secondary DB. The current workaround is to subsume control of the secondary DB instance that can then be managed with `cf`

### Steps to gain control of secondary DB instance

The status output from `cf service <fog instance>` has the information needed to subsume control over the secondary DB:

```bash
$ cf service example-fog
Showing info of service example-fog in org pivotal / space ernie as admin...

name:            example-fog
service:         csb-azure-mssql-failover-group
tags:            
plan:            small
description:     Manages auto failover group for managed Azure SQL on the Azure Platform
documentation:   https://docs.microsoft.com/en-us/azure/sql-database/sql-database-auto-failover-group/
dashboard:       

Showing status of last operation from service example-fog...

status:    create succeeded
message:   created failover group csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143 (id:
           /subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-primary/failoverGroups/csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143), primary db csb-db (id:
           /subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-primary/databases/csb-db) on server csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-primary (id:
           /subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-primary), secondary db csb-db (id:
           /subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-secondary/databases/csb-db) on server csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-secondary (id:
           /subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-secondary) URL:
           https://portal.azure.com/#@29248f74-371f-4db2-9a50-c62a6877a0c1/resource/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-primary/failoverGroup
started:   2020-09-21T18:43:51Z
updated:   2020-09-21T18:49:04Z

There are no bound apps for this service.
```

The secondary db id is the required piece of information to subsume control of the db instance.

```bash
$ cf create-service csb-masb-mssql-db-subsume current secondary-db -c '{"azure_db_id":"/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/csb-azsql-fog-8fa66105-2796-4bed-a2a3-d1691fbba143-secondary/databases/csb-db"}'
Creating service instance secondary-db in org pivotal / space ernie as admin...
OK

Create in progress. Use 'cf services' or 'cf service secondary-db' to check operation status.

Attention: The plan `current` of service `csb-masb-mssql-db-subsume` is not free.  The instance `secondary-db` will incur a cost.  Contact your administrator if you think this is in error.
```

The secondary db is now under `cf` control and `cf update` may be used to modify its configuration. 
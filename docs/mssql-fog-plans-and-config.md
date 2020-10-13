# Microsoft SQL Failover Group Plans and Config
## Applies to service *csb-azure-mssql-failover-group*
These are the default plans and configuration options for SQL Failover Group on Azure (not currently supported on GCP or AWS.)

## Plans

| Plan       | CPUs | Storage Size |
|------------|------|--------------|
| small       | 2    | 50GB         |
| medium      | 8    | 200GB        |
| large       | 32   | 500GB        |

## Configuration Options

The following options may be configured.

| Option Name | Values              | Default |
|-------------|---------------------|---------|
| max_storage_gb  |             | 50      |
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
| read_write_endpoint_failover_policy | string | Read/Write failover policy - `Automatic` or `Manual` | `Automatic` |
| failover_grace_minutes | number | grace period in minutes before failover with data loss is attempted | 60 |
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
Since the secondary (failover) database is implicitly created when the failover group is created, the broker doesn't currently gain any control over that db instance. This will manifest itself when a `cf update` is executed. Changes to SKU, etc won't get propagated to the secondary DB. The current workaround is to subsume control of the secondary DB instance that can then be managed with `cf`

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

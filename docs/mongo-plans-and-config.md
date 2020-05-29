# MongoDB Plans and Config
These are the default plans and configuration options for MongoDB across the supported cloud platforms (AWS, Azure and GCP.)

## Plans

| Plan            | API Version | Database Size | Throughput    |
|-----------------|-------------|---------------|---------------|
| small           | 3.2         | Unlimited     | Low           | 
| medium          | 3.2         | Unlimited     | Medium        | 
| large           | 3.2         | Unlimited     | High          | 

## Configuration Options

The following options can be configured across all supported platforms. Notes below document any platform specific information for mapping that might be required.

| Option Name     | Value Type   | Values                      | Default |
|-----------------|--------------|-----------------------------|---------|
| instance_name   | string       | `^[a-z][a-z0-9-]+$`         | csb-mongo-*instance_id* |
| db_name         | string       |                             | default_db |
| collection_name | string       |                             | default_collection |
| shard_key       | string       |                             | uniqueKey |

### Azure Notes - applies to service *csb-azure-mongodb*

Plan sizes mapped into [Azure sku's](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-vcore-resource-limits-single-databases) as follows:

| Plan   | Request Units      |
|--------|--------------------|
| small  | 400 RU             |
| medium | 1,000 RU           |
| large  | 10,000 RU          | 

#### Azure specific config parameters

The following parameters (and those above) may be configured during service provisioning (`cf create-service csb-azure-mongodb ... -c '{...}'`

| Option Name     | Value Type   | Values                      | Default                |
|-----------------|--------------|-----------------------------|------------------------|
| request_units   | integer      | Provisioned throughput in [request units](https://docs.microsoft.com/en-us/azure/cosmos-db/request-units). 400-100,000 (multiples of 100)| plan default or `400`    |
| location        | string       | Azure location of Mongo instance | westus |
| resource_group  | string       | Resource group for Mongo instance | rg-*instance_name* |
| failover_locations          | string       | Comma separated list of [Azure regions](https://docs.microsoft.com/en-us/azure/cosmos-db/regional-presence) (in failover priority) to deploy service instance.  The first region is the default write-enable region. |    |
| ip_range_filter | string | [CosmosDB Firewall Support](https://docs.microsoft.com/en-us/azure/cosmos-db/firewall-support). This value specifies the set of IP addresses or IP address ranges in CIDR form to be included as the allowed list of client IP's for a given database account. IP addresses/ranges must be comma separated and must not contain any spaces. 0.0.0.0 allows access from Azure networks.  An empty string "" allows access from all public networks. | `0.0.0.0` |
| consistency_level | string | The [Consistency Level](https://docs.microsoft.com/en-us/azure/cosmos-db/consistency-levels) to use for this CosmosDB Account - can be either BoundedStaleness, Eventual, Session, Strong or ConsistentPrefix | `Session` |
| max_interval_in_seconds | integer | (Optional) When used with the Bounded Staleness consistency level, this value represents the time amount of staleness (in seconds) tolerated. Accepted range for this value is 5 - 86400 (1 day). Defaults to 5. Required when consistency_level is set to BoundedStaleness. | `5` |
| max_staleness_prefix | integer | (Optional) When used with the Bounded Staleness consistency level, this value represents the number of stale requests tolerated. Accepted range for this value is 10 – 2147483647. Defaults to 100. Required when consistency_level is set to BoundedStaleness. | `100` |
| enable_multiple_write_locations | boolean | Enable multi-master support for this Cosmos DB account. | `false` |
| enable_automatic_failover | boolean | Enable automatic fail over for this Cosmos DB account. | `false` |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| skip_provider_registration | boolean | `true` to skip automatic Azure provider registration, set if service principal being used does not have rights to register providers | `false` |


### AWS Notes
Plan sizes mapped into [AWS DocumentDB On-Demand instance types](https://aws.amazon.com/documentdb/pricing/) as follows:

| Plan  | Instance type |
|-------|----------|
| small | db.r5.large |
| medium | db.r5.2xlarge |
| large | db.r5.4xlarge |

#### AWS specific config parameters

| Parameter | Value |
|-----------|--------|
| region  | [AWS region](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions) to deploy service instance |

### GCP Notes

MySQL minimum storage size is 10GB.

CPU/memory size mapped into [GCP tiers](https://cloud.google.com/sql/pricing#2nd-gen-pricing) as follows:

| Plan  | Tier     |
|-------|----------|
| small | db-n1-standard-2 |
| medium | db-n1-standard-4 |
| large | db-n1-standard-8 |

#### Core to service tier mapping

| Cores | Service Tier |
|-------|---------------|
| 1     | db-n1-standard-1  |
| 2     | db-n1-standard-2  |
| 4     | db-n1-standard-4  |
| 8     | db-n1-standard-8  |
| 16    | db-n1-standard-16  |
| 32    | db-n1-standard-32  |
| 64    | db-n1-standard-64  |

#### GCP specific config parameters

| Parameter | Value |
|-----------|--------|
| region  | [GCP region](https://cloud.google.com/compute/docs/regions-zones) to deploy service instance into |
| authorized_network | compute network to connect the instance to |

## Binding Credentials

The binding credentials for MySQL have the following shape:

```json
{
    "uri" : "mongodb connection URI"
}
```
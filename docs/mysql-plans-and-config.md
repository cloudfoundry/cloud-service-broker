# MySQL Plans and Config

These are the default plans and configuration options for MySQL across the supported cloud platforms (AWS, Azure and GCP.)

## Plans

| Plan | Version | CPUs | Memory Size | Disk Size |
|------|---------|------|-------------|-----------|
|small | 5.7     | 2    | min 4GB     | 5GB       |
|medium| 5.7     | 4    | min 8GB     | 10GB      |
|large | 5.7     | 8    | min 16GB    | 20GB      |

## Configuration Options

The following options can be configured across all supported platforms. Notes below document any platform specific information for mapping that might be required.

| Option Name | Values    | Default |
|-------------|-----------|---------|
| mysql_version | 5.6, 5.7| 5.7    |
| storage_gb  | 5 - 4096  | 5      |
| cores       | 1,2,4,8,16,32,64 | 1      |
| db_name     |           | csb-db |
 
### Azure Notes - applies to *csb-azure-mysql*
CPU/memory size mapped into [Azure sku's](https://docs.microsoft.com/en-us/azure/mysql/concepts-pricing-tiers) as follows:

| Plan   | Sku       |
|--------|-----------|
| small  | GP_Gen5_2 |
| medium | GP_Gen5_4 |
| large  | GP_Gen5_8 |

> Note that the maximum vCores is dependent on the Service Tier. B_ = Basic, GP_ = General Purpose and MO_ = Memory Optimized. See below for details.

> Note in order for `cf update-service -p <new plan>` to work, the sku's must be the same family (B, GP, or MO.) Otherwise Azure will refuse the update request.

#### Storage
[Storage auto grow](https://docs.microsoft.com/en-us/azure/mysql/concepts-pricing-tiers#storage-auto-grow) is enabled on Azure. Initial storage sizes are per plan.

#### Core to sku mapping

| Cores | Instance class |
|-------|----------------|
| 1     | GP_Gen5_1      |
| 2     | GP_Gen5_2      |
| 4     | GP_Gen5_4      |
| 8     | GP_Gen5_8      |
| 16    | GP_Gen5_16     |
| 32    | GP_Gen5_32     |
| 64    | GP_Gen5_64     |

#### Azure specific config parameters

The following parameters (as well as those above) may be configured during service provisioning (`cf create-service csb-azure-mysql ... -c '{...}'`
)
| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| instance_name | string |name of Azure instance to create | csb-mysql-*instance_id* |
| location  | string |Azure location to deploy service instance | westus |
| resource_group | string |The Azure resource group in which to create the instance | rg-*instance_name* |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| authorized_network  | string| Subnet ID (the long version) of the VNET/Subnet that is attached to this instance to allow remote access. By default no VNETs are allowed access. || sku_name | string |[Azure sku](https://docs.microsoft.com/en-us/azure/mysql/concepts-pricing-tiers) (typically, tier [`B`,`GP`,`MO`] + family [`Gen4`,`Gen5`] + *cores*, e.g. `B_Gen4_1`, `GP_Gen5_8`) *overrides* `cores` conversion into sku per table above
| use_tls | boolean |Use TLS for DB connections | `true` |
| skip_provider_registration | boolean | `true` to skip automatic Azure provider registration, set if service principal being used does not have rights to register providers | `false` |

Note: Currently MySQL is not available in all regions. The enum in the YML lists all the valid regions as of 2/12/2020

### AWS Notes
CPU/memory size mapped into [AWS DB instance types](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) as follows:

| Plan  | Instance class |
|-------|----------|
| small | db.t2.medium |
| medium | db.m4.xlarge |
| large | db.m4.2xlarge |

#### Core to instance class mapping

| Cores | Instance class |
|-------|---------------|
| 1     | db.t2.small  |
| 2     | db.t3.medium  |
| 4     | db.m5.xlarge  |
| 8     | db.m5.2xlarge |
| 16    | db.m5.4xlarge |
| 32    | db.m5.8xlarge |
| 64    | db.m5.16xlarge|

#### AWS specific config parameters

| Parameter | Type | Description | Default |
|-----------|------|------|---------|
 instance_name | string | name of AWS instance to create | csb-mysql-*instance_id* |
| region  | string | [AWS region](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions) to deploy service  | us-west-2 |
| vpc_id | string | The VPC to connect the instance to | the default vpc |
| aws_access_key_id | string | ID of Azure tenant for instance | config file value `aws.access_key_id` |
| aws_secret_access_key | string | ID of Azure subscription for instance | config file value `aws.secret_access_key` |
| instance_class | string | explicit [instance class](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) *overrides* `cores` conversion into instance class per table above | | 
| multi-az | boolean | If `true`, create multi-az ([HA](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.MultiAZ.html)) instance | `false` | 
| publicly_accessible | boolean | If `true`, make instance available to public connections | `false ` |
| storage_autoscale | boolean | If `true`, storage will autoscale to max of *storage_autoscale_limit_gb* | `false` |
| storage_autoscale_limit_gb | number | if *storage_autoscale* is `true`, max size storage will scale up to ||
| storage_encrypted | boolean | If `true`, DB storage will be encrypted | `false`|
| parameter_group_name | string | MySQL parameter group name for instance | `default.mysql.<mysql version>` |
| rds_subnet_group | string | Name of subnet to attach DB instance to, overrides *aws_vpc_id* | |
| rds_vpc_security_group_ids | comma delimited string | Security group ID's to assign to DB instance | |
| use_tls | boolean |Use TLS for DB connections | `true` |
| allow_major_version_upgrade | bool | Indicates that major version upgrades are allowed. | `true` |
| auto_minor_version_upgrade  | bool | Indicates that minor engine upgrades will be applied automatically to the DB instance during the maintenance window| `true` |
| maintenance_day | integer | Day of week for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html) |
| maintenance_start_hour | integer | Start hour for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html)|
| maintenance_start_min | integer | Start minute for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html)|
| maintenance_end_hour | integer | End hour for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html)|
| maintenance_end_min | integer | End minute for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html)|

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
    "name" : "database name",
    "hostname" : "database server host",
    "port" : "database server port",
    "username" : "authentication user name",
    "password" : "authentication password",
    "uri" : "database connection URI",
    "jdbcUrl" : "jdbc format connection URI"
}
```

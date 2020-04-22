# PostgreSQL Plans and Config

These are the default plans and configuration options for PostgreSQL across the supported cloud platforms (AWS, Azure and GCP.)

## Plans

| Plan | Version | CPUs | Memory Size | Disk Size |
|------|---------|------|-------------|-----------|
|small | 11      | 2    | min 4GB     | 5GB       |
|medium| 11      | 4    | min 8GB     | 10GB      |
|large | 11      | 8    | min 16GB    | 20GB      |


## Configuration Options

The following options can be configured across all supported platforms. Notes below document any platform specific information for mapping that might be required.

| Option Name | Values | Default |
|-------------|--------|---------|
| postgres_version | 9.5, 9.6, 10, 11 | 11    |
| storage_gb  | 5 - 4096| 5      |
| cores       | 1,2,4,8,16,32,64 | 1      |
| db_name     | | csb-db |

### Azure Notes - applies to *csb-azure-postgresql*
CPU/memory size mapped into [Azure sku's](https://docs.microsoft.com/en-us/azure/mysql/concepts-pricing-tiers) as follows:

| Plan   | Sku       | Memory | Storage | vCores |
|--------|-----------|--------|---------|--------|
| small  | B_Gen5_2  | 4GB    | 50GB     | 2      |
| medium | GP_Gen5_4 | 10GB   | 200GB    | 4      |
| large  | MO_Gen5_8 | 20GB   | 500GB    | 8      |

#### Core to sku mapping

| Cores | Instance class |
|-------|---------------|
| 1     | B_Gen5_1    |
| 2     | B_Gen5_2    |
| 4     | GP_Gen5_4   |
| 8     | MO_Gen5_8   |
| 16    | MO_Gen5_16  |
| 32    | MO_Gen5_32  |
| 64    | GP_Gen5_64  |

#### Azure specific config parameters

The following parameters (as well as those above) may be configured during service provisioning (`cf create-service csb-azure-postgresql ... -c '{...}'`
)
| Parameter | Description | Default |
|-----------|-------------|---------|
| instance_name | name of Azure instance to create | csb-mysql-*instance_id* |
| location  | Azure location to deploy service instance | westus |
| resource_group | The Azure resource group in which to create the instance | rg-*instance_name* |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| authorized_network  | Subnet ID (the long version) of the VNET/Subnet that is attached to this instance to allow remote access. By default no VNETs are allowed access. ||
| sku_name | [Azure sku](https://docs.microsoft.com/en-us/azure/mysql/concepts-pricing-tiers) (typically, tier [`B`,`GP`,`MO`] + family [`Gen4`,`Gen5`] + *cores*, e.g. `B_Gen4_1`, `GP_Gen5_8`) *overrides* `cores` conversion into sku per table above

Note: Currently MySQL is not available in all regions. The enum in the YML lists all the valid regions as of 2/12/2020

### AWS Notes - applies to *csb-aws-postgresql*

CPU/memory size mapped into [AWS DB instance types](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) as follows:

| Plan  | Instance class |
|-------|----------|
| small | db.t2.medium |
| medium | db.m4.xlarge |
| large | db.m4.2xlarge |

#### Core to instance class mapping

| Cores | Instance class |
|-------|---------------|
| 1     | db.m1.medium  |
| 2     | db.t2.medium  |
| 4     | db.m4.xlarge  |
| 8     | db.m4.2xlarge |
| 16    | db.m4.4xlarge |
| 32    | db.m5.8xlarge |
| 64    | db.m5.16xlarge|

#### AWS specific config parameters

The following parameters (as well as those above) may be configured during service provisioning (`cf create-service csb-aws-postgresql ... -c '{...}'`

| Parameter | Type | Description | Default |
|-----------|------|------|---------|
 instance_name | string | name of Azure instance to create | csb-mysql-*instance_id* |
| region  | string | [AWS region](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions) to deploy service  | us-west-2 |
| vpc_id | string | The VPC to connect the instance to | the default vpc |
| aws_access_key_id | string | ID of Azure tenant for instance | config file value `aws.access_key_id` |
| aws_secret_access_key | string | ID of Azure subscription for instance | config file value `aws.secret_access_key` |
| instance_class | string | explicit [instance class](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) *overrides* `cores` conversion into instance class per table above | | 
| multi-az | boolean | If `true`, create multi-az ([HA](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.MultiAZ.html)) instance | `false` | 
| publicly_accessible | boolean | If `true`, make instance available to public connections | `false ` |

## Binding Credentials

The binding credentials for PostgreSQL have the following shape:

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

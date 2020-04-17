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

| Option Name | Values | Default |
|-------------|--------|---------|
| mysql_version | 5.6, 5.7| 5.7    |
| storage_gb  | 5 - 4096| 5      |
| cores       | 1,2,4,8,16,32,64 | 1      |
| db_name     | | csb-db |
 
### Azure Notes - applies to *csb-azure-mysql*
CPU/memory size mapped into [Azure sku's](https://docs.microsoft.com/en-us/azure/mysql/concepts-pricing-tiers) as follows:

| Plan   | Sku       | Memory | Storage | vCores |
|--------|-----------|--------|---------|--------|
| small  | B_Gen5_2  | 4GB    | 50GB     | 2      |
| medium | GP_Gen5_4 | 10GB   | 200GB    | 4      |
| large  | MO_Gen5_8 | 20GB   | 500GB    | 8      |

> Note that the maximum vCores is dependant on the Service Tier. B_ = Basic, GP_ = General Purpose and MO_ = Memory Optimized. See below for details.

> Note that the maximum vCores is dependant on the Service Tier. B_ = Basic, GP_ = General Puspose and MO_ = Memory Optimized. See below for details.

#### Azure specific config parameters

The following parameters (as well as those above) may be configured during service provisioning (`cf create-service csb-azure-mysql ... -c '{...}'`
)
| Parameter | Description | Default |
|-----------|-------------|---------|
| instance_name | name of Azure instance to create | csb-mysql-*instance_id* |
| location  | Azure location to deploy service instance | westus |
| pricing_tier | One of B, GP, MO | |
| resource_group | The Azure resource group in which to create the instance | rg-*instance_name* |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |

Note: Currently MySQL is not available in all regions. The enum in the YML lists all the valid regions as of 2/12/2020

Each of the so-called Pricing Tiers in Azure has a min and max cores:
| Pricing Tier | Max vCores |
|--------------|------------|
| B - Basic | 2 |
| GP - General Purpose | 64 |
| MO - Memory Optimized | 32 |

| Parameter | Value |
|-----------|--------|
| authorized_network  | Subnet ID (the long version) of the VNET/Subnet that is attached to this instance to allow remote access. By default no VNETs are allowed access. |

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
| 1     | db.m1.medium  |
| 2     | db.t2.medium  |
| 4     | db.m4.xlarge  |
| 8     | db.m4.2xlarge |
| 16    | db.m4.4xlarge |
| 32    | db.m5.8xlarge |
| 64    | db.m5.16xlarge|

#### AWS specific config parameters

| Parameter | Type | Description | Default |
|-----------|------|------|---------|
 instance_name | string | name of Azure instance to create | csb-mysql-*instance_id* |
| region  | string | [AWS region](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions) to deploy service  | us-west-2 |
| vpc_id | string | The VPC to connect the instance to | the default vpc |
| aws_access_key_id | string | ID of Azure tenant for instance | config file value `aws.access_key_id` |
| aws_secret_access_key | string | ID of Azure subscription for instance | config file value `aws.secret_access_key` |
| instance_class | string | explicit [instance class](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) *overrides* `cores` conversion into instance class per table above | | 

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

```
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

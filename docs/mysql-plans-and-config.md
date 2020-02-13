# MySQL Plans and Config
These are the default plans and configuration options for MySQL across the supported cloud platforms (AWS, Azure and GCP.)

## Plans

| Plan | Version | CPUs | Memory Size | Disk Size |
|------|---------|------|-------------|-----------|
|Small | 5.7     | 2    | min 4GB     | 5GB       |
|Medium| 5.7     | 4    | min 8GB     | 10GB      |
|Large | 5.7     | 8    | min 16GB    | 20GB      |

## Configuration Options

The following options can be configured across all supported platforms. Notes below document any platform specific information for mapping that might be required.

| Option Name | Values | Default |
|-------------|--------|---------|
| mysql_version | 5.6, 5.7| 5.7    |
| storage_gb  | 5 - 4096| 5      |
| cores       | 1,2,4,8,16,32,64 | 1      |

### Azure Notes
CPU/memory size mapped into [Azure sku's](https://docs.microsoft.com/en-us/azure/sql-database/sql-database-vcore-resource-limits-single-databases) as follows:

| Plan  | Sku      |
|-------|----------|
| small | B_Gen5_2 |
| medium | GP_Gen5_4 |
| large | MO_Gen5_8 |

#### Azure specific config parameters

| Parameter | Value |
|-----------|--------|
| location  | Azure region to deploy service instance |

TODO: document how core count is mapped to an Azure sku.

### AWS Notes
CPU/memory size mapped into [AWS DB instance types](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) as follows:

| Plan  | Instance type |
|-------|----------|
| small | db.t2.medium |
| medium | db.m4.xlarge |
| large | db.m4.2xlarge |

#### Core to instance type mapping

| Cores | Instance type |
|-------|---------------|
| 1     | db.m1.medium  |
| 2     | db.t2.medium  |
| 4     | db.m4.xlarge  |
| 8     | db.m4.2xlarge |
| 16    | db.m4.4xlarge |
| 32    | db.m5.8xlarge |
| 64    | db.m5.16xlarge|

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

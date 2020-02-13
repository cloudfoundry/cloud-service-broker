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

| Plan  | Sku      | Memory | Storage | vCores |
|-------|----------|--------|---------|--------|
| small | B_Gen5_2 | 4GB | 50GB | 2 |
| medium | GP_Gen5_4 | 10GB | 200GB | 4 |
| large | MO_Gen5_8 | 20GB | 500GB | 8 |

#### Azure specific config parameters

| Parameter | Value |
|-----------|--------|
| region  | Azure region to deploy service instance |
Note: Currently MySQL is not available in all regions. The enum in the YML lists all the valid regions as of 2/12/2020

Each of the so-called Pricing Tiers in Azure has a min and max cores:
| Pricing Tier | Max vCores |
|--------------|------------|
| Basic | 2 |
| General Purpose | 64 |
| Memory Optimized | 32 |


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

### GCP Notes
CPU/memory size mapped into [GCP tiers](https://cloud.google.com/sql/pricing#2nd-gen-pricing) as follows:

| Plan  | Tier     |
|-------|----------|
| small | db-n1-standard-2 |
| medium | db-n1-standard-4 |
| large | db-n1-standard-8 |

TODO: document how core count is mapped to a GCP tier.

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

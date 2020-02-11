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
| version     | 5.6, 5.7| 5.7    |
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
CPU/memory size mapped into [AWS instance types](https://aws.amazon.com/ec2/instance-types/) as follows:

| Plan  | Instance type |
|-------|----------|
| small | t3.medium |
| medium | t3.xlarge |
| large | t3a.2xlarge |

TODO: document how core count is mapped to an AWS instance type.

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

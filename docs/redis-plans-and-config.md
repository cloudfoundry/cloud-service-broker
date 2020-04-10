# Redis Plans and Config
These are the default plans and configuration options for Redis across the supported cloud platforms (AWS, Azure and GCP.)

## Plans

| Plan | Cache Size | HA | 
|------|------------|----|
| small | min 1GB | no |
| medium | min 4GB | no |
| large | min 16GB | no |

## Configuration Options

The following options can be configured across all supported platforms. Notes below document any platform specific information for mapping that might be required.

### GCP Notes
Cache size mapped into [GCP memory store tiers](https://cloud.google.com/memorystore/pricing) as follows:

| Plan | GCP Service Tier | Memory Size |
|------|------------------| ------------|
| small | Basic           | 1GB |
| medium | Basic          | 4GB |
| large | Basic           | 16GB |

| Option Name | Values | Default |
|-------------|--------|---------|
| cache_size  | 1,2,4,8,16,32,64 | 1    |

### AWS Notes
Cache size mapped into [AWS ElastiCash node types](https://aws.amazon.com/elasticache/pricing/
) as follows:

| Plan | AWS Cache Node Type |
|------|---------------------|
| small | cache.t2.small |
| medium | cache.m5.large |
| large | cache.r4.xlarge |

TODO: document how cache_size is mapped to an AWS node type

### Azure Notes - applies to *csb-azure-redis*

Cache size mapped into [Azure sku's for Redis](https://azure.microsoft.com/en-us/pricing/details/cache/) as follows:

#### Plan Parameters

The following parameters plan options

| Parameter | Description | Default |
|-----------|-------------|---------|
| sku_name | Basic, Standard, Premium | |
| family | C, P | |
| capacity | relative size of cache | 1 |

#### Basic Plans:
| Plan | Sku | Family | Capacity | Memory Size | HA | 
|------|--------|-----|------------| ------------| ---- |
| small | Basic | C | 1 | 1GB | no |
| medium | Basic | C | 3 | 6GB | no |
| large | Basic | C | 5 | 26GB | no |

#### High Availability Plans:

| Plan | Sku | Family | Capacity | Memory Size | HA | 
|------|--------|-----|------------| ------------| ---- |
| ha-small | Standard | C | 1 | 1GB | yes |
| ha-medium | Standard | C | 3 | 6GB | yes |
| ha-large | Standard | C | 5 | 26GB | yes |


#### Configuration Options

The following parameters (as well as those above) may be configured at service provisioning time (`cf create-service csb-azure-redis ... -c '{...}'`)

| Parameter | Type | Description | Default |
|-----------|------|------|---------|
| instance_name | string | name of Azure instance to create | vsb-redis-*instance_id* |
| location  | string | Azure location to deploy service instance | westus |
| resource_group | string | The Azure resource group in which to create the instance | rg-*instance_name* |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |

#### Notes
For consuming Azure Redis, the TLS port is used in place of the standard port.  The key for the TLS port is "tls_port".  The standard port is disabled for both the Azure Basic Plans as well as Azure High Availability Plans.

## Binding Credentials

The binding credentials for Redis have the following shape:

```
{
    "host" : "redis server hostname",
    "port" : "redis server port",
    "password" : "authentication password",
    "tls_port" : "redis TLS port"
}
```

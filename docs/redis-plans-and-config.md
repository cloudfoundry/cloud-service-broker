# Redis plans and Config
These are the default plans and configuration options for Redis across the supported cloud platforms (AWS, Azure and GCP.)

## Plans

| Plan | Cache Size | HA | 
|------|------------|----|
| small | min 1GB | no |
| medium | min 4GB | no |
| large | min 16GB | no |


### GCP Notes
Cache size mapped into [GCP memory store tiers](https://cloud.google.com/memorystore/pricing) as follows:

| Plan | GCP Service Tier | Memory Size |
|------|------------------| ------------|
| small | Basic           | 1GB |
| medium | Basic          | 4GB |
| large | Basic           | 16GB |

### AWS Notes
Cache size mapped into [AWS ElastiCash node types](https://aws.amazon.com/elasticache/pricing/
) as follows:

| Plan | AWS Cache Node Type |
|------|---------------------|
| small | cache.t2.small |
| medium | cache.m5.large |
| large | cache.r4.xlarge |

TODO: document how cache size is mapped to an AWS node type

### Azure Notes
Cache size mapped into [Azure sku's for Redis](https://azure.microsoft.com/en-us/pricing/details/cache/) as follows:

| Plan | Family | Cache Name |
|------|--------|------------|
| small | Basic | C1 |
| medium | Basic | C3 |
| large | Basic | C5 |

TODO: document how cache size is mapped to an Azure cache name

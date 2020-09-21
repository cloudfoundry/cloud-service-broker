# Azure Storage Account Config
## Applies to service *csb-azure-storage-account*

*csb-azure-storage-account* manages an Azure Storage Account (not currently supported on GCP or AWS.)

## Plans

| Plan | Description |
|------|-------------|
| standard | General-purpose V2 account. Locally redundant, standard tier |

## Config parameters

The following parameters may be configured during service provisioning (`cf create-service csb-azure-storage-account ... -c '{...}'`

| Parameter | Type | Description | Default |
|-----------|------|------|---------|
| storage_account_type | string | Account type - `BlobStorage`, `BlockBlobStorage`, `FileStorage`, `Storage`, `StorageV2` | `StorageV2` |
| tier | string | Storage tier to use - `Standard`, `Premium` | `Standard` |
| replication_type | string | Replication type - `LRS`, `GRS`, `RAGRS`, `ZRS` | `LRS` |
| location  | Azure location to deploy service instance | westus |
| resource_group | The Azure resource group in which to create the instance | rg-*account_name* (account_name is always generated) |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| skip_provider_registration | boolean | `true` to skip automatic Azure provider registration, set if service principal being used does not have rights to register providers | `false` |
| authorized_networks | list (string) | A list of resource ids for subnets of the Azure Vnet authorized | `[]`

## Binding Credentials

The binding credentials for Azure Storage Account has the following shape:

```json
{
    "storage_account_name" : "storage account name",
    "primary_access_key" : "primary access key",
    "secondary_access_key" : "secondary access key",
}
```

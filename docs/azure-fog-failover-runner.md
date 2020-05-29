# Azure SQL Failover Group Failover Runner

## Applies to *csb-azure-mssql-fog-run-failover* service

The run failover service is a pseudo service that upon provisioning will cause a fail over group secondary server to become the primary. Upon de-provisioning, the failover will be undone and the original primary server will once again be primary. Not currently supported on GCP or AWS.

## Plans

The only plan is *standard*.

## Provision Parameters
 
 The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql-fog-run-failover standard ... -c '{...}'`)

| Parameter Name | Type | Description | Default |
|----------------|------|-------------|---------|
| fog_instance_name | string | instance name for failover group to target |
| server_pair_name | string | server pair from *server_pairs* |
| server_pairs | JSON | list of failover group server pairs, *server_pair* must match one of *name*. Format: `{ "name": { "primary":{"server_name":"...", "resource_group":..."}, "secondary":{"server_name":"...", "resource_group":..."}, ...}` |
| azure_tenant_id | string | ID of Azure tenant for instance | config file value `azure.tenant_id` |
| azure_subscription_id | string | ID of Azure subscription for instance | config file value `azure.subscription_id` |
| azure_client_id | string | ID of Azure service principal to authenticate for instance creation | config file value `azure.client_id` |
| azure_client_secret | string | Secret (password) for Azure service principal to authenticate for instance creation | config file value `azure.client_secret` |
| skip_provider_registration | boolean | `true` to skip automatic Azure provider registration, set if service principal being used does not have rights to register providers | `false` |

## Binding Credentials

There are no binding credentials for this service. 
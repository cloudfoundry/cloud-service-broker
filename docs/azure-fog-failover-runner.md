# Azure SQL Failover Group Failover Runner

## Applies to *csb-azure-mssql-fog-run-failover* service

The run failover service is a pseudo service that upon provisioning will cause a fail over group secondary server to become the primary. Upon de-provisioning, the failover will be undone and the original primary server will once again be primary. Not currently supported on GCP or AWS.

## Plans

The only plan is *standard*.

## Provision Parameters
 
 The following parameters may be configured during service provisioning (`cf create-service csb-azure-mssql-fog-run-failover standard ... -c '{...}'`)

| Parameter Name | Type | Description |
|----------------|------|-------------|
| fog_instance_name | string | instance name for failover group to target |
| server_pair_name | string | server pair from *server_pairs* |
| server_pairs | JSON | list of failover group server pairs, *server_pair* must match one of *name*. Format: `{ "name": { "primary":{"server_name":"...", "resource_group":..."}, "secondary":{"server_name":"...", "resource_group":..."}, ...}` |

## Binding Credentials

There are no binding credentials for this service. 
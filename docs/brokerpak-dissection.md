# Dissecting a Brokerpak

So you want to build a brokerpak from scratch? As long as you can write some terraform to control whatever resource you're interested in controlling, and the lifecycle can be mapped into the [OSBAPI API](https://www.openservicebrokerapi.org/), you should be able to write a brokerpak to fulfill your needs.

Hopefully this document helps familiarize you enough with layout and details of the brokerpak specification to help you along the way.

## Requirements

### Talent
* familiarity with [yaml](https://yaml.org/spec/1.2/spec.html) - yaml glues all the pieces together.
* knowledge of [terraform](https://www.terraform.io/docs/index.html) - terraform scripts do the heavy lifting of managing resource lifecycle.

## References
* [Brokerpak Introduction](./brokerpak-intro.md)
* [Brokerpak Specification](./brokerpak-specification.md)

## The Azure Brokerpak

Browse the brokerpak contents [here](https://github.com/pivotal/cloud-service-broker/tree/master/azure-brokerpak)

Have a look at the Azure brokerpak, starting with *[manifest.yml](https://github.com/pivotal/cloud-service-broker/blob/master/azure-brokerpak/manifest.yml)*

The file should resemble:

```yaml
packversion: 1
name: azure-services
version: 0.1.0
metadata:
  author: VMware
platforms:
- os: linux
  arch: amd64
- os: darwin
  arch: amd64
terraform_binaries:
- name: terraform
  version: 0.12.26
  source: https://github.com/hashicorp/terraform/archive/v0.12.26.zip  
- name: terraform-provider-azurerm
  version: 2.20.0
  source: https://github.com/terraform-providers/terraform-provider-azurerm/archive/v2.20.0.zip
- name: terraform-provider-random
  version: 2.2.1
  source: https://releases.hashicorp.com/terraform-provider-random/2.2.1/terraform-provider-random_2.2.1_linux_amd64.zip
- name: terraform-provider-mysql
  version: 1.9.0
  source: https://releases.hashicorp.com/terraform-provider-mysql/1.9.0/terraform-provider-mysql_1.9.0_linux_amd64.zip 
- name: terraform-provider-null
  version: 2.1.2
  source: https://releases.hashicorp.com/terraform-provider-null/2.1.2/terraform-provider-null_2.1.2_linux_amd64.zip
- name: psqlcmd
  version: 0.1.0
  source: https://packages.microsoft.com/config/rhel/7/packages-microsoft-prod.rpm
  url_template: ../build/${name}_${version}_${os}_${arch}.zip
- name: sqlfailover
  version: 0.1.0
  source: https://packages.microsoft.com/config/rhel/7/packages-microsoft-prod.rpm
  url_template: ../build/${name}_${version}_${os}_${arch}.zip  
- name: terraform-provider-postgresql
  version: 1.5.0
  source: https://github.com/terraform-providers/terraform-provider-postgresql/archive/v1.5.0.zip
env_config_mapping:
  ARM_SUBSCRIPTION_ID: azure.subscription_id
  ARM_TENANT_ID: azure.tenant_id
  ARM_CLIENT_ID: azure.client_id
  ARM_CLIENT_SECRET: azure.client_secret
service_definitions:
- azure-redis.yml
- azure-mysql.yml
- azure-mssql.yml
- azure-mssql-failover.yml
- azure-mongodb.yml
- azure-eventhubs.yml
- azure-mssql-db.yml
- azure-mssql-server.yml
- azure-mssql-db-failover.yml
- azure-mssql-fog-run-failover.yml
- azure-resource-group.yml
- azure-postgres.yml
- azure-storage-account.yml
- azure-cosmosdb-sql.yml
```

Let's break it down.

### Header

```yaml
packversion: 1
name: azure-services
version: 0.1.0
metadata:
  author: VMware
```

Metadata about the brokerpak.

| Field | Value |
|-------|-------|
| packversion | should always be 1 |
| name | name of brokerpak |
| version | version of brokerpak |
| metadata | a map of metadata to add to broker |

Besides *packversion* (which should always be 1,) these values are left to the brokerpak author to describe the brokerpak.

### Platforms

```yaml
platforms:
- os: linux
  arch: amd64
- os: darwin
  arch: amd64
```

Describes which platforms the brokerpak should support. Typically *os: linux* is the minimum required for `cf push`ing the broker into CloudFoundry. For local development on OSX, adding *os: darwin* allows the broker to run locally.

### Binaries

```yaml
terraform_binaries:
- name: terraform
  version: 0.12.26
  source: https://github.com/hashicorp/terraform/archive/v0.12.26.zip  
- name: terraform-provider-azurerm
  version: 2.20.0
  source: https://github.com/terraform-providers/terraform-provider-azurerm/archive/v2.20.0.zip
- name: terraform-provider-random
  version: 2.2.1
  source: https://releases.hashicorp.com/terraform-provider-random/2.2.1/terraform-provider-random_2.2.1_linux_amd64.zip
- name: terraform-provider-mysql
  version: 1.9.0
  source: https://releases.hashicorp.com/terraform-provider-mysql/1.9.0/terraform-provider-mysql_1.9.0_linux_amd64.zip 
- name: terraform-provider-null
  version: 2.1.2
  source: https://releases.hashicorp.com/terraform-provider-null/2.1.2/terraform-provider-null_2.1.2_linux_amd64.zip
- name: psqlcmd
  version: 0.1.0
  source: https://packages.microsoft.com/config/rhel/7/packages-microsoft-prod.rpm
  url_template: ../build/${name}_${version}_${os}_${arch}.zip
- name: sqlfailover
  version: 0.1.0
  source: https://packages.microsoft.com/config/rhel/7/packages-microsoft-prod.rpm
  url_template: ../build/${name}_${version}_${os}_${arch}.zip  
- name: terraform-provider-postgresql
  version: 1.5.0
  source: https://github.com/terraform-providers/terraform-provider-postgresql/archive/v1.5.0.zip
  ```

This section defines all the binaries and terraform providers that will be bundled into the brokerpak when its built. The *os* and *arch* parameters are substituted from the platforms section above.

| Field | Value |
|-------|-------|
| name  | name of artifact|
| version | version of artifact |
| source | URL for source code archive of artifact |
| url_template | non-standard location to copy artifact from (default: *https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip*)|

### Environment Config Mapping

The broker can be supplied runtime configuration through environment variables and/or a configuration file. Those values can be made available for use in the brokerpak (see [here](brokerpak-specification.md#functions)) via `config("config.key")`

To map values supplied as environment variables (often when `cf push`ing the broker) into config keys that may be referenced in the brokerpak, add them to the *environment_config_mapping* section of the manifest:

```yaml
env_config_mapping:
  ARM_SUBSCRIPTION_ID: azure.subscription_id
  ARM_TENANT_ID: azure.tenant_id
  ARM_CLIENT_ID: azure.client_id
  ARM_CLIENT_SECRET: azure.client_secret
  ```

These make the runtime environment variables *ARM_SUBSCRIPTION_ID*, *ARM_TENANT_ID*, *ARM_CLIENT_ID* and *ARM_CLIENT_SECRET* available as config values *azure.subscription_id*, *azure.tenant_id*, *azure.client_id*, and *azure.client_secret* respectively.

### Service Definitions

The final section references the yml files that describe the services that will be supported by the brokerpak.

```yaml
service_definitions:
- azure-redis.yml
- azure-mysql.yml
- azure-mssql.yml
- azure-mssql-failover.yml
- azure-mongodb.yml
- azure-eventhubs.yml
- azure-mssql-db.yml
- azure-mssql-server.yml
- azure-mssql-db-failover.yml
- azure-mssql-fog-run-failover.yml
- azure-resource-group.yml
- azure-postgres.yml
- azure-storage-account.yml
- azure-cosmosdb-sql.yml
```

Each of theses service yml files and their requisite terraform will be bundled into the brokerpak.

## A Service Definition

Now lets dive into one of the service yaml files, *[azure-mssql-db.yml](https://github.com/pivotal/cloud-service-broker/blob/master/azure-brokerpak/azure-mssql-db.yml)*

```yaml
version: 1
name: csb-azure-mssql-db
id: 6663f9f1-33c1-4f7d-839c-d4b7682d88cc
description: Manage Azure SQL Databases on pre-provisioned database servers
display_name: Azure SQL Database
image_url: https://msdnshared.blob.core.windows.net/media/2017/03/azuresqlsquaretransparent1.png
documentation_url: https://docs.microsoft.com/en-us/azure/sql-database/
support_url: https://docs.microsoft.com/en-us/azure/sql-database/
tags: [azure, mssql, sqlserver, preview]
plan_updateable: true
plans:
- name: small
  id: fd07d12b-94f8-4f69-bd5b-e2c4e84fafc1
  description: 'SQL Server latest version. Instance properties: General Purpose - Serverless ; 0.5 - 2 cores ; Max Memory: 6gb ; 50 GB storage ; auto-pause enabled after 1 hour of inactivity'
  display_name: "Small"
  properties:
    subsume: false
- name: medium
  id: 3ee14bce-33e8-4d02-9850-023a66bfe120
  description: 'SQL Server latest version. Instance properties: General Purpose - Provisioned ; Provisioned Capacity ; 8 cores ; 200 GB storage'
  display_name: "Medium"
  properties:
    cores: 8
    max_storage_gb: 200
    subsume: false
- name: large
  id: 8f1c9c7b-80b2-49c3-9365-a3a059df9907
  description: 'SQL Server latest version. Instance properties: Business Critical ; Provisioned Capacity ; 32 cores ; 500 GB storage'
  display_name: "Large"
  properties:
    cores: 32
    max_storage_gb: 500
    subsume: false
- name: extra-large
  id: 09096759-58a8-41d0-96bf-39b02a0e4104
  description: 'SQL Server latest version. Instance properties: Business Critical ; Provisioned Capacity ; 80 cores ; 1 TB storage'
  display_name: "Extra Large"
  properties:
    cores: 80
    max_storage_gb: 1024
    subsume: false
- name: subsume
  id: 7781fa41-f486-447a-942c-ded8cccb8299
  description: 'Subsume control of an existing SQL Database'
  display_name: "Subsume"
  properties:
    subsume: true
provision:
  import_inputs:
  - field_name: azure_db_id
    type: string
    details: Azure resource id for database to subsume
    tf_resource: azurerm_mssql_database.azure_sql_db
  import_parameter_mappings:
  - tf_variable: sku_name
    parameter_name: local.sku_name
  - tf_variable: max_size_gb
    parameter_name: var.max_storage_gb 
  - tf_variable: tags
    parameter_name: var.labels
  - tf_variable: retention_days
    parameter_name: var.short_term_retention_days
  import_parameters_to_delete: [ "azurerm_mssql_database.azure_sql_db.id", 
                                 "azurerm_mssql_database.azure_sql_db.min_capacity",
                                 "azurerm_mssql_database.azure_sql_db.long_term_retention_policy",
                                 "azurerm_mssql_database.azure_sql_db.extended_auditing_policy"]
  plan_inputs:
  - field_name: subsume
    type: boolean
    details: Subsume existing DB
  user_inputs:
  - field_name: cores
    type: number
    default: 2
    details: Number vcores for the instance (upto the maximum allowed for the service tier)
    constraints:
      maximum: 80
      minimum: 1
      multipleOf: 2
  - field_name: max_storage_gb
    type: number
    default: 5
    details: Maximum storage allocated to the database instance in GB      
  - field_name: db_name
    type: string
    details: Name for your database
    default: csb-db-${request.instance_id}
    constraints:
      maxLength: 64
  - field_name: server
    type: string
    details: Name of server from server_credentials to create database upon
    required: true
  - field_name: server_credentials
    type: object
    details: 'JSON has of server credentials. { "name1":{"server_name":"...", "server_resource_group":"...", "admin_username":"...", "admin_password":"..."},"name2":{...}...}'
    required: true
  - field_name: azure_tenant_id
    type: string
    details: Azure Tenant to create resource in
    default: ${config("azure.tenant_id")}      
  - field_name: azure_subscription_id
    type: string
    details: Azure Subscription to create resource in
    default: ${config("azure.subscription_id")}      
  - field_name: azure_client_id
    type: string
    details: Client ID of Azure principal 
    default: ${config("azure.client_id")}      
  - field_name: azure_client_secret
    type: string
    details: Client secret for Azure principal
    default: ${config("azure.client_secret")}
  - field_name: skip_provider_registration
    type: boolean
    details: Skip automatic Azure provider registration, set to true if service principal being used does not have rights to register providers
    default: false    
  - field_name: sku_name
    type: string
    details: Azure sku (typically, tier [GP_S,GP,BC,HS] + family [Gen4,Gen5] + cores, e.g. GP_S_Gen4_1, GP_Gen5_8) Will be computed from cores if empty.
    default: ""       
  - field_name: short_term_retention_days
    type: number
    details: Retention period in days for short term retention (Point in Time Restore) policy
    default: 7
    constraints:
      maximum: 35
  template_refs:
    outputs: terraform/azure-mssql-db/mssql-db-outputs.tf
    provider: terraform/azure-mssql-db/azure-provider.tf
    variables: terraform/azure-mssql-db/mssql-db-variables.tf
    main: terraform/azure-mssql-db/mssql-db-main.tf
    data: terraform/azure-mssql-db/mssql-db-data.tf
  computed_inputs:
  - name: labels
    default: ${json.marshal(request.default_labels)}
    overwrite: true
    type: object
  outputs:
  - field_name: sqlServerName
    type: string
    details: Hostname of the Azure SQL Server
  - field_name: sqldbName
    type: string
    details: The name of the database.    
  - field_name: sqlServerFullyQualifiedDomainName
    type: string
    details: The fully qualifief domain name (FQDN) of the Azure SQL Server
  - field_name: hostname
    type: string
    details: Hostname of the Azure SQL Server
  - field_name: port
    type: integer
    details: The port number to connect to the database on
  - field_name: name
    type: string
    details: The name of the database.
  - field_name: username
    type: string
    details: The username to authenticate to the database server.
  - field_name: password
    type: string
    details: The password to authenticate to the database server.
bind:
  plan_inputs: []
  user_inputs: []
  computed_inputs:
  - name: mssql_db_name
    type: string
    default: ${instance.details["name"]}
    overwrite: true
  - name: mssql_hostname
    type: string
    default: ${instance.details["hostname"]}
    overwrite: true
  - name: mssql_port
    type: integer
    default: ${instance.details["port"]}
    overwrite: true
  - name: admin_username
    type: string
    default: ${instance.details["username"]}
    overwrite: true
  - name: admin_password
    type: string
    default: ${instance.details["password"]}
    overwrite: true
  template_ref: terraform/azure-mssql/bind-mssql.tf
  outputs:
  - field_name: username
    type: string
    details: The username to authenticate to the database instance.
  - field_name: password
    type: string
    details: The password to authenticate to the database instance.  
  - field_name: uri
    type: string
    details: The uri to connect to the database instance and database.
  - field_name: jdbcUrl
    type: string
    details: The jdbc url to connect to the database instance and database.    
  - field_name: jdbcUrlForAuditingEnabled
    type: string
    details: The audit enabled JDBC URL to connect to the database server and database.    
  - field_name: databaseLogin
    type: string
    details: The username to authenticate to the database server.
  - field_name: databaseLoginPassword
    type: string
    details: The password to authenticate to the database server. 
```

Time to break it all down.

### Header

```yaml
version: 1
name: csb-azure-mssql-db
id: 6663f9f1-33c1-4f7d-839c-d4b7682d88cc
description: Manage Azure SQL Databases on pre-provisioned database servers
display_name: Azure SQL Database
image_url: https://msdnshared.blob.core.windows.net/media/2017/03/azuresqlsquaretransparent1.png
documentation_url: https://docs.microsoft.com/en-us/azure/sql-database/
support_url: https://docs.microsoft.com/en-us/azure/sql-database/
tags: [azure, mssql, sqlserver, preview]
plan_updateable: true
```

Metadata about the service.

| Field | Value |
|-------|-------|
| version | should always be 1 |
| name | name of service |
| id | a unique guid |
| description | human readable description of service |
| display_name | human readable name of the service |
| image_url | a link to an image that may be included in documentation |
| documentation_url | link to external documentation that may be included in documentation |
| support_url | link to external support site that may be included in documentation |
| tags | list of tags that will be provided in service bindings |
| plan_updateable | indicates if service support `cf update-service -p` |

Besides *version* (which should always be 1) these values are left to the brokerpak author to describe the service.

### Plans

Next is a list of plans that will be provided as defaults by the service.

```yaml
plans:
- name: small
  id: fd07d12b-94f8-4f69-bd5b-e2c4e84fafc1
  description: 'SQL Server latest version. Instance properties: General Purpose - Serverless ; 0.5 - 2 cores ; Max Memory: 6gb ; 50 GB storage ; auto-pause enabled after 1 hour of inactivity'
  display_name: "Small"
  properties:
    subsume: false
- name: medium
  ...
```

There may be zero or more plan entries.

| Field | Value |
|-------|-------|
| name | name of plan |
| id | a unique guid |
| description | human readable description of plan |
| display_name | human readable plan name |
| properties | list of property values for plan settings, property names must be defined in plan_inputs and user_inputs | 

### Provision and Bind 

The *provision* and *bind* sections contain the inputs, outputs and terraform for the provision and bind operation for the service. They are identical in form, the following sections apply to both.

```yaml
provision:
  plan_inputs:
    ...
  user_inputs:
    ...
  ...
bind:
  user_inputs:
    ...
  ...   
```

### Plan Inputs

Configuration parameters that can only be set as part of plan description. Users may not set these parameters through `cf create-service ... -c {...}` or `cf update-service ... -c {...}`

```yaml
  plan_inputs:
  - field_name: subsume
    type: boolean
    details: Subsume existing DB
```

| Field | Value |
|-------|-------|
| field_name | name of plan variable |
| type | field type |
| details | human readable description of field |

> The plan input *subsume* has special meaning. It is used to designate a plan for `tf import` to take over an existing resource. When *subsume* is true, all *import_parameter_mappings* values must be provided through `cf create-service ... -c {...}` and `cf update-service ... -p subsume` is prohibited.

> input fields must be declared as terraform *variables*. Failure to do so will result in failures to build brokerpak.
### User Inputs

Configuration parameters that my be set as part of a plan or set by the user through `cf create-service ... -c {...}` or `cf update-service ... -c {...}`

```yaml
  user_inputs:
  - field_name: cores
    type: number
    default: 2
    details: Number vcores for the instance (upto the maximum allowed for the service tier)
    constraints:
      maximum: 80
      minimum: 1
      multipleOf: 2
```

| Field | Value |
|-------|-------|
| field_name | name of user variable |
| type | field type |
| details | human readable description of field |
| default | (optional) default value for field |
| constraints | (optional) Holds additional JSONSchema validation for the field. The following keys are supported: `examples`, `const`, `multipleOf`, `minimum`, `maximum`, `exclusiveMaximum`, `exclusiveMinimum`, `maxLength`, `minLength`, `pattern`, `maxItems`, `minItems`, `maxProperties`, `minProperties`, and `propertyNames`.|

> input fields must be declared as terraform *variables*. Failure to do so will result in failures to build brokerpak.

### Import Inputs

In order to support `tf import` to subsume control of existing resources (instead of creating new resources) parameters that represent native resources and the terraform resources they apply to are described in the *import_inputs* section.

```yaml
  import_inputs:
  - field_name: azure_db_id
    type: string
    details: Azure resource id for database to subsume
    tf_resource: azurerm_mssql_database.azure_sql_db
```
| Field | Value |
|-------|-------|
| field_name | name of user variable |
| type | field type |
| details | human readable description of field |
| tf_resource | resource.instance of terraform resource to import |

A user will provide the values through `cf create-service ... -c {...}` and the broker will use them during the `tf import` phase.

### Import Parameter Mapping

Once `tf import` is run to generate matching terraform for the resource, some values may need to be parameterized so that the user may modify them later through `cf update-service`  The 

```yaml
  import_parameter_mappings:
  - tf_variable: max_size_gb
    parameter_name: var.max_storage_gb 
```

This will cause instances of *max_size_gb = ...* in the resulting imported terraform to be replaced with *max_size_gb = var.max_storage_gb* so that it may be updated by the user with `cf update-service ...`

### Import Parameters to Delete

`tf import` will return all current values for a resource, including those that are readonly and my not be set during `tf apply`  List those resource values in *import_parameters_to_delete* and they will be removed between `tf import` and `tf apply`

```yaml
  import_parameters_to_delete: [ "azurerm_mssql_database.azure_sql_db.id", 
                                 "azurerm_mssql_database.azure_sql_db.min_capacity",
                                 "azurerm_mssql_database.azure_sql_db.long_term_retention_policy",
                                 "azurerm_mssql_database.azure_sql_db.extended_auditing_policy"]
```

### Terraform Template References

The terraform that will be executed for provision or bind is referenced in *template_refs*

```yaml
  template_refs:
    outputs: terraform/azure-mssql-db/mssql-db-outputs.tf
    provider: terraform/azure-mssql-db/azure-provider.tf
    variables: terraform/azure-mssql-db/mssql-db-variables.tf
    main: terraform/azure-mssql-db/mssql-db-main.tf
    data: terraform/azure-mssql-db/mssql-db-data.tf
```

See [here](./brokerpak-specification.md#template-references) for details.

### Outputs

Outputs from terraform will be collected into binding credentials.

```yaml
  outputs:
  - field_name: username
    type: string
    details: The username to authenticate to the database instance.
```

| Field | Value |
|-------|-------|
| field_name | name of output field |
| type | field type |
| details | Human readable description of output field |

> output fields *must* be declared as *output* variables in terraform. Failure to do so will result in failures creating brokerpak

> binding credentials will contain all output variables from both the *provision* and *bind* portions of the service yaml.

> there is a special output parameter called *status* that may be declared in terraform, it does not need to be declared in the service manifest yaml. The *status* output value will be returned as the status message for the OSBAPI provision call and will be displayed to the user as the *message* portion of a `cf service <service name>` command. It is recommended that resource ID's and other information that may help a user identify the managed resource.
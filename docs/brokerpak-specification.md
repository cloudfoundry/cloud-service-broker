# Brokerpak V1 specification

This document will explain how a brokerpak is structured, and the schema it follows.

A brokerpak is comprised of a versioned Terraform binary and providers for one
or more platform, a manifest, one or more service definitions, and source code.
Here are the contents of an example brokerpak:

```
MODE        SIZE      NAME
drwx------  0         
drwxr-xr-x  0         bin
drwxr-xr-x  0         bin/linux
drwxr-xr-x  0         bin/linux/386
-rwxrwxr-x  75669952  bin/linux/386/terraform
-rwxr-xr-x  35176256  bin/linux/386/terraform-provider-google_v1.19.0_x4
drwxr-xr-x  0         bin/linux/amd64
-rwxrwxr-x  89397536  bin/linux/amd64/terraform
-rwxr-xr-x  41450144  bin/linux/amd64/terraform-provider-google_v1.19.0_x4
drwx------  0         definitions
-rw-------  5810      definitions/service-0.yml
-rw-------  467       manifest.yml
drwxr-xr-x  0         src
-rw-r--r--  8082786   src/terraform-provider-google.zip
-rw-r--r--  14418716  src/terraform.zip
```

You can create, inspect, validate, document and test brokerpaks using the `pak` sub-command.
Run the command `cloud-service-broker pak help` for more information about creating a pak.

## Manifest

Each brokerpak has a manifest file named `manifest.yml`. This file determines
which architectures your brokerpak will work on, which plugins it will use,
and which services it will provide.

### Schema

#### Manifest YAML file

| Field | Type | Description |
| --- | --- |--- |
| packversion* | int | The version of the schema the manifest adheres to. This MUST be set to `1` to be compatible with the brokerpak specification v1. |
| version* | string | The version of this brokerpak. It's RECOMMENDED you follow [semantic versioning](https://semver.org/) for your brokerpaks. |
| name* | string | The name of this brokerpak. It's RECOMMENDED that this be lower-case and include only alphanumeric characters, dashes, and underscores. |
| metadata | object | A free-form field for key/value pairs of additional information about this brokerpak. This could include the authors, creation date, source code repository, etc. |
| platforms* | array of platform | The platforms this brokerpak will be executed on. |
| terraform_binaries* | array of Terraform resource | The list of Terraform providers and Terraform that'll be bundled with the brokerpak. *The broker currently only supports terraform version v0.12.x and higher*|
| service_definitions* | array of string | Each entry points to a file relative to the manifest that defines a service as part of the brokerpak. |
| parameters | array of parameter | These values are set as environment variables when Terraform is executed. |
| required_env_variables | array of string | These are the required environment variables that will be passed through to the terraform execution environment. Use these to make terraform platform plugin auth credentials available for terraform execution. |
| env_config_mapping |map[string]string | List of mappings of environment variables into config keys, see [functions](#functions) for more information on how to use these |
| terraform_upgrade_path | array of Terraform Upgrade Path | List of Terraform version steps when performing upgrade in ascending order |
| terraform_state_provider_replacements | map of Terraform provider names | Map of terraform providers, where the key represents the old name of the provider and the value represents the new name of the provider. Can be used to replace the provider in the terraform state file when switching providers or upgrading to 0.13. |

#### Platform object

The platform OS and architecture follow Go's naming scheme.

| Field | Type | Description | Valid Values |
| --- | --- | --- | --- |
| os* | string | The operating system of the platform. | `linux`, `darwin` |
| arch* | string | The architecture of the platform. | `"386"`, `amd64` |

#### Terraform resource object

This structure holds information about a specific Terraform version or Resource.

| Field        | Type    | Description                                                                                                                                                                                                                               |
|--------------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name*        | string  | The name of this resource. e.g. `terraform-provider-google-beta`.                                                                                                                                                                         |
| version*     | string  | The version of the resource e.g. 1.19.0. *The broker currently only supports terraform version 0.12.x*                                                                                                                                    |
| source       | string  | (optional) The URL to a zip of the source code for the resource.                                                                                                                                                                          |
| url_template | string  | (optional) A custom URL template to get the release of the given tool. Available parameters are ${name}, ${version}, ${os}, and ${arch}. If unspecified the default Hashicorp Terraform download server is used. Can be a local file.     |
| provider     | string  | (optional) The provider in the form of `namespace/type` (e.g `cyrilgdn/postgresql`). This is required if the provider is not provided by Hashicorp. This should match the source of the provider in terraform.required_providers. |
| default      | boolean | (optional) Where there is more than one version of Terraform, this nominates the default version.                                                                                                                                         |

#### Parameter object

This structure holds information about an environment variable that the user can set on the Terraform instance.
These variables are first resolved from the configuration of the brokerpak then against a global set of values.

| Field | Type | Description |
| --- | --- | --- |
| name* | string | The environment variable that will be injected e.g. `PROJECT_ID`. |
| description* | string | A human readable description of what the variable represents. |

#### Terraform Upgrade Path object

This structure holds information about a step in the Terraform upgrade process

| Field | Type | Description |
| --- | --- | --- |
| version | semver | The terraform version to step through |


### Example

```yaml
packversion: 1
name: my-custom-services
version: 1.0.0
metadata:
  author: someone@my-company.com
platforms:
- os: linux
  arch: "386"
- os: linux
  arch: amd64
terraform_binaries:
- name: terraform
  version: 0.12.21
  source: https://github.com/hashicorp/terraform/archive/v0.12.21.zip
- name: terraform-provider-google
  version: 1.19.0
  source: https://github.com/terraform-providers/terraform-provider-google/archive/v1.19.0.zip
service_definitions:
- custom-cloud-storage.yml
- custom-redis.yml
- service-mesh.yml
parameters:
- name: TF_VAR_redis_version
  description: Set this to override the Redis version globally via injected Terraform variable.
env_config_mapping:
  ENV_CONFIG_VAL: env.config_val
required_env_variables:
- ENV_VARIABLE1
- ENV_VARIABLE2
```

## Services

### Schemas

#### Service YAML flie

| Field | Type | Description |
| --- | --- | --- |
| version* | int |  The version of the schema the service definition adheres to. This MUST be set to `1` to be compatible with the brokerpak specification v1. |
| name* | string | A CLI-friendly name of the service. MUST only contain alphanumeric characters, periods, and hyphens (no spaces). MUST be unique across all service objects returned in this response. MUST be a non-empty string. |
| id* | string | A UUID used to correlate this service in future requests to the Service Broker. This MUST be globally unique such that Platforms (and their users) MUST be able to assume that seeing the same value (no matter what Service Broker uses it) will always refer to this service. |
| description* | string | A short description of the service. MUST be a non-empty string. |
| tags | array of strings | Tags provide a flexible mechanism to expose a classification, attribute, or base technology of a service, enabling equivalent services to be swapped out without changes to dependent logic in applications, buildpacks, or other services. E.g. mysql, relational, redis, key-value, caching, messaging, amqp. |
| display_name* | string | The name of the service to be displayed in graphical clients. |
| image_url* | string | The URL to an image or a data URL containing an image. |
| documentation_url* | string | Link to documentation page for the service. |
| support_url* | string | Link to support page for the service. |
| plan_updateable | boolean | Set to `true` if service supports `cf update-service` 
| plans* | array of [plan objects](#plan-object) | A list of plans for this service, schema is defined below. MUST contain at least one plan. |
| provision* | [action object](#action-object) | Contains configuration for the provision operation, schema is defined below. |
| bind* | [action object](#action-object) | Contains configuration for the bind operation, schema is defined below. |
| examples* | [example object](#example) | Contains examples for the service, used in documentation and testing.  MUST contain at least one example. |

#### Plan object

A service plan in a human-friendly format that can be converted into an OSB compatible plan.

| Field | Type | Description |
| --- | --- | --- |
| name* | string | The CLI-friendly name of the plan. MUST only contain alphanumeric characters, periods, and hyphens (no spaces). MUST be unique within the service. MUST be a non-empty string. |
| id* | string | A GUID for this plan in UUID format. This MUST be globally unique such that Platforms (and their users) MUST be able to assume that seeing the same value (no matter what Service Broker uses it) will always refer to this plan. |
| description* | string | A short description of the plan. MUST be a non-empty string. |
| display_name* | string | The name of the plan to be displayed in graphical clients. |
| bullets | array of string | Features of this plan, to be displayed in a bulleted-list. |
| free | boolean | When false, Service Instances of this plan have a cost. The default is false. |
| properties* | map of string:any | Constant values for the provision and bind calls. They take precedent over any other definition of the same field. |
| provision_overrides | map of string:any | Constant values to be overwritten for the provision calls. |
| bind_overrides | map of string:aany |  Constant values to be overwritten for the bind calls. |

#### Action object

The Action object contains a Terraform template to execute as part of a
provision or bind action, and the inputs and outputs to that template.

| Field | Type | Description |
| --- | --- | --- |
| import_inputs | array of [import-input](#import-input-object) | Defines the variables that will be passed to tf import command |
| import_parameter_mappings | array of [import-parameter-mappings](#import-parameter-mapping-object) | Defines how tf resource variables will be replaced with broker variables between `tf import` and `tf apply` |
| import_parameters_to_delete| array of string | list of `tf import` discovered values to remove before `tf apply`. `tf import` will return read-only values that cannot be set during `tf apply` so they should be listed here to be removed between import and apply |
| import_parameters_to_add | array of [import-parameter-mappings](#import-parameter-mapping-object) | Defines tf resource variables to add between `tf import` and `tf apply`
| plan_inputs | array of [variable](#variable-object) | Defines constraints and settings for the variables plans provide in their properties map. It is used to validate [plan objects](#plan-object) properties field.  |
| user_inputs | array of [variable](#variable-object) | Defines constraints and defaults for the variables users provide as part of their request. |
| computed_inputs | array of [computed variable](#computed-variable-object) | Defines default values or overrides that are executed before the template is run. |
| template | string | The complete HCL of the Terraform template to execute. |
| template_ref | string | A path to HCL of the Terraform template to execute. If present, this will be used to populate the `template` field. |
| templates | map | The complete HCL of the Terraform templates to execute. |
| template_refs | map | standard terraform file [snippet list](#template-references) |
| outputs | array of [variable](#variable-object) | Defines constraints and settings for the outputs of the Terraform template. This MUST match the Terraform outputs and the constraints WILL be used as part of integration testing. |

#### Import Input object

The import input object defines the mapping of an input parameter to a terraform resource on the `tf import` command. The presence of any import input values will trigger a `tf import` before `tf apply` upon `cf create-service`

| Field | Type | Description |
| --- | --- | --- |
| field_name | string | the name of the user input variable to use |
| type | string | The JSON type of the field. This MUST be a valid JSONSchema type excepting null |
| details | string | A description of what this field is |
| tf_resource | string | The tf resource to import given this value. |

Given:
```yaml
  import_inputs:
  - field_name: azure_db_id
    type: string
    details: Azure resource id for database to subsume
    tf_resource: azurerm_mssql_database.azure_sql_db
```

A create service call:
```bash
cf create-service my-service my-plan my-instance -c '{"azure_db_id":"some-id"}'
```

Will result in terraform import:

```bash
terraform import azurerm_mssql_database.azure_sql_db some-id
```

#### Import Parameter Mapping object

The import parameter mapping object defines the tf variable to input variable mapping that will occur between `tf import` and `tf apply`

`tf import` will return current values for all variables for the service instance. In order to allow user configuration of these values to support `cf update-service`, it is necessary to enumerate the terraform resource variables that should be parameterized with broker input variables.

| Field | Type | Description |
| --- | --- | --- |
| tf_variable | string | the terraform resource variable name |
| parameter_name | string | the broker input variable name |

Given:
```yaml
  - tf_variable: requested_service_objective_name
    parameter_name: var.service_objective 
```

Will convert the resulting `tf import`:
```tf
resource "azurerm_mssql_database" "azure_sql_db" {
    requested_service_objective_name = S0
}
```

Into:
```tf
resource "azurerm_mssql_database" "azure_sql_db" {
    requested_service_objective_name = var.service_objective
}
```
Between running `tf import` and `tf apply`

So that:
```bash
cf update-service my-instance -c '{"service_objective":"S1"}'
```

Will successfully update the `requested_service_objective_name` for the instance.

#### Removing TF Values

`tf import` will often return read only values that cannot be set during `tf apply`  The *import_parameters_to_delete* field is used to specify which values to remove before `tf apply` is run.

Given:

```yaml
import_parameters_to_delete: [ "azurerm_mssql_database.azure_sql_db.id" ]
```

Will convert the resulting `tf import`

```tf
resource "azurerm_mssql_database" "azure_sql_db" {
    id = /subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server
    requested_service_objective_name = S0
}
```

into

```tf
resource "azurerm_mssql_database" "azure_sql_db" {
    requested_service_objective_name = S0
}
```

So that `tr apply` will not fail trying to set the read-only field *id*

#### Template References

It is possible to break terraform code into sections to aid reusability and better follow [terraform best practices](https://www.terraform-best-practices.com/code-structure#getting-started-with-structuring-of-terraform-configurations). It is also required to support `tf import` as main.tf is a special case during import.

Given:
```hcl
  template_refs:
    outputs: terraform/subsume-masb-mssql-db/mssql-db-outputs.tf
    provider: terraform/subsume-masb-mssql-db/azure-provider.tf
    variables: terraform/subsume-masb-mssql-db/mssql-db-variables.tf
    main: terraform/subsume-masb-mssql-db/mssql-db-main.tf
    data: terraform/subsume-masb-mssql-db/mssql-db-data.tf
```

Will result is a terraform workspace with the following structure:
* outputs.tf gets contents of *terraform/subsume-masb-mssql-db/mssql-db-outputs.tf*
* provider.tf gets contents of *terraform/subsume-masb-mssql-db/azure-provider.tf*
* variables.tf gets contents of *terraform/subsume-masb-mssql-db/mssql-db-variables.tf*
* main.tf gets contents of *terraform/subsume-masb-mssql-db/mssql-db-main.tf*
* data.tf gets contents of *terraform/subsume-masb-mssql-db/mssql-db-data.tf*

> If there are [import inputs](#import-input-object), a `tf import` will be run for each import input value before `tf apply` is run. Once all the import calls are complete, `tf show` is run to generate a new *main.tf*. So it is important not to put anything into *main.tf* that needs to be preserved. Put them in one of the other tf files.
> 
#### Variable object

The variable object describes a particular input or output variable. The
structure is turned into a JSONSchema to validate the inputs or outputs.
Outputs are _only_ validated on integration tests.

| Field           | Type | Description                                                                                                                                                                                                                                                                                                                                                                                           |
|-----------------| --- |-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| required        | boolean | Should the user request fail if this variable isn't provided?                                                                                                                                                                                                                                                                                                                                         |
| field_name*     | string | The name of the JSON field this variable serializes/deserializes to.                                                                                                                                                                                                                                                                                                                                  |
| type*           | string | The JSON type of the field. This MUST be a valid JSONSchema type excepting `null`.                                                                                                                                                                                                                                                                                                                    |
| details*        | string | Provides explanation about the purpose of the variable.                                                                                                                                                                                                                                                                                                                                               |
| default         | any | The default value for this field. If `null`, the field MUST be marked as required. If a string, it will be executed as a HIL expression and cast to the appropriate type described in the `type` field. See the [Expression language reference](#expression-language-reference) section for more information about what's available.                                                                  |
| enum            | map of any:string | Valid values for the field and their human-readable descriptions suitable for displaying in a drop-down list.                                                                                                                                                                                                                                                                                         |
| constraints     | map of string:any | Holds additional JSONSchema validation for the field. Feature flag `enable-catalog-schemas` controls whether to serve Json schemas in catalog. The following keys are supported: `examples`, `const`, `multipleOf`, `minimum`, `maximum`, `exclusiveMaximum`, `exclusiveMinimum`, `maxLength`, `minLength`, `pattern`, `maxItems`, `minItems`, `maxProperties`, `minProperties`, and `propertyNames`. |
| tf_attribute    | string | The tf resource attribute from which the value of this field can be extracted from (e.g. `azurerm_mssql_database.azure_sql_db.name`). To be specified for subsume use cases only.                                                                                                                                                                                                                     |
| prohibit_update | boolean | Defines if the field value can be updated on update operation.                                                                                                                                                                                                                                                                                                                                        |

#### Computed Variable Object

Computed variables allow you to evaluate arbitrary HIL expressions against
variables or metadata about the provision or bind call.

| Field | Type | Description |
| --- | --- | --- |
| name* | string | The name of the variable. |
| default* | any | The value to set the variable to. If it's a string, it will be evaluated by the expression engine and cast to the provided type afterwards. See the "Expression language reference" section for more information about what's available. |
| overwrite | boolean | If a variable already exists with the same name, should this one replace it? |
| type | string | The JSON type of the field it will be cast to if evaluated as an expression. If defined, this MUST be a valid JSONSchema type excepting `null`. |

### Example

```yaml
version: 1
name: example-service
id: 00000000-0000-0000-0000-000000000000
description: a longer service description
display_name: Example Service
image_url: https://example.com/icon.jpg
documentation_url: https://example.com
support_url: https://example.com/support.html
tags: [gcp, example, service]
plans:
- name: example-email-plan
  id: 00000000-0000-0000-0000-000000000001
  description: Builds emails for example.com.
  display_name: example.com email builder
  bullets:
  - information point 1
  - information point 2
  - some caveat here
  properties:
    domain: example.com
provision:
  plan_inputs:
  - required: true
    field_name: domain
    type: string
    details: The domain name
  user_inputs:
  - required: true
    field_name: username
    type: string
    details: The username to create
    tf_attribute: resourceType.resourceName.user
  computed_inputs: []
  template: |-
    variable domain {type = string}
    variable username {type = string}
    output email {value = "${var.username}@${var.domain}"}

  outputs:
  - required: true
    field_name: email
    type: string
    details: The combined email address
bind:
  plan_inputs: []
  user_inputs: []
  computed_inputs:
  - name: address
    default: ${instance.details["email"]}
    overwrite: true
  template: |-
    resource "random_string" "password" {
      length = 16
      special = true
      override_special = "/@\" "
    }

    output uri {value = "smtp://${var.address}:${random_string.password.result}@smtp.mycompany.com"}

  outputs:
  - required: true
    field_name: uri
    type: string
    details: The uri to use to connect to this service
examples:
- name: Example
  description: Examples are used for documenting your service AND as integration tests.
  plan_id: 00000000-0000-0000-0000-000000000001
  provision_params:
    username: my-account
  bind_params: {}

```

## Expression language reference

The broker uses the [HIL expression language](https://github.com/hashicorp/hil) with a limited set of built-in functions.

### Functions

The following string interpolation functions are available for use:

* `assert(condition_bool, message_string) -> bool`
  * If the condition is false, then an error will be raised to the user containing `message_string`.
  * Avoid using this function. Instead, try to make it so your users can't get into a bad state to begin with. See the "design guidelines" section.  
  * In the words of [PEP-20](https://www.python.org/dev/peps/pep-0020/):
    * If the implementation is hard to explain, it's a bad idea.
    * If the implementation is easy to explain, it may be a good idea.
* `time.nano() -> string`
  * This function returns the current time as a Unix time, the number of nanoseconds elapsed since January 1, 1970 UTC, as a decimal string.
  * The result is undefined if the Unix time in nanoseconds cannot be represented by an int64 (a date before the year 1678 or after 2262).
* `regexp.matches(regex_string, string) -> bool`
  * Checks if the string matches the given regex.
* `str.truncate(count, string) -> string`
  * Trims the given string to be at most `count` characters long.
  * If the string is already shorter, nothing is changed.
* `counter.next() -> int`
  * Provides a counter that increments once per call within the same call context.
  * The counter is reset on restart of the application.
* `rand.base64(count) -> string`
  * Generates `count` bytes of cryptographically secure randomness and converts it to [URL Encoded Base64](https://tools.ietf.org/html/rfc4648).
  * The randomness makes it suitable for using as passwords.
* `json.marshal(type) -> string`
  * Returns a JSON marshaled string of the given type.
* `map.flatten(keyValueSeparator, tupleSeparator, map)`
  * Converts a map into a string with each key/value pair separated by `keyValueSeparator` and each entry separated by `tupleSeparator`.
  * The output is deterministic.
  * Example: if `labels = {"key1":"val1", "key2":"val2"}` then `map.flatten(":", ";", labels)` produces `key1:val1;key2:val2`.
* `env("ENV_VAR_NAME")`
  * Returns value for environment variable `ENV_VAR_NAME`
* `config("config.key")`
  * Returns value for config key `config.key`. These will come from the config file or be mapped from environment variables by the *env_config_mapping* section of the root *manifest.yml*.

### Variables

The broker makes additional variables available to be used during provision and bind calls.

#### Resolution

The order of combining all plan properties before invoking Terraform is as follows:
1. Operator default variables loaded from the environment.
   * `GSB_PROVISION_DEFAULTS` values first and then `GSB_SERVICE_*SERVICE_NAME*_PROVISION_DEFAULTS`. 
1. User defined variables provided during provision/bind call.
   * The variable constraints are defined in `service_definitions.provision.user_inputs` or `service_definitions.bind.user_inputs` 
   * These values overwrite the above values if set.
1. (If the operation is an update) User defined variables provided during update call.
   * The variable constraints are defined in `service_definitions.provision.user_inputs` or `service_definitions.bind.user_inputs`
   * These values overwrite the above values if set.
1. Default variables and values as configured in `service_definitions.plans[0].provision_overrides`. 
   * These values overwrite the above values if set.
1. Default properties and values for any variables that is defined in `service_definitions.provision.user_input` and the user has not provided. 
   * These values do not override user defined params if they are already set by step 2 or 3.
1. Constant properties and values for any field that is defined in `service_definitions.plans[0].properties`. 
   * These values overwrite the above values if set.
1. Computed fields that are defined in `service_definitions.provision.computed_inputs`.
   * The service definition for `computed_inputs` specifies if these values will overwrite previous steps.

#### Provision/Deprovision

* `request.service_id` - _string_ The GUID of the requested service.
* `request.plan_id` - _string_ The ID of the requested plan. Plan IDs are unique within an instance.
* `request.instance_id` - _string_ The ID of the requested instance. Instance IDs are unique within a service.
* `request.default_labels` - _map[string]string_ A map of labels that should be applied to the created infrastructure for billing/accounting/tracking purposes. See [billing documentation](billing.md) for full details about mapping process including special handling of characters in Osb request. 
   * `request.default_labels.pcf-organization-guid` - _string_ Mapped from [cloudfoundry context](https://github.com/openservicebrokerapi/servicebroker/blob/master/profile.md#cloud-foundry-context-object) `organization_guid`
   * `request.default_labels.pcf-space-guid` - _string_ Mapped from [cloudfoundry context](https://github.com/openservicebrokerapi/servicebroker/blob/master/profile.md#cloud-foundry-context-object) `space_guid`
   * `request.default_labels.pcf-instance-id` - _string_ Mapped from the ID of the requested instance. 
* `request.context` - _map[string]any_ Mapped from [cloudfoundry context](https://github.com/openservicebrokerapi/servicebroker/blob/master/profile.md#cloud-foundry-context-object).
* `request.x_broker_api_originating_identity` - _map[string]any_ Mapped from [cloudfoundry `x_broker_api_originating_identity` header](https://github.com/openservicebrokerapi/servicebroker/blob/master/profile.md#originating-identity-header)
   
#### Bind/Unbind

* `request.binding_id` - _string_ The ID of the new binding.
* `request.instance_id` - _string_ The ID of the existing instance to bind to.
* `request.service_id` - _string_ The GUID of the service this binding is for.
* `request.plan_id` - _string_ The ID of plan the instance was created with.
* `request.plan_properties` - _map[string]string_ A map of properties set in the service's plan.
* `request.app_guid` - _string_ The ID of the application this binding is for.
* `instance.name` - _string_ The name of the instance.
* `instance.details` - _map[string]any_ Output variables of the instance as specified by ProvisionOutputVariables.

## File format

The brokerpak itself is a zip file with the extension `.brokerpak`.
In the root is the `manifest.yml` file, which will specify the version of the pak.

There are three directories in the pak's root:

* `src/` an unstructured directory that holds source code for the bundled binaries, this is for complying with 3rd party licenses.
* `bin/` contains binaries under `bin/{os}/{arch}` sub-directories for each supported platform.
* `definitions/` contain the service definition YAML files.

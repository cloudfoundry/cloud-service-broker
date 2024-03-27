# OpenTofu upgrades

This document will explain how a OpenTofu upgrade can be implemented and triggered.

> Note: OpenTofu replaced Terraform in the CSB starting with version 1.0.0.
> There may still be some references to Terraform in the codebase.

### Broker Configuration

In order for OpenTofu upgrades to be carried over by the broker, this capability needs to be turned on.
To do so, enable the relevant flags as documented in [Feature Flags Configuration](configuration.md#feature-flags-configuration)
and restage the app. 

The brokerpak manifest must also include a `terraform_upgrade_path` section and the new version of the OpenTofu binary. See below.

### Brokerpak specification

Each brokerpak can instruct the broker to upgrade the OpenTofu version independently.
The manifest needs to be updated to include a `terraform_upgrade_path` section and the new version of the OpenTofu binary.
Optionally, the `terraform_state_provider_replacements` has to be provided. This is specially needed for the upgrade from 
older brokerpak versions.
For more information on these manifest sections, see [brokerpak-specification](brokerpak-specification.md#manifest-yaml-file) and [OpenTofu Upgrade Path object](brokerpak-specification.md#OpenTofu-upgrade-Path-object)

> **Note:** OpenTofu does not recommend making OpenTofu language changes at the same time that performing an upgrade.
> Hence, ideally these changes should be included in a separate release of your brokerpak and all existing instances should be upgraded before installing a subsequent release.

#### Example Brokerpak manifest additions
```
...
terraform_upgrade_path:
- version: 0.12.30
- version: 0.13.7
terraform_binaries:
- name: terraform
  version: 0.12.30
  source: https://github.com/hashicorp/terraform/archive/v0.12.30.zip  
- name: terraform
  version: 0.13.7
  source: https://github.com/hashicorp/terraform/archive/v0.13.7.zip
  default: true
terraform_state_provider_replacements:
  registry.terraform.io/-/aws: "registry.terraform.io/hashicorp/aws"
...
```
> **Note:** `terraform_upgrade_path`s must be in ascending order and one entry in the `terraform_binaries` list must be marked `default: true`.

### Triggering an upgrade

Upgrades are not performed automatically. For an upgrade to be initiated for a service instance, a request to `update` the instance without any parameters must be made or a `cf upgrade-service <instance_name>` has to be executed.

When an upgrade is available for an instance, all other operations are blocked until the upgrade is performed.
Update, Bind and Unbind will return an error message indicating that an upgrade operation needs to be performed first.

Service instance upgrades will also upgrade existing bindings to those instances.

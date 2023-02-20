# Terraform upgrades

This document will explain how a terraform upgrade can be implemented and triggered.

### Broker Configuration

In order for terraform upgrades to be carried over by the broker, this capability needs to be turned on.
To do so, enable the relevant flags as documented in [Feature Flags Configuration](configuration.md#feature-flags-configuration)
and restage the app. 

The broker manifest must also include a `terraform_upgrade_path` section and the new version of the terraform binary. See below.

### Brokerpak specification

Each brokerpak can instruct the broker to upgrade the terraform version independently.
The manifest needs to be updated to include a `terraform_upgrade_path` section and the new version of the terraform binary.
Optionally, the `terraform_state_provider_replacements` has to be provided. This is specially needed for the upgrade from 
terraform 0.12 to 0.13 and it is otherwise not needed.
For more information on these manifest sections, see [brokerpak-specification](brokerpak-specification.md#manifest-yaml-file) and [Terraform Upgrade Path object](brokerpak-specification.md#terraform-upgrade-Path-object)

> :warning: **Note:** Upgrade is only supported for Terraform versions >= 0.12.0.

> **Note:** Terraform does not recommend making HCL changes at the same time that performing a terraform upgrade (see [docs](https://www.terraform.io/language/upgrade-guides/0-13#before-you-upgrade)). Hence ideally these changes should be included in a separate release of your brokerpak and all existing instances should be upgraded before installing a subsequent release.

#### Example Brokerpak manifest additions
```
...
terraform_upgrade_path:
- version: 0.13.7
- version: 0.12.30
terraform_binaries:
- name: terraform
  version: 0.12.30
  source: https://github.com/hashicorp/terraform/archive/v0.12.30.zip  
- name: terraform
  version: 0.13.7
  source: https://github.com/hashicorp/terraform/archive/v0.13.7.zip
terraform_state_provider_replacements:
  registry.terraform.io/-/aws: "registry.terraform.io/hashicorp/aws"
...
```

### Triggering an upgrade

Upgrades are not performed automatically. For an upgrade to be initiated for a service instance, a request to `update` the instance without any parameters must be made or a `cf upgrade-service <instance_name>` has to be executed.

When an upgrade is available for an instance, all other operations are blocked until the upgrade is performed.
Update, Bind and Unbind will return an error message indicating that an upgrade operation needs to be performed first.

Service instance upgrades will also upgrade existing bindings to those instances.

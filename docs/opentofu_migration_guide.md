# Upgrading to OpenTofu

OpenTofu replaced Terraform in the CSB starting with version 1.0.0. This guide will explain how to migrate from a Terraform based brokerpak to use OpenTofu.

> Important: OpenTofu is a drop in replacement for terraform versions 1.5.x. As such, there
> is no guarantee of migrations from earlier/later versions will work correctly. Upgrade to 1.5.x 
> terraform version should be performed and all existing instances upgraded before upgrading to CSB 1.0.0+.
> See the [upgrade guide](terraform-upgrades.md)

## Broker Configuration

In order for OpenTofu upgrades to be carried over by the broker, this capability needs to be turned on.
To do so, ensure the feature flags `TERRAFORM_UPGRADES_ENABLED` and `BROKERPAK_UPDATES_ENABLED` are set to `true` and
the application has been restaged to ensure the values have been updated in the broker.


## Brokerpak Configuration

### Specifying OpenTofu cli (a.k.a tofu)
  
  The manifest needs to be updated to include a `terraform_upgrade_path` section and the new version of the OpenTofu binary (`tofu`).
Specifying `url_template` is optional, it will default for the OpenTofu download url if not specified. However, we recommend you specify it to reduce the impact of future broker changes.
As there is only one `tofu` version the `default` property can be ignored. It is included in this example for completeness

#### Example tofu configuration
```yaml
...
terraform_upgrade_path:
- version: 1.6.0
terraform_binaries:
- name: tofu
  version: 1.6.0
  source: https://github.com/opentofu/opentofu/archive/refs/tags/v1.6.0.zip
  url_template: https://github.com/opentofu/opentofu/releases/download/v${version}/tofu_${version}_${os}_${arch}.zip
  default: true  
...
```

### Specifying Providers

Resources under `terraform_binaries` other than `tofu` are most frequently providers. The broker version 2.0.0 and onwards currently defaults to 
`registry.terraform.io` for providers. If you wish to change to the OpenTofu registry instead, then the `url_template` pointing to
the correct registry domain needs to be specified. In any case, it is recommended to always be explicit about the registry in use for clarity.

> Warning: There is currently a mismatch between the way that Hashicorp
> and OpenTofu make the providers available to download. The difference
> in structure of the artifacts published make it impossible for the
> broker to download the OpenTofu published providers at the moment. 
> To be able to use OpenTofu published providers it is currently necessary to provide
> a custom download location (using the url_template property) where artifacts are published
> in the same format as in `https://releases.hashicorp.com/`. That is, a zip containing
> only a binary file which includes provider name, version and architecture in its name.

For broker version 1.x.x, the broker would incorrectly default to OpenTofu registry causing providers in the state to be referenced as coming from OpenTofu unless otherwise specified in the yaml file. It is recommended you skip the 1.x versions altogether, otherwise, follow the above recommendation of explicitely specifying the registry domain in all providers

The following example demonstrates the two options.

#### Example providers configuration
```yaml
...
terraform_binaries:
- name: tofu
  ...
- name: terraform-provider-google
  version: 1.19.0
  provider: registry.terraform.io/hashicorp/random
  source: https://github.com/terraform-providers/terraform-provider-google/archive/v1.19.0.zip
  url_template: https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip   
- name: terraform-provider-random
  version: v3.6.0
  provider: registry.opentofu.org/hashicorp/random
  source: https://github.com/opentofu/terraform-provider-random/archive/v3.6.0.zip
  url_template: <custom-download-location>
...
```

Whenever the Terraform registry is in use, the domain needs to be included in the `required_providers` in the HCL files, as this is resolved directly by OpenTofu. It is recommended that you always include the `required_providers` block, specifying all providers in use with their fully qualified name.

#### Example providers block

For the above manifest configuration, the following block needs to be specified. Notice the domain `registry.terraform.io` is included in the `google` provider block and the `registry.opentofu.org` in the `random` provider block. The `opentofu` registry domain can be ignored as OpenTofu defaults to that registry, however
it is recommended to fully qualify the providers to avoid confusion and errors.

```hcl
terraform {
  required_providers {
    google = {
      source  = "registry.terraform.io/hashicorp/google"
      version = ">= 1.18"
    }
    random = {
      source  = "registry.opentofu.org/hashicorp/random"
      version = ">= 3.3.2"
    }
  }
}
```

### State changes

State stored by the broker (`terraform.tfstate`) contains fully qualified names of all the providers, including registry domain. This means that a provider name of `terraform-provider-random` will appear as `registry.terraform.io/terraform-provider-random` in the state. If you are changing from the `Terraform` (`registry.terraform.io`) registry to the `OpenTofu` (`registry.opentofu.org`) one by specifying the OpenTofu download urls in the `url_template` in the manifest for a provider, then state for existing instances needs to be ammended. This can be achieved by adding a `terraform_state_provider_replacements` [block in the manifest](brokerpak-specification.md#manifest-yaml-file).

> Note: for higher versions of OpenTofu, OpenTofu handles the renaming themselves. However due to to some issues identified already in their implementation, we recommend to explicitely specify this renames.

> Warning: As mentioned before in this guide, a custom donwload location must be currently specified in the `url_template` property for downloading OpenTofu providers as current published format in OpenTofu doesn't match Hashicorp's format.

#### Example providers change in state

The following example specifies the `random` provider needs to be fetched from
the opentofu registry instead of the `terraform` one. If there are existing
instances already created with a previous version of the brokerpak, then
the provider will need to be renamed in their state. 

```yaml
...
terraform_binaries:
- name: tofu
  ...
- name: terraform-provider-google
  version: 1.19.0
  provider: registry.terraform.io/hashicorp/random
  source: https://github.com/terraform-providers/terraform-provider-google/archive/v1.19.0.zip
  url_template: https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip   
- name: terraform-provider-random
  version: v3.6.0
  provider: registry.opentofu.org/hashicorp/random
  source: https://github.com/opentofu/terraform-provider-random/archive/v3.6.0.zip
  url_template: <custom-download-location>
terraform_state_provider_replacements:
  registry.terraform.io/hashicorp/random: "registry.opentofu.org/hashicorp/random"
...
```

 > Note: The provider replacement step is executed before doing a `tofu init`
 Currently, if the `hcl` for a binding o provision hcl is empty (there's no hcl specified in the offering yaml file for binding or provision) or specified in-line, this operation will fail. To overcome this issue, specify the hcl as `.tf` files in the `terraform` folder and use the `template_refs` property to specify the templates (not `template_ref` nor `template`). If the hcl for an operation is a noop, then a workaround is specifying `template_refs` anyway and just include a file with a comment (see [example](https://github.com/cloudfoundry/csb-brokerpak-aws/blob/b0d567fba8c447e8a7d4bfc18982f8ba4f77c0ea/aws-redis.yml#L324)) 

## Triggering an upgrade

Upgrades are not performed automatically. For an upgrade to be initiated for a service instance, a request to `update` the instance without any parameters must be made or a `cf upgrade-service <instance_name>` has to be executed.

When an upgrade is available for an instance, all other operations are blocked until the upgrade is performed.
Update, Bind and Unbind will return an error message indicating that an upgrade operation needs to be performed first.

Service instance upgrades will also upgrade existing bindings to those instances.

## Release notes for next release:

### Breaking changes
- The "config migrate-env" subcommand was removed as it appeared not to work and has assumed zero usage

### Features:

- Terraform lifecycle meta-argument `prevent_destroy` is now supported to protect resources during a service update. The
  property is unset during a deprovision.
- A tutorial on authoring brokerpaks has been added.
- Terraform Upgrades (feature flagged)
    - Maintenance info is set for every plan. The version is set to the same version as the default Terraform version.
    - Update endpoint can perform upgrades when the correct maintenance info information is passed and no other changes
      are requested. This will upgrade the instance and all related bindings.
    - Update, bind, unbind and delete operations are blocked if an upgrade has not happened first.
    - Terraform modules can also be upgraded
- The `tf list` subcommand now prints the version of Terraform for each workspace state
- A new "purge" subcommand can be used to remove a service instance from the database
- brokerpaktestframework: extra folders needed for brokerpak build are now supported.

### Fixes:
- Broker checks the database deployment workspace readability aat startup before attempting encryption or removing salt.
- Brokerpaks no longer include superfluous source code, but if needed it can be including by adding the --include-source
  option when building
- Terraform Upgrades (feature flagged)
    - Upgrade are no longer performed automatically when update or delete is called on an instance.
- Broker fails to delete service when create is in progress.
- brokerpaktestframework.TerraformMock.ReturnTFState() has been superseded by to SetTFState(). The original method works
  but is deprecated. The goal of this change is to be more precise in terms of the functionality of the method.
- brokerpaktestframework.TestInstance.BrokerUrl() has been superseded by BrokerURL(). The original method works but is
  deprecated. This is to match the Go pseudo-standard on
  initialisms:  https://github.com/golang/go/wiki/CodeReviewComments#initialisms
- brokerpaktestframework.TestInstance.BrokerUrl() has been superseded by BrokerURL(). The original method works but is deprecated. This is to match the Go pseudo-standard on initialisms:  https://github.com/golang/go/wiki/CodeReviewComments#initialisms
- Checks the database deployment workspace readability before attempting encryption or removing salt
- Brokerpaks no longer include superfluous source code, but if needed it can be including by adding the --include-source option when building
- brokerpaktestframework.TerraformMock.ReturnTFState() has been superseded by to SetTFState(). The original method works but is deprecated. The goal of this change is to be more precise in terms of the functionality of the method.
- Broker fails to delete service when create is in progress.
- Fixes concatenation of words in Terraform error messages
- Fixes an issue where a service instance update was incorrectly classified as invalid
- brokerpaktestfrawork now handles "terraform_upgrade_path", as part of this change the signature for `FindServicePlanGUIDs()` was changed to return an error as errors were previously silent



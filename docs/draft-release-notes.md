## Release notes for next release:

### Features:

- Terraform lifecycle meta-argument `prevent_destroy` is now supported to protect resources during a service update. The
  property is unset during a deprovision.
- A tutorial on authoring brokerpaks has been added.
- Terraform Upgrades (feature flagged)
    - Maintenance info is set for every plan. The version is set to the same version as the default Terraform version.
    - Update endpoint can perform upgrades when the correct maintenance info information is passed and no other changes
      are requested.
    - Update, bind, unbind and delete operations are blocked if an upgrade has not happened first.

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



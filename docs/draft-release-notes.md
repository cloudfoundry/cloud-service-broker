## Release notes for next release:


### Features:
- Terraform lifecycle meta-argument `prevent_destroy` is now supported to protect resources during a service update. The property is unset during a deprovision.

### Fixes:
- brokerpaktestframework.TestInstance.BrokerUrl() has been superseded by BrokerURL(). The original method works but is deprecated. This is to match the Go pseudo-standard on initialisms:  https://github.com/golang/go/wiki/CodeReviewComments#initialisms
- Checks the database deployment workspace readability before attempting encryption or removing salt
- Brokerpaks no longer include superfluous source code, but if needed it can be including by adding the --include-source option when building
- brokerpaktestframework.TerraformMock.ReturnTFState() has been superseded by to SetTFState(). The original method works but is deprecated. The goal of this change is to be more precise in terms of the functionality of the method.

### Notes:
- A tutorial on authoring brokerpaks has been added

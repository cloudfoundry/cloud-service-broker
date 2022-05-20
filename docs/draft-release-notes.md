## Release notes for next release:


### Features:


### Fixes:
- brokerpaktestframework.TestInstance.BrokerUrl() has been superseded by BrokerURL(). The original method works but is deprecated. This is to match the Go pseudo-standard on initialisms:  https://github.com/golang/go/wiki/CodeReviewComments#initialisms
- Checks the database deployment workspace readability before attempting encryption or removing salt

### Chores:
- brokerpaktestframework.TerraformMock.ReturnTFState() has been superseded by to SetTFState(). The original method works but is deprecated. The goal of this change is to be more precise in terms of the functionality of the method.
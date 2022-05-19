## Release notes for next release:


### Features:


### Fixes:
- brokerpaktestframework.TestInstance.BrokerUrl() has been superseded by BrokerURL(). The original method works but is deprecated. This is to match the Go pseudo-standard on initialisms:  https://github.com/golang/go/wiki/CodeReviewComments#initialisms
- Checks the database deployment workspace readability before attempting encryption or removing salt
- Brokerpaks no longer include superfluous source code, but if needed it can be including by adding the --include-source option when building

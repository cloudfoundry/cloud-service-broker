## Release notes for next release:

### Features:
- Allow usage of non-Hashicorp providers.
- `beta` tagged services are disabled by default. These can be enabled by supplying the environment variable: `GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=true`
- brokerpaks can rename terraform providers in the instance state by specifying renames in the manifest using `terraform_state_provider_replacements`


### Fixes:
- If multiple versions of Terraform are specified, the nominated default version must be highest version
- Fixed missing insert in message when building brokerpak from current working directory
- Error messages during encryption tell you how to fix the issue
- Error messages during encryption log DB row ID
- Checks the database for readability of all fields before attempting encryption or removing salt
- Better logging when setting up database connection leading to improved debug experience


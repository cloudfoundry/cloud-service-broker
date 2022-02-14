## Release notes for next release:

### New feature:
- Brokerpaks service provision properties can define `tf_attribute` field to be copied from the imported resources. 
This is necessary for subsume operations when brokerpak update is enabled, since the subsume operation does not store the required HCL variables. 
- Allow `terraform_upgrade_path` to be specified in manifest

### Fix:
- When a service has no plans, there is now an error on broker startup.
- A 30s timeout has been added for connecting to the broker state database

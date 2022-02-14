## Release notes for next release:

### New feature:
- Brokerpaks service provision properties can define `tf_attribute` field to be copied from the imported resources. 
This is necessary for subsume operations when brokerpak update is enabled, since the subsume operation does not store the required HCL variables. 

### Fix:


## Release notes for next release:

### New feature:
* Provision and update operations return errors when attempting to override plan defined properties. Previously these commands would accept the property change request, but override it with the plan defined value.

### Fix:
* Bind parameters are now stored on bind and used during unbind operations. This means that brokerpaks that define bind parameters without any default value can succedd during unbind operations. 

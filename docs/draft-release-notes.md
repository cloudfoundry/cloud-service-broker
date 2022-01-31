## Release notes for next release:

### New feature:
* Provision and update operations return errors when attempting to override plan defined properties. Previously these commands would accept the property change request, but override it with the plan defined value.

### Fix:
* Bind parameters are now stored on bind and used during unbind operations. This means that brokerpaks that define bind parameters without any default value can succedd during unbind operations.
* When additional JSON parameters are supplied as part of a `cf create-service`, `cf update-service`, `cf bind-service` or `cf create-service-key` they are now validated against the list of supported parameters; the operation will fail with an error if a parameter is unknown.

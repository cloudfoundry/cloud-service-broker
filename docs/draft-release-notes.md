## Release notes for next release:

# Features
- introduced the ability for typed input fields to optionally be set to `null`.
- **pak build now has --target flag** which allows the build to target a specific platform/architecture or "current".
- **pak build now has --compress flag** which allows the compression of the brokerpak to be controlled.
- **provider_display_name is now configurable:** the `provider_display_name` can now be configured in the brokerpak services.
   and it will be returned by the catalog endpoint making it available for display in visual tools. 

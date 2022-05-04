## Release notes for next release:

### Features:
- new "/info" endpoint for diagnostics reports broker version and uptime
- allow users to configure the TLS `skip-verify` option when using custom certificates
- "*.pem" files get stored alongside Terraform state - experimental and subject to change
- Feature flagged: Terraform is upgraded for a service instance when `cf update-service` is invoked, or `cf delete-service`. The Terraform binding state is also updated when `cf delete-service-binding` is invoked. This is dependent on a Terraform upgrade path being specified, and the environment variable `TERRAFORM_UPGRADES_ENABLED=true` is set.
### Fixes:
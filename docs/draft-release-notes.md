## Release notes for next release:

### Breaking Changes:
- TLS is now enforced by default on connection to the Database. This can be configured via the `DB_TLS` environment variable.

### Features:
- new "/info" endpoint for diagnostics reports broker version and uptime
- new environment variable `TLS_SKIP_VERIFY` can be configured to skip TLS verification on connection to the Database. 

### Fixes:



# Broker Configuration
The broker can be configured though environment variables or configuration files or a combo of both.

## Configuration File
A configuration file can be provided at run time to the broker.
```bash
cloud-service-broker server --config <config file name>
```

A configuration file can be YAML or JSON. Config file values that are `.` delimited represent hierarchy in the config file.

Example:
```
db:
  host: hostname
```
represents a config file value of `db.host`

## Database Configuration Properties

Connection details for the backing database for the service broker.

You can configure the following values:

| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| <tt>DB_HOST</tt> <b>*</b> | db.host | string | <p>Database host </p>|
| <tt>DB_USERNAME</tt> | db.user | string | <p>Database username </p>|
| <tt>DB_PASSWORD</tt> | db.password | secret | <p>Database password </p>|
| <tt>DB_PORT</tt> <b>*</b> | db.port | string | <p>Database port (defaults to 3306)  Default: <code>3306</code></p>|
| <tt>DB_NAME</tt> <b>*</b> | db.name | string | <p>Database name  Default: <code>servicebroker</code></p>|
| <tt>CA_CERT</tt> | db.ca.cert | text | <p>Server CA cert </p>|
| <tt>CLIENT_CERT</tt> | db.client.cert | text | <p>Client cert </p>|
| <tt>CLIENT_KEY</tt> | db.client.key | text | <p>Client key </p>|

## Broker Service Configuration

Broker service configuration values:
| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| <tt>SECURITY_USER_NAME</tt> <b>*</b> | api.user | string | <p>Broker authentication username</p>|
| <tt>SECURITY_USER_PASSWORD</tt> <b>*</b> | api.password | string | <p>Broker authentication password</p>|
| <tt>PORT</tt> | api.port | string | <p>Port to bind broker to</p>|

## Brokerpak Configuration

Brokerpak configuration values:
| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| <tt>GSB_BROKERPAK_BUILTIN_PATH</tt> | brokerpak.builtin.path | string | <p>Path to search for .brokerpak files, default: <code>./</code></p>|


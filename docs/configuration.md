# Broker Configuration
The broker can be configured though environment variables or configuration files or a combo of both.

## Configuration File
A configuration file can be provided at run time to the broker.
```bash
cloud-service-broker serve --config <config file name>
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

## Credhub Configuration
The broker supports passing credentials to apps via credhub references, thus keeping them private to the application (they won't show up in `cf env app_name` output.)

| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| CH_CRED_HUB_URL           |credhub.url    | URL | credhub service URL - usually `https://credhub.service.cf.internal:8844`|
| CH_UAA_URL                |credhub.uaa_url | URL | uaa service URL - usually `https://uaa.service.cf.internal:8443`|
| CH_UAA_CLIENT_NAME        |credhub.uaa_client_name| string | uaa username - usually `credhub_admin_client`|
| CH_UAA_CLIENT_SECRET      |credhub.uaa_client_secret| string | uaa client secret - "*Credhub Admin Client Credentials*" from *Operations Manager > PAS > Credentials* tab. |
| CH_SKIP_SSL_VALIDATION    |credhub.skip_ssl_validation| boolean | skip SSL validation if true | 
| CH_CA_CERT_FILE           |credhub.ca_cert_file| path | path to cert file |

## Brokerpak Configuration

Brokerpak configuration values:
| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| <tt>GSB_BROKERPAK_BUILTIN_PATH</tt> | brokerpak.builtin.path | string | <p>Path to search for .brokerpak files, default: <code>./</code></p>|
|<tt>GSB_BROKERPAK_CONFIG</tt>|brokerpak.config| string | JSON global config for broker pak services|
||service.*service-name*.provision.defaults| string | JSON provision defaults override for *service-name*|
||services.*service-name*.plans| string | JSON plan collection to augment plans for *service-name*|

### Global Config Example

Services for a given IaaS should have common parameter names for service wide platform resources (like regions)

Azure services support global region and resource group parameters:

```yaml
provision:
  defaults: '{
    "region": "eastus2", 
    "resource_group": "sb-acceptance-test-rg"
  }'
```

### Provision Default Example

The Azure MS SQL DB service (azure-mssql-db) provisions databases on an existing MS SQL server. Configuring the server credentials looks like this:
```yaml
service:
  azure-mssql-db:
    provision:
      defaults: '{
          "server_name": "vsb-azsql-svr-52539613-83bc-4f57-9ed8-8a98ebc394e5",
          "admin_username": "KlpWlZCYHEyqdwuf",
          "admin_password": "KZe-.-rTuhK2ucDCx5UYQJyjsbum65SlC8_LTZg~Klr.2.1Yut-1weBdF1Xk-uo.",
          "resource_group": "vsb-azsql-svr-52539613-83bc-4f57-9ed8-8a98ebc394e5"
        }' 
```

### Plans Example

The Azure MS SQL DB service (azure-mssql-db) can also have its plans augmented to support more than one existing DB server:
```yaml
service:
  azure-mssql-db:
    plans: '[
      {
        "id":"881de5d9-e078-44e7-bed5-26faadabda3c",
        "name":"small",
        "description":"2cores, 10GB storage DB on server vsb-azsql-test-db4",
        "pricing_tier":"GP",
        "cores":"2",
        "storage_gb":"10",
        "server_name":"vsb-azsql-test-db4",
        "admin_username":"eqVrU6vcTBvgfiqj",
        "admin_password":"BI@G9a9nCnXIV4CV",
        "resource_group":"vsb-azsql-test-db4"
      },
      {
        "id":"1a1de5d9-e078-44e7-bed5-266aadabdaa6",
        "name":"small",
        "description":"2cores, 10GB storage DB on server vsb-azsql-svr-52539613-83bc-4f57-9ed8-8a98ebc394e5",
        "pricing_tier":"GP",
        "cores":"2",
        "storage_gb":"10",
        "server_name": "vsb-azsql-svr-52539613-83bc-4f57-9ed8-8a98ebc394e5",
        "admin_username": "KlpWlZCYHEyqdwuf",
        "admin_password": "KZe-.-rTuhK2ucDCx5UYQJyjsbum65SlC8_LTZg~Klr.2.1Yut-1weBdF1Xk-uo.",
        "resource_group": "vsb-azsql-svr-52539613-83bc-4f57-9ed8-8a98ebc394e5"
      }
    ]'
```




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

| Environment Variable | Config File Value | Type    | Description                                                                                                                                   |
|----------------------|------|---------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| <tt>DB_HOST</tt> <b>*</b> | db.host | string  | <p>Database host </p>                                                                                                                         |
| <tt>DB_USERNAME</tt> | db.user | string  | <p>Database username </p>                                                                                                                     |
| <tt>DB_PASSWORD</tt> | db.password | secret  | <p>Database password </p>                                                                                                                     |
| <tt>DB_PORT</tt> <b>*</b> | db.port | string  | <p>Database port (defaults to 3306)  Default: <code>3306</code></p>                                                                           |
| <tt>DB_NAME</tt> <b>*</b> | db.name | string  | <p>Database name  Default: <code>servicebroker</code></p>                                                                                     |
| <tt>DB_TLS</tt> <b>*</b>      | db.tls   | string  | <p>Enforce TLS on connection to Database. Allowed values:<code>true</code>,<code>false</code>,<code>skip-verify</code>,<code>custom</code></p> |
| <tt>CUSTOM_CERT_TLS_SKIP_VERIFY</tt> <b>*</b>      | db.custom_certs.tls_skip_verify   | bool    | <p>Skip TLS verification when using custom certificates. Default: <code>true</code></p> |
| <tt>CA_CERT</tt> | db.ca.cert | text    | <p>Server CA cert </p>                                                                                                                        |
| <tt>CLIENT_CERT</tt> | db.client.cert | text    | <p>Client cert </p>                                                                                                                           |
| <tt>CLIENT_KEY</tt> | db.client.key | text    | <p>Client key </p>                                                                                                                            |
| <tt>ENCRYPTION_ENABLED</tt> | db.encryption.enabled | Boolean | <p>Enable encryption of sensitive data in the database </p>                                                                                   |
| <tt>ENCRYPTION_PASSWORDS</tt> | db.encryption.passwords | text    | <p>JSON collection of passwords </p>                                                                                                          |

Example:
```
db:
  host: hostname
  encryption:
    enabled: true
    passwords: "[{\"label\":\"first-password\",{\"password\":{\"secret\": \"veryStrongSecurePassword\"}},\"primary\": true}]"
```

Example Encryption Passwords JSON object: 
```
[
  {
    "label": "first-password",
    "password": {
      "secret": "veryStrongSecurePassword"
    },
    "primary": true
  }
]
```

### Enabling first time encryption
1. Set `encryption.enabled` to `true` and add a password to the collection of passwords and mark it as primary.
1. Restart the CSB app.

### Rotating encryption keys

1. Add a new password to the collection of passwords and mark it as primary. The previous primary password should still be provided and 
no longer marked as primary.
1. Restart the CSB app.
1. Once the app has successfully started, the old password(s) can be removed from the configuration.

### Disabling encryption (after it was enabled)
1. Set `encryption.enabled` to `false`. The previous primary password should still be provided and no longer marked as primary.
1. Restart the CSB app.
1. Once the app has successfully started, the old password(s) can be removed from the configuration.

## Broker Service Configuration

Broker service configuration values:
| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| <tt>SECURITY_USER_NAME</tt> <b>*</b> | api.user | string | <p>Broker authentication username</p>|
| <tt>SECURITY_USER_PASSWORD</tt> <b>*</b> | api.password | string | <p>Broker authentication password</p>|
| <tt>PORT</tt> | api.port | string | <p>Port to bind broker to</p>|
| <tt>TLS_CERT_CHAIN</tt> | api.certCaBundle | string | <p>File path to a pem encoded certificate chain</p>|
| <tt>TLS_PRIVATE_KEY</tt> | api.tlsKey | string | <p>File path to a pem encoded private key</p>|

## Feature flags Configuration

Feature flags can be toggled through the following configuration values. See also [source code occurences of "toggles.Features.Toggle"](https://github.com/cloudfoundry/cloud-service-broker/search?q=toggles.Features.Toggle&type=code)
| Environment Variable | Config File Value | Type | Description | Default |
|----------------------|------|-------------|------------------|----------|
| <tt>GSB_COMPATIBILITY_ENABLE_BUILTIN_BROKERPAKS</tt> <b>*</b> | compatibility.enable_builtin_brokerpaks | Boolean | <p>Load brokerpaks that are built-in to the software.</p>| "true" |
| <tt>GSB_COMPATIBILITY_ENABLE_CATALOG_SCHEMAS</tt> <b>*</b> | compatibility.enable_catalog_schemas | Boolean | <p>Enable generating JSONSchema for the service catalog.</p>| "false" |
| <tt>GSB_COMPATIBILITY_ENABLE_CF_SHARING</tt> <b>*</b> | compatibility.enable_cf_sharing | Boolean | <p>Set all services to have the Sharable flag so they can be shared</p>| "false" |
| <tt>GSB_COMPATIBILITY_ENABLE_EOL_SERVICES</tt> <b>*</b> | compatibility.enable_eol_services | Boolean | <p>Enable broker services that are end of life.</p>| "false" |
| <tt>GSB_COMPATIBILITY_ENABLE_BETA_SERVICES</tt> <b>*</b> | compatibility.enable_beta_services | Boolean | <p>Enable services that are in Beta. These have no SLA or support</p>| "false" |
| <tt>GSB_COMPATIBILITY_ENABLE_GCP_DEPRECATED_SERVICES</tt> <b>*</b> | compatibility.enable_gcp_deprecated_services | Boolean | <p>Enable services that use deprecated GCP components.</p>| "false" |
| <tt>GSB_COMPATIBILITY_ENABLE_PREVIEW_SERVICES</tt> <b>*</b> | compatibility.enable_preview_services | Boolean | <p>Enable services that are new to the broker this release.</p>| "true" |
| <tt>GSB_COMPATIBILITY_ENABLE_TERRAFORM_SERVICES</tt> <b>*</b> | compatibility.enable_terraform_services | Boolean | <p>Enable services that use the experimental, unstable, Terraform back-end.</p>| "false" |
| <tt>GSB_COMPATIBILITY_ENABLE_UNMAINTAINED_SERVICES</tt> <b>*</b> | compatibility.enable_unmaintained_services | Boolean | <p>Enable broker services that are unmaintained.</p>| "false" |
| <tt>TERRAFORM_UPGRADES_ENABLED</tt> <b>*</b> | brokerpak.terraform.upgrades.enabled | Boolean | <p>Enables terraform version upgrades when brokerpak specifies an upgrade path and an upgrade is requested for an instance.</p>| "false" |
| <tt>BROKERPAK_UPDATES_ENABLED</tt> <b>*</b> | brokerpak.updates.enabled | Boolean | <p>Enable update of HCL of existing instances on update. When false, any update will be executed with the same HCL the instance was created with. If true, updates will be executed with newest specification in the brokerpak.</p>| "false" |

## Credhub Configuration
The broker supports passing credentials to apps via [credhub references](https://github.com/cloudfoundry-incubator/credhub/blob/master/docs/secure-service-credentials.md#service-brokers), thus keeping them private to the application (they won't show up in `cf env app_name` output.)

| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| CH_CRED_HUB_URL           |credhub.url    | URL | credhub service URL - usually `https://credhub.service.cf.internal:8844`|
| CH_UAA_URL                |credhub.uaa_url | URL | uaa service URL - usually `https://uaa.service.cf.internal:8443`|
| CH_UAA_CLIENT_NAME        |credhub.uaa_client_name| string | uaa username - usually `credhub_admin_client`|
| CH_UAA_CLIENT_SECRET      |credhub.uaa_client_secret| string | uaa client secret - "*Credhub Admin Client Credentials*" from *Operations Manager > PAS > Credentials* tab. |
| CH_SKIP_SSL_VALIDATION    |credhub.skip_ssl_validation| boolean | skip SSL validation if true | 
| CH_CA_CERT_FILE           |credhub.ca_cert_file| path | path to cert file |

### Credhub Config Example (Azure) 
```
azure:
  subscription_id: your subscription id
  tenant_id: your tenant id
  client_id: your client id
  client_secret: your client secret
db:
  host: your mysql host
  password: your mysql password
  user: your mysql username
api:
  user: someusername
  password: somepassword
credhub:
  url: ...
  uaa_url: ...
  uaa_client_name: ...
  uaa_client_secret: ...
 ```

## Brokerpak Configuration

Brokerpak configuration values:
| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| <tt>GSB_BROKERPAK_BUILTIN_PATH</tt> | brokerpak.builtin.path | string | <p>Path to search for .brokerpak files, default: <code>./</code></p>|
|<tt>GSB_BROKERPAK_CONFIG</tt>|brokerpak.config| string | JSON global config for broker pak services|
|<tt>GSB_PROVISION_DEFAULTS</tt>|provision.defaults| string | JSON global provision defaults|
|<tt>GSB_SERVICE_*SERVICE_NAME*_PROVISION_DEFAULTS</tt>|service.*service-name*.provision.defaults| string | JSON provision defaults override for *service-name*|
|<tt>GSB_SERVICE_*SERVICE_NAME*_PLANS</tt>|service.*service-name*.plans| string | JSON plan collection to augment plans for *service-name*|



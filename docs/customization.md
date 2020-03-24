# Installation Customization

*NOTE this document left for here for reference. Current configuration information should be found [here](./configuration.md)*


This file documents the various environment variables you can set to change the functionality of the service broker.
If you are using the PCF Tile deployment, then you can manage all of these options through the operator forms.
If you are running your own, then you can set them in the application manifest of a PCF deployment, or in your pod configuration for Kubernetes.


## Root Service Account

Please paste in the contents of the json keyfile (un-encoded) for your service account with owner credentials.

You can configure the following environment variables:

| Environment Variable | Type | Description |
|----------------------|------|-------------|
| <tt>ROOT_SERVICE_ACCOUNT_JSON</tt> <b>*</b> | text | <p>Root Service Account JSON </p>|


## Database Properties

Connection details for the backing database for the service broker.

You can configure the following environment variables:

| Environment Variable | Type | Description |
|----------------------|------|-------------|
| <tt>DB_HOST</tt> <b>*</b> | string | <p>Database host </p>|
| <tt>DB_USERNAME</tt> | string | <p>Database username </p>|
| <tt>DB_PASSWORD</tt> | secret | <p>Database password </p>|
| <tt>DB_PORT</tt> <b>*</b> | string | <p>Database port (defaults to 3306)  Default: <code>3306</code></p>|
| <tt>DB_NAME</tt> <b>*</b> | string | <p>Database name  Default: <code>servicebroker</code></p>|
| <tt>CA_CERT</tt> | text | <p>Server CA cert </p>|
| <tt>CLIENT_CERT</tt> | text | <p>Client cert </p>|
| <tt>CLIENT_KEY</tt> | text | <p>Client key </p>|


## Brokerpaks

Brokerpaks are ways to extend the broker with custom services defined by Terraform templates.
A brokerpak is an archive comprised of a versioned Terraform binary and providers for one or more platform, a manifest, one or more service definitions, and source code.

You can configure the following environment variables:

| Environment Variable | Type | Description |
|----------------------|------|-------------|
| <tt>GSB_BROKERPAK_CONFIG</tt> <b>*</b> | text | <p>Global Brokerpak Configuration A JSON map of configuration key/value pairs for all brokerpaks. If a variable isn't found in the specific brokerpak's configuration it's looked up here. Default: <code>{}</code></p>|


## Feature Flags

Service broker feature flags.

You can configure the following environment variables:

| Environment Variable | Type | Description |
|----------------------|------|-------------|
| <tt>GSB_COMPATIBILITY_ENABLE_BUILTIN_BROKERPAKS</tt> <b>*</b> | boolean | <p>enable-builtin-brokerpaks Load brokerpaks that are built-in to the software. Default: <code>true</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_BUILTIN_SERVICES</tt> <b>*</b> | boolean | <p>enable-builtin-services Enable services that are built in to the broker i.e. not brokerpaks. Default: <code>true</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_CATALOG_SCHEMAS</tt> <b>*</b> | boolean | <p>enable-catalog-schemas Enable generating JSONSchema for the service catalog. Default: <code>false</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_CF_SHARING</tt> <b>*</b> | boolean | <p>enable-cf-sharing Set all services to have the Sharable flag so they can be shared across spaces in PCF. Default: <code>false</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_EOL_SERVICES</tt> <b>*</b> | boolean | <p>enable-eol-services Enable broker services that are end of life. Default: <code>false</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_GCP_BETA_SERVICES</tt> <b>*</b> | boolean | <p>enable-gcp-beta-services Enable services that are in GCP Beta. These have no SLA or support policy. Default: <code>true</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_GCP_DEPRECATED_SERVICES</tt> <b>*</b> | boolean | <p>enable-gcp-deprecated-services Enable services that use deprecated GCP components. Default: <code>false</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_PREVIEW_SERVICES</tt> <b>*</b> | boolean | <p>enable-preview-services Enable services that are new to the broker this release. Default: <code>true</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_TERRAFORM_SERVICES</tt> <b>*</b> | boolean | <p>enable-terraform-services Enable services that use the experimental, unstable, Terraform back-end. Default: <code>false</code></p>|
| <tt>GSB_COMPATIBILITY_ENABLE_UNMAINTAINED_SERVICES</tt> <b>*</b> | boolean | <p>enable-unmaintained-services Enable broker services that are unmaintained. Default: <code>false</code></p>|



## Custom Plans

You can specify custom plans for the following services.
The plans MUST be an array of flat JSON objects stored in their associated environment variable e.g. <code>[{...}, {...},...]</code>.
Each plan MUST have a unique UUID, if you modify the plan the UUID should stay the same to ensure previously provisioned services continue to work.
If you are using the PCF tile, it will generate the UUIDs for you.
DO NOT delete plans, instead you should change their labels to mark them as deprecated.

### Configure Brokerpaks

Configure Brokerpaks
To specify a custom plan manually, create the plan as JSON in a JSON array and store it in the environment variable: <tt>GSB_BROKERPAK_SOURCES</tt>.

For example:
<code>
[{"id":"00000000-0000-0000-0000-000000000000", "name": "custom-plan-1", "uri": setme, "service_prefix": setme, "excluded_services": setme, "config": setme, "notes": setme},...]
</code>

<table>
<tr>
  <th>JSON Property</th>
  <th>Type</th>
  <th>Label</th>
  <th>Details</th>
</tr>
<tr>
  <td><tt>id</tt></td>
  <td><i>string</i></td>
  <td>Plan UUID</td>
  <td>
    The UUID of the custom plan, use the <tt>uuidgen</tt> CLI command or [uuidgenerator.net](https://www.uuidgenerator.net/) to create one.
    <ul><li><b>Required</b></li></ul>
  </td>
</tr>
<tr>
  <td><tt>name</tt></td>
  <td><i>string</i></td>
  <td>Plan CLI Name</td>
  <td>
    The name of the custom plan used to provision it, must be lower-case, start with a letter a-z and contain only letters, numbers and dashes (-).
    <ul><li><b>Required</b></li></ul>
  </td>
</tr>


<tr>
  <td><tt>uri</tt></td>
  <td><i>string</i></td>
  <td>Brokerpak URI</td>
  <td>
  The URI to load. Supported protocols are http, https, gs, and git.
				Cloud Storage (gs) URIs follow the gs://<bucket>/<path> convention and will be read using the service broker service account.

				You can validate the checksum of any file on download by appending a checksum query parameter to the URI in the format type:value.
				Valid checksum types are md5, sha1, sha256 and sha512. e.g. gs://foo/bar.brokerpak?checksum=md5:3063a2c62e82ef8614eee6745a7b6b59


<ul>
  <li><b>Required</b></li>
</ul>


  </td>
</tr>

<tr>
  <td><tt>service_prefix</tt></td>
  <td><i>string</i></td>
  <td>Service Prefix</td>
  <td>
  A prefix to prepend to every service name. This will be exact, so you may want to include a trailing dash.


<ul>
  <li><i>Optional</i></li>
</ul>


  </td>
</tr>

<tr>
  <td><tt>excluded_services</tt></td>
  <td><i>text</i></td>
  <td>Excluded Services</td>
  <td>
  A list of UUIDs of services to exclude, one per line.


<ul>
  <li><i>Optional</i></li>
</ul>


  </td>
</tr>

<tr>
  <td><tt>config</tt></td>
  <td><i>text</i></td>
  <td>Brokerpak Configuration</td>
  <td>
  A JSON map of configuration key/value pairs for the brokerpak. If a variable isn't found here, it's looked up in the global config.


<ul>
  <li><b>Required</b></li>
  <li>Default: <code>{}</code></li>
</ul>


  </td>
</tr>

<tr>
  <td><tt>notes</tt></td>
  <td><i>text</i></td>
  <td>Notes</td>
  <td>
  A place for your notes, not used by the broker.


<ul>
  <li><i>Optional</i></li>
</ul>


  </td>
</tr>

</table>



---------------------------------------

_Note: **Do not edit this file**, it was auto-generated by running <code>cloud-service-broker generate customization</code>. If you find an error, change the source code in <tt>customization-md.go</tt> or file a bug._

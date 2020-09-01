# GCP Stack Driver Trace Config
## Applies to service *csb-google-stackdriver-trace*

*csb-google-stackdriver-trace* manages stack driver trace account credentials on GCP (not currently supported on AWS or Azure.)

## Plans

| Plan | Description |
|------|-------------|
| default | default service account to use stack driver trace service|

## Config parameters

The following are bind time parameters, they may be set with:

```bash
cf bind-service <app name> <service instance> -c '{"<param name>":"<param value>", ...}'
```


| Parameter | Type | Description | Default |
|-----------|------|------|---------|
| name| string | Name of service account to create | csb-*binding_id* | credentials | string | GCP service account JSON string | config file value `gcp.credentials` |
| credentials | string | GCP service account JSON string | config file value `gcp.project` |
| role | string | GCP role for service account | `cloudtrace.agent` |

## Binding Credentials

The binding credentials for the stackdriver trace have the following shape:

```json
{
    "Name": "service instance name",
    "Email": "email address for service accunt",
    "UniqueId": "service instance unique id",
    "PrivateKeyData": "base64 encoded service account key json",
    "ProjectId": "servcie account project id"
}
```


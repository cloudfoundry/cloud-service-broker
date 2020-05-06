# K8S Manifests

*Work in progress, currently only tested with minikube*

Uses kustomize to inject configuration and secrets.

## Generate Config with Credentials

Start by generating a config file with IaaS credentials.

### For AWS:
Required environment variables
| Variable | Value |
|----------|-------|
| AWS_ACCESS_KEY_ID | access key id |
| AWS_SECRET_ACCESS_KEY | secret access key |

```bash
./build-config.sh aws
```
### For Azure:
Required environment variables
| Variable | Value |
|----------|-------|
|ARM_TENANT_ID|ID for tenant that resources will be created in|
|ARM_SUBSCRIPTION_ID|ID for subscription that resources will be created in|
|ARM_CLIENT_ID|service principal client ID|
|ARM_CLIENT_SECRET|service principal secret|

```bash
./build-config.sh azure
```
### For GCP:
Required environment variables
| Variable | Value |
|----------|-------|
|GOOGLE_CREDENTIALS| gcp service account json |
|GOOGLE_PROJECT| gcp project |

```bash
./build-config.sh gcp
```

The resulting config file - `./config-files/broker-config.yaml` - can be modified as needed (see [docs](../docs/configuration.md) for further configuration options.)

## Deploy

Deploy the broker:

```bash
kubectl apply -k ./
```

A mysql pod and a broker pod should be deployed, `kubectl get pods` should look something like:

```
NAME                         READY   STATUS    RESTARTS   AGE
csb-6df5cf46db-skln4         1/1     Running   3          110s
csb-mysql-7fff9c5697-8f4x8   1/1     Running   0          110s
```

### Run Client Example Tests

**NOTE: when using minikube, make sure `minikube tunnel` is running.**

To run the *cloud-service-broker* example tests:

```bash
./run-client-examples.sh
```


# Installing the broker on AWS

The broker service and the AWS brokerpak can be pushed and registered on a foundation running on AWS.

Documentation for broker configuration can be found [here](./configuration.md).

## Requirements

### CloudFoundry running on AWS.
The AWS brokerpak services are provisioned with firewall rules that only allow internal connectivity. This allows `cf push`ed applications access, while denying any public access.

### AWS Service Credentials
The services need to be provisioned in the same AWS account that the foundation is running in. To do this, the broker needs the following service principal credentials to manage resources within that account:
- access key id
- secret access key

#### Required IAM Policies
The AWS account represented by the access key needs the following permission policies:
```json
{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Action": [
                    "s3:CreateBucket",
                    "s3:DeleteBucket",
                    "s3:PutBucketAcl",
                    "s3:PutBucketLogging",
                    "s3:PutBucketPolicy",
                    "s3:PutBucketTagging",
                    "s3:GetObject",
                    "s3:ListBucket",
                    "iam:CreateAccessKey",
                    "iam:CreateUser",
                    "iam:GetUser",
                    "iam:DeleteAccessKey",
                    "iam:DeleteUser",
                    "iam:DeleteUserPolicy",
                    "iam:ListAccessKeys",
                    "iam:ListAttachedUserPolicies",
                    "iam:ListUserPolicies",
                    "iam:ListPolicies",
                    "iam:PutUserPolicy",
                    "iam:GetPolicy",
                    "iam:GetAccountAuthorizationDetails",
                    "rds:CreateDBCluster",
                    "rds:CreateDBInstance",
                    "rds:DeleteDBCluster",
                    "rds:DeleteDBInstance",
                    "rds:DescribeDBClusters",
                    "rds:DescribeDBInstances",
                    "rds:DescribeDBSnapshots",
                    "rds:DeleteDBSnapshot",
                    "rds:CreateDBParameterGroup",
                    "rds:ModifyDBParameterGroup",
                    "rds:DeleteDBParameterGroup",
                    "dynamodb:ListTables",
                    "dynamodb:DeleteTable",
                    "dynamodb:DescribeTable",
                    "sqs:CreateQueue",
                    "sqs:DeleteQueue",
                    "ec2:DescribeVpcs",
                    "ec2:DescribeVpcAttribute",
                    "ec2:DescribeSubnets",
                    "ec2:CreateSecurityGroup",
                    "ec2:DescribeSecurityGroups",
                    "ec2:DescribeNetworkInterfaces",
                    "ec2:DeleteSecurityGroup",
                    "ec2:RevokeSecurityGroupEgress",
                    "ec2:AuthorizeSecurityGroupIngress",
                    "rds:CreateDBSubnetGroup",
                    "rds:DescribeDBSubnetGroups",
                    "rds:ListTagsForResource",
                    "rds:DeleteDBSubnetGroup",
                    "rds:AddTagsToResource",
                    "ec2:RevokeSecurityGroupIngress",
                ],
                "Effect": "Allow",
                "Resource": "*"
            }
        ]
    }
```

### MySQL Database for Broker State
The broker keeps service instance and binding information in a MySQL database. 

#### Binding a MySQL Database
If there is an existing broker in the foundation that can provision a MySQL instance use `cf create-service` to create a new MySQL instance. Then use `cf bind-service` to bind that instance to the service broker.

#### Manually Provisioning a MySQL Database
If a MySQL instance needs to be manually provisioned, it must be accessible to applications running within the foundation so that the `cf push`ed broker can access it. The following configuration parameters will be needed:
- `DB_HOST`
- `DB_USERNAME`
- `DB_PASSWORD`

It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`).

## Step By Step From a Pre-build Release with a Bound MySQL Instance

Fetch a pre-built broker and brokerpak and bind it to a `cf create-service` managed MySQL.

### Requirements

The following tools are needed on your workstation:
- [cf cli](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html)

### Assumptions

The `cf` CLI has been used to authenticate with a foundation (`cf api` and `cf login`,) and an org and space have been targeted (`cf target`)

### Fetch A Broker and AWS Brokerpak

Download a release from https://github.com/pivotal/cloud-service-broker/releases. Find the latest release matching the name pattern `sb-0.1.0-rc.XXX-aws-0.0.1-rc.YY`. This will have a broker and brokerpak that have been tested together. Follow the hyperlink into that release and download `cloud-servic-broker` and `aws-services-0.1.0-rc.YY.brokerpak` into the same directory on your workstation.

### Create a MySQL instance with AWS broker
The following command will create a basic MySQL database instance named `csb-sql`
```bash
cf create-service aws-rds-mysql basic csb-sql
```
### Build Config File
To avoid putting any sensitive information in environment variables, a config file can be used.

Create a file named `config.yml` in the same directory the broker and brokerpak have been downloaded to. Its contents should be:

```yaml
aws:
  access_key_id: your access key id
  secret_access_key: your secret access key

api:
  user: someusername
  password: somepassword
```

### Push and Register the Broker

Push the broker as a binary application:

```bash
SECURITY_USER_NAME=someusername
SECURITY_USER_PASSWORD=somepassword
APP_NAME=cloud-service-broker

chmod +x cloud-service-broker
cf push "${APP_NAME}" -c './cloud-service-broker serve --config config.yml' -b binary_buildpack --random-route --no-start
```

Bind the MySQL database and start the service broker:
```bash
cf bind-service cloud-service-broker csb-sql
cf start "${APP_NAME}"
```
Register the service broker:
```bash
BROKER_NAME=csb-$USER

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
```
Once this completes, the output from `cf marketplace` should include:
```
csb-aws-mysql    small, medium, large            Amazon RDS for MySQL
```

## Step By Step From a Pre-built Release with a Manually Provisioned MySQL Instance

Fetch a pre-built broker and brokerpak and configure with a manually provisioned MySQL instance.

Requirements and assumptions are the same as above. Follow instructions above to [fetch the broker and brokerpak](#Fetch-A-Broker-and-AWS-Brokerpak)

### Create a MySQL Database
Its an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access. The database connection values (hostname, user name and password) will be needed in the next step. It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`).

### Build Config File
To avoid putting any sensitive information in environment variables, a config file can be used.

Create a file named `config.yml` in the same directory the broker and brokerpak have been downloaded to. Its contents should be:

```yaml
aws:
  access_key_id: your access key id
  secret_access_key: your secret access key

db:
  host: your mysql host
  password: your mysql password
  user: your mysql username

api:
  user: someusername
  password: somepassword
```

### Push and Register the Broker

Push the broker as a binary application:

```bash
SECURITY_USER_NAME=someusername
SECURITY_USER_PASSWORD=somepassword
APP_NAME=cloud-service-broker

chmod +x cloud-service-broker
cf push "${APP_NAME}" -c './cloud-service-broker serve --config config.yml' -b binary_buildpack --random-route
```

Register the service broker:
```bash
BROKER_NAME=csb-$USER

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
```

Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Step By Step From Source with Bound MySQL
Grab the source code, build and deploy.

### Requirements

The following tools are needed on your workstation:
- [go 1.14](https://golang.org/dl/)
- make
- [cf cli](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html)

The Pivotal AWS Service Broker must be installed in your foundation.

### Assumptions

The `cf` CLI has been used to authenticate with a foundation (`cf api` and `cf login`,) and an org and space have been targeted (`cf target`)

### Clone the Repo

The following commands will clone the service broker repository and cd into the resulting directory.
```bash
git clone https://github.com/pivotal/"${APP_NAME}".git
cd "${APP_NAME}"
```
### Set Required Environment Variables

Collect the AWS service credentials for your account and set them as environment variables:
```bash
export AWS_SECRET_ACCESS_KEY=your secret access key
export AWS_ACCESS_KEY_ID=your access key id

```
Generate username and password for the broker - Cloud Foundry will use these credentials to authenticate API calls to the service broker.
```bash
export SECURITY_USER_NAME=someusername
export SECURITY_USER_PASSWORD=somepassword
```
### Create a MySQL instance

The following command will create a basic MySQL database instance named `csb-sql`
```bash
cf create-service aws-rds-mysql basic csb-sql
```
### Use the Makefile to Deploy the Broker
There is a make target that will build the broker and brokerpak and deploy to and register with Cloud Foundry as a space scoped broker. This will be local and private to the org and space your `cf` CLI is targeting.
```bash
make push-broker-aws
```
Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Step By Step Slightly Harder Way

Requirements and assumptions are the same as above. Follow instructions for the first two steps above ([Clone the Repo](#Clone-the-Repo) and [Set Required Environment Variables](Set-Required-Environment-Variables))

### Create a MySQL Database
Its an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access. It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`). Set the following environment variables with information about that MySQL instance:
```bash
export DB_HOST=mysql server host
export DB_USERNAME=mysql server username
export DB_PASSWORD=mysql server password
```

### Build the Broker and Brokerpak
Use the makefile to build the broker executable and brokerpak.
```bash
make build-aws-brokerpak
```
### Pushing the Broker
All the steps to push and register the broker:
```bash
APP_NAME=cloud-service-broker

cf push --no-start

cf set-env "${APP_NAME}" SECURITY_USER_PASSWORD "${SECURITY_USER_PASSWORD}"
cf set-env "${APP_NAME}" SECURITY_USER_NAME "${SECURITY_USER_NAME}"

cf set-env "${APP_NAME}" AWS_ACCESS_KEY_ID "${AWS_ACCESS_KEY_ID}"
cf set-env "${APP_NAME}" AWS_SECRET_ACCESS_KEY "${AWS_SECRET_ACCESS_KEY}"

cf set-env "${APP_NAME}" DB_HOST "${DB_HOST}"
cf set-env "${APP_NAME}" DB_USERNAME "${DB_USERNAME}"
cf set-env "${APP_NAME}" DB_PASSWORD "${DB_PASSWORD}"

cf set-env "${APP_NAME}" GSB_BROKERPAK_BUILTIN_PATH ./AWS-brokerpak

cf start "${APP_NAME}"

BROKER_NAME=csb-$USER

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
```
Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Uninstalling the Broker
First, make sure there are all service instances created with `cf create-service` have been destroyed with `cf delete-service` otherwise removing the broker will fail.

### Unregister the Broker
```bash
cf delete-service-broker csb-$USER
```

### Uninstall the Broker
```bash
cf delete cloud-service-broker
```



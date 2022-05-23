# Writing Your First Brokerpak

In this tutorial, we are going to write a Brokerpak to expose a MySQL database on Amazon Web Services (“AWS”).
Without a Brokerpak, the Cloud Service Broker doesn’t know what MySQL or AWS are. A Brokerpak tells the Cloud
Service Broker what to do: it’s what makes the Cloud Service Broker useful.

The Cloud Service Broker should arguably be called the Terraform Broker. It is an opinionated way of running
Terraform to create things in the cloud. To write a Brokerpak, you therefore need to have a working knowledge
of Terraform.

## Provisioning and Binding

In CloudFoundry, services have a lifecycle. This is specified in the Open Service Broker API. There are two phases:
1. **Provisioning and Deprovisioning**. These correspond to the `cf create-service` and `cf delete-service`
   commands. Typically they involve the creation and deletion of the service. In this tutorial, provisioning
   will create a MySQL database and deprovisioning will delete the database. From a Terraform perspective
   they will be a `terraform apply` and a `terraform destroy` operation.
1. **Binding and Unbinding**. These correspond to the `cf bind-service` and `cf unbind-service` commands.
   They also correspond to the `cf create-service-key`, `cf delete-service-key` commands, because service
   keys and service bindings are very similar from the perspective of OSBAPI. Typically they involve
   the creation of an account with credentials to access the service. In this tutorial, they will
   correspond to creating and deleting an account in the MySQL database. From a Terraform perspective
   they are also `terraform apply` and `terraform destroy` operations.

## Developing the Terraform
In order to get going quickly, here are the Terraform files that we are going to use in this tutorial.

Firstly, `main.tf` will define the key resources that we are going to create. In this case that’s a
MySQL database, and the networking required to access it.
```hcl
resource "aws_security_group" "rds_sg" {
  name   = format("%s-sg", var.instance_name)
  vpc_id = var.aws_vpc_id
}

resource "aws_db_subnet_group" "rds_private_subnet" {
  name       = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnets.all.ids
}

resource "aws_security_group_rule" "rds_inbound_access" {
  from_port         = 3306
  protocol          = "tcp"
  security_group_id = aws_security_group.rds_sg.id
  to_port           = 3306
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "random_string" "username" {
  length  = 16
  special = false
  number  = false
}

resource "random_password" "password" {
  length           = 40
  special          = false
  override_special = "~_-."
}

resource "aws_db_instance" "default" {
  identifier             = var.instance_name
  db_name                = "main"
  engine                 = "mysql"
  username               = random_string.username.result
  password               = random_password.password.result
  db_subnet_group_name   = aws_db_subnet_group.rds_private_subnet.name
  vpc_security_group_ids = [aws_security_group.rds_sg.id]
  allocated_storage      = 10
  instance_class         = "db.t3.micro"
  apply_immediately      = true
  skip_final_snapshot    = true
}
```

You’ll notice that we refer to data, so we should define them too in `data.tf`:
```hcl
data "aws_subnets" "all" {
  filter {
    name   = "vpc-id"
    values = [var.aws_vpc_id]
  }
}
```

We also need to define the Terraform provider versions in use in `versions.tf`:
```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.13.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.1.3"
    }
  }
}
```

And configure the providers in `providers.tf`:
```hcl
provider "aws" {
  region     = "us-west-2"
  access_key = var.access_key
  secret_key = var.secret_key
}
```

And finally in `variables.tf` we define all the variables that we have used. It’s important to
parameterise the Terraform like this so that the values can be injected by the CSB.
Anything related to a secret, a name that could clash, or a configuration of Cloud Foundry
should be parameterized.
```hcl
variable "access_key" {
  type      = string
  sensitive = true
}

variable "secret_key" {
  type      = string
  sensitive = true
}

variable "instance_name" {
  type = string
}

variable "aws_vpc_id" {
  type = string
}
```

## Creating a Brokerpak
Create a directory. You may want to initialize the directory as a git repository. In that directory, create
a “terraform” subdirectory with a “provision” subdirectory. Place all the Terraform files above into that
subdirectory so that it looks like this:
```console
$ tree
.
└── terraform
    └── provision
        ├── data.tf
        ├── main.tf
        ├── providers.tf
        ├── variables.tf
        └── versions.tf

2 directories, 5 files

```

We are now ready to build these Terraform files into a Brokerpak. In order to do this we will need a manifest
file and a service definition file. The CSB can create examples of these files by running the
`cloud-service-broker pak init` command. For this tutorial, use the examples below:

First we will define the manifest. The snippet below defines the Terraform version, and also the versions
of the Terraform providers. These must match the versions in `versions.tf`. It also defines the mapping
between environment variables and config keys.
```yaml
packversion: 1
name: mysql-brokerpak
version: 1.0.0
platforms:
  - os: linux
    arch: amd64
terraform_binaries:
  - name: terraform
    version: 1.1.9
    source: 'https://github.com/hashicorp/terraform/archive/v1.1.9.zip'
  - name: terraform-provider-aws
    version: 4.13.0
    source: 'https://github.com/terraform-providers/terraform-provider-aws/archive/v4.13.0.zip'
  - name: terraform-provider-random
    version: 3.1.3
    source: 'https://github.com/terraform-providers/terraform-provider-random/archive/v3.1.3.zip'
service_definitions:
  - aws-mysql-tutorial.yml
env_config_mapping:
  AWS_ACCESS_KEY_ID: aws.access_key_id
  AWS_SECRET_ACCESS_KEY: aws.secret_access_key
  AWS_VPC_ID: aws.vpc_id
```
This YAML should be placed into the file `manifest.yml` at the top level of the Brokerpak.

Services are also defined in the Brokerpak in YAML. Many services can be defined in a single Brokerpak.
Whilst the service itself must be defined in the Brokerpak, the plans for that service, can either be
defined in the service definition, or later as an environment variable.

When defining a service and plans, the service IDs and plan IDs must be unique; ideally a UUID.

This file also defines the mapping between parameters and Terraform variables. Parameters can be set
in the plan, or defined by the user. They can also be templated: you can see in the below example that
the instance name is templated from the Cloud Foundry instance ID. This makes it easier to identify
which resource in AWS corresponds to which Cloud Foundry instance.
```yaml
version: 1
name: aws-mysql-tutorial
id: 30cc6f3e-d52d-11ec-8808-27cc122bc6dc
description: MySQL on AWS created as part of a tutorial
display_name: MySQL on AWS (Tutorial)
documentation_url: 'https://example.com'
image_url: 'https://example.com'
support_url: 'https://example.com'
tags: ['aws', 'example', 'mysql']
plans:
  - name: basic
    id: 606b8cc0-d52d-11ec-a67e-d7e5728aaf2a
    display_name: basic
    description: Basic plan
provision:
  user_inputs:
    - field_name: instance_name
      type: string
      details: Name for your mysql instance
      default: 'csb-mysql-${request.instance_id}'
      constraints:
        maxLength: 98
        minLength: 6
        pattern: '^[a-z][a-z0-9-]+$'
      prohibit_update: true
    - field_name: access_key
      type: string
      details: AWS access key
      default: '${config("aws.access_key_id")}'
    - field_name: secret_key
      type: string
      details: AWS secret key
      default: '${config("aws.secret_access_key")}'
    - field_name: aws_vpc_id
      details: VPC ID for instance
      default: '${config("aws.vpc_id")}'
  template_refs:
    providers: terraform/provision/providers.tf
    versions: terraform/provision/versions.tf
    main: terraform/provision/main.tf
    data: terraform/provision/data.tf
    variables: terraform/provision/variables.tf
```

This YAML should be placed into the file `aws-mysql-tutotial.yml` at the top level of the Brokerpak.

Finally we need the Cloud Service Broker itself which can be downloaded from
[the releases page](https://github.com/cloudfoundry/cloud-service-broker/releases).
For the purpose of this tutorial, download the linux binary and place it in the Brokerpak directory
renamed to `cloud-service-broker`. 

In order to be able to push the Cloud Service Broker as an app, we need a Cloud Foundry manifest:
```yaml
applications:
- name: cloud-service-broker-tutorial
  command: ./cloud-service-broker serve
  memory: 1G
  buildpacks:
  - binary_buildpack
  random-route: true
  env:
    DB_HOST: <DB hostname>
    DB_USERNAME: <DB username>
    DB_PASSWORD: <DB password>
    DB_NAME: <DB name>
    DB_TLS: "skip-verify"
    AWS_VPC_ID: <VPC ID for Cloud Foundry (or blank)>
    AWS_SECRET_ACCESS_KEY: <AWS secret access key>
    AWS_ACCESS_KEY_ID: <AWS secret access ID>
    SECURITY_USER_NAME: <username for the broker>
    SECURITY_USER_PASSWORD: <password for the broker>
```

This YAML should be placed into the file `cf-manifest.yml` at the top level of the Brokerpak. 

You will need to provide credentials for a MySQL database that the CSB can store state in, credentials
for AWS and a username and password for the broker itself.

Your directory should now look like:
```console
$ tree
.
├── aws-mysql-tutorial.yml
├── cf-manifest.yml
├── cloud-service-broker
├── manifest.yml
└── terraform
    └── provision
        ├── data.tf
        ├── main.tf
        ├── providers.tf
        ├── variables.tf
        └── versions.tf

2 directories, 9 files
```

To test out the Brokerpak:

1. Build the Brokerpak: `cloud-service-broker pak build`
   Note that to build the Brokerpak you will need a CSB binary that works on your system (e.g Mac),
   which may be different to the CSB binary that you push to Cloud Foundry (Linux).
   
1. Push the broker: `cf push -f cf-manifest.yml`

1. Register the broker: `cf create-service-broker mybroker <username> <password> https://<app URL>`

1. Make the services available: `cf enable-service-access aws-mysql-tutorial`

1. Check that you can see the new service: `cf marketplace`

1. Create a service: `cf create-service aws-mysql-tutorial basic mydb`

After this you will be able to see the new MySQL database in the AWS console.
To confirm which one it is, you can get the GUID of the service instance
(`cf service mydb --guid`). Once you are satisfied, delete the service instance:
```shell
cf delete-service mydb -f
```

## Adding the Ability to Bind
A database is only useful if you can connect to it. We saw in the Terraform files above that our MySQL is
created with a username and password. We could expose this username and password. For some resources,
this may be the right solution. But because this is an administrator user, it may give excessive
privilege to users. A better solution would be to use this admin user to create non-admin users
for each binding. When a binding is created, a new username and password is created. When the binding
is deleted, the associated username and password are revoked. And because the binding user is not
an administrator, it is not able to create other users. This gives us good control over access to
the database, and the admin user is managed within the Brokerpak and never exposed.

Firstly we must allow the admin username and password to be used by the binding phase.
To do this we declare outputs in  a new file `terraform/provision/outputs.tf`:
```hcl
output "hostname" { value = aws_db_instance.default.address }
output "username" { value = aws_db_instance.default.username }
output "password" {
  value     = aws_db_instance.default.password
  sensitive = true
}
output "status" {
  value = format("created db %s (id: %s) on server %s",
    aws_db_instance.default.db_name,
    aws_db_instance.default.id,
  aws_db_instance.default.address)
}
```

Then we need to create Terraform files for the binding. Create a new subdirectory under `terraform`
called `bind` and add the following files. Firstly, `main.tf` which generates credentials and creates the account.
```hcl
resource "random_string" "username" {
  length  = 16
  special = false
  number  = false
}

resource "random_password" "password" {
  length           = 64
  override_special = "~_-."
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "mysql_user" "binding_user" {
  user               = random_string.username.result
  plaintext_password = random_password.password.result
  host               = "%"
}

resource "mysql_grant" "binding_user" {
  user       = mysql_user.binding_user.user
  database   = "main"
  host       = "%"
  privileges = ["ALL"]
}
```

`versions.tf` defines the terraform provider versions. Note that we are using a new Terraform provider here.
The Hashicorp MySQL terraform provider is no longer maintained, and this is a fork with many fixed issues.
```hcl
terraform {
  required_providers {
    mysql = {
      source  = "petoju/mysql"
      version = "3.0.12"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.1.3"
    }
  }
}
```

`providers.tf` tells the MySQL provider how to connect to the MySQL instance.
```hcl
provider "mysql" {
  endpoint = format("%s:3306", var.hostname)
  username = var.admin_username
  password = var.admin_password
}
```

`variables.tf` defines the variables that have been used. The values will come from the outputs of the
provision operation.
```hcl
variable "admin_username" {
  type = string
}

variable "admin_password" {
  type      = string
  sensitive = true
}

variable "hostname" {
  type = string
}
```

And finally `outputs.tf` provides the data that an app will need to connect to the database.
```hcl
output "username" { value = random_string.username.result }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "hostname" { value = var.hostname }
output "name" { value = "main" }
output "port" { value = 3306 }
```

You should end up with a directory structure looking like:
```console
$ tree terraform/bind/
terraform/bind/
├── main.tf
├── outputs.tf
├── providers.tf
├── variables.tf
└── versions.tf

0 directories, 5 files
```

We will also need to modify the `manifest.yml` file to add the new provider.
Because it’s not a Hashicorp provider, this requires a bit more information.
Add the following into the `terraform_binaries` block:
```yaml
  - name: terraform-provider-mysql
    version: 3.0.12
    source: 'https://github.com/petoju/terraform-provider-mysql/archive/v3.0.12.zip'
    provider: petoju/mysql
    url_template: 'https://github.com/petoju/${name}/releases/download/v${version}/${name}_${version}_${os}_${arch}.zip'
```

We also need to add properties  to the `aws-mysql-tutorial.yml`, starting with the output from the
provision operation. Append the following to the existing `aws-mysql-tutorial.yml` file. The indentation
of the first added line should match the indentation of the last pre-existing line.
```yaml
    outputs: terraform/provision/outputs.tf
  outputs:
    - field_name: hostname
      type: string
      details: 'Hostname or IP address of the exposed mysql endpoint used by clients to connect to the service.'
    - field_name: username
      type: string
      details: The username to authenticate to the database instance.
    - field_name: password
      type: string
      details: The password to authenticate to the database instance.
bind:
  computed_inputs:
    - name: hostname
      type: string
      default: '${instance.details["hostname"]}'
      overwrite: true
    - name: admin_username
      type: string
      default: '${instance.details["username"]}'
      overwrite: true
    - name: admin_password
      type: string
      default: '${instance.details["password"]}'
      overwrite: true
  template_refs:
    versions: terraform/bind/versions.tf
    providers: terraform/bind/providers.tf
    main: terraform/bind/main.tf
    variables: terraform/bind/variables.tf
    outputs: terraform/bind/outputs.tf
  outputs:
    - field_name: username
      type: string
      details: The username to authenticate to the database instance
    - field_name: password
      type: string
      details: The password to authenticate to the database instance
    - field_name: hostname
      type: string
      details: The hostname of the database instance
    - field_name: name
      type: string
      details: The name of the database
    - field_name: port
      type: number
      details: The port of the database instance
```

Now it’s time to build and test it out:

1. Build the Brokerpak: `cloud-service-broker pak build`

1. Push the broker: `cf push -f cf-manifest.yml`

1. Create a service: `cf create-service aws-mysql-tutorial basic mydb`
   (note, the service broker should still be registered and the service access enabled from the previous steps)

1. Create a service key: `cf create-service-key mydb mykey`

1. Create a service key: `cf service-key mydb mykey`

This will show the credentials for connecting to the database. Because of the way that we have
created the database, you will not be able to connect from outside Cloud Foundry. But an app that
has been pushed will be able to connect.

Well done! You’ve now written your first Brokerpak. There are many other features of
Cloud Service Broker that allow more sophisticated Brokerpaks to be written.
Refer to the documentation to find out more.







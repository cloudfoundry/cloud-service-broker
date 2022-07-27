# Extending a service

In this tutorial, we are going to add a feature to an existing service in the [AWS Brokerpak](https://github.com/cloudfoundry/csb-brokerpak-aws) for the S3 service. Our goal is to be able to add a feature to the service definition, write the accompanying Terraform HCL, and add adequate testing so that a PR can be raised against the Brokerpak.

## Before you Start
It is recommended that you have first familiarised yourself with what is required to create a brokerpak in [this tutorial](../writing-your-first-brokerpak). 

In order to complete this tutorial, you will need:
- Access to a Cloud Foundry environment
- Access to AWS including a secret access ID and key
- A MySQL instance accessible from a Cloud Foundry app, to act as the state store
- A development environment with the ability to create files and run commands

## Understanding the existing service

Before thinking about adding a feature to a service, it's important to be able to understand a service definition and identify if the feature is already exposed. 

In our use case we're using the AWS S3 service and the feature we want to enable on our bucket is server-side encryption. 

First we navigate to the service definition file `aws-s3-bucket.yml`. The structure of the Brokerpak dictates that these files exist at the top level of the directory and are defined in YAML. 

Given that we're interested in what properties can be defined when provisioning a service we can navigate within the file to the `provision` block:

```
provision:
  user_inputs:
    - field_name: bucket_name
      type: string
      details: Name of bucket
      default: csb-${request.instance_id}
      plan_updateable: true
      prohibit_update: true
    - field_name: region
      type: string
      details: The region of AWS.
      default: us-west-2
      constraints:
        examples:
          - us-west-2
          - eu-west-1
        pattern: ^[a-z][a-z0-9-]+$
      prohibit_update: true
    - field_name: aws_access_key_id
      type: string
      details: AWS access key
      default: ${config("aws.access_key_id")}
    - field_name: aws_secret_access_key
      type: string
      details: AWS secret key
      default: ${config("aws.secret_access_key")}
```

Here we're able to see what properties are exposed in order for us to define them when provisioning a service. 
As we can see in this example, the only fields which will be accepted by the Cloud Service Broker (CSB), when provisioning an S3 service are: `bucket_name`, `region`, `aws_access_key_id`, `aws_secret_access_key`. 

We want the ability to toggle server side encryption, but there are no existing fields related to this property, so we're going to have to add one. 

## Identifying a feature you want to add

We have determined that there are no exposed properties to allow us to configure server side encryption. However, it may be possible that this field is actually defined in the Terraform HCL for this service, it just hasn't been exposed as a configurable property. 

To understand this we need to look at the `main.tf` Terraform HCL file for the S3 service. Within the CSB-Brokerpak-AWS this file can be found here:
```console
$ tree
.
└── terraform
    └── s3
        ├── bind
        └── provision
            ├── main.tf
            ├── outputs.tf
            ├── provider.tf
            └── variables.tf
```

Examining this file we see the following content:

```terraform/s3/provision/main.tf
...
resource "aws_s3_bucket" "b" {
  bucket = var.bucket_name

  tags = var.labels

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.b.id
  acl = var.acl
}

resource "aws_s3_bucket_versioning" "bucket_versioning" {
  bucket = aws_s3_bucket.b.id
  versioning_configuration {
    status = var.enable_versioning ? "Enabled" : "Disabled"
  }
}
```

To understand this file we need a bit more context on the Terraform provider being used, so we can understand how the service is being created. 

We can tell by looking at the `provider.tf` file that we are using the AWS Terraform provider. Looking at the [Terraform provider documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) will help us determine which fields configure server side encryption, which is the value we want to define. 

The [documentation shows](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_server_side_encryption_configuration) that in order to create an S3 resource with server side encryption, a resource of type `aws_s3_bucket_server_side_encryption_configuration` needs to be created. 
We can see in our existing `main.tf` file there is no reference to this resource, so we will need to add it in order to provision a bucket with this feature enabled. 

## Adding a property to the HCL
## Adding the field to the service definition
## Testing
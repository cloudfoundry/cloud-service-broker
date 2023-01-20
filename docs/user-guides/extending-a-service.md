# Extending a service

In this tutorial, we are going to add a feature to an existing service in the [AWS Brokerpak](https://github.com/cloudfoundry/csb-brokerpak-aws) for the S3 service. Our goal is to be able to add a feature to the service definition, write the accompanying Terraform HCL, and add adequate testing so that a PR can be raised against the Brokerpak.

## Before you Start
Before you start, you should be familiar with Brokerpak concepts, for example by reading the [writing your first brokerpak](./writing-your-first-brokerpak.md) tutorial. 

In order to complete this tutorial, you will need:
- Access to a Cloud Foundry environment
- Access to AWS including a secret access ID and key
- A MySQL instance accessible from a Cloud Foundry app, to act as the state store
- A development environment with the ability to create files and run commands

## Understanding the existing service

Before thinking about adding a feature to a service, it's important to be able to understand a service definition and identify if the feature is already exposed. 

In our use case we're using the AWS S3 service and the feature we want to enable on our bucket is for the requester to pay for requests to the bucket. 

First we navigate to the service definition file `aws-s3-bucket.yml`. By convention, these files exist at the top level of the Brokerpak directory and are defined in YAML. 

Given that we're interested in which properties can be defined when provisioning a service we can navigate within the file to the `provision` block:

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
As we can see in this snippet, the fields which will be accepted by the Cloud Service Broker (CSB), when provisioning
an S3 service are: `bucket_name`, `region`, `aws_access_key_id`, `aws_secret_access_key`.

We want the ability to configure the bucket so that the requester pays, but there are no existing fields related to this property, so we're going to have to add one. 

## Identifying a feature you want to add

We have determined that there are no exposed properties to allow us to configure the requester to pay. However, it may be possible that this field is actually defined in the Terraform HCL for this service, it just hasn't been exposed as a configurable property. 

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

Examining this file we see the bucket definition and definitions for other resources:

```terraform
resource "aws_s3_bucket" "b" {
  bucket = var.bucket_name

  tags = var.labels

  lifecycle {
    prevent_destroy = true
  }
}
```

To understand this file we need a bit more context on the Terraform provider being used, so we can understand how the service is being created. 

We can tell by looking at the `provider.tf` file that we are using the AWS Terraform provider. Looking at the [Terraform provider documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) will help us determine how to configure the bucket so that the requester pays.

The [documentation shows](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_request_payment_configuration)
that in order for the requester to pay, it's necessary to create a resource of type `aws_s3_bucket_request_payment_configuration`
with appropriate details. We can see in our existing `main.tf` file there is no reference to this
resource, so we will need to add it in order to provision a bucket with this feature enabled. 

## Adding a property to the HCL
Append the following to `main.tf`:
```terraform
resource "aws_s3_bucket_request_payment_configuration" "request_payer" {
  bucket = aws_s3_bucket.b.bucket
  payer  = var.request_payer
}
```
This adds the required `aws_s3_bucket_request_payment_configuration` resource and associates it
with the bucket. It allows us to pass in the `payer` field as a variable. This means that we can
maintain the default of `BucketOwner`, or set it to `Requester`. Also, if an additional value were
added in the future, then we could pass that in too without having to make modifications. An
alternative would have been to use a ternary operator such as:
```terraform
payer = var.request_payer ? "Requester" : "BucketOwner"
```
This approach is less flexible, but may be appropriate in situations where it's unlikely that
another value could ever be added.

Because we are using a variable, we also need to define it. Append this to `varaibles.tf`:
```terraform
variable "request_payer" { type = string }
```

## Adding the field to the service definition
All variables defined in HCL need an equivalent field in the service definition.
Insert the following into the `user_inputs` block, making sure that it's correctly aligned:
```yaml
    - field_name: request_payer
      type: string
      default: BucketOwner
      details: Who pays for requests to the bucket. One of `BucketOwner` or `Requester`.
```
This will allow the `request_payer` property to be set when creating or updating a bucket.
The default value is `BucketOwner` which aligns with the Terraform Provider default and the
default in the AWS console. Note that any string value can be provided. We could use an
`enum` to restrict the strings that can be provided. This is sometimes appropriate, but it
can make the service more rigid. For instance, if AWS ever enabled another potential value
then an `enum` would need to be updated, but a string field will just work with the new value.

## Testing

The aims of testing are to make sure that the new property works as expected, and to check
that other existing functionality has not been broken. It's not practical to test everything,
but the AWS brokerpak has some tests that are worth running and updating in order to improve confidence
in the change that we have made.

### Examples tests
Example tests are defined in the service definition file. In `aws-s3-bucket.yml` we see:
```yaml
examples:
- name: s3-default
  description: Default S3 Bucket
  plan_id: f64891b4-5021-4742-9871-dfe1a9051302
  provision_params: {}
  bind_params: {}
```
An example test will provision the service with the specified plan and parameters,
create a service binding, delete the service binding, and then deprovision the service.
By putting a service through the lifecycle it will detect major errors in the service
definition. Because running these tests creates artifacts in AWS and hence incurs a
cost, it's not practical to test every possible value for every input. Running the example
test will check that the Terraform HCL works, and that the default value for
`request_payer` also works.

To run the example tests, you should start the broker in one terminal with `make run`.
In another terminal you should run the test with
`make run-examples service_name=csb-aws-s3-bucket example_name=s3-default`.


### Integration tests
The tests are located in the `integration-tests/` directory. They test that values passed in to the
broker are correctly translated into Terraform variables. The tests are written in Go using
the [Ginkgo](https://github.com/onsi/ginkgo) test framework. It is possible to extend these
tests with minimal understanding of Go and Ginkgo. If we look at the S3 test file we see
something like:
```go
var _ = Describe("S3", Label("s3"), func() {
    ...
	Describe("provisioning", func() {
		It("should provision a plan", func() {
			instanceID, err := broker.Provision(s3ServiceName, customS3Plan["name"].(string), nil)
			...
				SatisfyAll(
					HaveKeyWithValue("bucket_name", "csb-"+instanceID),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
					HaveKeyWithValue("aws_access_key_id", awsAccessKeyID),
					HaveKeyWithValue("aws_secret_access_key", awsSecretAccessKey),
					...
```
Without understanding all the details of the test, we can see a point where we could add the new
property. In this case we could test that the default value is correctly set. You could add the
following into this section:
```go
HaveKeyWithValue("request_payer", "BucketOwner"),
```
This ensures that the Terraform variable takes the correct default value. You can examine other tests
in this file, and could add a test to make sure that the property can be set during provision, and
can be updated.

To run the integration tests, you can run `make run-integration-tests` in the top level of the brokerpak.

### Terraform tests

Terraform tests are `unit` tests that run with the system's **installed** Terraform binary on the Terraform files directly.
These tests will download the latest versions of the providers that comply with the restrictions in definition files, 
usually the `provider.tf` files. They won't look at the service offering definitions (.yml) or the manifest.yml files.

The tests are located in the `terraform-tests/` directory, and they aim to verify that given a set of variables inputs,
Terraform sends the expected values to the providers. The language used to write the tests is Go and 
[Ginkgo](https://github.com/onsi/ginkgo) the test framework.

If you look at the S3 test file we see something like this:
```go
package terraformtests

var _ = Describe("S3", Label("S3-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	requestPayer := "BucketOwner"
	bucketName := "csb-s3-test"
	defaultVars := map[string]any{
		"aws_access_key_id":      awsAccessKeyID,
		"aws_secret_access_key":  awsSecretAccessKey,
		"bucket_name":            bucketName,
		"region":                 "us-west-2",
		"request_payer":          requestPayer,
	}
	
	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "s3/provision") // where the Terraform files are
		Init(terraformProvisionDir)
	})
	
	Context("with default values", func() {
        BeforeAll(func() {
            plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
        })
    
        It("should create the right resources", func() {
            Expect(plan.ResourceChanges).To(HaveLen(6))
    
            Expect(ResourceChangesTypes(plan)).To(ConsistOf(
                "aws_s3_bucket",
                "aws_s3_bucket_acl",
                "aws_s3_bucket_versioning",
                "aws_s3_bucket_ownership_controls",
                "aws_s3_bucket_public_access_block",
                "aws_s3_bucket_request_payment_configuration", // A new resource was added
            ))
        })
    
        It("should create an S3 bucket with the correct properties", func() {
            Expect(AfterValuesForType(plan, "aws_s3_bucket")).To(
                MatchKeys(IgnoreExtras, Keys{
                    "bucket":              Equal("csb-s3-test"),
                    "object_lock_enabled": BeFalse(),
                    "tags": MatchAllKeys(Keys{
                        "k1": Equal("v1"),
                    }),
                }),
            )
        })

        // Check that the new resource will be created with the correct values!!
        It("should create an S3 bucket request payment configuration resource with the right values", func() {
            Expect(AfterValuesForType(plan, "aws_s3_bucket_request_payment_configuration")).To(
                MatchKeys(IgnoreExtras, Keys{
                    "bucket": Equal(bucketName),
                    "payer":  Equal(requestPayer),
                }))
        })
    })
})
```

Without understanding all the details of the test, you can see the dynamics of the tests. In this way it is easy to test
the logic applied in your Terraform code in which ternary operations or local variables are usually added. Thanks to
this type of test, you add an extra layer of security for future changes.

As you see in the next section, you can check that the new Terraform resource added will be created with the
expected values:

```go
package terraformtests

var _ = Describe("S3", Label("S3-terraform"), Ordered, func() {

	requestPayer := "BucketOwner"
	bucketName := "csb-s3-test"
	defaultVars := map[string]any{
		"aws_access_key_id":      awsAccessKeyID,
		"aws_secret_access_key":  awsSecretAccessKey,
		"bucket_name":            bucketName,
		"request_payer":          requestPayer,
	}

	// ...

	Context("with default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})
		// ...

		// Check that the new resource will be created with the correct values!!
		It("should create an S3 bucket request payment configuration resource with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_s3_bucket_request_payment_configuration")).To(
				MatchKeys(IgnoreExtras, Keys{
					"bucket": Equal(bucketName),
					"payer":  Equal(requestPayer),
				}))
		})
	})
})
```
You can examine other tests in this folder, and add your examples by analyzing the rest.

To run the Terraform tests, run `make run-terraform-tests` in the top level of the brokerpak, but before doing it,
you will need to set the environment variables to connect with the provider.

Take a look at the file `terraform_tests_suite_test.go`. You will find the mandatory configuration to run the tests.

```go
package terraformtests

...

var (
	...
	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsAccessKeyID     = os.Getenv("AWS_ACCESS_KEY_ID")
	...
)
```

Two environment variables are mandatory on AWS: `AWS_SECRET_ACCESS_KEY` and `AWS_ACCESS_KEY_ID`, both used to provide the credentials.
Therefore, the final command you will execute will look like this:

```shell
AWS_ACCESS_KEY_ID=<AWS-ACCESS-KEY-ID> AWS_SECRET_ACCESS_KEY=<AWS-SECRET-ACCESS-KEY> make run-terraform-tests
```
Obviously, you must replace the placeholder with the credentials associated with your account in AWS.
For more information see the following link:
[Terraform: Authentication and Configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication-and-configuration)

**Note:** Each provider needs its environment variables. To know what they are, you should see the configuration in
the following file `terraform_tests_suite_test.go`.
At the time of writing this text, the necessary variables categorized by the provider are as follows:
* AWS:
    * AWS_ACCESS_KEY_ID
    * AWS_SECRET_ACCESS_KEY
* AZURE:
    * ARM_CLIENT_ID
    * ARM_CLIENT_SECRET
    * ARM_SUBSCRIPTION_ID
    * ARM_TENANT_ID
* GCP:
    * GOOGLE_CREDENTIALS
    * GOOGLE_PROJECT

## Acceptance Tests
These test are located in the `acceptance_tests` directory. These are end-to-end tests which need to be run
in coordination with a CloudFoundry environment. They are also written in Go with the Ginkgo framework.
In general there is a single test for each service, and the test is not extended for every additional property.

To run the acceptance test, run `ginkgo -v --label-filter s3` in the `acceptance_tests`
directory. You will need to be logged into CloudFoundry. There is also an Upgrade acceptance
test which tests the upgrade from a previous brokerpak to the current version.
This is located in the `acceptance_tests/upgrade` directory, which contains a README
explaining how to run them.

### Completion
Once you are satisfied that the change is good, you can submit a Pull Request
against the AWS brokerpak in GitHub. The Cloud Service Broker team aim to merge
pull requests that are well authored and have the potential to be used by multiple
users.

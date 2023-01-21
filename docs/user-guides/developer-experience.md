## Cloud Service Broker Developer Experience

Improve your developer experience through examples and valuable tips that make your day-to-day easy.

### Prerequisites

* Familiarity with the command line.
* Install the latest [Go version](https://go.dev/).
* Install the shell extension [direnv](https://direnv.net/).
* Set environment variables. See [Necessary environment variables section](#necessary-environment-variables).


### Local Development Experience

This section will explain how to use the Cloud Service Broker, hereafter referred to as CSB, in the local environment.

#### Build CSB binary

The first step you need to do is to build the CSB.
Make sure you have an up-to-date version of Go installed, then run the following commands in your command line:

1. Clone the repository:
  ```shell
  {▸} ~/workspace/csb ✓ git clone git@github.com:cloudfoundry/cloud-service-broker.git
  ```
2. Install the binary

The next command will build the CSB, and install it in your local machine with the name `csb`.
  ```shell
  {▸} ~/workspace/csb/cloud-service-broker on main ✓ make install
  go build -o csb -ldflags "-X github.com/cloudfoundry/cloud-service-broker/utils.Version=v0.14.0-1b000966"
  mv csb /usr/local/bin/csb
  ```
If you do not have the latest version of Go, you will get an error similar to this:

  ```shell
  {▪} ~/workspace/csb/cloud-service-broker on main ✓ make install
  Go version does not match: expected: go1.19.5, got go1.18.2
  make: *** [Makefile:45: deps-go-binary] Error 1
  ```
After installing the CSB binary, you will be ready to advance to the next step.

##### Check the installation

Explore the available commands using the command `csb help`.

```shell
{▸} ~/workspace/csb/cloud-service-broker on main ✓ csb help              
An OSB compatible service broker for that leverages Terraform for service brokering

Usage:
  cloud-service-broker [flags]
  cloud-service-broker [command]

Available Commands:
  client             A CLI client for the service broker
  completion         Generate the autocompletion script for the specified shell
  config             Show system configuration
  create-service     EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: create a service instance
  create-service-key EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: create a service instance key
  delete-service     EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: delete a service instance
  delete-service-key EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: delete a service key
  help               Help about any command
  marketplace        EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list services and plans
  pak                interact with user-defined service definition bundles
  purge              purge a service instance from the database
  serve              Start the service broker
  serve-docs         Just serve the docs normally available on the broker
  service-keys       EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list service keys for a service instance
  services           EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list service instances
  tf                 Interact with the Terraform backend
  update-service     EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: update a service instance
  version            Show the version info of the broker

Flags:
      --config string   Configuration file to be read
  -h, --help            help for cloud-service-broker

Use "cloud-service-broker [command] --help" for more information about a command.

```

Or check the installed version by running `csb version`:
```shell
{▸} ~/workspace/csb/cloud-service-broker on main ✓ csb version
v0.14.0-1b000966
```

The `csb version` command will show you the installed version with the commit sha you got, which will help you to
identify if you have the latest changes.

#### Use a brokerpak

If you are not familiar with the concept of brokerpak, read the documentation you will find in the following links:
* [Brokerpak Intro](../brokerpak-intro.md)
* [Brokerpak Specification](../brokerpak-specification.md)
* [Brokerpak Dissection](../brokerpak-dissection.md)

You can also check the brokerpaks developed by the CSB team:
* [Brokerpak for AWS](https://github.com/cloudfoundry/csb-brokerpak-aws)
* [Brokerpak for GCP](https://github.com/cloudfoundry/csb-brokerpak-gcp)
* [Brokerpak for Azure](https://github.com/cloudfoundry/csb-brokerpak-azure)

Clone the brokerpak for AWS and navigate to the root of the project:

```shell
cd .. && git clone git@github.com:cloudfoundry/csb-brokerpak-aws.git && cd csb-brokerpak-aws
```

```shell
{▸} ~/workspace/csb/cloud-service-broker on main ✓ cd .. && git clone git@github.com:cloudfoundry/csb-brokerpak-aws.git && cd csb-brokerpak-aws
...
...
direnv: error ~/workspace/csb/csb-brokerpak-aws/.envrc is blocked. Run `direnv allow` to approve its content
{▸} ~/workspace/csb/csb-brokerpak-aws on main ✓ direnv allow                          
direnv: loading ~/workspace/csb/csb-brokerpak-aws/.envrc                                                                                                                                                                                
direnv: export +GSB_COMPATIBILITY_ENABLE_BETA_SERVICES \
+GSB_SERVICE_CSB_AWS_AURORA_MYSQL_PLANS \
+GSB_SERVICE_CSB_AWS_AURORA_POSTGRESQL_PLANS \
+GSB_SERVICE_CSB_AWS_MYSQL_PLANS \
+GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS \
+GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS \
+PAK_BUILD_CACHE_PATH
```

As you noticed, as you come into the directory, the `direnv` tool tells you the environment variables that have been set.
These variables are necessary for the proper functioning of the CSB, for example, to define the plans associated with
the different services, the path for the cache, etc.

#### Check the available services

You can check the available services by executing the next command:

```shell
csb marketplace
```

```shell
{▸} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb marketplace        
2023/01/16 21:25:36 Packing brokerpak version "0.1.0" with CSB version "v0.14.0-1b000966"...
2023/01/16 21:25:36 Using temp directory: /tmp/brokerpak2435728932
2023/01/16 21:25:36 Packing binaries...
2023/01/16 21:25:36      https://releases.hashicorp.com/terraform/0.12.30/terraform_0.12.30_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64/0.12.30 (from cache)
2023/01/16 21:25:36      https://releases.hashicorp.com/terraform/0.13.7/terraform_0.13.7_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64/0.13.7 (from cache)
2023/01/16 21:25:36      https://releases.hashicorp.com/terraform/0.14.11/terraform_0.14.11_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64/0.14.11 (from cache)
2023/01/16 21:25:36      https://releases.hashicorp.com/terraform/1.0.11/terraform_1.0.11_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64/1.0.11 (from cache)
2023/01/16 21:25:36      https://releases.hashicorp.com/terraform/1.1.9/terraform_1.1.9_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64/1.1.9 (from cache)
2023/01/16 21:25:36      https://releases.hashicorp.com/terraform-provider-aws/4.50.0/terraform-provider-aws_4.50.0_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64 (from cache)
2023/01/16 21:25:36      https://releases.hashicorp.com/terraform-provider-random/3.4.3/terraform-provider-random_3.4.3_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64 (from cache)
2023/01/16 21:25:36      https://releases.hashicorp.com/terraform-provider-mysql/1.9.0/terraform-provider-mysql_1.9.0_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64 (from cache)
2023/01/16 21:25:36      https://github.com/cyrilgdn/terraform-provider-postgresql/releases/download/v1.18.0/terraform-provider-postgresql_1.18.0_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64 (from cache)
2023/01/16 21:25:36      https://github.com/cloudfoundry/terraform-provider-csbpg/releases/download/v1.0.1/terraform-provider-csbpg_1.0.1_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64 (from cache)
2023/01/16 21:25:36      https://github.com/cloudfoundry/terraform-provider-csbmysql/releases/download/v1.0.0/terraform-provider-csbmysql_1.0.0_linux_amd64.zip -> /tmp/brokerpak2435728932/bin/linux/amd64 (from cache)
2023/01/16 21:25:36 Packing definitions...
2023/01/16 21:25:36     /aws-mysql.yml -> /tmp/brokerpak2435728932/definitions/service0-csb-aws-mysql.yml
2023/01/16 21:25:36     /aws-redis-cluster.yml -> /tmp/brokerpak2435728932/definitions/service1-csb-aws-redis.yml
2023/01/16 21:25:36     /aws-postgresql.yml -> /tmp/brokerpak2435728932/definitions/service2-csb-aws-postgresql.yml
2023/01/16 21:25:36     /aws-s3-bucket.yml -> /tmp/brokerpak2435728932/definitions/service3-csb-aws-s3-bucket.yml
2023/01/16 21:25:36     /aws-dynamodb.yml -> /tmp/brokerpak2435728932/definitions/service4-csb-aws-dynamodb.yml
2023/01/16 21:25:36     /aws-aurora-postgresql.yml -> /tmp/brokerpak2435728932/definitions/service5-csb-aws-aurora-postgresql.yml
2023/01/16 21:25:36     /aws-aurora-mysql.yml -> /tmp/brokerpak2435728932/definitions/service6-csb-aws-aurora-mysql.yml
2023/01/16 21:25:36 Creating archive: aws-services-0.1.0.brokerpak
created: aws-services-0.1.0.brokerpak
{"timestamp":"1673904337.076221704","source":"cloud-service-broker","message":"cloud-service-broker.starting","log_level":1,"data":{"version":"v0.14.0-1b000966-dirty"}}
{"timestamp":"1673904337.076258659","source":"cloud-service-broker","message":"cloud-service-broker.WARNING: DO NOT USE SQLITE3 IN PRODUCTION!","log_level":1,"data":{}}
{"timestamp":"1673904337.123361111","source":"cloud-service-broker","message":"cloud-service-broker.database-encryption","log_level":1,"data":{"primary":"none"}}
{"timestamp":"1673904337.123420477","source":"brokerpak-registration","message":"brokerpak-registration.registering","log_level":1,"data":{"excluded-services":null,"location":"/tmp/csb-220611276/aws-services-0.1.0.brokerpak","name":"builtin-0","notes":"This pak was automatically loaded because the toggle GSB_COMPATIBILITY_ENABLE_BUILTIN_BROKERPAKS was enabled","prefix":""}}
{"timestamp":"1673904337.462893724","source":"brokerpak-registration","message":"brokerpak-registration.registration-successful","log_level":1,"data":{"version":"0.1.0"}}
{"timestamp":"1673904337.463354349","source":"cloud-service-broker","message":"cloud-service-broker.service catalog","log_level":1,"data":{"catalog":[{"id":"7446e75e-2a09-11ed-8816-23072dae39dc","name":"csb-aws-aurora-mysql","plans":[{"id":"10b2bd92-2a0b-11ed-b70f-c7c5cf3bb719","name":"default"}],"tags":["aws","aurora","mysql"]},{"id":"36203e40-2945-11ed-8980-eb81bd131a02","name":"csb-aws-aurora-postgresql","plans":[{"id":"d20c5cf2-29e1-11ed-93da-1f3a67a06903","name":"default"}],"tags":["aws","aurora","postgresql","postgres"]},{"id":"bf1db66a-1316-11eb-b959-e73b704ea230","name":"csb-aws-dynamodb","plans":[{"id":"52b109ee-1318-11eb-851b-dbe6aa707e6b","name":"ondemand"},{"id":"591808b4-1318-11eb-b932-cbf259c3124c","name":"provisioned"}],"tags":["aws","dynamodb","beta"]},{"id":"fa22af0f-3637-4a36-b8a7-cfc61168a3e0","name":"csb-aws-mysql","plans":[{"id":"0f3522b2-f040-443b-bc53-4aed25284840","name":"default"}],"tags":["aws","mysql"]},{"id":"fa6334bc-5314-4b63-8a74-c0e4b638c950","name":"csb-aws-postgresql","plans":[{"id":"de7dbcee-1c8d-11ed-9904-5f435c1e2316","name":"default"}],"tags":["aws","postgresql","postgres"]},{"id":"e9c11b1b-0caa-45c9-b9b2-592939c9a5a6","name":"csb-aws-redis","plans":[{"id":"ad963fcd-19f7-4b79-8e6d-645756e84f7a","name":"small"},{"id":"df41095a-43e8-4be4-b4d6-ae2d8a35068d","name":"medium"},{"id":"da4dc49c-a64f-4d2a-8490-5e456cbb0577","name":"large"},{"id":"70544df7-0ac4-4580-ba51-c1fbdd6fdfd0","name":"small-ha"},{"id":"a4235008-80f4-4053-924b-defcce17cb63","name":"medium-ha"},{"id":"f26cda6f-d4b4-473a-966c-32d238f723ef","name":"large-ha"}],"tags":["aws","redis","beta"]},{"id":"ffe28d48-c235-4e07-9c51-ddff5699e48c","name":"csb-aws-s3-bucket","plans":[{"id":"f64891b4-5021-4742-9871-dfe1a9051302","name":"default"}],"tags":["aws","s3"]}]}}
{"timestamp":"1673904337.474073887","source":"cloud-service-broker","message":"cloud-service-broker.Serving","log_level":1,"data":{"port":"34521"}}

Service Offering           Plans
----------------           -----
csb-aws-aurora-mysql       default
csb-aws-aurora-postgresql  default
csb-aws-dynamodb           ondemand, provisioned
csb-aws-mysql              default
csb-aws-postgresql         default
csb-aws-redis              small, medium, large, small-ha, medium-ha, large-ha
csb-aws-s3-bucket          default


```

You may be thinking about what the command is doing before showing the available services and plans.
Basically, it downloads the necessary third parties such as Terraform providers, and packs the definitions of the
available services.

Now, you are ready to advance to the next step and create a new service on AWS.

##### Create an AWS S3 Bucket

Before creating an AWS S3 Bucket, you should understand the structure of the command you should execute:

```shell
 csb create-service <Service Offering> <Service plan> <Service name> -c '{extra params in JSON format}'
```

You have to select the service offering, the plan, and write a name for your service. If you want to change some
default values associated with the plan you have selected, use the -c option to add extra parameters.

Run the following command to create an S3 Bucket using the default plan and default values:

```shell
csb create-service csb-aws-s3-bucket default my-bucket 
```

```shell
{▪} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb create-service csb-aws-s3-bucket default my-bucket         
.........
{"timestamp":"1673958847.473781347","source":"cloud-service-broker","message":"cloud-service-broker.lastOperation.done-check-for-operation","log_level":1,"data":{"correlation-id":"c985a332-9ba3-45b6-8116-8b19766899f6","instance-id":"6d792d62-7563-6b65-742e-2e2e2e2e2e2e","requestIdentity":"","session":"46","state":"succeeded"}}
```

You will wait until seeing the result `"state":"succeeded"` that indicates the operation was executed correctly,
but in the meantime, there is a pulling operation to retrieve the last operation status. The time of pulling depends
on the service you are creating. In this example, the S3 service is tremendously fast.

If you do not set the environment variables necessary for the selected provider, you will get an error similar to this:

```shell
{▪} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb create-service csb-aws-s3-bucket default my-bucket         
.........
2023/01/16 21:40:21 unexpected status code 500: {"description":"1 error(s) occurred: couldn't compute the value for \"aws_secret_access_key\", template: \"${config(\\\"aws.secret_access_key\\\")}\", config: missing config value aws.secret_access_key"}
```

Check the environment variables [section](#necessary-environment-variables) to know the names of the variables.


##### Check the created service

You can see the created service by running the command:

```shell
csb services
```

```shell
{▸} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb services                                          
....

Name       Service offering   Plan
----       ----------------   ----
my-bucket  csb-aws-s3-bucket  default

```

The executed command shows you the bucket `my-bucket` you created in the previous step.

Next step: you can modify the service `my-bucket` and continue with your learning path.

##### Update your service

In this section, you will learn how to update the `my-bucket` service or any service you have previously created.

You can review the AWS S3 service definition for available properties.

Check the section `user_inputs` in the file `aws-s3-bucket.yml` in the project `csb-brokerpak-aws`.
Remember that you can change any property that does not have the configuration `prohibit_update: true` and
is not defined in the plan. For example:

```yaml
  - field_name: enable_versioning
    type: boolean
    details: Enable bucket versioning
    default: false
```

As you created your bucket using the default plan, a plan without configuration,
the value for the `enable_versioning` property is false because it is the default value used in the property definition.

> Note: check the definition of the plan in the `.envrc` file.
> The environment variable called `GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS` defines an array of plans for AWS S3.

You can enable this functionality by running the command:

```shell
csb update-service my-bucket -c '{"enable_versioning": true}'
```

```shell
{▪} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb update-service my-bucket -c '{"enable_versioning": true}'
...
...
{"timestamp":"1673965049.968899488","source":"cloud-service-broker","message":"cloud-service-broker.lastOperation.done-check-for-operation","log_level":1,"data":{"correlation-id":"3257b4c0-c75a-422e-b19a-d6809a5e179e","instance-id":"6d792d62-7563-6b65-742e-2e2e2e2e2e2e","requestIdentity":"","session":"59","state":"succeeded"}}
```

If everything goes correct, the last message you will receive will contain `"state":"succeeded"`.
You can always check the AWS console to analyze your service.

If you do not remember how to update a service, remember you can execute the command without parameters:

```shell
{▪} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb update-service     
Error: accepts 1 arg(s), received 0
Usage:
  cloud-service-broker update-service NAME [flags]

Flags:
  -c, --c string   parameters as JSON
  -h, --help       help for update-service
  -p, --p string   change service plan for a service instance

Global Flags:
      --config string   Configuration file to be read

accepts 1 arg(s), received 0

```

###### Identify the service in the AWS console

Well, you probably are asking yourself, how to identify your service in the AWS console.
The local development uses an SQLite database stored in the root of the project of the brokerpak you are using.
If you navigate to the root folder of the `csb-brokerpak-aws` project, you will see a file called `.csb.db`.

Configure your favourite SQL client and see the details of your service instance, by running the command:

```sql
select id, json_extract(other_details, '$.bucket_name') as bucket_name, service_id, plan_id from service_instance_details
```

```shell
id                                  ,bucket_name                             ,service_id                          ,plan_id
6d792d62-7563-6b65-742e-2e2e2e2e2e2e,csb-6d792d62-7563-6b65-742e-2e2e2e2e2e2e,ffe28d48-c235-4e07-9c51-ddff5699e48c,f64891b4-5021-4742-9871-dfe1a9051302
```

You can navigate to the S3 console and search your bucket by name `csb-6d792d62-7563-6b65-742e-2e2e2e2e2e2e`.
You will see on the `Properties` tab that Bucket Versioning is Enabled.


##### Create a service key

Since you are not using Cloud Foundry as an intermediate layer to manage your service or applications,
you cannot use the binary to bind a service with an app on Cloud Foundry, but you can create a service key.

From the point of view of the Open Service Broker API, a service key represents the same concept as making a binding.
In other words, you can create the necessary credentials to connect with your service by running the command:

```shell
csb create-service-key my-bucket my-first-key
```

```shell
{▸} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb create-service-key my-bucket my-first-key         
...
...
Service Key: {"credentials":{"access_key_id":"<YOU WILL SEE YOUR ACCESS KEY HERE>","arn":"arn:aws:s3:::csb-6d792d62-7563-6b65-742e-2e2e2e2e2e2e","bucket_domain_name":"csb-6d792d62-7563-6b65-742e-2e2e2e2e2e2e.s3.amazonaws.com","bucket_name":"csb-6d792d62-7563-6b65-742e-2e2e2e2e2e2e","region":"us-west-2","secret_access_key":"<YOU WILL SEE YOUR SECRET ACCESS KEY HERE>"}}
```

Remember you can see the result stored in the database by executing:

```sql
select
    sbc.id, sbc.other_details as binding_details
from service_binding_credentials sbc
```

```shell
id,    binding_details
1,     "{"access_key_id":"<YOU WILL SEE YOUR ACCESS KEY HERE>","secret_access_key":"<YOU WILL SEE YOUR SECRET ACCESS KEY HERE>"}"
```

##### List your service keys 

You can list the service keys associated with your service by running the command:

```shell
csb service-keys my-bucket 
```

```shell
{▸} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb service-keys my-bucket  
Name
----
my-first-key
```

##### Delete your service key

In the same way, you created the service key, you can delete it. 

Check the help to know the expected parameters: 

```shell
{▸} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb delete-service-key                       
Error: accepts 2 arg(s), received 0
Usage:
  cloud-service-broker delete-service-key SERVICE_INSTANCE SERVICE_KEY [flags]

Flags:
  -h, --help   help for delete-service-key

Global Flags:
      --config string   Configuration file to be read

accepts 2 arg(s), received 0
```

And now, delete your key by executing the command:

```shell
csb delete-service-key my-bucket my-first-key
```

```shell
{▸} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb delete-service-key my-bucket my-first-key
...
...
...
```

##### Delete your service

Finally, try deleting the created service, in other words, delete your bucket by running the command:

```shell
csb delete-service my-bucket
```

```shell
{▸} ~/workspace/csb/csb-brokerpak-aws on main ✓ csb delete-service my-bucket
{"........."state":"succeeded"}}
```

If you check the AWS console you can see that the bucket no longer exists, and if you search for service
instance ID in the database you won't receive any rows:

```sql
select * from service_instance_details where id='6d792d62-7563-6b65-742e-2e2e2e2e2e2e'
```
```shell
O rows
```

Remember that you previously got the service instance ID by executing the SQL
[command](#identify-the-service-in-the-aws-console) in the section "Identify the service in the AWS console". 

### Necessary environment variables

Each provider needs its environment variables.
At the time of writing this text, the necessary variables categorized by the provider are as follows:
* AWS:
    * AWS_ACCESS_KEY_ID: AWS access key.
    * AWS_SECRET_ACCESS_KEY: AWS secret.
* Azure:
    * ARM_CLIENT_ID: service principal client ID.
    * ARM_CLIENT_SECRET: service principal secret.
    * ARM_SUBSCRIPTION_ID: ID for the subscription that resources will be created in.
    * ARM_TENANT_ID: ID for the tenant that resources will be created in.
* Google
  * GOOGLE_CREDENTIALS: the string version of the credentials file created for the Owner level Service Account.
  * GOOGLE_PROJECT: GCP project id.
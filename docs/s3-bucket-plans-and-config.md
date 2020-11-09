# Amazon S3 Bucket Config
## Applies to service *csb-aws-s3-bucket*

*csb-aws-s3-bucket* manages individual S3 buckets on AWS (not currently supported on GCP or Azure.)

## Plans

| Plan | Description |
|------|-------------|
| private | a private S3 bucket |
| public-read | a publicly readable S3 bucket |

## Config parameters

The following parameters may be configured during service provisioning (`cf create-service csb-aws-s3-bucket ... -c '{...}'`

| Parameter | Type | Description | Default |
|-----------|------|------|---------|
| bucket_name| string | Name of bucket to create | csb-*instance_id* |
| region  | string | [AWS region](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions) to deploy service  | us-west-2 |
| aws_access_key_id | string | AWS Access Key to use for instance | config file value `aws.access_key_id` |
| aws_secret_access_key | string | Corresponding secret for the AWS Access Key to use for instance | config file value `aws.secret_access_key` |

## Binding Credentials

The binding credentials for the S3 bucket have the following shape:

```json
{
    "arn" : "bucket arn",
    "bucket_domain_name" : "bucket FQDN",
    "region" : "bucket region",
    "bucket_name" : "bucket name",
    "access_key_id" : "access key for bucket",
    "secret_access_key" : "secret key for bucket",
}
```

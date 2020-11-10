variable dynamodb_table_arn { type = string }
variable dynamodb_table_id { type = string }
variable aws_access_key_id { type = string }
variable aws_secret_access_key { type = string }
variable region { type = string }
variable user_name { type = string }


provider "aws" {
  version = "~> 3.0"
  region  = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}  


data "aws_iam_policy_document" "user_policy" {
    statement {
        sid = "dynamoAccess"
        actions = [
    		"dynamodb:*",
        ]
        resources = [
            var.dynamodb_table_arn 
        ]
    }


}

resource "aws_iam_user" "user" {
    name = var.user_name
    path = "/cf/"
}

resource "aws_iam_access_key" "access_key" {
    user = aws_iam_user.user.name
}

resource "aws_iam_user_policy" "user_policy" {
    name = format("%s-p", var.user_name)

    user = aws_iam_user.user.name

    policy = data.aws_iam_policy_document.user_policy.json
}

output access_key_id { value = aws_iam_access_key.access_key.id}
output secret_access_key { value = aws_iam_access_key.access_key.secret }
output dynamodb_table_arn { value = var.dynamodb_table_arn }
output dynamodb_table_id { value = var.dynamodb_table_arn }
output region { value = var.region }
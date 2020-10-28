
variable aws_access_key_id { type = string }
variable aws_secret_access_key { type = string }
variable region { type = string }
provider "aws" {
  version = "~> 3.0"
  region  = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}    


variable region { type = string }
variable labels { type = map }
variable aws_access_key_id { type = string }
variable aws_secret_access_key { type = string }
variable aws_vpc_id { type = string }
variable instance_name { type = string }

variable billing_mode { type = string }
variable hash_key { type = string }
variable range_key { type = string }
variable tabel_name { type = string }

variable server_side_encryption_kms_key_arn { type = string }
variable attributes { type = list(map(string)) }
variable local_secondary_indexes { type = any }
variable global_secondary_indexes { type = any }

variable ttl_attribute_name { type = string }
variable ttl_enabled { type = bool }
variable stream_enabled { type = bool }
variable stream_view_type { type = string }
variable server_side_encryption_enabled { type = bool }

variable write_capacity { type = number }
variable read_capacity { type = number }
   

provider "aws" {
  version = "~> 3.0"
  region  = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
} 



resource "aws_dynamodb_table" "this" {
  
  name             = var.tabel_name
  billing_mode     = var.billing_mode
  hash_key         = var.hash_key
  range_key        = var.range_key
  read_capacity    = var.read_capacity
  write_capacity   = var.write_capacity
  stream_enabled   = var.stream_enabled
  stream_view_type = var.stream_view_type

  ttl {
    enabled        = var.ttl_enabled
    attribute_name = var.ttl_attribute_name
  }

 #attribute = var.attributes

  dynamic "attribute" {
    for_each = var.attributes

    content {
      name = attribute.value.name
      type = attribute.value.type
    }
  }

  dynamic "local_secondary_index" {
    for_each = var.local_secondary_indexes

    content {
      name               = local_secondary_index.value.name
      range_key          = local_secondary_index.value.range_key
      projection_type    = local_secondary_index.value.projection_type
      non_key_attributes = lookup(local_secondary_index.value, "non_key_attributes", null)
    }
  }

  dynamic "global_secondary_index" {
    for_each = var.global_secondary_indexes

    content {
      name               = global_secondary_index.value.name
      hash_key           = global_secondary_index.value.hash_key
      projection_type    = global_secondary_index.value.projection_type
      range_key          = lookup(global_secondary_index.value, "range_key", null)
      read_capacity      = lookup(global_secondary_index.value, "read_capacity", null)
      write_capacity     = lookup(global_secondary_index.value, "write_capacity", null)
      non_key_attributes = lookup(global_secondary_index.value, "non_key_attributes", null)
      
    }
  }



#   dynamic "replica" {
#     for_each = var.replica_regions

#     content {
#       region_name = replica.value
#     }
#   }

  server_side_encryption {
    enabled     = var.server_side_encryption_enabled
    kms_key_arn = var.server_side_encryption_kms_key_arn
  }

tags = var.labels


}

output region { value = var.region }
output dynamodb_table_arn { value =element(concat(aws_dynamodb_table.this.*.arn, list("")), 0)}
output dynamodb_table_id { value = element(concat(aws_dynamodb_table.this.*.id, list("")), 0) }
output dynamodb_table_name { value = aws_dynamodb_table.this.name }

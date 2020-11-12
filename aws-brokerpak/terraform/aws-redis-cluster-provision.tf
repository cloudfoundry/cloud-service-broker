# Copyright 2020 Pivotal Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

variable cache_size { type = number }
variable redis_version { type = string }
variable instance_name { type = string }
variable region { type = string }
variable labels { type = map }
variable aws_access_key_id { type = string }
variable aws_secret_access_key { type = string }
variable aws_vpc_id { type = string }
variable node_type { type = string }
variable node_count { type = number }

provider "aws" {
  version = "~> 3.0"
  region  = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
} 

data "aws_vpc" "default" {
  default = true
}

locals {
  instance_types = {
    // https://aws.amazon.com/elasticache/pricing
    1 = "cache.t2.small"
    2 = "cache.t3.medium"
    4 = "cache.m5.large"
    8 = "cache.m5.xlarge"
    16 = "cache.r4.xlarge"
    32 = "cache.r4.2xlarge"
    64 = "cache.r4.4xlarge"
    128 = "cache.r4.8xlarge"
    256 = "cache.r5.12xlarge"
  }   

  parameter_group_names = {
    "3.2" = "default.redis3.2"
    "4.0" = "default.redis4.0"
    "5.0" = "default.redis5.0"
    "6.0" = "default.redis6.x"
  }

  vpc_id = length(var.aws_vpc_id) == 0 ? data.aws_vpc.default.id : var.aws_vpc_id
  node_type = length(var.node_type) == 0 ? local.instance_types[var.cache_size] : var.node_type
  port = 6379
}

data "aws_subnet_ids" "all" {
  vpc_id = local.vpc_id
}

resource "aws_security_group" "sg" {
  name   = format("%s-sg", var.instance_name)
  vpc_id = local.vpc_id
}

resource "aws_elasticache_subnet_group" "subnet_group" {
  name = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnet_ids.all.ids
}

resource "aws_security_group_rule" "inbound_access" {
  from_port         = local.port
  protocol          = "tcp"
  security_group_id = aws_security_group.sg.id
  to_port           = local.port
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "random_password" "auth_token" {
  length = 64
  // https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/auth.html
  override_special = "!&#$^<>-"
  min_upper = 2
  min_lower = 2
  min_special = 2  
}

resource "aws_elasticache_replication_group" "redis" {
  automatic_failover_enabled    = var.node_count > 1
  replication_group_id          = var.instance_name
  replication_group_description = format("%s redis", var.instance_name)
  node_type                     = local.node_type
  number_cache_clusters         = var.node_count
  parameter_group_name          = local.parameter_group_names[var.redis_version]
  port                          = local.port
  tags                = var.labels
  security_group_ids = [aws_security_group.sg.id]
  subnet_group_name = aws_elasticache_subnet_group.subnet_group.name  
  transit_encryption_enabled = true
  auth_token = random_password.auth_token.result
}

output name { value = aws_elasticache_replication_group.redis.id }
output host { value = aws_elasticache_replication_group.redis.primary_endpoint_address }
output password { value = random_password.auth_token.result }
output tls_port { value = local.port }

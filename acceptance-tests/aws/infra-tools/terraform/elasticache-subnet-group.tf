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

variable region { type = string }
variable vpc_id { type = string }
variable name { type = string }

provider "aws" {
  version = "~> 3.0"
  region  = var.region
}

locals {
  ports = {
    "redis" = 6379
  }
}

data "aws_subnet_ids" "all" {
  vpc_id = var.vpc_id
}

resource "aws_security_group" "rds-sg" {
  name   = format("%s-elasticache-sg", var.name)
  vpc_id = var.vpc_id
}

resource "aws_elasticache_subnet_group" "subnet_group" {
  name = format("%s-elasticache-sn", var.name)
  subnet_ids = data.aws_subnet_ids.all.ids
}

resource "aws_security_group_rule" "redis" {
  from_port         = local.ports["redis"]
  protocol          = "tcp"
  security_group_id = aws_security_group.rds-sg.id
  to_port           = local.ports["redis"]
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

output name { value = aws_elasticache_subnet_group.subnet_group.name }
output security_group_id { value = aws_security_group.rds-sg.id }



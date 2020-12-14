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
    "mysql" = 3306
    "postgres" = 5432
  }
}

data "aws_subnet_ids" "all" {
  vpc_id = var.vpc_id
}

resource "aws_security_group" "rds-sg" {
  name   = format("%s-rds-sg", var.name)
  vpc_id = var.vpc_id
}

resource "aws_db_subnet_group" "rds-private-subnet" {
  name = format("%s-rds-sn", var.name)
  subnet_ids = data.aws_subnet_ids.all.ids
}

resource "aws_security_group_rule" "postgres" {
  from_port         = local.ports["postgres"]
  protocol          = "tcp"
  security_group_id = aws_security_group.rds-sg.id
  to_port           = local.ports["postgres"]
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "mysql" {
  from_port         = local.ports["mysql"]
  protocol          = "tcp"
  security_group_id = aws_security_group.rds-sg.id
  to_port           = local.ports["mysql"]
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

output name { value = aws_db_subnet_group.rds-private-subnet.name }
output security_group_id { value = aws_security_group.rds-sg.id }



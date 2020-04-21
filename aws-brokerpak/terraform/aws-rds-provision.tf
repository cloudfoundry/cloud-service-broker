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

variable cores { type = number }
variable instance_name { type = string }
variable db_name { type = string }
variable region { type = string }
variable labels { type = map }
variable storage_gb { type = number }
variable aws_access_key_id { type = string }
variable aws_secret_access_key { type = string }
variable aws_vpc_id { type = string }
variable publicly_accessible { type = bool }
variable multi_az { type = bool }
variable instance_class { type = string }
variable engine { type = string }
variable engine_version { type = string }

provider "aws" {
  version = "~> 2.0"
  region  = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}    

resource "random_string" "username" {
  length = 16
  special = false
  number = false
}

resource "random_password" "password" {
  length = 32
  special = false 
  // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html#RDS_Limits.Constraints
  override_special = "~_-."
}

data "aws_vpc" "default" {
  default = true
}

locals {
  instance_types = {
    // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html
    1 = "db.m1.medium"
    2 = "db.t2.medium"
    4 = "db.m4.xlarge"
    8 = "db.m4.2xlarge"
    16 = "db.m4.4xlarge"
    32 = "db.m5.8xlarge"
    64 = "db.m5.16xlarge"
  }   

  ports = {
      "mysql" = 3306
      "postgres" = 5432
  }
  
  vpc_id = length(var.aws_vpc_id) == 0 ? data.aws_vpc.default.id : var.aws_vpc_id
  instance_class = length(var.instance_class) == 0 ? local.instance_types[var.cores] : var.instance_class
}

data "aws_subnet_ids" "all" {
  vpc_id = local.vpc_id
}

resource "aws_security_group" "rds-sg" {
  name   = format("%s-sg", var.instance_name)
  vpc_id = local.vpc_id
}

resource "aws_db_subnet_group" "rds-private-subnet" {
  name = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnet_ids.all.ids
}

resource "aws_security_group_rule" "rds_inbound_access" {
  from_port         = local.ports[var.engine]
  protocol          = "tcp"
  security_group_id = aws_security_group.rds-sg.id
  to_port           = local.ports[var.engine]
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_db_instance" "db_instance" {
  allocated_storage    = var.storage_gb
  storage_type         = "gp2"
  skip_final_snapshot  = true
  engine               = var.engine
  engine_version       = var.engine_version
  instance_class       = local.instance_class
  identifier           = var.instance_name
  name                 = var.db_name
  username             = random_string.username.result
  password             = random_password.password.result
  parameter_group_name = format("default.%s%s",var.engine,var.engine_version)
  tags                 = var.labels
  vpc_security_group_ids = [aws_security_group.rds-sg.id]
  db_subnet_group_name = aws_db_subnet_group.rds-private-subnet.name
  publicly_accessible  = var.publicly_accessible
  multi_az             = var.multi_az
}

output name { value = aws_db_instance.db_instance.name }
output hostname { value = aws_db_instance.db_instance.address }
output port { value = local.ports[var.engine] }
output username { value = aws_db_instance.db_instance.username }
output password { value = aws_db_instance.db_instance.password }

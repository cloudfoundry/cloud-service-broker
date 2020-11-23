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

data "aws_vpc" "vpc" {
  default = length(var.aws_vpc_id) == 0
  id = length(var.aws_vpc_id) == 0 ? null : var.aws_vpc_id
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
  
  instance_class = length(var.instance_class) == 0 ? local.instance_types[var.cores] : var.instance_class
}

data "aws_subnet_ids" "all" {
  vpc_id = data.aws_vpc.vpc.id
}

resource "aws_security_group" "rds-sg" {
  name   = format("%s-sg", var.instance_name)
  vpc_id = data.aws_vpc.vpc.id
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

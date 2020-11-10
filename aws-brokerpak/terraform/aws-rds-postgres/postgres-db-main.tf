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
  allow_major_version_upgrade = true
  apply_immediately = true
}
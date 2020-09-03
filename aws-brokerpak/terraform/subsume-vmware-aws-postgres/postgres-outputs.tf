output name { value = aws_db_instance.db_instance.name }
output hostname { value = aws_db_instance.db_instance.address }
output port { value = 5432}
output username { value = aws_db_instance.db_instance.username }
output password { value = var.admin_password}

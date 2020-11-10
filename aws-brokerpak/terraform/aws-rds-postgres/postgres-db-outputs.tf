output name { value = aws_db_instance.db_instance.name }
output hostname { value = aws_db_instance.db_instance.address }
output port { value = local.ports[var.engine] }
output username { value = aws_db_instance.db_instance.username }
output password { value = "${ var.admin_password != "" ? var.admin_password : aws_db_instance.db_instance.password  }" }


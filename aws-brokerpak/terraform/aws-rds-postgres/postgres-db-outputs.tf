output name { value = aws_db_instance.db_instance.name }
output hostname { value = aws_db_instance.db_instance.address }
output port { value = local.ports[var.engine] }
output username { value = aws_db_instance.db_instance.username }
output password { value = "${ var.admin_password != "" ? var.admin_password : aws_db_instance.db_instance.password  }" }
output status {value = format("created db %s (id: %s) on server %s URL: https://%s.console.aws.amazon.com/rds/home?region=%s#database:id=%s;is-cluster=false", 
                               aws_db_instance.db_instance.name, 
                               aws_db_instance.db_instance.id, 
                               aws_db_instance.db_instance.address, 
                               var.region,
                               var.region,
                               aws_db_instance.db_instance.id)}


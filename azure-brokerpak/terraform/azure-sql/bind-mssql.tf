variable mssql_db_name { type = string }
variable mssql_hostname { type = string }
variable mssql_port { type = number }
variable admin_username { type = string }
variable admin_password { type = string }

resource "random_string" "username" {
  length = 16
  special = false
}

resource "random_password" "password" {
  length = 64
  special = true
  override_special = "_@!$#%"
  depends_on = [random_string.username]
}    

resource "null_resource" "create-sql-login" {

  provisioner "local-exec" {
    command = "sqlcmd -S ${var.mssql_hostname} -U ${var.admin_username} -P ${var.admin_password} -d master -Q \"CREATE LOGIN ${random_string.username.result} with PASSWORD='${random_string.password.result}'\"" 

  }

    provisioner "local-exec" {
	when = destroy
    command = "sqlcmd -S ${var.mssql_hostname} -U ${var.admin_username} -P ${var.admin_password} -d master -Q \"DROP LOGIN ${random_string.username.result}\""
  }
  depends_on = [random_password.password]
}

resource "null_resource" "create-sql-user-and-permissions" {

  provisioner "local-exec" {
    command = "sqlcmd -S ${var.mssql_hostname} -U ${var.admin_username} -P ${var.admin_password} -d ${var.mssql_db_name} -Q \"CREATE USER ${random_string.username.result} from LOGIN ${random_string.username.result};ALTER ROLE db_owner ADD MEMBER ${random_string.username.result};\"" 

  }

    provisioner "local-exec" {
	when = destroy
    command = "sqlcmd -S ${var.mssql_hostname} -U ${var.admin_username} -P ${var.admin_password} -d ${var.mssql_db_name} -Q \"ALTER ROLE db_owner DROP MEMBER ${random_string.username.result};DROP USER ${random_string.username.result};\""
  }

  depends_on = [null_resource.create-sql-login]
}

output username { value = random_string.username.result }
output password { value = random_password.password.result }
output "jdbcUrl" {
  value = format("jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30", 
                 var.mssql_hostname, 
                 var.mssql_port,
                 var.mssql_db_name, 
                 random_string.username.result, 
                 random_password.password.result)
}
output "jdbcUrlForAuditingEnabled" {
  value = format("jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30", 
                 var.mssql_hostname, 
                 var.mssql_port,
                 var.mssql_db_name, 
                 random_string.username.result,
                 random_password.password.result)
}
output "uri" {
  value = format("mssql://%s:%d/%s?encrypt=true&TrustServerCertificate=false&HostNameInCertificate=*.database.windows.net", 
                 var.mssql_hostname, 
                 var.mssql_port,
                 var.mssql_db_name)
}

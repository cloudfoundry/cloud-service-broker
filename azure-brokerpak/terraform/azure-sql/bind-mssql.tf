variable mssql_db_name { type = string }
variable mssql_hostname { type = string }
variable mssql_port { type = number }
variable admin_username { type = string }
variable admin_password { type = string }

provider "sqlserver" {
  address = var.mssql_hostname
  port = var.mssql_port
  username = var.admin_username
  password = var.admin_password
}

resource "random_string" "username" {
  length = 16
  special = false
}

resource "random_password" "password" {
  length = 64
  special = true
  override_special = "_@!$#%"
}    

resource "sqlserver_login" "newuser" {
  name             = random_string.username.result
  password         = random_password.password.result
  default_database = var.mssql_db_name
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

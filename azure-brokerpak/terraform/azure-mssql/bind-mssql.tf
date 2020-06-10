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

variable mssql_db_name { type = string }
variable mssql_hostname { type = string }
variable mssql_port { type = number }
variable admin_username { type = string }
variable admin_password { type = string }

resource "random_string" "username" {
  length = 16
  special = false
  number = false
}

resource "random_password" "password" {
  length = 64
  override_special = "~_-."
  min_upper = 2
  min_lower = 2
  min_special = 2
  depends_on = [random_string.username]
}    

resource "null_resource" "create-sql-login" {

  provisioner "local-exec" {
    command = format("psqlcmd %s %d %s %s master \"CREATE LOGIN [%s] with PASSWORD='%s'\"",
                     var.mssql_hostname,
                     var.mssql_port,
                     var.admin_username,
                     var.admin_password,
                     random_string.username.result,
                     random_password.password.result)
  }

  provisioner "local-exec" {
	  when = destroy
    command = format("psqlcmd %s %d %s %s master \"DROP LOGIN [%s]\"",
                     var.mssql_hostname,
                     var.mssql_port,
                     var.admin_username,
                     var.admin_password,
                     random_string.username.result)
  }
  depends_on = [random_password.password]
}

resource "null_resource" "create-sql-user" {
  provisioner "local-exec" {
    command = format("psqlcmd %s %d %s %s %s \"CREATE USER [%s] from LOGIN %s;\"", 
                     var.mssql_hostname,
                     var.mssql_port,
                     var.admin_username,
                     var.admin_password,
                     var.mssql_db_name,
                     random_string.username.result,
                     random_string.username.result)
  }

  provisioner "local-exec" {
	  when = destroy
    command = format("psqlcmd %s %d %s %s %s \"DROP USER [%s];\"",
                     var.mssql_hostname,
                     var.mssql_port,
                     var.admin_username,
                     var.admin_password,
                     var.mssql_db_name,
                     random_string.username.result)
  }

  depends_on = [null_resource.create-sql-login]
}

locals {
  roles = { "db_ddladmin" = "db_ddladmin"
            "db_datareader" = "db_datareader"
            "db_datawriter" = "db_datawriter" }
}

resource "null_resource" "add-user-roles" {
  # https://docs.microsoft.com/en-us/sql/relational-databases/security/authentication-access/database-level-roles?view=sql-server-ver15
  for_each = local.roles

  provisioner "local-exec" {
    command = format("psqlcmd %s %d %s %s %s \"ALTER ROLE %s ADD MEMBER [%s];\"", 
                     var.mssql_hostname,
                     var.mssql_port,
                     var.admin_username,
                     var.admin_password,
                     var.mssql_db_name,
                     each.key,
                     random_string.username.result)
  }

  provisioner "local-exec" {
	  when = destroy

    command = format("psqlcmd %s %d %s %s %s \"ALTER ROLE %s DROP MEMBER [%s]\"",
                     var.mssql_hostname,
                     var.mssql_port,
                     var.admin_username,
                     var.admin_password,
                     var.mssql_db_name,
                     each.key,
                     random_string.username.result)
  }

  depends_on = [null_resource.create-sql-user]
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
output "databaseLogin" {value = random_string.username.result}
output "databaseLoginPassword" {value = random_password.password.result}


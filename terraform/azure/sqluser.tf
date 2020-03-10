variable address { type = string }
variable database { type = string }
variable admin_username { type = string }
variable admin_password { type = string }
variable database_username {type = string}
variable database_password {type = string}
resource "null_resource" "create-sql-login" {

  provisioner "local-exec" {
    command = "sqlcmd -S ${var.address} -U ${var.admin_username} -P ${var.admin_password} -d master -Q \"CREATE LOGIN ${var.database_username} with PASSWORD='${var.database_password}'\"" 

  }

    provisioner "local-exec" {
	when = destroy
    command = "sqlcmd -S ${var.address} -U ${var.admin_username} -P ${var.admin_password} -d master -Q \"DROP LOGIN ${var.database_username}\""
  }
}

resource "null_resource" "create-sql-user-and-permissions" {

  provisioner "local-exec" {
    command = "sqlcmd -S ${var.address} -U ${var.admin_username} -P ${var.admin_password} -d ${var.database} -Q \"CREATE USER ${var.database_username} from LOGIN ${var.database_username};ALTER ROLE db_owner ADD MEMBER ${var.database_username};\"" 

  }

    provisioner "local-exec" {
	when = destroy
    command = "sqlcmd -S ${var.address} -U ${var.admin_username} -P ${var.admin_password} -d ${var.database} -Q \"ALTER ROLE db_owner DROP MEMBER ${var.database_username};DROP USER ${var.database_username};\""
  }

  depends_on = [null_resource.create-sql-login]
}

variable address { type = string }
variable database { type = string }
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
}    

resource "null_resource" "create-sql-login" {

  provisioner "local-exec" {
    command = "sqlcmd -S ${var.address} -U ${var.admin_username} -P ${var.admin_password} -d master -Q \"CREATE LOGIN ${random_string.username.result} with PASSWORD='${random_password.password.result}'\"" 

  }

    provisioner "local-exec" {
	when = destroy
    command = "sqlcmd -S ${var.address} -U ${var.admin_username} -P ${var.admin_password} -d master -Q \"DROP LOGIN ${random_string.username.result}\""
  }
}

resource "null_resource" "create-sql-user-and-permissions" {

  provisioner "local-exec" {
    command = "sqlcmd -S ${var.address} -U ${var.admin_username} -P ${var.admin_password} -d ${var.database} -Q \"CREATE USER ${random_string.username.result} from LOGIN ${random_string.username.result};ALTER ROLE db_owner ADD MEMBER ${random_string.username.result};\"" 

  }

    provisioner "local-exec" {
	when = destroy
    command = "sqlcmd -S ${var.address} -U ${var.admin_username} -P ${var.admin_password} -d ${var.database} -Q \"ALTER ROLE db_owner DROP MEMBER ${random_string.username.result};DROP USER ${random_string.username.result};\""
  }

  depends_on = [null_resource.create-sql-login]
}
output username { value = random_string.username.result }
output password { value = random_password.password.result }

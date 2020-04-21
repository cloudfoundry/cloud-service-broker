variable db_name { type = string }
variable hostname { type = string }
variable port { type = number }
variable admin_username { type = string }
variable admin_password { type = string }

provider "postgresql" {
  host            = var.hostname
  port            = var.port
  username        = var.admin_username
  password        = var.admin_password
  superuser       = false
  database        = var.db_name
}

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
}    

resource "postgresql_role" "new_user" {
  name     = random_string.username.result
  login    = true
  password = random_password.password.result
  skip_reassign_owned = true
  skip_drop_role = true
}

resource "postgresql_grant" "all_access" {
  depends_on  = [ postgresql_role.new_user ]
  database    = var.db_name
  role        = random_string.username.result
  schema      = "public"
  object_type = "table"
  privileges  = ["ALL"]
}

output username { value = random_string.username.result }
output password { value = random_password.password.result }
output uri { 
  value = format("postgresql://%s:%s@%s:%d/%s", 
                  random_string.username.result, 
                  random_password.password.result, 
                  var.hostname, 
                  var.port,
                  var.db_name) 
}
output jdbcUrl { 
  value = format("jdbc:postgresql://%s:%d/%s?user=%s\u0026password=%s\u0026useSSL=false", 
                  var.hostname, 
                  var.port,
                  var.db_name, 
                  random_string.username.result, 
                  random_password.password.result) 
}

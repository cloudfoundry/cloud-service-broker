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

variable postgres_db_name { type = string }
variable postgres_hostname { type = string }
variable postgres_port { type = number }
variable admin_username { type = string }
variable admin_password { type = string }

locals {
  schema_name = "album" 
  #var.schema_name == null || var.schema_name == "" ? var.db_name : var.schema_name
}

provider "postgresql" {
  host            = var.postgres_hostname
  port            = var.postgres_port
  username        = var.admin_username
  password        = var.admin_password
  superuser       = false
  database        = var.postgres_db_name
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


resource "postgresql_schema" "schema" {
  name = local.schema_name
  database = var.postgres_db_name
  owner = var.admin_username

  policy {
    create_with_grant = true
    usage_with_grant = true
    role = postgresql_role.new_user.name
  }

}


resource "postgresql_grant" "all_access" {
  depends_on  = [ postgresql_role.new_user ]
  database    = var.postgres_db_name
  role        = random_string.username.result
  schema      = "public"
  object_type = "table"
  privileges  = ["ALL"]
}



// Adding Root user to new db

resource "postgresql_grant" "all_rootuser_access" {

  database    = var.postgres_db_name
  role        = var.admin_username
  schema      = "public"
  object_type = "table"
  privileges  = ["ALL"]
}

resource "postgresql_default_privileges" "db_newuser_tables" {
  depends_on  = [ postgresql_role.new_user ]
  database    = var.postgres_db_name
  object_type = "table"
  owner       = random_string.username.result
  privileges  = ["ALL"]
  role        = random_string.username.result
  schema      = "public"
}

resource "postgresql_default_privileges" "db_rootuser_tables" {
  database    = var.postgres_db_name
  object_type = "table"
  owner       = var.admin_username
  privileges  = ["ALL"]
  role        = var.admin_username
  schema      = "public"
} 

# resource "postgresql_default_privileges" "app" {
#   database = var.postgres_db_name
#   #schema = postgresql_schema.schema.name
#   owner = var.admin_username
#   role = postgresql_role.db_app.name
#   object_type = "table"
#   privileges = ["ALL"]
# }





output username { value = random_string.username.result }
output password { value = random_password.password.result }
output uri { 
  value = format("postgresql://%s:%s@%s:%d/%s", 
                  random_string.username.result, 
                  random_password.password.result, 
                  var.postgres_hostname, 
                  var.postgres_port,
                  var.postgres_db_name) 
}
output jdbcUrl { 
  value = format("jdbc:postgresql://%s:%d/%s?user=%s\u0026password=%s\u0026useSSL=false", 
                  var.postgres_hostname, 
                  var.postgres_port,
                  var.postgres_db_name, 
                  random_string.username.result, 
                  random_password.password.result) 
}
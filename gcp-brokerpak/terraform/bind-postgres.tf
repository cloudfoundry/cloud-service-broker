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
  
   table_privileges = [
    "DELETE",
    "INSERT",
    "REFERENCES",
    "SELECT",
    "TRIGGER",
    "TRUNCATE",
    "UPDATE"
  ]
  sequence_privileges = [
    "SELECT",
    "UPDATE",
    "USAGE"
  ]
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

// Create postgres role and db
resource "postgresql_role" "app_role" {
  login               = true
  name                = random_string.username.result
  password            = random_password.password.result
  skip_reassign_owned = true
  skip_drop_role = true
  
}

resource "postgresql_default_privileges" "app_tables" {
  database    = var.postgres_db_name
  depends_on  = [postgresql_role.app_role]
  object_type = "table"
  owner       = var.admin_username
  privileges  = local.table_privileges
  role        = postgresql_role.app_role.name
  schema      = "public"
}

resource "postgresql_default_privileges" "app_sequence" {
  database    = var.postgres_db_name
  depends_on  = [postgresql_role.app_role]
  object_type = "sequence"
  owner       = var.admin_username
  privileges  = local.sequence_privileges
  role        = postgresql_role.app_role.name
  schema      = "public"
}


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

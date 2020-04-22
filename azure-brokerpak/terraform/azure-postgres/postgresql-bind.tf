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

locals {
  username = format("%s@%s", random_string.username.result, var.hostname)
}
output username { value = local.username }
output password { value = random_password.password.result }
output uri { 
  value = format("%s://%s:%s@%s:%d/%s", 
                  "postgresql",
                  local.username, 
                  random_password.password.result, 
                  var.hostname, 
                  var.port,
                  var.db_name) 
}
output jdbcUrl { 
  value = format("jdbc:%s://%s:%s/%s?user=%s\u0026password=%s\u0026verifyServerCertificate=true\u0026useSSL=true\u0026requireSSL=false\u0026serverTimezone=GMT", 
                  "postgresql",
                  var.hostname, 
                  var.port,
                  var.db_name, 
                  local.username, 
                  random_password.password.result) 
}  

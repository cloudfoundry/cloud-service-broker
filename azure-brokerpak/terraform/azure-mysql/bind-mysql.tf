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

variable mysql_db_name { type = string }
variable mysql_hostname { type = string }
variable mysql_port { type = number }
variable admin_username { type = string }
variable admin_password { type = string }

provider "mysql" {
  endpoint = format("%s:%d", var.mysql_hostname, var.mysql_port)
  username = var.admin_username
  password = var.admin_password
}

resource "random_string" "username" {
  length = 16
  special = false
}

resource "random_password" "password" {
  length = 31
  override_special = "~_-."
  min_upper = 2
  min_lower = 2
  min_special = 2
}    

resource "mysql_user" "newuser" {
  user               = random_string.username.result
  plaintext_password = random_password.password.result
  host = "%"
}

resource "mysql_grant" "newuser" {
  user       = mysql_user.newuser.user
  database   = var.mysql_db_name
  host = mysql_user.newuser.host
  privileges = ["ALL"]
}

locals {
  username = format("%s@%s", random_string.username.result, var.mysql_hostname)
}
output username { value = local.username }
output password { value = random_password.password.result }
output uri { 
  value = format("%s://%s:%s@%s:%d/%s", 
                  "mysql",
                  local.username, 
                  random_password.password.result, 
                  var.mysql_hostname, 
                  var.mysql_port,
                  var.mysql_db_name) 
}
output jdbcUrl { 
  value = format("jdbc:%s://%s:%s/%s?user=%s\u0026password=%s\u0026verifyServerCertificate=true\u0026useSSL=true\u0026requireSSL=false\u0026serverTimezone=GMT", 
                  "mysql",
                  var.mysql_hostname, 
                  var.mysql_port,
                  var.mysql_db_name, 
                  local.username, 
                  random_password.password.result) 
}  
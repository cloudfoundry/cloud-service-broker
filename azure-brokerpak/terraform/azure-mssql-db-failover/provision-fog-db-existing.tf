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

variable instance_name { type = string }
variable server_credential_pairs { type = map }
variable server_pair { type = string }
variable db_name { type = string }

locals {
  serverFQDN = format("%s.database.windows.net", var.instance_name)
}

output "sqldbName" {value = var.db_name}
output "sqlServerName" {value = var.instance_name}
output "sqlServerFullyQualifiedDomainName" {value = local.serverFQDN}
output "hostname" {value = local.serverFQDN}
output "port" {value = 1433}
output "name" {value = var.db_name}
output "username" {value = var.server_credential_pairs[var.server_pair].admin_username}
output "password" {value = var.server_credential_pairs[var.server_pair].admin_password}
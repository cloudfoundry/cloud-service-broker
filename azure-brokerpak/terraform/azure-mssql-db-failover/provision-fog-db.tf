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
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable server_credential_pairs { type = map }
variable server_pair { type = string }
variable db_name { type = string }
variable labels { type = map }
variable pricing_tier { type = string }
variable cores { type = number }
variable storage_gb { type = number }

provider "azurerm" {
  version = "=2.9.0"
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id  
}

data "azurerm_sql_server" "primary_sql_db_server" {
  name                         = var.server_credential_pairs[var.server_pair].primary.server_name
  resource_group_name          = var.server_credential_pairs[var.server_pair].primary.resource_group
}

data "azurerm_sql_server" "secondary_sql_db_server" {
  name                         = var.server_credential_pairs[var.server_pair].secondary.server_name
  resource_group_name          = var.server_credential_pairs[var.server_pair].secondary.resource_group
}

resource "azurerm_sql_database" "azure_sql_db" {
  name                = var.db_name
  resource_group_name = var.server_credential_pairs[var.server_pair].primary.resource_group
  location            = data.azurerm_sql_server.primary_sql_db_server.location
  server_name         = data.azurerm_sql_server.primary_sql_db_server.name
  requested_service_objective_name = format("%s_Gen5_%d", var.pricing_tier, var.cores)
  max_size_bytes      = var.storage_gb * 1024 * 1024 * 1024
  tags                = var.labels
}

resource "azurerm_sql_failover_group" "failover_group" {
  name                = var.instance_name
  resource_group_name = var.server_credential_pairs[var.server_pair].primary.resource_group
  server_name         = data.azurerm_sql_server.primary_sql_db_server.name
  databases           = [azurerm_sql_database.azure_sql_db.id]
  partner_servers {
    id = data.azurerm_sql_server.secondary_sql_db_server.id
  }

  read_write_endpoint_failover_policy {
    mode          = "Automatic"
    grace_minutes = 60
  }
}

locals {
  serverFQDN = format("%s.database.windows.net", azurerm_sql_failover_group.failover_group.name)
}
output "sqldbName" {value = azurerm_sql_database.azure_sql_db.name}
output "sqlServerName" {value = azurerm_sql_failover_group.failover_group.name}
output "sqlServerFullyQualifiedDomainName" {value = local.serverFQDN}
output "hostname" {value = local.serverFQDN}
output "port" {value = 1433}
output "name" {value = azurerm_sql_database.azure_sql_db.name}
output "username" {value = var.server_credential_pairs[var.server_pair].admin_username}
output "password" {value = var.server_credential_pairs[var.server_pair].admin_password}
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
variable resource_group { type = string }
variable db_name { type = string }
variable location { type = string }
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable labels { type = map }
variable sku_name { type = string }
variable cores { type = number }
variable max_storage_gb { type = number }
variable authorized_network {type = string}
variable skip_provider_registration { type = bool }

provider "azurerm" {
  version = "=2.9.0"
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id  

  skip_provider_registration = var.skip_provider_registration
}

locals {
  instance_types = {
    1 = "GP_S_Gen5_1"
    2 = "GP_S_Gen5_2"
    4 = "GP_Gen5_4"
    8 = "GP_Gen5_8"
    16 = "GP_Gen5_16"
    32 = "HS_Gen5_32"
    80 = "HS_Gen5_80"
  }     
  sku_name = length(var.sku_name) == 0 ? local.instance_types[var.cores] : var.sku_name    
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
}

resource "azurerm_resource_group" "azure_sql" {
  name     = local.resource_group
  location = var.location
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0
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

resource "azurerm_sql_server" "azure_sql_db_server" {
  depends_on = [ azurerm_resource_group.azure_sql ]
  name                         = var.instance_name
  resource_group_name          = local.resource_group
  location                     = var.location
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags = var.labels
}

resource "azurerm_sql_database" "azure_sql_db" {
  name                = var.db_name
  resource_group_name = azurerm_sql_server.azure_sql_db_server.resource_group_name
  location            = var.location
  server_name         = azurerm_sql_server.azure_sql_db_server.name
  requested_service_objective_name = local.sku_name
  max_size_bytes      = var.max_storage_gb * 1024 * 1024 * 1024
  tags = var.labels
}

resource "azurerm_sql_virtual_network_rule" "allow_subnet_id" {
  name                = format("subnetrule-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.azure_sql_db_server.name
  subnet_id           = var.authorized_network
  count = var.authorized_network != "default" ? 1 : 0   
}

resource "azurerm_sql_firewall_rule" "sql_firewall_rule" {
  name                = format("firewallrule-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.azure_sql_db_server.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count = var.authorized_network == "default" ? 1 : 0  
}

output "sqldbResourceGroup" {value = azurerm_sql_server.azure_sql_db_server.resource_group_name}
output "sqldbName" {value = azurerm_sql_database.azure_sql_db.name}
output "sqlServerName" {value = azurerm_sql_server.azure_sql_db_server.name}
output "sqlServerFullyQualifiedDomainName" {value = azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name}
output "hostname" {value = azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name}
output "port" {value = 1433}
output "name" {value = azurerm_sql_database.azure_sql_db.name}
output "username" {value = random_string.username.result}
output "password" {value = random_password.password.result}


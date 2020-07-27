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

variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable db_name { type = string }
variable server { type = string }
variable server_credentials { type = map }
variable labels { type = map }
variable sku_name { type = string }
variable cores { type = number }
variable max_storage_gb { type = number }
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

data "azurerm_sql_server" "azure_sql_db_server" {
  name                         = var.server_credentials[var.server].server_name
  resource_group_name          = var.server_credentials[var.server].server_resource_group
}

locals {
  instance_types = {
    1 = "GP_Gen5_1"
    2 = "GP_Gen5_2"
    4 = "GP_Gen5_4"
    8 = "GP_Gen5_8"
    16 = "GP_Gen5_16"
    32 = "GP_Gen5_32"
    80 = "GP_Gen5_80"
  }     
  sku_name = length(var.sku_name) == 0 ? local.instance_types[var.cores] : var.sku_name  
}

resource "azurerm_sql_database" "azure_sql_db" {
  name                = var.db_name
  resource_group_name = data.azurerm_sql_server.azure_sql_db_server.resource_group_name
  location            = data.azurerm_sql_server.azure_sql_db_server.location
  server_name         = data.azurerm_sql_server.azure_sql_db_server.name
  requested_service_objective_name = local.sku_name
  max_size_bytes      = var.max_storage_gb * 1024 * 1024 * 1024
  tags                = var.labels
}

locals {
  serverFQDN = data.azurerm_sql_server.azure_sql_db_server.fqdn
}

output "sqldbName" {value = azurerm_sql_database.azure_sql_db.name}
output "sqlServerName" {value = data.azurerm_sql_server.azure_sql_db_server.name}
output "sqlServerFullyQualifiedDomainName" {value = local.serverFQDN}
output "hostname" {value = local.serverFQDN}
output "port" {value = 1433}
output "name" {value = azurerm_sql_database.azure_sql_db.name}
output "username" {value = var.server_credentials[var.server].admin_username}
output "password" {value = var.server_credentials[var.server].admin_password}

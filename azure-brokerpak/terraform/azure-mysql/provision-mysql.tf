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
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable db_name { type = string }
variable mysql_version { type = string }
variable location { type = string }
variable labels { type = map }
variable cores {type = string }
variable sku_name { type = string }
variable storage_gb {type = string }
variable authorized_network {type = string}
variable use_tls { type = bool }

provider "azurerm" {
  version = "=2.9.0"
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id  
}

locals {
  instance_types = {
    1 = "B_Gen5_1"
    2 = "B_Gen5_2"
    4 = "GP_Gen5_4"
    8 = "MO_Gen5_8"
    16 = "MO_Gen5_16"
    32 = "MO_Gen5_32"
    64 = "GP_Gen5_64"
  }     
  sku_name = length(var.sku_name) == 0 ? local.instance_types[var.cores] : var.sku_name    
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
}

resource "azurerm_resource_group" "azure-msyql" {
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

resource "random_string" "servername" {
  length = 8
  special = false
}

resource "random_password" "password" {
  length = 31
  override_special = "~_-."
  min_upper = 2
  min_lower = 2
  min_special = 2
}

resource "azurerm_mysql_server" "instance" {
  depends_on = [ azurerm_resource_group.azure-msyql ]
  name                = lower(random_string.servername.result)
  location            = var.location
  resource_group_name = local.resource_group
  sku_name = local.sku_name

  storage_profile {
    storage_mb            = var.storage_gb * 1024
    backup_retention_days = 7
    geo_redundant_backup  = "Disabled"
  }

  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  version                      = var.mysql_version
  ssl_enforcement              = var.use_tls ? "Enabled" : "Disabled"
  tags                         = var.labels
}

resource "azurerm_mysql_database" "instance-db" {
  name                = var.db_name
  resource_group_name = local.resource_group
  server_name         = azurerm_mysql_server.instance.name
  charset             = "utf8"
  collation           = "utf8_unicode_ci"
}

resource "azurerm_mysql_virtual_network_rule" "allow_subnet_id" {
  name                = format("subnetrule-%s", lower(random_string.servername.result))
  resource_group_name = local.resource_group
  server_name         = azurerm_mysql_server.instance.name
  subnet_id           = var.authorized_network
  count = var.authorized_network != "default" ? 1 : 0      
}

resource "azurerm_mysql_firewall_rule" "allow_azure" {
  name                = format("firewall-%s", lower(random_string.servername.result))
  resource_group_name = local.resource_group
  server_name         = azurerm_mysql_server.instance.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count = var.authorized_network == "default" ? 1 : 0
}    

output name { value = azurerm_mysql_database.instance-db.name }
output hostname { value = azurerm_mysql_server.instance.fqdn }
output port { value = 3306 }
output username { value = format( "%s@%s", azurerm_mysql_server.instance.administrator_login, azurerm_mysql_server.instance.name ) }
output password { value = azurerm_mysql_server.instance.administrator_login_password }
output use_tls { value = var.use_tls }
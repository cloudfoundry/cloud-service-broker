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
variable resource_group { type = string }
variable db_name { type = string }
variable location { type = string }
variable failover_location { type = string }
variable labels { type = map }
variable sku_name { type = string }
variable cores { type = number }
variable max_storage_gb { type = number }
variable authorized_network {type = string}
variable skip_provider_registration { type = bool }
variable read_write_endpoint_failover_policy { type = string }
variable failover_grace_minutes { type = number }

provider "azurerm" {
  version = "~> 2.31.0"
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id  

  skip_provider_registration = var.skip_provider_registration
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
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
}

resource "azurerm_resource_group" "azure-sql-fog" {
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

resource "azurerm_sql_server" "primary_azure_sql_db_server" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                         = format("%s-primary", var.instance_name)
  resource_group_name          = local.resource_group
  location                     = var.location
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags = var.labels
}

locals {
  default_pair = {
    // https://docs.microsoft.com/en-us/azure/best-practices-availability-paired-regions
    "eastasia" = "southeastasia"
    "southeastasia" = "eastasia"
    "centralus" = "eastus2"
    "eastus" = "westus"
    "eastus2" = "centralus"
    "westus" = "eastus"
    "northcentralus" = "southcentralus"
    "southcentralus" = "northcentralus"
    "northeurope" = "westeurope"
    "westeurope" = "northeurope"
    "japanwest" = "japaneast"
    "japaneast" = "japanwest"
    "brazilsouth" = "southcentralus"
    "australiaeast" = "australiasoutheast"
    "australiasoutheast" = "australiaeast"
    "australiacentral" = "australiacentral2"
    "australiacentral2" = "australiacentral"
    "southindia" = "centralindia"
    "centralindia" = "southindia"
    "westindia" = "southindia"
    "canadacentral" = "canadaeast"
    "canadaeast" = "canadacentral"
    "uksouth" = "ukwest"
    "ukwest" = "uksouth"
    "westcentralus" = "westus2"
    "westus2" = "westcentralus"
    "koreacentral" = "koreasouth"
    "koreasouth" = "koreacentral"
    "francecentral" = "francesouth"
    "francesouth" = "francecentral"
    "uaenorth" = "uaecentral"
    "uaecentral" = "uaenorth"
    "southafricanorth" = "southafricawest" 
    "southafricawest" = "southafricanorth"
    "germanycentral" = "germanynortheast"
    "germanynortheast" = "germanycentral"
  }   
}

resource "azurerm_sql_server" "secondary_sql_db_server" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                         = format("%s-secondary", var.instance_name)
  resource_group_name          = local.resource_group
  location                     = var.failover_location != "default" ? var.location : local.default_pair[var.location]
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags                         = var.labels
}

resource "azurerm_sql_database" "azure_sql_db" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                = var.db_name
  resource_group_name = local.resource_group
  location            = var.location
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  requested_service_objective_name = local.sku_name
  max_size_bytes      = var.max_storage_gb * 1024 * 1024 * 1024
  tags                = var.labels
}

resource "azurerm_sql_failover_group" "failover_group" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                = var.instance_name
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  databases           = [azurerm_sql_database.azure_sql_db.id]
  partner_servers {
    id = azurerm_sql_server.secondary_sql_db_server.id
  }

  read_write_endpoint_failover_policy {
    mode          = var.read_write_endpoint_failover_policy
    grace_minutes = var.failover_grace_minutes
  }
}

resource "azurerm_sql_virtual_network_rule" "allow_subnet_id1" {
  name                = format("subnetrule1-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  subnet_id           = var.authorized_network
  count = var.authorized_network != "default" ? 1 : 0   
}

resource "azurerm_sql_virtual_network_rule" "allow_subnet_id2" {
  name                = format("subnetrule2-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.secondary_sql_db_server.name
  subnet_id           = var.authorized_network
  count = var.authorized_network != "default" ? 1 : 0   
}

resource "azurerm_sql_firewall_rule" "server1" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                = format("firewallrule1-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count = var.authorized_network == "default" ? 1 : 0  
}

resource "azurerm_sql_firewall_rule" "server2" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                = format("firewallrule2-%s", lower(var.instance_name))
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.secondary_sql_db_server.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count = var.authorized_network == "default" ? 1 : 0  
}

locals {
    serverFQDN = format("%s.database.windows.net", azurerm_sql_failover_group.failover_group.name)
}
output sqldbName {value = azurerm_sql_database.azure_sql_db.name}
output sqlServerName {value = azurerm_sql_failover_group.failover_group.name}
output sqlServerFullyQualifiedDomainName {value = local.serverFQDN}
output hostname {value = local.serverFQDN}
output port {value = 1433}
output name {value = azurerm_sql_database.azure_sql_db.name}
output username {value = random_string.username.result}
output password {value = random_password.password.result}
output status {value = format("created failover group %s (id: %s), primary db %s (id: %s) on server %s (id: %s), secondary db %s (id: %s/databases/%s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
                              azurerm_sql_failover_group.failover_group.name, azurerm_sql_failover_group.failover_group.id,
                              azurerm_sql_database.azure_sql_db.name, azurerm_sql_database.azure_sql_db.id,
                              azurerm_sql_server.primary_azure_sql_db_server.name, azurerm_sql_server.primary_azure_sql_db_server.id,
                              azurerm_sql_database.azure_sql_db.name, azurerm_sql_server.secondary_sql_db_server.id, azurerm_sql_database.azure_sql_db.name,
                              azurerm_sql_server.secondary_sql_db_server.name, azurerm_sql_server.secondary_sql_db_server.id,
                              var.azure_tenant_id,
                              azurerm_sql_server.primary_azure_sql_db_server.id)}

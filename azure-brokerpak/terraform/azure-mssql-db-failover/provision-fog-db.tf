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
variable sku_name { type = string }
variable cores { type = number }
variable max_storage_gb { type = number }
variable skip_provider_registration { type = bool }
variable existing { type = bool }
variable read_write_endpoint_failover_policy { type = string }
variable failover_grace_minutes { type = number }

provider "azurerm" {
  version = "~> 2.20.0"
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id  

  skip_provider_registration = var.skip_provider_registration
}

data "azurerm_sql_server" "primary_sql_db_server" {
  name                         = var.server_credential_pairs[var.server_pair].primary.server_name
  resource_group_name          = var.server_credential_pairs[var.server_pair].primary.resource_group
}

data "azurerm_sql_server" "secondary_sql_db_server" {
  name                         = var.server_credential_pairs[var.server_pair].secondary.server_name
  resource_group_name          = var.server_credential_pairs[var.server_pair].secondary.resource_group
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

resource "azurerm_sql_database" "azure_sql_db_primary" {
  name                = var.db_name
  resource_group_name = var.server_credential_pairs[var.server_pair].primary.resource_group
  location            = data.azurerm_sql_server.primary_sql_db_server.location
  server_name         = data.azurerm_sql_server.primary_sql_db_server.name
  requested_service_objective_name = local.sku_name
  max_size_bytes      = var.max_storage_gb * 1024 * 1024 * 1024
  tags                = var.labels
  count = var.existing ? 0 : 1
}

resource "azurerm_sql_failover_group" "failover_group" {
  name                = var.instance_name
  resource_group_name = var.server_credential_pairs[var.server_pair].primary.resource_group
  server_name         = data.azurerm_sql_server.primary_sql_db_server.name
  databases           = [azurerm_sql_database.azure_sql_db_primary[0].id]
  partner_servers {
    id = data.azurerm_sql_server.secondary_sql_db_server.id
  }

  read_write_endpoint_failover_policy {
    mode          = var.read_write_endpoint_failover_policy
    grace_minutes = var.failover_grace_minutes
  }
  count = var.existing ? 0 : 1
}

locals {
  serverFQDN = format("%s.database.windows.net", var.instance_name)
}

output sqldbName {value = var.db_name}
output sqlServerName {value = var.instance_name}
output sqlServerFullyQualifiedDomainName {value = local.serverFQDN}
output hostname {value = local.serverFQDN}
output port {value = 1433}
output name {value = var.db_name}
output username {value = var.server_credential_pairs[var.server_pair].admin_username}
output password {value = var.server_credential_pairs[var.server_pair].admin_password}
output status {
    value = var.existing ? format("connected to existing failover group - primary server %s (id: %s) secondary server %s (%s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
                                              data.azurerm_sql_server.primary_sql_db_server.name, data.azurerm_sql_server.primary_sql_db_server.id,
                                              data.azurerm_sql_server.secondary_sql_db_server.name, data.azurerm_sql_server.secondary_sql_db_server.id,
                                              var.azure_tenant_id,
                                              data.azurerm_sql_server.primary_sql_db_server.id) : format("created failover group %s (id: %s), primary db %s (id: %s) on server %s (id: %s), secondary db %s (id: %s/databases/%s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
                                              azurerm_sql_failover_group.failover_group[0].name, azurerm_sql_failover_group.failover_group[0].id,
                                              azurerm_sql_database.azure_sql_db_primary[0].name, azurerm_sql_database.azure_sql_db_primary[0].id,
                                              data.azurerm_sql_server.primary_sql_db_server.name, data.azurerm_sql_server.primary_sql_db_server.id,
                                              azurerm_sql_database.azure_sql_db_primary[0].name, data.azurerm_sql_server.secondary_sql_db_server.id, azurerm_sql_database.azure_sql_db_primary[0].name,
                                              data.azurerm_sql_server.secondary_sql_db_server.name, data.azurerm_sql_server.secondary_sql_db_server.id,
                                              var.azure_tenant_id,
                                              data.azurerm_sql_server.primary_sql_db_server.id)
}
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

variable storage_account_type { type = string }
variable tier { type = string }
variable replication_type { type = string }
variable location { type = string }
variable labels { type = map }
variable resource_group { type = string }
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
# variable authorized_network {type = string}
variable skip_provider_registration { type = bool }
variable authorized_networks { type = list(string) }

provider "azurerm" {
  version = "~> 2.31.0"
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id  

  skip_provider_registration = var.skip_provider_registration
}

resource "random_string" "account_name" {
  length = 24
  special = false
  upper = false
}

locals {
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", random_string.account_name.result) : var.resource_group
}

resource "azurerm_resource_group" "azure-storage" {
  name     = local.resource_group
  location = var.location
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0
}

resource "azurerm_storage_account" "account" {
  depends_on = [ azurerm_resource_group.azure-storage ]
  name                     = random_string.account_name.result
  resource_group_name      = local.resource_group
  location                 = var.location
  account_tier             = var.tier
  account_replication_type = var.replication_type
  account_kind = var.storage_account_type

  tags = var.labels
}

resource "azurerm_storage_account_network_rules" "account_network_rule" {
  count = length(var.authorized_networks) != 0 ? 1 : 0

  resource_group_name  = local.resource_group
  storage_account_name = azurerm_storage_account.account.name

  default_action             = "Deny"
  virtual_network_subnet_ids = var.authorized_networks[*]
}

output primary_access_key { value = azurerm_storage_account.account.primary_access_key }
output secondary_access_key { value = azurerm_storage_account.account.secondary_access_key }
output storage_account_name { value = azurerm_storage_account.account.name }
output status { value = format("created storage account %s (id: %s) URL:  https://portal.azure.com/#@%s/resource%s",
                               azurerm_storage_account.account.name,
                               azurerm_storage_account.account.id,
                               var.azure_tenant_id,
                               azurerm_storage_account.account.id)}
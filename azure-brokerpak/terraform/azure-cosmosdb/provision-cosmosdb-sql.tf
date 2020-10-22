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
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable failover_locations {type = list(string) }
variable location { type = string }
variable ip_range_filter { type = string }
variable request_units { type = number }
variable enable_automatic_failover { type = bool }
variable enable_multiple_write_locations { type = bool }
variable consistency_level { type = string }
variable max_interval_in_seconds { type = number }
variable max_staleness_prefix {	type= number }
variable labels { type = map }
variable skip_provider_registration { type = bool }

provider "azurerm" {
  version = "~> 2.33.0"
  features {}

  subscription_id = var.azure_subscription_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
  tenant_id       = var.azure_tenant_id  

  skip_provider_registration = var.skip_provider_registration
}

locals {
	resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
}

resource "azurerm_resource_group" "rg" {
	name     = local.resource_group
	location = var.location
	tags     = var.labels
	count    = length(var.resource_group) == 0 ? 1 : 0
}

resource "azurerm_cosmosdb_account" "cosmosdb-account" {
	depends_on = [ azurerm_resource_group.rg ]	
	name                = var.instance_name
	location            = var.location
	resource_group_name = local.resource_group
	offer_type          = "Standard"
	kind                = "GlobalDocumentDB"

	consistency_policy {
		consistency_level       = var.consistency_level
		max_interval_in_seconds = var.max_interval_in_seconds
		max_staleness_prefix    = var.max_staleness_prefix
	}

	dynamic "geo_location" {
		for_each = var.failover_locations
		content {
				location = geo_location.value
				failover_priority = index(var.failover_locations,geo_location.value)
		}
	}

	enable_automatic_failover       = var.enable_automatic_failover
	enable_multiple_write_locations = var.enable_multiple_write_locations
	ip_range_filter                 = var.ip_range_filter
	tags                            = var.labels	
}

resource "azurerm_cosmosdb_sql_database" "db" {
  name                = var.db_name
  resource_group_name = azurerm_cosmosdb_account.cosmosdb-account.resource_group_name
  account_name        = azurerm_cosmosdb_account.cosmosdb-account.name
  throughput          = var.request_units
}

output cosmosdb_host_endpoint {value = azurerm_cosmosdb_account.cosmosdb-account.endpoint }
output cosmosdb_master_key {value = azurerm_cosmosdb_account.cosmosdb-account.primary_master_key }
output cosmosdb_readonly_master_key {value = azurerm_cosmosdb_account.cosmosdb-account.primary_readonly_master_key }
output cosmosdb_database_id { value = azurerm_cosmosdb_sql_database.db.name }
output status { value = format("created account %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s",
                               azurerm_cosmosdb_account.cosmosdb-account.name,
                               azurerm_cosmosdb_account.cosmosdb-account.id,
                               var.azure_tenant_id,
                               azurerm_cosmosdb_account.cosmosdb-account.id )}

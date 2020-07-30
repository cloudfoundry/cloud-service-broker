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

variable resource_group { type = string }
variable instance_name { type = string }
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable account_name { type = string }
variable db_name { type = string }
variable collection_name { type = string }
variable request_units {type = number }
variable failover_locations {type = list(string) }
variable location { type = string }
variable shard_key { type = string }
variable ip_range_filter { type = string }
variable enable_automatic_failover { type = bool }
variable enable_multiple_write_locations { type = bool }
variable consistency_level { type = string }
variable max_interval_in_seconds { type = number }
variable max_staleness_prefix {	type= number }
variable labels { type = map }
variable skip_provider_registration { type = bool }

provider "azurerm" {
  version = "~> 2.20.0"
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

resource "azurerm_cosmosdb_account" "mongo-account" {
	depends_on = [ azurerm_resource_group.rg ]	
	name                = var.account_name
	location            = var.location
	resource_group_name = local.resource_group
	offer_type          = "Standard"
	kind                = "MongoDB"

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

resource "azurerm_cosmosdb_mongo_database" "mongo-db" {
	name                = var.db_name
	resource_group_name = azurerm_cosmosdb_account.mongo-account.resource_group_name
	account_name        = azurerm_cosmosdb_account.mongo-account.name
	throughput          = var.request_units
}

resource "azurerm_cosmosdb_mongo_collection" "mongo-collection" {
	name                = var.collection_name
	resource_group_name = azurerm_cosmosdb_account.mongo-account.resource_group_name
	account_name        = azurerm_cosmosdb_account.mongo-account.name
	database_name       = azurerm_cosmosdb_mongo_database.mongo-db.name

	default_ttl_seconds = "777"
	shard_key           = var.shard_key
}

output uri { value = replace(azurerm_cosmosdb_account.mongo-account.connection_strings[0], "/?", "/${azurerm_cosmosdb_mongo_database.mongo-db.name}?")  }
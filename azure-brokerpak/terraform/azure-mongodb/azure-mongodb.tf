variable instance_prefix { type = string }
variable instance_name { type = string }
variable db_name { type = string }
variable collection_name { type = string }
variable request_units {type = number }
variable failover_locations {type = list(string) }
variable region { type = string }
variable shard_key { type = string }
variable ip_range_filter { type = string }
variable enable_automatic_failover { 
	type = bool
	default = false
}
variable enable_multiple_write_locations {
	type = bool
	default = false
}
variable consistency_level {
	type = string
	default = "Session"
}
variable max_interval_in_seconds {
	type = number
	default = 5
}
variable max_staleness_prefix {
	type= number
	default = 100
}

variable labels { type = map }



resource "random_string" "account_id" {
	upper = false
	special = false
	lower = true
	number = true
	length = 12
}

resource "azurerm_resource_group" "rg" {
	name     = coalesce(var.instance_name, "${var.instance_prefix}-${random_string.account_id.result}")
	location = var.region
	tags     = var.labels
}

resource "azurerm_cosmosdb_account" "mongo-account" {
	name                = coalesce(var.instance_name, "${var.instance_prefix}-${random_string.account_id.result}")
	location            = azurerm_resource_group.rg.location
	resource_group_name = azurerm_resource_group.rg.name
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

	enable_automatic_failover = var.enable_automatic_failover
	enable_multiple_write_locations = var.enable_multiple_write_locations
    ip_range_filter = var.ip_range_filter
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
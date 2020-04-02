variable instance_name {
	type = string
}

variable resource_group {
	type = string
}

variable location {
	type = string
}

variable sku {
	type = string
}

variable auto_inflate_enabled {
  	type = bool
}

variable partition_count {
	type = number
}

variable message_retention {
	type = number
}

variable labels {
	type = map
}

locals{
	tags = merge (var.labels,{"heritage": "cloud-service-broker"})
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

resource "azurerm_eventhub_namespace" "rg-namespace" {
	depends_on = [ azurerm_resource_group.rg ]	
	name                 = var.instance_name
	location             = var.location
	resource_group_name  = local.resource_group
	sku                  = var.sku
	capacity             = 1
	auto_inflate_enabled = var.auto_inflate_enabled
	tags                 = local.tags
}

resource "azurerm_eventhub" "eventhub" {
	name                = var.instance_name
	namespace_name      = azurerm_eventhub_namespace.rg-namespace.name
	resource_group_name = local.resource_group
	partition_count     = var.partition_count
	message_retention   = var.message_retention
}

output eventhub_rg_name {value=local.resource_group}
output namespace_name {value=azurerm_eventhub_namespace.rg-namespace.name}
output eventhub_name {value=azurerm_eventhub.eventhub.name}


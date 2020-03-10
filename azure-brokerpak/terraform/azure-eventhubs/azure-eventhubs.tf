variable prefix {
	 type = string
 }

variable location {
	type = string
 }

 variable sku {
	 type = string
 }
variable eventhub_name {
	type = string
}
variable namespace_name {
	type = string
	 default="pcf-namespace"
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
	default = {}
}


locals{
tags = merge (var.labels,{"heritage": "cloud-service-broker"})
}


resource "random_string" "eventhub_id" {
	upper = false
	special = false
	lower = true
	number = true
	length = 12
}

resource "random_string" "namespace_id" {
	upper = false
	special = false
	lower = true
	number = true
	length = 12
}

resource "azurerm_resource_group" "rg" {
  name     =  "${var.prefix}-${random_string.namespace_id.result}"
	location = var.location
}

resource "azurerm_eventhub_namespace" "rg-namespace" {
  name                = coalesce(var.namespace_name, "${var.prefix}-${random_string.namespace_id.result}")
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  sku                 = var.sku
  capacity            = 1
  auto_inflate_enabled = var.auto_inflate_enabled
  tags = local.tags
}

resource "azurerm_eventhub" "eventhub" {
  name                = coalesce(var.eventhub_name, "${var.prefix}-${random_string.eventhub_id.result}")
  namespace_name      = azurerm_eventhub_namespace.rg-namespace.name
  resource_group_name = azurerm_resource_group.rg.name
  partition_count     = var.partition_count
  message_retention   = var.message_retention
}


output eventhub_rg_name {value=azurerm_resource_group.rg.name}
output namespace_name {value=azurerm_eventhub_namespace.rg-namespace.name}
output eventhub_name {value=azurerm_eventhub.eventhub.name}


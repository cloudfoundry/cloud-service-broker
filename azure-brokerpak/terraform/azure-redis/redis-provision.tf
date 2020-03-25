variable resource_group { type = string }
variable sku_name { type = string }
variable family { type = string }
variable capacity { type = string }
variable instance_name { type = string }
variable region { type = string }
variable labels { type = map }

locals {
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
}

resource "azurerm_resource_group" "azure-redis" {
  name     = local.resource_group
  location = var.region
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0
}

resource "azurerm_redis_cache" "redis" {
  depends_on = [ azurerm_resource_group.azure-redis ]  
  name                = var.instance_name
  sku_name            = var.sku_name
  family              = var.family
  capacity            = var.capacity
  location            = var.region
  resource_group_name = local.resource_group
  tags                = var.labels
}

output name { value = azurerm_redis_cache.redis.name }
output hostname { value = azurerm_redis_cache.redis.hostname }
output port { value = azurerm_redis_cache.redis.port }
output password { value = azurerm_redis_cache.redis.primary_access_key }
output tls_port { value = azurerm_redis_cache.redis.ssl_port }
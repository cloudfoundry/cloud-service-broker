variable resource_group { type = string }
variable sku_name { type = string }
variable family { type = string }
variable capacity { type = string }
variable instance_name { type = string }
variable region { type = string }
variable labels { type = map }

resource "azurerm_resource_group" "azure-redis" {
  name     = var.resource_group
  location = var.region
  tags     = var.labels
}

resource "azurerm_redis_cache" "redis" {
  name                = var.instance_name
  sku_name            = var.sku_name
  family              = var.family
  capacity            = var.capacity
  location            = azurerm_resource_group.azure-redis.location
  resource_group_name = azurerm_resource_group.azure-redis.name
  tags                = var.labels
}

output name { value = "${azurerm_redis_cache.redis.name}" }
output hostname { value = "${azurerm_redis_cache.redis.hostname}" }
output port { value = azurerm_redis_cache.redis.port }
output password { value = "${azurerm_redis_cache.redis.primary_access_key}" }
output tls_port { value = "${azurerm_redis_cache.redis.ssl_port}" }
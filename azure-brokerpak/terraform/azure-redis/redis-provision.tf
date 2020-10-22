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
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable sku_name { type = string }
variable family { type = string }
variable capacity { type = string }
variable instance_name { type = string }
variable location { type = string }
variable labels { type = map }
variable skip_provider_registration { type = bool }
variable tls_min_version { type = string }
variable maxmemory_policy { type = string }

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

resource "azurerm_resource_group" "azure-redis" {
  name     = local.resource_group
  location = var.location
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0
}

resource "azurerm_redis_cache" "redis" {
  depends_on = [ azurerm_resource_group.azure-redis ]  
  name                = var.instance_name
  sku_name            = var.sku_name
  family              = var.family
  capacity            = var.capacity
  location            = var.location
  resource_group_name = local.resource_group
  minimum_tls_version = length(var.tls_min_version) == 0 ? "1.2" : var.tls_min_version
  tags                = var.labels
  redis_configuration {
    maxmemory_policy   = length(var.maxmemory_policy) == 0 ? "allkeys-lru" : var.maxmemory_policy
  }
}

output name { value = azurerm_redis_cache.redis.name }
output host { value = azurerm_redis_cache.redis.hostname }
# output port { value = azurerm_redis_cache.redis.port }
output password { value = azurerm_redis_cache.redis.primary_access_key }
output tls_port { value = azurerm_redis_cache.redis.ssl_port }
output status { value = format("created cache %s (id: %s) URL: URL: https://portal.azure.com/#@%s/resource%s",
                               azurerm_redis_cache.redis.name,
                               azurerm_redis_cache.redis.id,
                               var.azure_tenant_id,
                               azurerm_redis_cache.redis.id)}
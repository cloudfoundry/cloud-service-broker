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
variable sku_name { type = string }
variable family { type = string }
variable capacity { type = string }
variable instance_name { type = string }
variable location { type = string }
variable labels { type = map }

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
  tags                = var.labels
}

output name { value = azurerm_redis_cache.redis.name }
output hostname { value = azurerm_redis_cache.redis.hostname }
output port { value = azurerm_redis_cache.redis.port }
output password { value = azurerm_redis_cache.redis.primary_access_key }
output tls_port { value = azurerm_redis_cache.redis.ssl_port }
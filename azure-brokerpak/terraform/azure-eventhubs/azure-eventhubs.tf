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
variable azure_tenant_id { type = string }
variable azure_subscription_id { type = string }
variable azure_client_id { type = string }
variable azure_client_secret { type = string }
variable resource_group { type = string }
variable location {	type = string }
variable sku { type = string }
variable auto_inflate_enabled { type = bool }
variable partition_count { type = number }
variable message_retention { type = number }
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
output status {value=format("created event hub %s (id: %s)  URL: https://portal.azure.com/#@%s/resource%s",
               azurerm_eventhub.eventhub.name,
               azurerm_eventhub.eventhub.id,
               var.azure_tenant_id,
               azurerm_eventhub.eventhub.id)}


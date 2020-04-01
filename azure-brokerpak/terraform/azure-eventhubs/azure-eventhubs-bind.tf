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

variable eventhub_rg_name { type = string }
variable namespace_name { type = string }
variable eventhub_name { type = string }

data "azurerm_eventhub_namespace" "ns" {
  name                = var.namespace_name
  resource_group_name = var.eventhub_rg_name
}

output "event_hub_connection_string" {
  value = "${data.azurerm_eventhub_namespace.ns.default_primary_connection_string};EntityPath=${var.eventhub_name}"
}

output "event_hub_name" {
  value = var.eventhub_name
}

output "namespace_connection_string" {
  value = data.azurerm_eventhub_namespace.ns.default_primary_connection_string
}

output "namespace_name" {
  value = data.azurerm_eventhub_namespace.ns.name
}

output "shared_access_key_name" {
  value = "RootManageSharedAccessKey"
}

output "shared_access_key_value" {
  value = data.azurerm_eventhub_namespace.ns.default_primary_key
}


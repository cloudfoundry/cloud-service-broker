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

resource "azurerm_mssql_database" "primary_db" {
  name                = var.db_name
  server_id           = data.azurerm_sql_server.primary_sql_db_server.id
  sku_name            = local.sku_name
  max_size_gb         = var.max_storage_gb
  tags                = var.labels
  count = var.existing ? 0 : null
  short_term_retention_policy {
    retention_days = var.short_term_retention_days
  }
}

resource "azurerm_sql_failover_group" "failover_group" {
  name                = var.instance_name
  resource_group_name = var.server_credential_pairs[var.server_pair].primary.resource_group
  server_name         = data.azurerm_sql_server.primary_sql_db_server.name
  databases           = [azurerm_mssql_database.primary_db[0].id]
  partner_servers {
    id = data.azurerm_sql_server.secondary_sql_db_server.id
  }

  read_write_endpoint_failover_policy {
    mode          = var.read_write_endpoint_failover_policy
    grace_minutes = var.failover_grace_minutes
  }
  count = var.existing ? 0 : null
}

resource "azurerm_mssql_database" "secondary_db" {
  count = var.subsume ? null : 0
}
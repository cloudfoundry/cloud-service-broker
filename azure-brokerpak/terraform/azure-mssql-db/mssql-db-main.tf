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

resource "azurerm_mssql_database" "azure_sql_db" {
  name                = var.db_name
  server_id           = data.azurerm_sql_server.azure_sql_db_server.id
  sku_name            = local.sku_name
  max_size_gb         = var.max_storage_gb
  tags                = var.labels
  short_term_retention_policy {
    retention_days = var.short_term_retention_days
  }
}
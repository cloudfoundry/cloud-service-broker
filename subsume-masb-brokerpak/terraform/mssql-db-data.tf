
# data "azurerm_sql_server" "azure_sql_db_server" {
#   name                         = var.server_name
#   resource_group_name          = var.server_resource_group
# }

data "azurerm_sql_server" "azure_sql_db_server" {
  name                         = azurerm_sql_database.azure_sql_db.server_name
  resource_group_name          = azurerm_sql_database.azure_sql_db.resource_group_name
}
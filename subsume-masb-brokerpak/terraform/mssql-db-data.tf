
data "azurerm_sql_server" "azure_sql_db_server" {
  name                         = var.server_name
  resource_group_name          = var.server_resource_group
}
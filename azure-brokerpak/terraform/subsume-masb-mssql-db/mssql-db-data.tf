
data "azurerm_sql_server" "azure_sql_db_server" {
  name                         = var.server_credentials[var.server].server_name
  resource_group_name          = var.server_credentials[var.server].server_resource_group
}

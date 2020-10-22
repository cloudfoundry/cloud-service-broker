
data "azurerm_sql_server" "primary_sql_db_server" {
  name                         = var.server_credential_pairs[var.server_pair].primary.server_name
  resource_group_name          = var.server_credential_pairs[var.server_pair].primary.resource_group
}

data "azurerm_sql_server" "secondary_sql_db_server" {
  name                         = var.server_credential_pairs[var.server_pair].secondary.server_name
  resource_group_name          = var.server_credential_pairs[var.server_pair].secondary.resource_group
}

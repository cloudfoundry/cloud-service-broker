variable server_name { type = string }
variable region { type = string }
variable labels { type = map }

resource "random_string" "username" {
  length = 16
  special = false
  number = false
}

resource "random_password" "password" {
  length = 64
  override_special = "~_-."
  min_upper = 2
  min_lower = 2
  min_special = 2
}

resource "azurerm_resource_group" "azure_sql" {
  name     = var.server_name
  location = var.region
  tags = var.labels
}

resource "azurerm_sql_server" "azure_sql_db_server" {
  name                         = var.server_name
  resource_group_name          = azurerm_resource_group.azure_sql.name
  location                     = var.region
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags = var.labels
}

resource "azurerm_sql_firewall_rule" "example" {
  name                = "FirewallRule1"
  resource_group_name = azurerm_resource_group.azure_sql.name
  server_name         = azurerm_sql_server.azure_sql_db_server.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
}

output "sqldbResourceGroup" {value = "${azurerm_sql_server.azure_sql_db_server.resource_group_name}"}
output "sqlServerName" {value = "${azurerm_sql_server.azure_sql_db_server.name}"}
output "sqlServerFullyQualifiedDomainName" {value = "${azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name}"}
output "hostname" {value = "${azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name}"}
output "port" {value = 1433}
output "username" {value = "${random_string.username.result}"}
output "password" {value = "${random_password.password.result}"}
output "databaseLogin" {value = random_string.username.result}
output "databaseLoginPassword" {value = random_password.password.result}


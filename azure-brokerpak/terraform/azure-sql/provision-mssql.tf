variable server_name { type = string }
variable db_name { type = string }
variable region { type = string }
variable labels { type = map }

resource "random_string" "username" {
  length = 16
  special = false
}

resource "random_password" "password" {
  length = 16
  special = true
  override_special = "_@"
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

resource "azurerm_sql_database" "azure_sql_db" {
  name                = var.db_name
  resource_group_name = azurerm_resource_group.azure_sql.name
  location            = var.region
  server_name         = azurerm_sql_server.azure_sql_db_server.name
  tags = var.labels
}

output "sqldbName" {value = "${azurerm_sql_database.azure_sql_db.name}"}
output "sqlServerName" {value = "${azurerm_sql_server.azure_sql_db_server.name}"}
output "sqlServerFullyQualifiedDomainName" {value = "${azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name}"}
output "databaseLogin" {value = "${random_string.username.result}"}
output "databaseLoginPassword" {value = "${random_password.password.result}"}
output "jdbcUrl" {value = format("jdbc:sqlserver://%s:1433;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30", azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name, azurerm_sql_database.azure_sql_db.name, random_string.username.result, random_password.password.result)}
output "jdbcUrlForAuditingEnabled" {value = format("jdbc:sqlserver://%s:1433;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30", azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name, azurerm_sql_database.azure_sql_db.name, random_string.username.result, random_password.password.result)}
output "hostname" {value = "${azurerm_sql_server.azure_sql_db_server.name}"}
output "port" {value = 1433}
output "name" {value = "${azurerm_sql_database.azure_sql_db.name}"}
output "username" {value = "${random_string.username.result}"}
output "password" {value = "${random_password.password.result}"}
output "uri" {value = format("mssql://%s:1433/%s?encrypt=true&TrustServerCertificate=false&HostNameInCertificate=*.database.windows.net", azurerm_sql_server.azure_sql_db_server.fully_qualified_domain_name, azurerm_sql_database.azure_sql_db.name)}
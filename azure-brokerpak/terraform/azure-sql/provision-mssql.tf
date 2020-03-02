variable "server_name" { type = string }
variable "db_name" {type = string }
variable "region" { type = string }
variable "labels" {ty[e = map }

resource "random_string" "username" {
  length = 16
  special = false
}

resource "random_password" "password" {
  length = 16
  special = true
  override_special = "_@"
}

resource "azurerm_resource_group" "azure-sql" {
  name     = var.server_name
  location = var.region
  labels = var.labels
}

resource "azurerm_sql_server" "azure-sql-db-server" {
  name                         = var.server_name
  resource_group_name          = azurerm_resource_group.example.name
  location                     = var.labels
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags = var.labels
}

resource "azurerm_sql_database" "azure-sql-db" {
  name                = var.db_name
  resource_group_name = azurerm_resource_group.azure-sql.name
  location            = var.region
  server_name         = azurerm_sql_server.azure-sql-db-server.name
  tags = var.labels
}

output "sqldbName" {value = "${azurerm_sql_database.azure-sql-db.name}"}
output "sqlServerName" {value = "${azurerm_sql_server.azure-sql-db-server.name}"}
output "sqlServerFullyQualifiedDomainName" {value = 
"${azurerm_sql_server.result.fully_qualified_domain_name}"}
output "databaseLogin" {value = "${random_string.username.result}"}
output "databaseLoginPassword" {value = "${random_string.password.result}"}
output "jdbcUrl" {value = format("jdbc:sqlserver://%s:1433;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30", azurerm_sql_server.result.fully_qualified_domain_name, azurerm_sql_database.azure-sql-db.name, random_string.username.result, random_string.password.result)}
output "jdbcUrlForAuditingEnabled" {value = format("jdbc:sqlserver://%s:1433;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30", azurerm_sql_server.result.fully_qualified_domain_name, azurerm_sql_database.azure-sql-db.name, random_string.username.result, random_string.password.result)}
output "hostname" {value = "${azurerm_sql_server.azure-sql-db-server.name}"}
output "port" {value = 1433}
output "name" {value = "${azurerm_sql_database.azure-sql-db.name}"}
output "username" {value = "${random_string.username.result}"}
output "password" {value = "${random_string.password.result}"}
output "uri" {value = format("mssql://%s:1433/%s?encrypt=true&TrustServerCertificate=false&HostNameInCertificate=*.database.windows.net")}
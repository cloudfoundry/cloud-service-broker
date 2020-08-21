locals {
  serverFQDN = data.azurerm_sql_server.azure_sql_db_server.fqdn
}

output "sqldbName" {value = azurerm_sql_database.azure_sql_db.name}
output "sqlServerName" {value = data.azurerm_sql_server.azure_sql_db_server.name}
output "sqlServerFullyQualifiedDomainName" {value = local.serverFQDN}
output "hostname" {value = local.serverFQDN}
output "port" {value = 1433}
output "name" {value = azurerm_sql_database.azure_sql_db.name}
output "username" {value = data.azurerm_sql_server.azure_sql_db_server.administrator_login }
output "password" {value = var.admin_password}
locals {
  serverFQDN = data.azurerm_sql_server.azure_sql_db_server.fqdn
}

output sqldbName {value = azurerm_mssql_database.azure_sql_db.name}
output sqlServerName {value = data.azurerm_sql_server.azure_sql_db_server.name}
output sqlServerFullyQualifiedDomainName {value = local.serverFQDN}
output hostname {value = local.serverFQDN}
output port {value = 1433}
output name {value = azurerm_mssql_database.azure_sql_db.name}
output username {value = var.server_credentials[var.server].admin_username}
output password {value = var.server_credentials[var.server].admin_password}
output status {value = format("created db %s (id: %s) URL: URL: https://portal.azure.com/#@%s/resource%s",
                              azurerm_mssql_database.azure_sql_db.name,
                              azurerm_mssql_database.azure_sql_db.id,
                              var.azure_tenant_id,
                              azurerm_mssql_database.azure_sql_db.id)}

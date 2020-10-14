locals {
  serverFQDN = data.azurerm_sql_server.azure_sql_db_server.fqdn
}

output sqldbName {value = azurerm_mssql_database.azure_sql_db.name}
output sqlServerName {value = data.azurerm_sql_server.azure_sql_db_server.name}
output sqlServerFullyQualifiedDomainName {value = local.serverFQDN}
output hostname {value = local.serverFQDN}
output port {value = 1433}
output name {value = azurerm_mssql_database.azure_sql_db.name}
output username {value = data.azurerm_sql_server.azure_sql_db_server.administrator_login }
output password {value = var.admin_password}
output status {value = format("subsumed db %s (id: %s) on server %s (id: %s) edition: %s, service_objective: %s, URL: https://portal.azure.com/#@%s/resource%s",
                              azurerm_mssql_database.azure_sql_db.name, azurerm_mssql_database.azure_sql_db.id,
                              data.azurerm_sql_server.azure_sql_db_server.name, data.azurerm_sql_server.azure_sql_db_server.id,
                              azurerm_mssql_database.azure_sql_db.edition, azurerm_mssql_database.azure_sql_db.requested_service_objective_name,
                              var.azure_tenant_id, azurerm_mssql_database.azure_sql_db.id)}
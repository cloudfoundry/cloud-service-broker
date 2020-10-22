locals {
  serverFQDN = format("%s.database.windows.net", azurerm_sql_failover_group.failover_group.name)
}

output sqldbName {value = azurerm_mssql_database.primary_db.name}
output sqlServerName {value = azurerm_sql_failover_group.failover_group.name}
output sqlServerFullyQualifiedDomainName {value = local.serverFQDN}
output hostname {value = local.serverFQDN}
output port {value = 1433}
output name {value = azurerm_mssql_database.primary_db.name}
output username {value = var.server_credential_pairs[var.server_pair].admin_username}
output password {value = var.server_credential_pairs[var.server_pair].admin_password}
output status {value = format("created failover group %s (id: %s), primary db %s (id: %s) on server %s (id: %s), secondary db %s (id: %s/databases/%s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
                                              azurerm_sql_failover_group.failover_group.name, azurerm_sql_failover_group.failover_group.id,
                                              azurerm_mssql_database.primary_db.name, azurerm_mssql_database.primary_db.id,
                                              data.azurerm_sql_server.primary_sql_db_server.name, data.azurerm_sql_server.primary_sql_db_server.id,
                                              azurerm_mssql_database.primary_db.name, data.azurerm_sql_server.secondary_sql_db_server.id, azurerm_mssql_database.primary_db.name,
                                              data.azurerm_sql_server.secondary_sql_db_server.name, data.azurerm_sql_server.secondary_sql_db_server.id,
                                              var.azure_tenant_id,
                                              data.azurerm_sql_server.primary_sql_db_server.id)}
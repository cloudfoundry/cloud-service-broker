
locals {
  serverFQDN = format("%s.database.windows.net", var.instance_name)
}

output sqldbName {value = var.db_name}
output sqlServerName {value = var.instance_name}
output sqlServerFullyQualifiedDomainName {value = local.serverFQDN}
output hostname {value = local.serverFQDN}
output port {value = 1433}
output name {value = var.db_name}
output username {value = var.server_credential_pairs[var.server_pair].admin_username}
output password {value = var.server_credential_pairs[var.server_pair].admin_password}
output status {
    value = var.existing ? format("connected to existing failover group - primary server %s (id: %s) secondary server %s (%s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
                                              data.azurerm_sql_server.primary_sql_db_server.name, data.azurerm_sql_server.primary_sql_db_server.id,
                                              data.azurerm_sql_server.secondary_sql_db_server.name, data.azurerm_sql_server.secondary_sql_db_server.id,
                                              var.azure_tenant_id,
                                              data.azurerm_sql_server.primary_sql_db_server.id) : format("created failover group %s (id: %s), primary db %s (id: %s) on server %s (id: %s), secondary db %s (id: %s/databases/%s) on server %s (id: %s) URL: https://portal.azure.com/#@%s/resource%s/failoverGroup",
                                              azurerm_sql_failover_group.failover_group[0].name, azurerm_sql_failover_group.failover_group[0].id,
                                              azurerm_sql_failover_group.failover_group[0].name, azurerm_mssql_database.primary_db[0].id,
                                              data.azurerm_sql_server.primary_sql_db_server.name, data.azurerm_sql_server.primary_sql_db_server.id,
                                              azurerm_mssql_database.primary_db[0].name, data.azurerm_sql_server.secondary_sql_db_server.id, azurerm_mssql_database.primary_db[0].name,
                                              data.azurerm_sql_server.secondary_sql_db_server.name, data.azurerm_sql_server.secondary_sql_db_server.id,
                                              var.azure_tenant_id,
                                              data.azurerm_sql_server.primary_sql_db_server.id)
}

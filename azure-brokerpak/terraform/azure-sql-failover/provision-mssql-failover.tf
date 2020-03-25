variable instance_name { type = string }
variable resource_group { type = string }
variable db_name { type = string }
variable region { type = string }
variable failover_region { type = string }
variable labels { type = map }
variable pricing_tier { type = string }
variable cores { type = number }
variable storage_gb { type = number }

locals {
  resource_group = length(var.resource_group) == 0 ? format("rg-%s", var.instance_name) : var.resource_group
}
resource "azurerm_resource_group" "azure-sql-fog" {
  name     = local.resource_group
  location = var.region
  tags     = var.labels
  count    = length(var.resource_group) == 0 ? 1 : 0
}

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

resource "azurerm_sql_server" "primary_azure_sql_db_server" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                         = format("%s-primary", var.instance_name)
  resource_group_name          = local.resource_group
  location                     = var.region
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags = var.labels
}

locals {
  default_pair = {
    // https://docs.microsoft.com/en-us/azure/best-practices-availability-paired-regions
    "eastasia" = "southeastasia"
    "southeastasia" = "eastasia"
    "centralus" = "eastus2"
    "eastus" = "westus"
    "eastus2" = "centralus"
    "westus" = "eastus"
    "northcentralus" = "southcentralus"
    "southcentralus" = "northcentralus"
    "northeurope" = "westeurope"
    "westeurope" = "northeurope"
    "japanwest" = "japaneast"
    "japaneast" = "japanwest"
    "brazilsouth" = "southcentralus"
    "australiaeast" = "australiasoutheast"
    "australiasoutheast" = "australiaeast"
    "australiacentral" = "australiacentral2"
    "australiacentral2" = "australiacentral"
    "southindia" = "centralindia"
    "centralindia" = "southindia"
    "westindia" = "southindia"
    "canadacentral" = "canadaeast"
    "canadaeast" = "canadacentral"
    "uksouth" = "ukwest"
    "ukwest" = "uksouth"
    "westcentralus" = "westus2"
    "westus2" = "westcentralus"
    "koreacentral" = "koreasouth"
    "koreasouth" = "koreacentral"
    "francecentral" = "francesouth"
    "francesouth" = "francecentral"
    "uaenorth" = "uaecentral"
    "uaecentral" = "uaenorth"
    "southafricanorth" = "southafricawest" 
    "southafricawest" = "southafricanorth"
    "germanycentral" = "germanynortheast"
    "germanynortheast" = "germanycentral"
  }   
}

resource "azurerm_sql_server" "secondary_sql_db_server" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                         = format("%s-secondary", var.instance_name)
  resource_group_name          = local.resource_group
  location                     = var.failover_region != "default" ? var.region : local.default_pair[var.region]
  version                      = "12.0"
  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  tags                         = var.labels
}

resource "azurerm_sql_database" "azure_sql_db" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                = var.db_name
  resource_group_name = local.resource_group
  location            = var.region
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  requested_service_objective_name = format("%s_Gen5_%d", var.pricing_tier, var.cores)
  max_size_bytes      = var.storage_gb * 1024 * 1024 * 1024
  tags                = var.labels
}

resource "azurerm_sql_failover_group" "failover_group" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                = var.instance_name
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  databases           = [azurerm_sql_database.azure_sql_db.id]
  partner_servers {
    id = azurerm_sql_server.secondary_sql_db_server.id
  }

  read_write_endpoint_failover_policy {
    mode          = "Automatic"
    grace_minutes = 60
  }
}

resource "azurerm_sql_firewall_rule" "server1" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                = "FirewallRule1"
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.primary_azure_sql_db_server.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
}

resource "azurerm_sql_firewall_rule" "server2" {
  depends_on = [ azurerm_resource_group.azure-sql-fog ]
  name                = "FirewallRule1"
  resource_group_name = local.resource_group
  server_name         = azurerm_sql_server.secondary_sql_db_server.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
}

locals {
    serverFQDN = format("%s.database.windows.net", azurerm_sql_failover_group.failover_group.name)
}
output "sqldbName" {value = azurerm_sql_database.azure_sql_db.name}
output "sqlServerName" {value = azurerm_sql_failover_group.failover_group.name}
output "sqlServerFullyQualifiedDomainName" {value = local.serverFQDN}
output "hostname" {value = local.serverFQDN}
output "port" {value = 1433}
output "name" {value = azurerm_sql_database.azure_sql_db.name}
output "username" {value = random_string.username.result}
output "password" {value = random_password.password.result}

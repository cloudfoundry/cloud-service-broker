variable instance_name { type = string }
variable db_name { type = string }
variable mysql_version { type = string }
variable region { type = string }
variable labels { type = map }
variable pricing_tier { type = string }
variable cores {type = string }
variable storage_gb {type = string }
variable authorized_network {type = string}
resource "azurerm_resource_group" "azure-msyql" {
  name     = var.instance_name
  location = var.region
  tags     = var.labels
}

resource "random_string" "username" {
  length = 16
  special = false
  number = false
}

resource "random_string" "servername" {
  length = 8
  special = false
}

resource "random_password" "password" {
  length = 31
  override_special = "~!$^&()_-+={}[]/<>,.|"
  min_upper = 2
  min_lower = 2
  min_special = 2
}

resource "azurerm_mysql_server" "instance" {
  name                = lower(random_string.servername.result)
  location            = azurerm_resource_group.azure-msyql.location
  resource_group_name = azurerm_resource_group.azure-msyql.name
  sku_name = format("%s_Gen5_%d", var.pricing_tier, var.cores)


  storage_profile {
  storage_mb            = var.storage_gb * 1024
  backup_retention_days = 7
  geo_redundant_backup  = "Disabled"
  }

  administrator_login          = random_string.username.result
  administrator_login_password = random_password.password.result
  version                      = var.mysql_version
  ssl_enforcement              = "Disabled"
  tags                         = var.labels
}

resource "azurerm_mysql_database" "instance-db" {
    name                = var.db_name

    resource_group_name = azurerm_resource_group.azure-msyql.name
    server_name         = azurerm_mysql_server.instance.name
    charset             = "utf8"
    collation           = "utf8_unicode_ci"
}

resource "azurerm_mysql_virtual_network_rule" "allow_subnet_id" {
  name                = format("subnetrule-%s", lower(random_string.servername.result))
  resource_group_name = azurerm_resource_group.azure-msyql.name
  server_name         = azurerm_mysql_server.instance.name
  subnet_id           = var.authorized_network
  count = "${var.authorized_network != "default" ? 1 : 0}"      
}

resource "azurerm_mysql_firewall_rule" "allow_azure" {
  name                = format("firewall-%s", lower(random_string.servername.result))
  resource_group_name = azurerm_resource_group.azure-msyql.name
  server_name         = azurerm_mysql_server.instance.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
  count = "${var.authorized_network == "default" ? 1 : 0}"       
}    

output name { value = "${azurerm_mysql_database.instance-db.name}" }
output hostname { value = "${azurerm_mysql_server.instance.fqdn}" }
output port { value = 3306 }
output username { value = "${azurerm_mysql_server.instance.administrator_login}@${azurerm_mysql_server.instance.name}" }
output password { value = "${azurerm_mysql_server.instance.administrator_login_password}" }
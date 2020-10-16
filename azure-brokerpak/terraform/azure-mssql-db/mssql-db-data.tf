
data "azurerm_sql_server" "azure_sql_db_server" {
  name                         = var.server_credentials[var.server].server_name
  resource_group_name          = var.server_credentials[var.server].server_resource_group
}

locals {
  instance_types = {
    1 = "GP_Gen5_1"
    2 = "GP_Gen5_2"
    4 = "GP_Gen5_4"
    8 = "GP_Gen5_8"
    16 = "GP_Gen5_16"
    32 = "GP_Gen5_32"
    80 = "GP_Gen5_80"
  }     
  sku_name = length(var.sku_name) == 0 ? local.instance_types[var.cores] : var.sku_name  
}

variable credentials  { type = string }
variable project  { type = string }
variable labels { type = map }
variable ddl { type = list }
variable num_nodes { type = string }
variable instance_name { type = string }
variable config { type = string }

provider "google" {
  version = ">=3.17.0"
  credentials = var.credentials
  project     = var.project
  
}

locals {
  display_name = substr(var.instance_name, 4,29)
  db_name = replace(local.display_name ,"-","_")
}


#---------------------------------------------------
# Create google spanner instance
#---------------------------------------------------
resource "google_spanner_instance" "spanner_instance" {

    config          = var.num_nodes > 2 ? var.config : "regional-${var.config}"
    display_name    = local.display_name
    name            = lower(var.instance_name)
    num_nodes       = var.num_nodes
    project         = var.project
    labels          = var.labels

    lifecycle {
        ignore_changes = []
        create_before_destroy = true
    }
}


#---------------------------------------------------
# Create spanner database
#---------------------------------------------------
resource "google_spanner_database" "spanner_database" {


    instance    = google_spanner_instance.spanner_instance.name
    name        = local.db_name
    project     = var.project

    ddl       =  var.ddl

    lifecycle {
        ignore_changes = []
        create_before_destroy = true
    }

}

output instance { value = google_spanner_instance.spanner_instance.name }
output db_name { value =  join(",", google_spanner_database.spanner_database.*.name)}

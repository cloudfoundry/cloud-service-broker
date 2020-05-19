variable credentials  { type = string }
variable project  { type = string }
variable labels { type = map }
variable ddl { type = list }
variable num_nodes { type = string }
variable instance_name { type = string }
variable config { type = string }

provider "google" {
  version = "~> 3.5.0"
  credentials = var.credentials
  project     = var.project
  
}

#---------------------------------------------------
# Create google spanner instance
#---------------------------------------------------
resource "google_spanner_instance" "spanner_instance" {
  
    config          = var.num_nodes > 2 ? var.config : "regional-${var.config}"
    display_name    = var.instance_name
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
    name        = "${var.instance_name}-db"
    project     = var.project
   
    ddl       =  var.ddl

    lifecycle {
        ignore_changes = []
        create_before_destroy = true
    }

}

output instance { value = google_spanner_instance.spanner_instance.name }
output db_name { value = google_spanner_database.spanner_database.name }

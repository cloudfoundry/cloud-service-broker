variable credentials  { type = string }
variable project  { type = string }
variable labels { type = map }
variable region { type = string }
variable instance_name { type = string }

provider "google" {
  version = ">=3.5.0"
  credentials = var.credentials
  project     = var.project
  region      = var.region 
}

provider "google-beta" {
  version = ">=3.22.0"
  project     = var.project
}

resource "google_bigquery_dataset" "csb_dataset" {
  dataset_id    = replace(var.instance_name,"-","") 
  friendly_name = var.instance_name
  location      = var.region
  access {
    role          = "OWNER"
    special_group = "projectOwners"
  }
  access {
    role          = "WRITER"
    special_group = "projectWriters"
  }
  access {
    role          = "READER"
    special_group = "allAuthenticatedUsers"
  }

}

output dataset_id { value =  google_bigquery_dataset.csb_dataset.dataset_id }


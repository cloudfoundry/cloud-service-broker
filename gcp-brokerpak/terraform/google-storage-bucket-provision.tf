variable name {type = "string"}
variable region {type = "string"}
variable storage_class {type = "string"}
variable labels {type = "map"}
variable acl {type = "string"}

variable credentials  { type = string }
variable project  { type = string }

provider "google" {
  version = "~> 3.5.0"
  credentials = var.credentials
  project     = var.project
  
}


resource "google_storage_bucket" "bucket" {
    name     = "${var.name}"
    location = "${var.region}"
    storage_class = "${var.storage_class}"
    labels = "${var.labels}"
}

output id {value = "${google_storage_bucket.bucket.id}"}
output bucket_name {value = "${var.name}"}

# Copyright 2020 Pivotal Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

variable cores { type = number }
variable authorized_network { type = string }
variable instance_name { type = string }
variable db_name { type = string }

variable region { type = string }
variable labels { type = map }
variable storage_gb { type = number }
variable database_version { type = string }

variable credentials  { type = string }
variable project  { type = string }

provider "google" {
  version = ">=3.17.0"
  credentials = var.credentials
  project     = var.project
  
}

data "google_compute_network" "authorized-network" {
  name = var.authorized_network
}

data "google_compute_global_address" "private_ip_address" {
  name          = "csb-dataservices"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 24
  network       = data.google_compute_network.authorized-network.self_link
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = data.google_compute_network.authorized-network.self_link
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}
  
locals {
  service_tiers = {
    // https://cloud.google.com/sql/pricing#2nd-gen-pricing
    1 = "db-n1-standard-1"
    2 = "db-n1-standard-2"
    4 = "db-n1-standard-4"
    8 = "db-n1-standard-8"
    16 = "db-n1-standard-16"
    32 = "db-n1-standard-32"
    64 = "db-n1-standard-64"
    0.6   = "db-f1-micro" 
    1.7   = "db-g1-small" 
  }   

}

resource "google_sql_database_instance" "instance" {
  name             = var.instance_name
  database_version = var.database_version
  region           = var.region

  # depends_on = [google_service_networking_connection.private_vpc_connection]

  settings {
    tier = local.service_tiers[var.cores]
    disk_size = var.storage_gb
    user_labels = var.labels
    
    ip_configuration {
      ipv4_enabled    = false
      private_network = data.google_compute_network.authorized-network.self_link
    }
  }
}

resource "google_sql_database" "database" {
  name     = var.db_name
  instance = google_sql_database_instance.instance.name
}

resource "random_string" "username" {
  length = 16
  special = false
}

resource "random_password" "password" {
  length = 16
  special = true
  override_special = "_@"
}

resource "google_sql_user" "admin_user" {
  name     = random_string.username.result
  instance = google_sql_database_instance.instance.name
  password = random_password.password.result
}

output name { value = "${google_sql_database.database.name}" }
output hostname { value = "${google_sql_database_instance.instance.first_ip_address}" }

output port { value = (var.database_version == "POSTGRES_11" ? 5432 : 3306  ) }
output username { value = "${google_sql_user.admin_user.name}" }
output password { value = "${google_sql_user.admin_user.password}" }
variable dataset_id { type = string }
variable role { type = string }
variable service_account_name {type = "string"}
variable service_account_display_name {type = "string"}
variable project  { type = string }
variable credentials  { type = string }

provider "google-beta" {
  version = ">=3.22.0"
  project     = var.project
}
  
  
resource "google_service_account" "account" {
    account_id = var.service_account_name
    display_name = var.service_account_display_name
}
resource "google_service_account_key" "key" {
    service_account_id = google_service_account.account.name
}


locals {
    members = format("serviceAccount:%s",google_service_account.account.email)
}
    

// create memebers of the service account ..

#---------------------------------------------------
# Create spanner database iam
#---------------------------------------------------
// create memebers of the service account ..

#---------------------------------------------------
# Create spanner database iam
#---------------------------------------------------
data "google_iam_policy" "database_iam_policy" {
    binding {
        role = var.role
        members = [local.members]
    }
}

resource "google_bigquery_dataset_access" "access" {
  project       = var.project
  dataset_id    = var.dataset_id
  role          = "roles/${var.role}"
  user_by_email = google_service_account.account.email
}


#-------------------------------------------------------------------
# DM IAM
#-------------------------------------------------------------------
output "Name" {value = "${google_service_account.account.name}"}
output "Email" {value = "${google_service_account.account.email}"}
output "UniqueId" {value = "${google_service_account.account.unique_id}"}
output "PrivateKeyData" {value = "${google_service_account_key.key.private_key}"}
output "ProjectId" {value = "${google_service_account.account.project}"}
output dataset_id { value = var.dataset_id }



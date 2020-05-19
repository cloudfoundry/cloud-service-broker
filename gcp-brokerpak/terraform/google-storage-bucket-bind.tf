    variable role {type = "string"}
    variable service_account_name {type = "string"}
    variable service_account_display_name {type = "string"}
    variable bucket {type = "string"}
    resource "google_service_account" "account" {
      account_id = "${var.service_account_name}"
      display_name = "${var.service_account_display_name}"
    }
    resource "google_service_account_key" "key" {
      service_account_id = "${google_service_account.account.name}"
    }
    resource "google_storage_bucket_iam_member" "member" {
      bucket = "${var.bucket}"
      role   = "roles/${var.role}"
      member = "serviceAccount:${google_service_account.account.email}"
    }

    output "Name" {value = "${google_service_account.account.display_name}"}
    output "Email" {value = "${google_service_account.account.email}"}
    output "UniqueId" {value = "${google_service_account.account.unique_id}"}
    output "PrivateKeyData" {value = "${google_service_account_key.key.private_key}"}
    output "ProjectId" {value = "${google_service_account.account.project}"}
    
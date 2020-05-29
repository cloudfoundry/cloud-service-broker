    variable service_account_name {type = string}
    variable bucket {type = string}

    resource "google_service_account" "account" {
      account_id = var.service_account_name
      display_name = var.service_account_name
    }

    resource "google_service_account_key" "key" {
      service_account_id = google_service_account.account.name
    }

    resource "google_storage_bucket_iam_member" "member" {
      bucket = var.bucket
      role   = "roles/storage.objectAdmin"
      member = "serviceAccount:${google_service_account.account.email}"
    }

    resource "google_project_iam_member" "member" {
      role   = "roles/dataproc.editor"
      member = "serviceAccount:${google_service_account.account.email}"
    }


    output "name" {value = "${google_service_account.account.display_name}"}
    output "email" {value = "${google_service_account.account.email}"}
    output "private_key" {value = "${google_service_account_key.key.private_key}"}
    output "project_id" {value = "${google_service_account.account.project}"}
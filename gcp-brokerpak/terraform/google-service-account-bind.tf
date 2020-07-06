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

variable name {type = string}
variable credentials  { type = string }
variable project  { type = string }
variable role { type = string }

provider "google" {
  version = ">=3.17.0"
  credentials = var.credentials
  project     = var.project 
}

# data "google_compute_default_service_account" "default" {
# }

resource "google_service_account" "account" {
  account_id = substr(var.name, 0, 30)
  display_name = format("%s with role %s", var.name, var.role)
}

resource "google_service_account_key" "key" {
  service_account_id = google_service_account.account.name
}

# resource "google_service_account_iam_member" "member" {
#   service_account_id = google_service_account.account.name
#   role   = format("roles/%s", var.role)
#   member = format("serviceAccount:%s", google_service_account.account.email)
# }

resource "google_project_iam_member" "member" {
  project = var.project
  role    = format("roles/%s", var.role)
  member  = format("serviceAccount:%s", google_service_account.account.email)
}

output "Name" {value = google_service_account.account.name}
output "Email" {value = google_service_account.account.email}
output "UniqueId" {value = google_service_account.account.unique_id}
output "PrivateKeyData" {value = google_service_account_key.key.private_key}
output "ProjectId" {value = google_service_account.account.project}

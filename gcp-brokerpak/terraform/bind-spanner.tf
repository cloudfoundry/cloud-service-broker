variable instance { type = string } 
variable db_name { type = string }
variable role { type = string }
variable service_account_name {type = "string"}
variable service_account_display_name {type = "string"}

# provider "google-beta" {
#     version     = "2.7.0"
# }

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

resource "google_spanner_database_iam_policy" "spanner_database_iam_policy" {

    instance        =  var.instance
    database        =   var.db_name
    policy_data     = data.google_iam_policy.database_iam_policy.policy_data

    depends_on      = [data.google_iam_policy.database_iam_policy]

    lifecycle {
        ignore_changes = []
        create_before_destroy = true
    }
}

resource "google_spanner_database_iam_binding" "spanner_database_iam_binding" {


    instance    = var.instance
    database    = var.db_name
    role        = var.role

    members = [local.members] 

    lifecycle {
        ignore_changes = []
        create_before_destroy = true
    }
}

resource "google_spanner_database_iam_member" "spanner_database_iam_member" {


    instance    = var.instance
    database    = var.db_name
    role        = var.role
    member      = local.members

    lifecycle {
        ignore_changes = []
        create_before_destroy = true
    }
}


#-------------------------------------------------------------------
# DM IAM
#-------------------------------------------------------------------
output "policy_etag" { value =  join(",", google_spanner_database_iam_policy.spanner_database_iam_policy.*.etag)}
output "binding_etag"{ value =  join(",", google_spanner_database_iam_binding.spanner_database_iam_binding.*.etag)}
output "member_etag" { value =  join(",", google_spanner_database_iam_member.spanner_database_iam_member.*.etag)}
output "Name" {value = "${google_service_account.account.name}"}
output "Email" {value = "${google_service_account.account.email}"}
output "UniqueId" {value = "${google_service_account.account.unique_id}"}
output "PrivateKeyData" {value = "${google_service_account_key.key.private_key}"}
output "ProjectId" {value = "${google_service_account.account.project}"}
output instance { value = var.instance}
output db_name { value = var.db_name }



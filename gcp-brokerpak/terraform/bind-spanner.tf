variable instance { type = string }
variable db_name { type = string }
variable role { type = string }
variable members { type = list }

#---------------------------------------------------
# Create spanner database iam
#---------------------------------------------------
data "google_iam_policy" "database_iam_policy" {
    binding {
        role = var.role
        members = var.members
    }
}

resource "google_spanner_database_iam_policy" "spanner_database_iam_policy" {
    
    instance        =  var.instance
    database        = "${var.instance}-db"
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
    
    members     = var.members
    
    lifecycle {
        ignore_changes = []
        create_before_destroy = true
    }
} 

resource "google_spanner_database_iam_member" "spanner_database_iam_member" {
    

    instance    = var.instance
    database    = var.db_name
    role        = var.role
    member      = length(var.members) > 0 ? element(var.members, 0) : ""

    lifecycle {
        ignore_changes = []
        create_before_destroy = true
    }
}  


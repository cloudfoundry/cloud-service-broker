variable worker_machine_type {type = string}
variable master_machine_type {type = string}
variable worker_count {type = number}
variable master_count {type = number}
variable preemptible_count {type = number}

variable name {type = string}
variable region {type = string}
variable labels {type = map}

   variable credentials { type = string }
    variable project { type = string }  

    provider "google" {
      version = ">=3.17.0"
      credentials = var.credentials
      project     = var.project      
    }

    resource "google_dataproc_cluster" "cluster" {
      name   = var.name
      region = var.region
      labels = var.labels

      cluster_config {
        master_config {
          num_instances = var.master_count
          machine_type  = var.master_machine_type
        }

        worker_config {
          num_instances = var.worker_count
          machine_type  = var.worker_machine_type
        }

        preemptible_worker_config {
          num_instances = var.preemptible_count
        }
      }
    }

    output bucket_name {value = google_dataproc_cluster.cluster.cluster_config.0.bucket}
    output cluster_name {value = google_dataproc_cluster.cluster.name}
    output region {value = google_dataproc_cluster.cluster.region}
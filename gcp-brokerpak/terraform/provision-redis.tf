variable service_tier { type = string }
    variable authorized_network { type = string }
    variable display_name { type = string }
    variable instance_id { type = string }
    variable region { type = string }
    variable memory_size_gb { type = number }
    variable labels { type = map }
    variable credentials { type = string }
    variable project { type = string }  

    provider "google" {
      version = ">=3.17.0"
      credentials = var.credentials
      project     = var.project      
    }

    data "google_compute_network" "authorized-network" {
      name = var.authorized_network
    }

    resource "google_redis_instance" "instance" {
      name               = var.instance_id
      tier               = var.service_tier
      memory_size_gb     = var.memory_size_gb
      display_name       = var.display_name
      region             = var.region
      authorized_network = data.google_compute_network.authorized-network.self_link
      labels             = var.labels

      timeouts {
        create = "15m"
        update = "15m"
        delete = "15m"
      }
    }

    output memory_size_gb { value = google_redis_instance.instance.memory_size_gb }
    output service_tier { value = google_redis_instance.instance.tier }
    output redis_version { value = google_redis_instance.instance.redis_version }
    output host { value = google_redis_instance.instance.host }
    output port { value = google_redis_instance.instance.port }
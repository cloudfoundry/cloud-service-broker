terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
    }
  }
}

variable length { type = number }

resource "random_string" "fake" {
  length = var.length
  lifecycle {
    prevent_destroy = true
  }
}
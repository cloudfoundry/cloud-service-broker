terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
    }
  }
}

resource "random_integer" "priority" {
  min = 1
  max = 100
}
terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = ">=3.1.0"
    }
  }
}

provider "random" { }

resource "random_integer" "priority" {
  min = 1
  max = 100
}
terraform {
  required_providers {
    random = {
      source = "registry.terraform.io/hashicorp/random"
      version = "3.5.1"
    }
  }
}

resource "random_integer" "priority" {
  min = 1
  max = 100
}

variable alpha_input { type = string }
terraform {
  required_providers {
    random = {
      source = "hashicorp/random"
      version = "3.1.0"
    }
  }
}

resource "random_integer" "priority" {
  min = 1
  max = 100
}
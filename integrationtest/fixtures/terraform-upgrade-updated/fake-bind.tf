terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.1.0"
    }
  }
}

resource "random_integer" "priority" {
  min = 3
  max = 4
}

output "provision_output" {
  value = random_integer.priority.result
}
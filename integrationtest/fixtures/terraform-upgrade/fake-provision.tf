terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.1.0"
    }
  }
}

resource "random_integer" "priority" {
  min = 1
  max = 2
}

output "provision_output" {
  value = random_integer.priority.result
}

variable alpha_input { type = string }
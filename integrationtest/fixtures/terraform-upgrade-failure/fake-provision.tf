terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
    }
  }
}

variable "max" {type = number}

resource "random_integer" "priority" {
  min = 1
  max = var.max
}

output "provision_output" {
  value = random_integer.priority.result
}
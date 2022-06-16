terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.1.0"
    }
  }
}

resource "random_password" "password" {
  length = 5
}

output "provision_output" {
  value = random_password.password.result
}
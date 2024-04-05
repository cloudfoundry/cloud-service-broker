terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
    }
  }
}

resource "random_password" "password" {
  length = 5
}

output "provision_output" {
  value = random_password.password.result
  sensitive = true
}
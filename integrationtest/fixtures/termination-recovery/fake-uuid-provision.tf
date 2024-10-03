terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
    }
    null = {
      source  = "registry.terraform.io/hashicorp/null"
    }
  }
}
resource "null_resource" "sleeper" {
  provisioner "local-exec" {
    command = "sleep 10"
  }
}
resource "random_uuid" "random" {}

output provision_output { value = random_uuid.random.result }

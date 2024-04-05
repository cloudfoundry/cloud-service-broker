terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
    }
  }
}

resource "random_uuid" "random" {}

output provision_output { value = random_uuid.random.result }
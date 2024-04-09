// Fails as required "length" parameter is missing
terraform {
  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
    }
  }
}

resource "random_string" "random" {}

output bind_output { value = random_string.random.result }
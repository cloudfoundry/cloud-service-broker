## new versions of terraform fail with a different error to the one we want to trigger if this file is empty.
## adding some dummy terraform to test errors when a particular field is not present.

terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
}

resource "random_integer" "priority" {
  min = 3
  max = 4
}


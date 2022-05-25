provider "random" { }

variable length { type = number }

resource "random_string" "fake" {
  length = var.length
  lifecycle {
    prevent_destroy = true
  }
}
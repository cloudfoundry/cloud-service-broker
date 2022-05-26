provider "random" { }

resource "random_integer" "priority" {
  min = 1
  max = 100
}
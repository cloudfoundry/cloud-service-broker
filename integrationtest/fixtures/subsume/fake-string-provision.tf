resource "random_string" "random" {
  length  = 10
  special = false
}

output provision_output { value = random_string.random.result }
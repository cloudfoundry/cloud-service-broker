resource "random_uuid" "random" {
}

output provision_output { value = random_uuid.random.result }

output "status" {
  value = format("created random GUID: %s", random_uuid.random.result)
}
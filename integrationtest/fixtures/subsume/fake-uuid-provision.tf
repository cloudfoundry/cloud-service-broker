resource "random_uuid" "random" {
}

output provision_output { value = random_uuid.random.result }
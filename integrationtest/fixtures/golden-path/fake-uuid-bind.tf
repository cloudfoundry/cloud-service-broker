resource "random_uuid" "random" {}

output bind_output { value = random_uuid.random.result }
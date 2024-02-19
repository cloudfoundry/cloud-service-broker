// Fails as required "length" parameter is missing
resource "random_string" "random" {}

output provision_output { value = random_string.random.result }
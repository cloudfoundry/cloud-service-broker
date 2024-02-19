resource "random_string" "random" {
  length = 10
}

output bind_output { value = random_string.random.result }
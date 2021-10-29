variable update_output { type = string }
output provision_output { value = "provision output value" }
output update_output_output_updated { value = "${var.update_output}" }
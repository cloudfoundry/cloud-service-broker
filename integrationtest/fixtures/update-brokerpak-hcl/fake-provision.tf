variable update_input { type = string }
output provision_output { value = "provision output value" }
output update_output { value = "${var.update_input}" }
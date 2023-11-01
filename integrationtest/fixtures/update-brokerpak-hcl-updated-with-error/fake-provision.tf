variable update_input { type = string }
variable extra_input { type = string }
output provision_output { value = "provision output value" }
output update_output_updated { value = "${var.wrong_update_input == null ? "empty" : var.update_input}${var.extra_input == null ? "empty" : var.extra_input}" }
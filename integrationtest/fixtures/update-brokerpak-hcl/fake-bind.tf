variable provision_output { type = string }
output bind_output { value = "${var.provision_output} and bind output value" }
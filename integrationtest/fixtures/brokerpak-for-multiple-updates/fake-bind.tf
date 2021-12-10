variable alpha_output { type = string }
variable beta_output { type = string }
output bind_output { value = "${var.alpha_output};${var.beta_output}" }
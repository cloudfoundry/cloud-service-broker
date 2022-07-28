variable alpha_input { type = string }
variable beta_input { type = string }
output alpha_output { value = var.alpha_input }
output beta_output { value = var.beta_input == null ? "is_null" : var.beta_input }
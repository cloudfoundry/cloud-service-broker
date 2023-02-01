variable alpha_output { type = string }
variable beta_output { type = string }
output bind_output { value = "${var.alpha_output};${var.beta_output}" }
output bind_output_2 { value = tomap({foo = "bar"}) }
output bind_output_3 { value = tolist(["a", "b", "c"]) }
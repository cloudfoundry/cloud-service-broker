variable labels { type = map(any) }
output labels { value = jsonencode(var.labels) }
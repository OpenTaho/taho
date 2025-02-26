variable "test_input_1" {
  type = string
}

output "test_output_2" {
  value = var.test_input_2
}

variable "test_input_2" {
  type = string
}

output "test_output_1" {
  value = var.test_input_1
}
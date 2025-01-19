variable "used_2" {
  type = string
}

variable "used_1" {
  type = string
}

# tflint-ignore: terraform_unused_declarations
variable "unused_2" {
  type = string
}

# tflint-ignore: terraform_unused_declarations
variable "unused_1" {
  type = string
}

output "result" {
  value = "${var.used_1}-${var.used_2}"
}
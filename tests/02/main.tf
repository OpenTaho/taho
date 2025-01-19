locals {
  test_value_1 = var.test_input_1
  test_value_two =     var.test_input_2
}

variable "test_input_1" {
  type = string
}

provider "kubernetes" {

}

terraform {
  required_version = ">=1.6.2"
}

output "test_output_2" {
  value = local.test_value_two
}

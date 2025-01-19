// This is a slash style comment.

# This is a single line comment for the main file.

# This is about input_1
variable "input_1" {
  # This is an inner comment at the start of the block.

  # This is a comment about the type attribute.
  type = string

  # This is an inner comment.

  # This is a comment about the default attribute.
  default = null

  # This is an inner comment at end of block will be removed.
}

# This is a comment about output_1
output "output_1" {

  /* foo */
  value = var.input_1

  /* bar */
}
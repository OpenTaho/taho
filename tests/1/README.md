# Test Simple Main Tranformation

While it is legal to have a Terraform project with all content in `main.tf` (or
really in any file) we consider such a structure to be ugly.

The `taho` tool will transform this module to one where the proper files exist
with the correct content in each file.

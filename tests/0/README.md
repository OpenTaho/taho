# Test Empty Directory

When executed in an empty directory the program will create empty files for
`main.tf`, `variables.tf`, and `outputs.tf`.

In an empty directory a `terraform.tf` file will be created with the following
content to reach the minimal acceptable state for a Taho complient module.

```terraform
terraform {
  required_version = ">=0.0.1"
}
```

If executed in a directory that contains the files in the prior state the
program will set a zero exit code.

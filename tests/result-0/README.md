# Test Empty Directory

The program creates empty files for `main.tf`, `variables.tf`, and `outputs.tf`
when executed in an empty directory. A `terraform.tf` file will be created with
the following content to reach the minimal acceptable state for a Taho complient
module.

```terraform
terraform {
  required_version = ">=0.0.1"
}
```

The program will set a zero exit code when executed in a directory that contains
the files in the proper state.

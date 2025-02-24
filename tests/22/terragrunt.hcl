terraform {
  source = "git@gitlab.com:EXAMPLE"
}
inputs = {
  b = null
  a = null
  c = null
}
locals {
  b = null
  a = null
  c = null
}
remote_state {
  backend = "s3"
  config = {
    encrypt        = true
    bucket         = "name-state-${local.aws_region}-${local.account_id}"
    key            = "${path_relative_to_include()}/tf.tfstate"
    region         = local.aws_region
    dynamodb_table = "terraform-locks"
  }
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
}

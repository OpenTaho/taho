inputs = {
  a = null
  b = null
  c = null
}

locals {
  a = null
  b = null
  c = null
}

remote_state {
  backend = "s3"

  config = {
    bucket         = "name-state-${local.aws_region}-${local.account_id}"
    dynamodb_table = "terraform-locks"
    encrypt        = true
    key            = "${path_relative_to_include()}/tf.tfstate"
    region         = local.aws_region
  }

  generate = {
    if_exists = "overwrite_terragrunt"
    path      = "backend.tf"
  }
}

terraform {
  source = "git@gitlab.com:EXAMPLE"
}

terraform {
  required_version = ">=1.6.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">=5.84.0"
    }

    null = {
      source  = "hashicorp/null"
      version = ">=3.2.3"
    }
  }
}

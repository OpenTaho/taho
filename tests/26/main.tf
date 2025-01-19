terraform {
  required_version = ">= 1.5, < 2.0.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }

  backend "s3" {
    bucket         = "my-tf-bucket"
    dynamodb_table = "terraform-lock"
    encrypt        = true
    key            = "my-key"
    region         = "my-region"
  }
}

provider "aws" {
  default_tags {
    tags = var.tags
  }
  region = var.region
}

variable "region" {
  type = string 
}

variable "tags" {
  type = map(string)
}
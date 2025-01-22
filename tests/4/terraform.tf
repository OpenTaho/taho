locals {
  attribute2 = null
  attribute1 = null
}

terraform {
  required_version = ">=1.6.6"

  backend "s3" {}

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

resource "kubernetes_cluster_role" "planner" {
  metadata { 
    name = "sre3:planner" 
  }

  rule {
    resources  = ["*"]
    api_groups = ["*"]
    verbs      = ["list", "watch", "get"]
  }

  count = 0

  rule {
    non_resource_urls = ["/version"]
    verbs             = ["get"]
  }
}

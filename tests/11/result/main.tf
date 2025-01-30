locals {
  attribute1 = null
  attribute2 = null
}

resource "kubernetes_cluster_role" "planner" {
  count = 0

  aggregation_rule {}

  metadata {
    name = "sre3:planner"
  }

  rule {
    api_groups = ["*"]
    resources  = ["*"]
    verbs      = ["list", "watch", "get"]
  }

  rule {
    non_resource_urls = ["/version"]
    verbs             = ["get"]
  }

  lifecycle {
    create_before_destroy = true
  }

  lifecycle {
    create_before_destroy = true
  }
}

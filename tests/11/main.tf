locals {
  attribute1 = null
  attribute2 = null
}

resource "kubernetes_cluster_role" "planner" {
  count = 0

  metadata {
    name = "sre3:planner"
  }

  aggregation_rule {}

  lifecycle {
    create_before_destroy = true
  }

  connection {
    host = null
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
}

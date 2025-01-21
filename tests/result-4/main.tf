locals {
  attribute2 = null
  attribute1 = null
}

resource "kubernetes_cluster_role" "planner" {
  count = 0

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
}

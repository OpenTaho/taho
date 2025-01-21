locals {
  attribute2 = null
  attribute1 = null
}

resource "kubernetes_cluster_role" "planner" {
  metadata { 
    name = "sre3:planner" 
  }

  rule {
    api_groups = ["*"]
    resources  = ["*"]
    verbs      = ["list", "watch", "get"]
  }

  count = 0

  rule {
    non_resource_urls = ["/version"]
    verbs             = ["get"]
  }
}

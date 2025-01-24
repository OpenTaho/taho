resource "aws_ecr_repository" "test" {
  count    = 0
  for_each = []

  id   = null
  name = null

  tags = {
    account_id = null
  }
}

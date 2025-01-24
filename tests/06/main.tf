resource "aws_ecr_repository" "test" {
  count = 0
  name = null
  id = null
  for_each = []
  tags = {
    account_id = null
  }
}

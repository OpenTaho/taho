data "template_file" "test" {
  count    = 0
  for_each = []

  template = null

  vars = {
    account_id = null
  }
}

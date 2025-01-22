data "template_file" "test" {
  count = 0
  template = null
  for_each = []
  vars = {
    account_id = null
  }
}

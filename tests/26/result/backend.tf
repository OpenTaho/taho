terraform {
  backend "s3" {
    bucket         = "my-tf-bucket"
    dynamodb_table = "terraform-lock"
    encrypt        = true
    key            = "my-key"
    region         = "my-region"
  }
}

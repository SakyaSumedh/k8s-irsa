module "s3-bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "3.15.1"

  bucket                   = "${local.env}-bucket"
  acl                      = "private"
  control_object_ownership = true
  object_ownership         = "ObjectWriter"
}
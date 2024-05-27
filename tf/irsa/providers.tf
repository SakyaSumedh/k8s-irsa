provider "aws" {
  region = var.AWS_REGION

  assume_role {
    role_arn = var.AWS_ROLE_ARN
  }

  default_tags {
    tags = {
      Environment = local.env
      Management  = "terraform"
      Product     = local.config.product
      Service     = local.config.service
    }
  }
}
provider "aws" {
  region = "ap-southeast-1"

  #   assume_role {
  #     role_arn = var.AWS_ROLE_ARN
  #   }

  default_tags {
    tags = {
      Management = "terraform"
      Author     = "Sumedh"
    }
  }
}
terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "terraform-organization-name"

    workspaces {
      prefix = "demo-irsa-"
    }
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

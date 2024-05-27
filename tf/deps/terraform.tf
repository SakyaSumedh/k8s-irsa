terraform {
  backend "remote" {
    organization = "terraform-org-name"
    hostname     = "app.terraform.io"

    workspaces {
      prefix = "demo-irsa-"
    }
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    archive = {
      source = "hashicorp/archive"
    }
  }
}

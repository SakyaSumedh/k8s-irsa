variable "AWS_REGION" {}

variable "AWS_ROLE_ARN" {}

variable "SERVICE" {
  type    = string
  default = "demo-irsa"
}

variable "cluster_name" {}
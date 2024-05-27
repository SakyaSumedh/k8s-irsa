locals {
  env = "${var.SERVICE}"
  config = {
    product = "${var.SERVICE}"
    service = "irsa"
  }

  cluster_name = "${var.cluster_name}"
  namespace    = "demo-irsa"
  aws = {
    account_id = data.aws_caller_identity.current.account_id
    region     = data.aws_region.current.name
    eks = {
      oidc_url = data.aws_eks_cluster.cluster.identity[0]["oidc"][0]["issuer"]
    }
  }

  policies = {
    sqs = {
      Action = [
        "sqs:SendMessage",
      ]
      Effect = "Allow"
      Resource = [
        "arn:aws:sqs:${local.aws.region}:${local.aws.account_id}:${local.env}*"
      ]
    }

    s3 = {
      Action = [
        "s3:Get*",
        "s3:List*",
        "s3:Describe*",
        "s3:Put*",
        "s3:DeleteObject"
      ]
      Effect = "Allow"
      Resource = [
        "arn:aws:s3:::${local.env}*",
        "arn:aws:s3:::*${local.env}"
      ]
    }

    sns = {
      Action = [
        "sns:Publish",
      ]
      Effect = "Allow"
      Resource = [
        "arn:aws:sns:*:${local.aws.account_id}:${local.env}*"
      ]
    }

    lambda = {
      Action = [
        "lambda:InvokeFunction",
        "lambda:InvokeAsync"
      ]
      Effect = "Allow"
      Resource = [
        "arn:aws:lambda:*:${local.aws.account_id}:function:${local.env}*"
      ]
    }

    dynamodb = {
      Action = [
        "dynamodb:BatchGetItem",
        "dynamodb:ListGlobalTables",
        "dynamodb:PutItem",
        "dynamodb:DescribeTable",
        "dynamodb:ListTables",
        "dynamodb:DeleteItem",
        "dynamodb:GetItem",
        "dynamodb:Scan",
        "dynamodb:Query",
        "dynamodb:UpdateItem",
        "dynamodb:GetRecords"
      ]
      Effect = "Allow"
      Resource = [
        "arn:aws:dynamodb:*:${local.aws.account_id}:table/${local.env}*"
      ]
    }
  }

  services = {
    demo-app = {
      policies = [
        local.policies.sqs,
        local.policies.s3,
        local.policies.lambda,
        local.policies.dynamodb,
      ]
    }
  }
}

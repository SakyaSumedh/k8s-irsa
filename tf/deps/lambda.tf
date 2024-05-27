// zip the binary, as we can upload only zip files to AWS lambda
data "archive_file" "function_archive" {

  type        = "zip"
  source_file = local.binary_path
  output_path = local.archive_path
}


module "demo_irsa_lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "6.5.0"
  timeout = 60

  function_name  = "${local.env}-lambda"
  create_package = false
  publish        = true
  runtime        = "python3.12"

  handler                = "main.function_handler"
  local_existing_package = data.archive_file.function_archive.output_path

  allowed_triggers = {}

  attach_cloudwatch_logs_policy     = true
  cloudwatch_logs_retention_in_days = 1
  attach_policy_statements          = true

  #Controls whether policy_statements should be added to IAM role for Lambda Function
  policy_statements = {
    lambda_permission = {
      effect = "Allow",
      actions = [
        "lambda:InvokeFunction",
        "lambda:InvokeAsync"
      ],
      resources = ["*"]
    },
    cloudwatch_permission = {
      effect = "Allow"
      actions = [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]

      resources = [
        "arn:aws:logs:*:*:*",
      ]
    },
  }
}
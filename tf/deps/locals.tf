locals {
  env = "demo-irsa"
  function_name = "main.py"
  binary_path   = "${path.module}/lambda/main.py"
  archive_path  = "${path.module}/lambda.zip"
}


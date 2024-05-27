resource "aws_iam_policy" "irsa_policy" {
  for_each    = local.services
  name        = "${each.key}-${local.env}-policy"
  description = "IRSA IAM policy"

  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = local.services[each.key].policies
  })
}


module "irsa" {
  for_each = local.services

  source = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"

  role_name              = "${local.env}"
  allow_self_assume_role = true

  oidc_providers = {
    main = {
      # example arn
      # arn:aws:iam::<aws-account-id>:oidc-provider/oidc.eks.ap-southeast-1.amazonaws.com/id/34FBD71D5EE0E3F02D2746761C178AC7

      #https://oidc.eks.ap-southeast-1.amazonaws.com/id/34FBD71D5EE0E3F02D2746761C178AC7
      provider_arn               = "arn:aws:iam::${local.aws.account_id}:oidc-provider/${replace(local.aws.eks.oidc_url, "https://", "")}"
      namespace_service_accounts = ["${local.namespace}:${each.key}"]
    }
  }

  role_policy_arns = {
    policy = aws_iam_policy.irsa_policy[each.key].arn
  }

  depends_on = [aws_iam_policy.irsa_policy]
}
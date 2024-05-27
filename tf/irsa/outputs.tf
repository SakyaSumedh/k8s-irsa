output "oidc" {
  value = data.aws_eks_cluster.cluster.identity[0]["oidc"][0]["issuer"]
}

output "irsa_arns" {
  value = module.irsa[*]
}
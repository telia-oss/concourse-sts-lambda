# ------------------------------------------------------------------------------
# Output
# ------------------------------------------------------------------------------
output "role_arn" {
  value = module.lambda.role_arn
}

output "arn" {
  value = module.lambda.arn
}

output "name" {
  value = module.lambda.name
}


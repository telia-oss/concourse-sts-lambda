# ------------------------------------------------------------------------------
# Output
# ------------------------------------------------------------------------------
output "role_arn" {
  value = "${module.lambda.role_arn}"
}

output "function_arn" {
  value = "${module.lambda.function_arn}"
}

output "function_name" {
  value = "${module.lambda.function_name}"
}

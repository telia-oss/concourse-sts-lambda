# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------
variable "name_prefix" {
  description = "Prefix used for resource names."
}

variable "lambda_arn" {
  description = "ARN of the STS Lambda."
}

variable "config_bucket" {
  description = "Name of the config bucket."
}

variable "accounts" {
  description = "Valid JSON representation of a Team (see Go code)."
  type        = "list"
}

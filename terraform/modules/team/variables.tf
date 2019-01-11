# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------
variable "name" {
  description = "Name of the team (used to give descriptive name to resources)."
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

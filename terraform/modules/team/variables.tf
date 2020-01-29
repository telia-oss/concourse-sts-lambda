# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------
variable "name" {
  description = "Name of the team (used to give descriptive name to resources)."
  type        = string
}

variable "lambda_arn" {
  description = "ARN of the STS Lambda."
  type        = string
}

variable "config_bucket" {
  description = "Name of the config bucket."
  type        = string
}

variable "accounts" {
  description = "Valid JSON representation of a Team (see Go code)."
  type        = list(object({ name = string, roleArn = string, duration = string }))
}

variable "tags" {
  description = "A map of tags (key-value pairs) passed to resources."
  type        = map(string)
  default     = {}
}

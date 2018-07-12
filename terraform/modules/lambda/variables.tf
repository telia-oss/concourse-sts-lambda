# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------
variable "name_prefix" {
  description = "Prefix used for resource names."
}

variable "filename" {
  description = "Path to .zip file containing the handler. (I.e., output of make release)"
}

variable "role_prefix" {
  description = "Prefix of CI roles which the Lambda function will be allowed to assume. (should be the same in all accounts)."
}

variable "secrets_manager_prefix" {
  description = "Prefix used for secrets. The Lambda will be allowed to create and write secrets to any secret with this prefix."
  default     = "concourse"
}

variable "tags" {
  description = "A map of tags (key-value pairs) passed to resources."
  type        = "map"
  default     = {}
}

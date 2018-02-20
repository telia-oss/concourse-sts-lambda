# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------
variable "prefix" {
  description = "Prefix used for resource names."
}

variable "zip_file" {
  description = "Path to .zip file containing the handler. (I.e., output of make release)"
}

variable "role_prefix" {
  description = "Prefix of CI roles which the Lambda function will be allowed to assume. (should be the same in all accounts)."
}

variable "ssm_prefix" {
  description = "Prefix used for SSM Parameters. The Lambda will be allowed to write to any parameter with this prefix."
  default     = "concourse"
}

variable "region" {
  description = "Region to use for S3 and SSM clients."
}

variable "tags" {
  description = "A map of tags (key-value pairs) passed to resources."
  type        = "map"
  default     = {}
}

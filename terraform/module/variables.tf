# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------
variable "prefix" {
  description = "Prefix used for resource names."
}

variable "zip_file" {
  description = "Path to .zip file containing the handler. (I.e., output of make release)"
}

variable "ci_role_prefix" {
  description = "Prefix of CI roles which the Lambda function will be allowed to assume. (should be the same in all accounts)."
}

variable "ci_ssm_prefix" {
  description = "Prefix used for SSM Parameters. The Lambda will be allowed to write to any parameter with this prefix."
  default     = "concourse"
}

variable "config_region" {
  description = "Region to use for S3 and SSM clients."
}

variable "config_bucket" {
  description = "Name of bucket containing the config file."
}

variable "config_key" {
  description = "Bucket key for the config file."
}

variable "tags" {
  description = "A map of tags (key-value pairs) passed to resources."
  type        = "map"
  default     = {}
}

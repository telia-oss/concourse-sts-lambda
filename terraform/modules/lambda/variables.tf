# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------
variable "name_prefix" {
  description = "Prefix used for resource names."
}

variable "filename" {
  description = "Path to the handler zip-file."
  default     = ""
}

variable "s3_bucket" {
  description = "The bucket where the lambda function is uploaded."
  default     = ""
}

variable "s3_key" {
  description = "The s3 key for the Lambda artifact."
  default     = ""
}

variable "config_bucket_arn" {
  description = "The ARN of the config bucket."
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

terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  version = ">= 2.17"
  region  = var.region
}

resource "aws_s3_bucket" "config" {
  bucket_prefix = var.name_prefix
  acl           = "private"
  force_destroy = true
  tags          = var.tags
}

module "lambda" {
  source = "./modules/lambda"

  name_prefix            = var.name_prefix
  role_prefix            = "${var.name_prefix}-machine-role"
  secrets_manager_prefix = "concourse"
  config_bucket_arn      = aws_s3_bucket.config.arn
  tags                   = var.tags
}

# Each team will need their own Lambda trigger which is CRON triggered
# and passes that teams configuration to the function when it's invoked.
module "team" {
  source = "./modules/team"

  name          = "${var.name_prefix}-team"
  lambda_arn    = module.lambda.arn
  config_bucket = aws_s3_bucket.config.id

  accounts = [
    {
      name    = "example-account"
      roleArn = aws_iam_role.main.arn
    },
  ]
}

# The lambda function now has privileges to assume any role with the
# "machine-user" prefix, in any account. So if we create a role which
# allows the Lambda functions execution role to assume it we should be
# good to go.
resource "aws_iam_role" "main" {
  name                  = "${var.name_prefix}-machine-role"
  assume_role_policy    = data.aws_iam_policy_document.assume.json
  force_detach_policies = true
  tags                  = var.tags
}

data "aws_iam_policy_document" "assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = [module.lambda.role_arn]
    }
  }
}

resource "aws_iam_role_policy_attachment" "view_only_policy" {
  role       = aws_iam_role.main.name
  policy_arn = "arn:aws:iam::aws:policy/job-function/ViewOnlyAccess"
}

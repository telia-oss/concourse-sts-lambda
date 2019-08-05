# ------------------------------------------------------------------------------
# Resources
# ------------------------------------------------------------------------------
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

locals {
  s3_bucket = var.filename == "" && var.s3_bucket == "" ? "telia-oss-${data.aws_region.current.name}" : var.s3_bucket
  s3_key    = var.filename == "" && var.s3_key == "" ? "concourse-sts-lambda/v0.9.1.zip" : var.s3_key
}

module "lambda" {
  source  = "telia-oss/lambda/aws"
  version = "3.0.0"

  name_prefix = var.name_prefix
  filename    = var.filename
  s3_bucket   = local.s3_bucket
  s3_key      = local.s3_key
  policy      = data.aws_iam_policy_document.lambda.json
  handler     = "main"
  runtime     = "go1.x"

  environment = {
    SECRETS_MANAGER_PATH = "/${var.secrets_manager_prefix}/{{.Team}}/{{.Account}}"
  }

  tags = var.tags
}

data "aws_iam_policy_document" "lambda" {
  statement {
    effect = "Allow"

    actions = [
      "sts:AssumeRole",
    ]

    resources = [
      "arn:aws:iam::*:role/${var.role_prefix}*",
    ]
  }

  statement {
    effect  = "Allow"
    actions = ["s3:GetObject"]

    resources = [
      "${var.config_bucket_arn}/*",
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      "*",
    ]
  }

  // https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_CreateSecret.html
  // https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_UpdateSecret.html
  statement {
    effect = "Allow"

    actions = [
      "secretsmanager:CreateSecret",
      "secretsmanager:UpdateSecret",
    ]

    resources = [
      "arn:aws:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:/${var.secrets_manager_prefix}/*",
    ]
  }
}

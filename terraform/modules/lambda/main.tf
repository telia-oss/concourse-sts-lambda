# ------------------------------------------------------------------------------
# Resources
# ------------------------------------------------------------------------------
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

module "lambda" {
  source  = "telia-oss/lambda/aws"
  version = "0.2.0"

  name_prefix = "${var.name_prefix}"
  filename    = "${var.filename}"
  policy      = "${data.aws_iam_policy_document.lambda.json}"
  handler     = "main"
  runtime     = "go1.x"

  environment {
    REGION               = "${data.aws_region.current.name}"
    SECRETS_MANAGER_PATH = "/${var.secrets_manager_prefix}/{{.Team}}/{{.Account}}"
  }

  tags = "${var.tags}"
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

# ------------------------------------------------------------------------------
# Resources
# ------------------------------------------------------------------------------
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

module "lambda" {
  source = "github.com/TeliaSoneraNorge/divx-terraform-modules//lambda/function?ref=0.4.0"

  prefix   = "${var.prefix}"
  policy   = "${data.aws_iam_policy_document.lambda.json}"
  zip_file = "${var.zip_file}"
  handler  = "main"
  runtime  = "go1.x"

  variables {
    REGION   = "${var.region}"
    SSM_PATH = "/${var.ssm_prefix}/{{.Team}}/{{.Account}}"
  }

  tags {
    environment = "dev"
    terraform   = "True"
  }
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

  statement {
    effect = "Allow"

    actions = [
      "ssm:PutParameter",
    ]

    resources = [
      "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter/${var.ssm_prefix}*",
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "kms:Encrypt",
    ]

    resources = [
      "*",
    ]
  }
}

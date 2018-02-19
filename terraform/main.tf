provider "aws" {
  region = "eu-west-1"
}

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

module "sts-lambda" {
  source = "./module"

  prefix         = "assume-role"
  zip_file       = "../main.zip"
  ci_role_prefix = "machine-user"
  ci_ssm_prefix  = "concourse"
  config_region  = "${data.aws_region.current.name}"
  config_bucket  = "<bucket>"
  config_key     = "example.json"

  tags {
    environment = "dev"
    terraform   = "True"
  }
}

# The lambda function now has privileges to assume any role with the
# "machine-user" prefix, in any account. So if we create a role which
# allows the Lambda functions execution role to assume it we should be
# good to go.
resource "aws_iam_role" "main" {
  name                  = "machine-user-example"
  assume_role_policy    = "${data.aws_iam_policy_document.assume.json}"
  force_detach_policies = "true"
}

data "aws_iam_policy_document" "assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type = "AWS"

      identifiers = [
        "${module.sts-lambda.role_arn}",
      ]
    }
  }
}

resource "aws_iam_role_policy_attachment" "view_only_policy" {
  role       = "${aws_iam_role.main.name}"
  policy_arn = "arn:aws:iam::aws:policy/job-function/ViewOnlyAccess"
}

output "lambda_arn" {
  value = "${module.sts-lambda.function_arn}"
}

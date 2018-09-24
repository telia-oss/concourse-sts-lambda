provider "aws" {
  region = "eu-west-1"
}

module "sts-lambda" {
  source = "./modules/lambda"

  name_prefix            = "assume-role"
  role_prefix            = "machine-user"
  secrets_manager_prefix = "concourse"

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

# Each team will need their own Lambda trigger which is CRON triggered
# and passes that teams configuration to the function when it's invoked.
module "sts-lambda-trigger" {
  source = "./modules/trigger"

  name_prefix = "example-team"
  lambda_arn  = "${module.sts-lambda.arn}"

  team_config = <<EOF
{
  "name": "example-team",
  "accounts": [{
    "name": "example-account",
    "roleArn": "${aws_iam_role.main.arn}"
  }]
}
EOF
}

output "lambda_arn" {
  value = "${module.sts-lambda.arn}"
}

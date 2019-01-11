provider "aws" {
  region = "eu-west-1"
}

resource "aws_s3_bucket" "config" {
  bucket        = "sts-lambda-config-bucket-example"
  acl           = "private"
  force_destroy = "true"

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
        "${module.lambda.role_arn}",
      ]
    }
  }
}

resource "aws_iam_role_policy_attachment" "view_only_policy" {
  role       = "${aws_iam_role.main.name}"
  policy_arn = "arn:aws:iam::aws:policy/job-function/ViewOnlyAccess"
}

module "lambda" {
  source = "./modules/lambda"

  name_prefix            = "assume-role"
  role_prefix            = "machine-user"
  secrets_manager_prefix = "concourse"
  config_bucket_arn      = "${aws_s3_bucket.config.arn}"

  tags {
    environment = "dev"
    terraform   = "True"
  }
}

# Each team will need their own Lambda trigger which is CRON triggered
# and passes that teams configuration to the function when it's invoked.
module "team" {
  source = "./modules/team"

  name_prefix   = "example-team"
  lambda_arn    = "${module.lambda.arn}"
  config_bucket = "${aws_s3_bucket.config.id}"

  accounts = [
    {
      name    = "example-account"
      roleArn = "${aws_iam_role.main.arn}"
    },
  ]
}

output "lambda_arn" {
  value = "${module.lambda.arn}"
}

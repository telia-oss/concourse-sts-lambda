# ------------------------------------------------------------------------------
# Resources
# ------------------------------------------------------------------------------
resource "aws_s3_bucket_object" "main" {
  bucket  = var.config_bucket
  key     = "${var.name}.json"
  content = local.team_config
}

resource "aws_cloudwatch_event_rule" "main" {
  name                = "concourse-${var.name}-sts-${local.config_hash}"
  description         = "STS Lambda team configuration and trigger."
  schedule_expression = "rate(30 minutes)"
}

resource "aws_lambda_permission" "main" {
  statement_id  = "concourse-${var.name}-sts-lambda-permission"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_arn
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.main.arn
}

resource "aws_cloudwatch_event_target" "main" {
  rule = aws_cloudwatch_event_rule.main.name
  arn  = var.lambda_arn

  input = <<EOF
  {
    "bucket": "${var.config_bucket}",
    "key": "${aws_s3_bucket_object.main.key}"
  }
EOF
}


locals {
  team_config = <<EOF
  {
    "name": "${var.name}",
    "accounts": ${jsonencode(var.accounts)}
  }
EOF
  config_hash = substr(md5(local.team_config), 0, 7)
}

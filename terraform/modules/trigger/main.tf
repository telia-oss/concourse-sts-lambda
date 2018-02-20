# ------------------------------------------------------------------------------
# Resources
# ------------------------------------------------------------------------------
resource "aws_cloudwatch_event_rule" "main" {
  name                = "${var.prefix}-cron-trigger"
  description         = "STS Lambda team configuration and trigger."
  schedule_expression = "rate(50 minutes)"
}

resource "aws_cloudwatch_event_target" "main" {
  rule  = "${aws_cloudwatch_event_rule.main.name}"
  arn   = "${var.lambda_arn}"
  input = "${var.team_config}"
}

resource "aws_lambda_permission" "main" {
  statement_id  = "${var.prefix}-lambda-permission"
  action        = "lambda:InvokeFunction"
  function_name = "${var.lambda_arn}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.main.arn}"
}

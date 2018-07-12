# ------------------------------------------------------------------------------
# Resources
# ------------------------------------------------------------------------------
resource "aws_cloudwatch_event_rule" "main" {
  name                = "${var.name_prefix}-sts-${substr(md5(var.team_config), 0, 7)}"
  description         = "STS Lambda team configuration and trigger."
  schedule_expression = "rate(30 minutes)"
}

resource "aws_cloudwatch_event_target" "main" {
  rule  = "${aws_cloudwatch_event_rule.main.name}"
  arn   = "${var.lambda_arn}"
  input = "${var.team_config}"
}

resource "aws_lambda_permission" "main" {
  statement_id  = "${var.name_prefix}-sts-lambda-permission"
  action        = "lambda:InvokeFunction"
  function_name = "${var.lambda_arn}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.main.arn}"
}

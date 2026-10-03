# Alerts before the bill surprises anyone: this project is meant to run for a
# few dollars a month and the EC2 node is the only real cost.
resource "aws_budgets_budget" "monthly" {
  name         = "${var.project_name}-monthly"
  budget_type  = "COST"
  limit_amount = var.monthly_budget_usd
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = compact([var.budget_alert_email])
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = compact([var.budget_alert_email])
  }

  # CI (PR plan, nightly drift check) does not know the address, so it plans
  # with an empty list. Without this every CI run would report a diff.
  lifecycle {
    ignore_changes = [notification]
  }
}

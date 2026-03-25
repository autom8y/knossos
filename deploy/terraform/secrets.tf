# Secrets Manager — placeholder entries for Clew secrets.
#
# These are created EMPTY. The operator must populate them manually
# via the AWS Console or CLI before the first deploy:
#
#   aws secretsmanager put-secret-value \
#     --secret-id clew/slack-signing-secret \
#     --secret-string "YOUR_VALUE"

resource "aws_secretsmanager_secret" "slack_signing_secret" {
  name                    = "clew/slack-signing-secret"
  description             = "Slack app signing secret for HMAC-SHA256 request verification"
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret" "slack_bot_token" {
  name                    = "clew/slack-bot-token"
  description             = "Slack bot OAuth token (xoxb-*) for API calls"
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret" "anthropic_api_key" {
  name                    = "clew/anthropic-api-key"
  description             = "Anthropic Claude API key for reasoning pipeline"
  recovery_window_in_days = 7
}

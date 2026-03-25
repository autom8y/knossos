# Clew Infrastructure -- Operator Runbook

## Prerequisites

- [ ] AWS CLI v2 installed and configured (`aws sts get-caller-identity`)
- [ ] Terraform >= 1.5 installed (`terraform -version`)
- [ ] AWS account with permissions: IAM, ECS, ECR, ALB, Secrets Manager, CloudWatch, VPC
- [ ] Existing VPC with at least 2 public subnets (with internet gateway)
- [ ] ACM certificate issued and validated for your custom domain (e.g., `clew.yourdomain.com`)
- [ ] Custom domain -- required because ALB DNS cannot have an ACM certificate
- [ ] GitHub repo: `autom8y/knossos` (or adjust `github_org` / `github_repo`)

## 1. Configure Variables

```bash
cd deploy/terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` with real values:

```hcl
aws_account_id      = "YOUR_12_DIGIT_ACCOUNT_ID"
aws_region          = "us-east-1"
vpc_id              = "vpc-xxxxxxxxxxxxxxxxx"
public_subnet_ids   = ["subnet-xxxxxxxxxxxxxxxxx", "subnet-yyyyyyyyyyyyyyyyy"]
acm_certificate_arn = "arn:aws:acm:us-east-1:ACCOUNT_ID:certificate/CERT_ID"
```

**Never commit `terraform.tfvars` -- it is gitignored.**

## 2. Initialize and Apply

```bash
terraform init
terraform plan -out=tfplan
```

Review the plan (~20 resources). Then apply and save outputs:

```bash
terraform apply tfplan
```

```bash
terraform output -json > terraform-outputs.json
export ALB_DNS=$(terraform output -raw alb_dns_name)
export ECR_URL=$(terraform output -raw ecr_repository_url)
export OIDC_ROLE=$(terraform output -raw oidc_role_arn)
export ECS_CLUSTER=$(terraform output -raw ecs_cluster_name)
export ECS_SERVICE=$(terraform output -raw ecs_service_name)
```

**Potential blocker**: If a GitHub Actions OIDC provider already exists in this AWS account, `terraform apply` will fail with `EntityAlreadyExists`. Import it:

```bash
EXISTING_ARN=$(aws iam list-open-id-connect-providers --query 'OpenIDConnectProviderList[?contains(Arn, `github`)].Arn' --output text)
terraform import aws_iam_openid_connect_provider.github_actions "$EXISTING_ARN"
# Then re-run terraform apply
```

## 3. Create Slack App

Create the Slack app per `deploy/SETUP.md` Phase 3. **Do NOT set the Request URL during app creation** -- it must be configured after the app is deployed and healthy (see Step 8).

1. Go to https://api.slack.com/apps
2. Click **Create New App** > **From an app manifest**
3. Paste `deploy/slack-app-manifest.yml` -- remove the `request_url` line before submitting
4. Click **Install to Workspace** and approve the scopes

Collect credentials:

| Value | Location | Format |
|-------|----------|--------|
| **Signing Secret** | Basic Information > App Credentials | 64-character hex string |
| **Bot User OAuth Token** | OAuth & Permissions > Bot User OAuth Token | `xoxb-...` |

## 4. Populate Secrets

All three secrets are created empty by Terraform. Populate before the first deploy:

```bash
aws secretsmanager put-secret-value \
  --secret-id clew/slack-signing-secret \
  --secret-string "YOUR_SLACK_SIGNING_SECRET"

aws secretsmanager put-secret-value \
  --secret-id clew/slack-bot-token \
  --secret-string "xoxb-YOUR-SLACK-BOT-TOKEN"

aws secretsmanager put-secret-value \
  --secret-id clew/anthropic-api-key \
  --secret-string "sk-ant-YOUR-ANTHROPIC-KEY"
```

## 5. Configure GitHub

Set the secret and variable required by `.github/workflows/deploy-clew.yml`:

```bash
# Secret (GitHub CLI)
gh secret set AWS_ACCOUNT_ID --body "YOUR_12_DIGIT_ACCOUNT_ID"

# Variable (may be empty to disable tracing)
gh variable set OTEL_ENDPOINT --body ""
```

The workflow derives other values from naming conventions matching Terraform resources.

## 6. Validate Infrastructure

```bash
# ECR repository exists
aws ecr describe-repositories --repository-names clew

# ECS cluster is ACTIVE
aws ecs describe-clusters --clusters clew-cluster \
  --query 'clusters[0].status'

# ALB is provisioned (DNS should resolve)
echo "ALB endpoint: https://$ALB_DNS"
curl -sk "https://$ALB_DNS/ready" || echo "Expected: 503 until first deploy"

# Secrets exist (values not shown)
aws secretsmanager list-secrets \
  --filters Key=name,Values=clew/ \
  --query 'SecretList[].Name'

# CloudWatch log group exists
aws logs describe-log-groups \
  --log-group-name-prefix /ecs/clew \
  --query 'logGroups[].logGroupName'
```

ECS shows 0 running tasks until the first CI deploy -- this is expected.

## 7. DNS Setup

Point your domain's CNAME (or Route 53 alias) at the ALB DNS name:

```
clew.yourdomain.com  CNAME  ${ALB_DNS}
```

**Do NOT update the Slack app's Request URL yet.** That happens in Step 8 after the first deploy.

## 8. First Deploy

```bash
gh workflow run "Deploy Clew"   # or push to main on a monitored path
gh run watch                     # monitor progress
```

After deploy completes, verify the application is healthy:

```bash
aws ecs describe-services --cluster clew-cluster --services clew-service \
  --query 'services[0].{status:status,running:runningCount,deployments:length(deployments)}'
# Expected: status: ACTIVE, running: 1, deployments: 1

curl -s https://clew.yourdomain.com/health | jq .
# Expected: {"status":"ok"}

curl -s https://clew.yourdomain.com/ready | jq .
# Expected: {"status":"ready",...} with all checks "ok"
```

## 9. Request URL Activation (MUST BE LAST)

**This step MUST happen after the application is running and healthy.** Slack sends a `url_verification` challenge within 3 seconds of setting the Request URL. If the app is not running, verification will fail.

1. Go to https://api.slack.com/apps > your Clew app > **Event Subscriptions**
2. Toggle **Enable Events** to On (if not already)
3. Set **Request URL** to: `https://clew.yourdomain.com/slack/events`
4. Wait for the green checkmark confirming verification passed

**Verification**:

```bash
aws logs filter-log-events \
  --log-group-name /ecs/clew \
  --filter-pattern '"url_verification"' \
  --start-time $(date -v-5M +%s000 2>/dev/null || date -d '5 minutes ago' +%s000)
# Expected: log entry showing successful challenge response
```

If verification fails, check:
- Is the app running? `curl -s https://clew.yourdomain.com/ready`
- Does DNS resolve? `dig clew.yourdomain.com`
- Is the ALB security group open on 443?
- Is the TLS certificate valid? `openssl s_client -connect clew.yourdomain.com:443`

## Known Limitations

- **Local state** -- migrate to S3+DynamoDB before adding a second operator
- **Single task** -- `desired_count=1`, no multi-task redundancy (circuit breaker handles bad deploys)
- **No WAF** -- ALB accepts all internet traffic; add WAF if needed
- **OIDC provider** -- one per account; import existing if conflict arises
- **Deletion protection off** -- enable on ALB after confirming stack stability
- **No secret rotation** -- rotate manually when credentials change

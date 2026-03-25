# Clew Deployment Setup Guide

Complete operator walkthrough to deploy [@clew](https://github.com/autom8y/knossos) on AWS ECS Fargate.

---

## Prerequisites

Before starting, confirm you have:

- [ ] **AWS Account** with admin access (or permissions to create IAM roles, ECS, ALB, ECR, Secrets Manager, CloudWatch)
- [ ] **Terraform >= 1.5** installed locally (`terraform version`)
- [ ] **AWS CLI v2** configured with credentials (`aws sts get-caller-identity`)
- [ ] **Existing VPC** with at least 2 public subnets (with internet gateway route)
- [ ] **ACM Certificate** for your domain (validated, in the same region as deployment)
- [ ] **GitHub admin access** to autom8y/knossos (for secrets/variables configuration)
- [ ] **Slack workspace admin access** (for app creation)

---

## Step 1: Provision Infrastructure with Terraform

```bash
cd deploy/terraform

# Copy and fill in your variables
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your real values:
#   aws_account_id, vpc_id, public_subnet_ids, acm_certificate_arn

# Initialize Terraform (downloads AWS provider)
terraform init

# Preview what will be created
terraform plan

# Apply (creates ~20 resources)
terraform apply
```

**Save the outputs** -- you will need them in subsequent steps:

```bash
terraform output
# alb_dns_name         = "clew-alb-XXXXXXXXX.us-east-1.elb.amazonaws.com"
# ecr_repository_url   = "123456789012.dkr.ecr.us-east-1.amazonaws.com/clew"
# oidc_role_arn        = "arn:aws:iam::123456789012:role/clew-github-actions"
# ecs_cluster_name     = "clew-cluster"
# ecs_service_name     = "clew-service"
```

### What Terraform Creates

| Resource | Name | Purpose |
|----------|------|---------|
| ECR Repository | `clew` | Docker image storage |
| ECS Cluster | `clew-cluster` | Fargate compute cluster |
| ECS Service | `clew-service` | Runs 1 Fargate task |
| ECS Task Definition | `clew` | Container config (256 CPU, 512 MB) |
| ALB | `clew-alb` | HTTPS termination + routing |
| Target Group | `clew-tg` | Health-checked target (/ready on 8080) |
| CloudWatch Log Group | `/ecs/clew` | 30-day log retention |
| Secrets Manager | 3 secrets | Empty placeholders (populated in Step 3) |
| IAM Roles | 3 roles | Execution, task, GitHub Actions OIDC |
| Security Groups | 2 SGs | ALB (443 in), ECS (8080 from ALB) |

---

## Step 2: Create the Slack App

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Click **Create New App** > **From an app manifest**
3. Select your workspace
4. Paste the contents of `deploy/slack-app-manifest.yml`
5. **Replace `${ALB_DNS}`** with the `alb_dns_name` from Terraform output
6. Review and click **Create**

After creation, collect these values from the Slack app settings:

| Value | Location in Slack UI | Destination |
|-------|---------------------|-------------|
| **Signing Secret** | Basic Information > App Credentials | `clew/slack-signing-secret` |
| **Bot User OAuth Token** | OAuth & Permissions > Bot User OAuth Token | `clew/slack-bot-token` |

---

## Step 3: Store Secrets in AWS Secrets Manager

Populate the three empty secrets that Terraform created:

```bash
# Slack signing secret (from Step 2)
aws secretsmanager put-secret-value \
  --secret-id clew/slack-signing-secret \
  --secret-string "YOUR_SLACK_SIGNING_SECRET"

# Slack bot token (from Step 2)
aws secretsmanager put-secret-value \
  --secret-id clew/slack-bot-token \
  --secret-string "xoxb-YOUR-SLACK-BOT-TOKEN"

# Anthropic API key (from console.anthropic.com)
aws secretsmanager put-secret-value \
  --secret-id clew/anthropic-api-key \
  --secret-string "sk-ant-YOUR-ANTHROPIC-API-KEY"
```

Verify all three are populated:

```bash
for secret in clew/slack-signing-secret clew/slack-bot-token clew/anthropic-api-key; do
  echo -n "$secret: "
  aws secretsmanager get-secret-value --secret-id "$secret" --query 'Name' --output text 2>/dev/null \
    && echo "OK" || echo "MISSING"
done
```

---

## Step 4: Configure GitHub Secrets and Variables

In the GitHub repository settings (Settings > Secrets and variables > Actions):

### Secrets (Settings > Secrets > Actions > New repository secret)

| Name | Value | Source |
|------|-------|--------|
| `AWS_ACCOUNT_ID` | Your 12-digit AWS account ID | `aws sts get-caller-identity --query Account --output text` |

### Variables (Settings > Variables > Actions > New repository variable)

| Name | Value | Source |
|------|-------|--------|
| `OTEL_ENDPOINT` | OTLP HTTP endpoint (e.g., `http://collector:4318`) or empty string | Your observability stack |

---

## Step 5: DNS Configuration

Point your domain at the ALB:

1. Get the ALB DNS name: `terraform -chdir=deploy/terraform output alb_dns_name`
2. Create a CNAME record in your DNS provider:
   - **Name**: `clew.yourdomain.com` (must match ACM certificate)
   - **Type**: CNAME
   - **Value**: the ALB DNS name

3. Update the Slack app's Request URL:
   - Go to [api.slack.com/apps](https://api.slack.com/apps) > your app > Event Subscriptions
   - Set Request URL to: `https://clew.yourdomain.com/slack/events`
   - Slack will send a challenge request -- the app must be running to verify

---

## Step 6: First Deploy

### Option A: Via GitHub Actions (recommended)

Trigger the workflow manually:

1. Go to Actions > "Deploy Clew" > "Run workflow"
2. Select `main` branch
3. Click "Run workflow"

Or push any change to files in the trigger paths to `main`.

### Option B: Manual first deploy

If the CI pipeline is not yet configured, push the first image manually:

```bash
# Authenticate with ECR
ECR_REGISTRY=$(terraform -chdir=deploy/terraform output -raw ecr_repository_url | cut -d/ -f1)
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REGISTRY

# Build and push
docker build -f deploy/Dockerfile -t $ECR_REGISTRY/clew:latest .
docker push $ECR_REGISTRY/clew:latest

# Force a new deployment of the service
aws ecs update-service \
  --cluster clew-cluster \
  --service clew-service \
  --force-new-deployment
```

---

## Step 7: Verification

### Check ECS service health

```bash
aws ecs describe-services \
  --cluster clew-cluster \
  --services clew-service \
  --query 'services[0].{status:status,desired:desiredCount,running:runningCount}'
# Expected: {"status": "ACTIVE", "desired": 1, "running": 1}
```

### Check application health endpoints

```bash
# Liveness (always 200 if process is running)
curl -s https://clew.yourdomain.com/health | jq .
# Expected: {"status":"ok"}

# Readiness (200 only when all checks pass)
curl -s https://clew.yourdomain.com/ready | jq .
# Expected: {"status":"ready","checks":{"slack":"ok","reasoning":"ok","catalog":"ok","search_index":"ok","claude_api":"ok"}}
```

### Test Slack integration

1. Open Slack and DM @Clew
2. Send: "What does our codebase do?"
3. Verify you receive a response within 30 seconds

### Check logs

```bash
aws logs tail /ecs/clew --follow --since 5m
# Look for: "server configured" with port=8080, pipeline_ready=true
```

---

## Troubleshooting

### ECS task won't start

```bash
# Check stopped task reason
aws ecs list-tasks --cluster clew-cluster --desired-status STOPPED --query 'taskArns[0]' --output text \
  | xargs -I{} aws ecs describe-tasks --cluster clew-cluster --tasks {} \
    --query 'tasks[0].{reason:stoppedReason,exitCode:containers[0].exitCode}'

# Common causes:
# - "CannotPullContainerError" -> ECR image doesn't exist, push one first
# - "ResourceInitializationError" -> Secrets Manager access denied, check execution role
# - Exit code 1 -> Check /ecs/clew logs for application startup errors
```

### Health check failing

```bash
# See which readiness check fails
curl -s https://clew.yourdomain.com/ready | jq .
# If a specific check fails:
#   claude_api  -> Verify ANTHROPIC_API_KEY secret is set and valid
#   slack       -> Verify SLACK_BOT_TOKEN is valid (not expired)
#   catalog     -> Non-critical, domain catalog may be empty on first run
```

### Slack not receiving events

1. Verify Request URL is correct in Slack app settings
2. Check that the ALB security group allows 443 inbound
3. Check that DNS resolves: `dig clew.yourdomain.com`
4. Look for Slack verification failures in logs:
   ```bash
   aws logs filter-log-events \
     --log-group-name /ecs/clew \
     --filter-pattern '"verification failed"' \
     --start-time $(date -d '1 hour ago' +%s000)
   ```

### Rollback

ECS circuit breaker handles automatic rollback on health check failure during deploy. For manual rollback:

```bash
# List recent task definitions
aws ecs list-task-definitions --family-prefix clew --sort DESC --max-items 5

# Roll back to a specific revision
aws ecs update-service \
  --cluster clew-cluster \
  --service clew-service \
  --task-definition clew:PREVIOUS_REVISION \
  --force-new-deployment

# Wait for stability
aws ecs wait services-stable --cluster clew-cluster --services clew-service
```

---

## Architecture

```
                    Internet
                       |
                  [ Route 53 ]
                       |
              [ ALB (HTTPS:443) ]     <-- ACM cert, clew-alb-sg
                       |
              [ Target Group :8080 ]  <-- /ready health check
                       |
              [ ECS Fargate Task ]    <-- clew-ecs-sg (8080 from ALB only)
                   /       \
          [ Slack API ]  [ Claude API ]   <-- outbound via public subnet
                              |
                    [ Secrets Manager ]    <-- signing secret, bot token, API key
```

---

## Cost Estimate (MVP)

| Resource | Monthly Cost (approx.) |
|----------|----------------------|
| Fargate (256 CPU, 512 MB, 1 task, 24/7) | ~$9 |
| ALB (low traffic) | ~$16 + $0.008/LCU-hour |
| CloudWatch Logs (30-day, <1 GB) | ~$0.50 |
| Secrets Manager (3 secrets) | ~$1.20 |
| ECR (< 1 GB images) | ~$0.10 |
| **Total** | **~$27/month** |

Claude API costs are billed separately via your Anthropic account.

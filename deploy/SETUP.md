# Clew Deployment Setup Guide

Complete operator walkthrough to deploy [@clew](https://github.com/autom8y/knossos) on AWS ECS Fargate.

---

## Phase 1: Prerequisites

Complete ALL prerequisites before running any Terraform commands. The ACM certificate is on the critical path and can take up to 72 hours with email validation -- use DNS validation for fastest turnaround.

- [ ] **AWS Account** with admin access (or permissions to create IAM roles, ECS, ALB, ECR, Secrets Manager, CloudWatch, VPC)
- [ ] **Terraform >= 1.5** installed locally (`terraform version`)
- [ ] **AWS CLI v2** configured with credentials (`aws sts get-caller-identity`)
- [ ] **Existing VPC** with at least 2 public subnets (with internet gateway route for Fargate public IP egress)
- [ ] **ACM Certificate** issued and validated for your custom domain (e.g., `clew.yourdomain.com`). Must be in the same region as deployment. Certificate CN/SAN must match your custom domain.
- [ ] **Custom domain** -- required because ALB DNS (`*.elb.amazonaws.com`) cannot have an ACM certificate. The Slack Request URL MUST use a custom domain with a valid CA-signed certificate.
- [ ] **Anthropic API key** -- obtain from console.anthropic.com (format: `sk-ant-...`)
- [ ] **GitHub admin access** to autom8y/knossos (for secrets/variables configuration)
- [ ] **Slack workspace admin access** (for app creation and installation)

**Critical ordering note**: The ACM certificate MUST be issued and validated before `terraform apply`. The ALB HTTPS listener creation will fail without it.

| Prerequisite | How to Verify |
|-------------|---------------|
| AWS account access | `aws sts get-caller-identity` |
| Terraform version | `terraform version` |
| VPC subnets | `aws ec2 describe-subnets --filters Name=vpc-id,Values=vpc-xxx` |
| ACM certificate | `aws acm describe-certificate --certificate-arn <arn>` -- Status: ISSUED |
| Custom domain | `dig clew.yourdomain.com` (after DNS setup in Phase 6) |

---

## Phase 2: Infrastructure Provisioning

```bash
cd deploy/terraform

# Copy and fill in your variables
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` with real values:

```hcl
aws_account_id      = "123456789012"
aws_region          = "us-east-1"
vpc_id              = "vpc-xxxxxxxxxxxxxxxxx"
public_subnet_ids   = ["subnet-xxxxxxxxxxxxxxxxx", "subnet-yyyyyyyyyyyyyyyyy"]
acm_certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/YOUR-CERT-ID"
```

```bash
# Initialize Terraform (downloads AWS provider)
terraform init

# Preview what will be created (~20 resources)
terraform plan -out=tfplan

# Apply
terraform apply tfplan

# Save outputs (you will need these in later phases)
terraform output -json > terraform-outputs.json
export ALB_DNS=$(terraform output -raw alb_dns_name)
export ECR_URL=$(terraform output -raw ecr_repository_url)
export OIDC_ROLE=$(terraform output -raw oidc_role_arn)
export ECS_CLUSTER=$(terraform output -raw ecs_cluster_name)
export ECS_SERVICE=$(terraform output -raw ecs_service_name)
```

**Verification**:

```bash
# ECR repository exists
aws ecr describe-repositories --repository-names clew --query 'repositories[0].repositoryUri'
# Expected: "123456789012.dkr.ecr.us-east-1.amazonaws.com/clew"

# ECS cluster is ACTIVE
aws ecs describe-clusters --clusters clew-cluster --query 'clusters[0].status' --output text
# Expected: ACTIVE

# Secrets exist (empty)
aws secretsmanager list-secrets --filters Key=name,Values=clew/ --query 'SecretList[].Name' --output table
# Expected: clew/slack-signing-secret, clew/slack-bot-token, clew/anthropic-api-key
```

**Expected state after Phase 2**: ECS service exists but shows 0 running tasks (no container image yet). This is expected -- the deployment circuit breaker handles this gracefully.

**Potential blocker**: If a GitHub Actions OIDC provider already exists in this AWS account, `terraform apply` will fail with `EntityAlreadyExists`. Resolution:

```bash
# Import the existing provider
EXISTING_ARN=$(aws iam list-open-id-connect-providers --query 'OpenIDConnectProviderList[?contains(Arn, `github`)].Arn' --output text)
terraform import aws_iam_openid_connect_provider.github_actions "$EXISTING_ARN"
# Then re-run terraform apply
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
| Secrets Manager | 3 secrets | Empty placeholders (populated in Phase 4) |
| IAM Roles | 3 roles | Execution, task, GitHub Actions OIDC |
| Security Groups | 2 SGs | ALB (443 in), ECS (8080 from ALB) |

---

## Phase 3: Slack App Creation (WITHOUT Request URL)

**IMPORTANT: Remove or leave blank the `request_url` line before creating the app.** The Request URL can only be set after the app is deployed and healthy (Phase 7), because Slack immediately sends a `url_verification` challenge that must be answered within 3 seconds.

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Click **Create New App** > **From an app manifest**
3. Select your target workspace
4. Paste the contents of `deploy/slack-app-manifest.yml`
5. **Before submitting**: Remove the entire `request_url` line, or leave it blank
6. Review scopes and events, then click **Create**
7. Click **Install to Workspace** and approve the requested scopes

After creation and installation, collect credentials:

| Value | Location in Slack UI | Format |
|-------|---------------------|--------|
| **Signing Secret** | Basic Information > App Credentials > Signing Secret | 64-character hex string |
| **Bot User OAuth Token** | OAuth & Permissions > Bot User OAuth Token | `xoxb-...` (only available after installing to workspace) |

**Verification**: Both values are visible in the Slack app settings UI. The Bot Token only appears after clicking "Install to Workspace."

---

## Phase 4: Secrets Population

Populate the three empty secrets that Terraform created in Phase 2:

```bash
# Slack signing secret (from Phase 3)
aws secretsmanager put-secret-value \
  --secret-id clew/slack-signing-secret \
  --secret-string "YOUR_SLACK_SIGNING_SECRET_HERE"

# Slack bot token (from Phase 3)
aws secretsmanager put-secret-value \
  --secret-id clew/slack-bot-token \
  --secret-string "xoxb-YOUR-SLACK-BOT-TOKEN-HERE"

# Anthropic API key (from Prerequisites)
aws secretsmanager put-secret-value \
  --secret-id clew/anthropic-api-key \
  --secret-string "sk-ant-YOUR-ANTHROPIC-API-KEY-HERE"
```

**Verification**:

```bash
for secret in clew/slack-signing-secret clew/slack-bot-token clew/anthropic-api-key; do
  echo -n "$secret: "
  aws secretsmanager get-secret-value --secret-id "$secret" --query 'Name' --output text 2>/dev/null \
    && echo "OK" || echo "MISSING"
done
# Expected: all three show "OK"
```

---

## Phase 5: CI/CD Configuration

In the GitHub repository settings (Settings > Secrets and variables > Actions):

### Secrets (Settings > Secrets > Actions > New repository secret)

| Name | Value | Source |
|------|-------|--------|
| `AWS_ACCOUNT_ID` | Your 12-digit AWS account ID | `aws sts get-caller-identity --query Account --output text` |

### Variables (Settings > Variables > Actions > New repository variable)

| Name | Value | Source |
|------|-------|--------|
| `OTEL_ENDPOINT` | OTLP HTTP endpoint (e.g., `http://collector:4318`) or empty string | Your observability stack (leave empty to disable tracing) |

**Verification**:

```bash
# Using GitHub CLI
gh secret list | grep AWS_ACCOUNT_ID
# Expected: AWS_ACCOUNT_ID  Updated <timestamp>

gh variable list | grep OTEL_ENDPOINT
# Expected: OTEL_ENDPOINT  <value or empty>  Updated <timestamp>
```

---

## Phase 6: DNS Configuration and First Deploy

These can be done in parallel.

### 6A. DNS Setup

Create a CNAME record pointing your custom domain to the ALB:

| Record | Type | Value |
|--------|------|-------|
| `clew.yourdomain.com` | CNAME | Value of `$ALB_DNS` (e.g., `clew-alb-xxx.us-east-1.elb.amazonaws.com`) |

If using Route 53, create an Alias record instead of CNAME (avoids CNAME at zone apex issues).

**Verification**:

```bash
dig clew.yourdomain.com CNAME +short
# Expected: clew-alb-xxx.us-east-1.elb.amazonaws.com.
```

### 6B. First Deploy

**Option A -- Via GitHub Actions (recommended)**:

```bash
# Trigger manually
gh workflow run "Deploy Clew"
gh run watch
```

Or push any change to files in the trigger paths to `main`.

**Option B -- Manual first deploy** (if CI is not yet configured):

```bash
# Authenticate with ECR
ECR_REGISTRY=$(echo "$ECR_URL" | cut -d/ -f1)
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin "$ECR_REGISTRY"

# Build and push
docker build -f deploy/Dockerfile -t "$ECR_URL:latest" .
docker push "$ECR_URL:latest"

# Force new deployment
aws ecs update-service \
  --cluster clew-cluster \
  --service clew-service \
  --force-new-deployment
```

**Wait for ECS service stability** (~2-5 minutes):

```bash
aws ecs wait services-stable --cluster clew-cluster --services clew-service
```

**Verification**:

```bash
# ECS service health
aws ecs describe-services \
  --cluster clew-cluster \
  --services clew-service \
  --query 'services[0].{status:status,desired:desiredCount,running:runningCount}'
# Expected: {"status": "ACTIVE", "desired": 1, "running": 1}

# Application health (requires DNS to have propagated)
curl -s https://clew.yourdomain.com/health | jq .
# Expected: {"status":"ok"}

curl -s https://clew.yourdomain.com/ready | jq .
# Expected: {"status":"ready","checks":{"slack":"ok","reasoning":"ok","catalog":"ok","search_index":"ok","claude_api":"ok"}}

# TLS certificate validation
openssl s_client -connect clew.yourdomain.com:443 -servername clew.yourdomain.com </dev/null 2>/dev/null | openssl x509 -noout -dates
# Expected: shows valid notBefore/notAfter dates
```

**Check logs for successful startup**:

```bash
aws logs tail /ecs/clew --follow --since 5m
# Look for: "server configured" with port=8080, pipeline_ready=true
```

---

## Phase 7: Request URL Activation (MUST BE LAST)

**This phase MUST happen after the application is running and healthy.** The Request URL can only be set after the app is deployed because Slack immediately sends a `url_verification` challenge that must be answered within 3 seconds. If the app is not running, verification will fail and Slack will reject the URL.

1. Go to [api.slack.com/apps](https://api.slack.com/apps) > your Clew app
2. Navigate to **Event Subscriptions**
3. Toggle **Enable Events** to On (if not already)
4. Set **Request URL** to: `https://clew.yourdomain.com/slack/events`
5. Slack will immediately send a `url_verification` challenge
6. Wait for the green checkmark confirming verification passed

**Verification**:

```bash
# Check logs for the challenge
aws logs filter-log-events \
  --log-group-name /ecs/clew \
  --filter-pattern '"url_verification"' \
  --start-time $(date -v-5M +%s000 2>/dev/null || date -d '5 minutes ago' +%s000)
# Expected: log entry showing successful challenge response
```

**If verification fails**, check:
- Is the app running? `curl -s https://clew.yourdomain.com/ready`
- Does DNS resolve? `dig clew.yourdomain.com`
- Is the ALB security group open on 443? Check inbound rules.
- Is the TLS certificate valid? `openssl s_client -connect clew.yourdomain.com:443`

---

## Phase 8: Smoke Test

Run all four tiers to confirm end-to-end functionality.

### Tier 1 -- Infrastructure (Pre-Slack)

| Test | Command | Expected |
|------|---------|----------|
| Health endpoint | `curl -s https://clew.yourdomain.com/ready \| jq .` | `{"status":"ready",...}` with all checks "ok" |
| TLS valid | `openssl s_client -connect clew.yourdomain.com:443 </dev/null 2>/dev/null \| head -5` | Valid certificate chain |
| Secrets accessible | Check CloudWatch for startup errors | No `ResourceInitializationError` |

### Tier 2 -- Slack Connectivity

| Test | Method | Expected |
|------|--------|----------|
| URL verification | Completed in Phase 7 | Green checkmark in Slack |
| Auth test | Check logs for `auth.test` response | Bot user info in logs |
| Invalid signature | Send crafted request with wrong signature | 401 Unauthorized |

### Tier 3 -- Event Delivery

| Test | Method | Expected |
|------|--------|----------|
| Thread started | Open Clew in Slack assistant (top bar or DM) | Suggested prompts appear within 1-2 seconds |
| Message processing | Send "What is the architecture of this project?" | Response with citations within 30 seconds |
| Status indicator | Send any message | "Searching knowledge..." status visible during processing |
| Thread title | Send a message | Thread title set (first 60 chars of question) |

### Tier 4 -- Edge Cases

| Test | Method | Expected |
|------|--------|----------|
| Concurrent requests | Send 6+ messages rapidly | 5 processed, 6th gets rate-limited response |
| Empty message | Send message with only whitespace | No pipeline invocation (check logs) |
| Duplicate event | Check logs during normal operation | "duplicate event filtered" entries for Slack retries |

---

## Common Failure Modes

| Symptom | Likely Cause | Resolution |
|---------|-------------|------------|
| No events received | URL verification failed or Request URL not set | Complete Phase 7; check ALB DNS, SGs, TLS cert |
| Events received but no response | Bot token invalid or missing | Verify `xoxb-` token in Secrets Manager; redeploy ECS task |
| `missing_scope` error | Manifest scopes not applied | Reinstall app to workspace |
| 401 on all requests | Signing secret mismatch | Re-copy signing secret to Secrets Manager; redeploy |
| Duplicate responses | Dedup map lost on restart | Expected during deploys; accept for MVP |
| Status stuck on "thinking" | Pipeline error without status clear | Check `processMessage` error handling in logs |
| `CannotPullContainerError` | No image in ECR | Push first image (Phase 6B Option B) |
| `ResourceInitializationError` | Secrets Manager access denied | Check execution role IAM policy |

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

1. Verify Request URL is correct in Slack app settings (Phase 7)
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

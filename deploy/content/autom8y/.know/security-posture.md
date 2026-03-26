---
domain: security-posture
generated_at: "2026-03-11T11:00:00Z"
expires_after: "14d"
source_scope:
  - ".github/workflows/*.yml"
  - "tools/pip-audit-gate.py"
  - "terraform/environments/production/guardduty.tf"
  - ".gitleaks.toml"
  - ".semgrep-security.yml"
  - ".pre-commit-config.yaml"
generator: architect
source_hash: "e571a47"
confidence: 0.90
format_version: "1.0"
---

# Security Posture

Post-campaign security posture for the autom8y platform.
Governed by ADR-SEC-GATE-POLICY (CONSTITUTIONAL).

## SARIF Pipeline Architecture

### Producer Inventory

| # | Tool | SARIF Category | Workflow | Status | EPSS? |
|---|------|---------------|----------|--------|-------|
| 1 | Gitleaks | `gitleaks` | sdk-ci.yml | ACTIVE | No |
| 2 | Trivy Container | `trivy-container` | service-build.yml, security-scan.yml | ACTIVE | Yes |
| 3 | Checkov | `checkov-{service}` | terraform-plan-reusable.yml | ACTIVE | No |
| 4 | CodeQL Python | `/language:python` | codeql-analysis.yml | ACTIVE | No |
| 5 | Semgrep Security | `semgrep-security` | sdk-ci.yml | TRIAL | No |
| 6 | zizmor | `zizmor` | zizmor.yml | ACTIVE | No |

Cross-repo producers (a8 repository):

| # | Tool | SARIF Category | Status |
|---|------|---------------|--------|
| 7 | Gitleaks | `gitleaks` | ACTIVE |
| 8 | Trivy IaC | `trivy-iac` | ACTIVE |
| 9 | CodeQL Go | `codeql-go` | ACTIVE |
| 10 | govulncheck | `govulncheck` | ACTIVE |

### Non-SARIF Tools

| Tool | Integration | Notes |
|------|------------|-------|
| TruffleHog | Weekly scheduled scan | Full git history, verified secrets only |
| dependency-review-action | PR advisory comment | License + vulnerability checks |
| pip-audit (gate script) | `tools/pip-audit-gate.py` | EPSS + CISA KEV enrichment |

### Pipeline Flow

All producers output SARIF 2.1.0, uploaded via `github/codeql-action/upload-sarif@v3`
(SHA-pinned, Renovate-managed). Each producer uses a unique, stable `category` string.
Upload condition: `if: always() && github.event.repository.visibility == 'public'`
(DG-3 Option C — conditional upload for private repo compatibility).

## Gate Policy Engine

Governed by ADR-SEC-GATE-POLICY (CONSTITUTIONAL, location: `a8/.ledge/decisions/`).

### Severity-to-Action Matrix

| Severity | CISA KEV? | EPSS | Action | SLA |
|----------|-----------|------|--------|-----|
| CRITICAL | -- | -- | BLOCK (unconditional) | 24h |
| Any | Yes | -- | BLOCK (override to CRITICAL) | 24h |
| HIGH | -- | > 0.1 | BLOCK | 7d |
| HIGH | -- | <= 0.1 | WARN (PR annotation) | 7d |
| HIGH | -- | N/A (non-CVE) | BLOCK (fail safe) | 7d |
| MEDIUM | -- | -- | REPORT (Security tab) | 30d |
| LOW/INFO | -- | -- | REPORT (dashboard) | None |

### EPSS Enrichment

- **Threshold:** 0.1 (FIRST.org "high activity" cutoff)
- **Noise reduction:** ~80% of CVEs filtered (cluster near EPSS 0.0)
- **API:** `https://api.first.org/data/v1/epss` (free, no auth)
- **Fallback:** CVSS-only (all HIGH → BLOCK) if API unreachable

### CISA KEV Override

- **Feed:** CISA Known Exploited Vulnerabilities catalog
- **Behavior:** Any matching CVE → CRITICAL (unconditional, cannot be exempted)
- **Fallback:** Proceed without KEV enrichment + warning if feed unreachable

## SEC-* Invariants

| ID | Statement | Status | Enforcement |
|----|-----------|--------|-------------|
| SEC-01 | All tools produce SARIF | MET | 10 producers active |
| SEC-02 | Gate policy consistency | MET | ADR-SEC-GATE-POLICY governs all tools |
| SEC-03 | Config travels with fork | MET | All configs in-repo (.gitleaks.toml, workflows, pre-commit) |
| SEC-04 | No permanent advisory | MET | Semgrep TRIAL has 8-sprint max; Parliament graduated to blocking |
| SEC-05 | Scar annotations | MET | 30+ SEC-FP-* annotations on gosec findings with rationale |

### Exemption Rules

- No permanent exemptions (mandatory expiry)
- Max expiry: HIGH 30d, MEDIUM 90d, CRITICAL cannot be exempted
- Format: `// SEC-ACCEPT-{ID}: {rationale} (expires YYYY-MM-DD)`
- False positive: `// SEC-FP-{ID}: {rationale} (expires YYYY-MM-DD)`
- Trial graduation: SNR > 3:1 within 4 sprints, max 8 sprints TRIAL

## Cost Model

| Item | Monthly | Notes |
|------|---------|-------|
| SARIF tools | $0 | Open source (Gitleaks, Trivy, govulncheck, zizmor, Semgrep OSS) |
| CodeQL | $0 | Included with GitHub plan |
| EPSS + KEV APIs | $0 | Public APIs, no auth |
| harden-runner | $0 | Free tier audit mode |
| GuardDuty ECS | ~$2-32 | Depends on ECS task volume (1.75 vCPU fleet) |
| **Total** | **~$2-32** | Budget: $0-648/year |

## Runtime Monitoring

| Tool | Scope | Status |
|------|-------|--------|
| GuardDuty ECS Runtime | Container runtime threats | Terraform created, needs apply |
| harden-runner | Egress auditing | Audit mode on release + service-build |
| Schemathesis | API fuzzing | Auth service only (max_examples=200) |
| CodeRabbit | AI code review | Config created, app install pending |

## Cross-References

- **ADR:** `a8/.ledge/decisions/ADR-SEC-GATE-POLICY.md` (CONSTITUTIONAL)
- **Spec:** `.ledge/specs/SARIF-PIPELINE-ARCHITECTURE.md`
- **Constraints:** `.know/design-constraints.md` (TENSION-P014, LOAD-P009, RISK-P007/P008)
- **Scars:** `.know/scar-tissue.md` (CRITICAL-001 through CRITICAL-004, SEC-FP-* patterns)
- **Retrospective:** `.sos/wip/frames/campaign-retrospective.md`

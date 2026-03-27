---
domain: release/history
generated_at: "2026-03-09T20:00:00Z"
source_scope:
  - "./.know/release/"
generator: pipeline-monitor
source_hash: "c9d08d0"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 0
---

## Release Log

### 2026-03-09 -- ADOT sidecar AWS baseline alignment (PATCH)

| Detail | Value |
|--------|-------|
| Scope | Terraform module: ecs-otel-sidecar (health check, image pin, memory) |
| Services | auth, ads, data, asana + 9 additional services (batch push) |
| Complexity | PATCH (terraform-only, no SDK, no satellite code) |
| Outcome | PASS (13/14 services green, 1 collateral failure) |
| Duration | ~7 min (push + CI verification) |
| Commits | 23 commits pushed (c88df1a..a6ba5d7), top: a6ba5d7 |

**Changes deployed:**
- Health check: `wget CMD-SHELL` -> `/healthcheck` binary (distroless compat)
- Image: `:latest` -> `:v0.47.0` (pinned, matching gateway collector)
- Memory: 128 MB -> 512 MB (AWS minimum recommendation)
- Health check intervals: 30s/5s/3/15s -> 5s/6s/5/1s (AWS official)

**All 4 target services GREEN**: auth, ads, data, asana
**9 bonus services GREEN**: log-forwarder, auth-mysql-sync, observability, otlp-collector, reconcile-spend, codeartifact, sms, slack-alert, pull-payments

**Collateral failure (out of scope):**
- grafana apply: RED — 409 Conflict on contact point deletion (`platform-alerts-email-only`, `platform-alerts-critical`). Notification policy still references these contact points. Requires terraform fix before retry.

**Carry-forward:**
- Fix grafana notification policy -> contact point dependency before next terraform apply
- autom8y-sms: tighten autom8y-ai pin from >=0.1.0 to >=1.1.0 (from prior release)

---

### 2026-03-07 -- Satellite deploy wave (telemetry 0.5.2 bump + asana source changes)

| Detail | Value |
|--------|-------|
| Package | autom8y-telemetry@0.5.2 (SDK bump to all 4 satellites) |
| Services | autom8y-ads (ECS), autom8y-data (ECS), autom8y-sms (Lambda), autom8y-asana (ECS + Lambda) |
| Complexity | RELEASE (4 satellites, 2-phase gated execution) |
| Outcome | PASS (4/4 green, 0 failures) |
| Duration | ~32 min (Phase 1: ~14 min, Phase 2: ~13 min, monitoring overhead: ~5 min) |
| Commits | 4389088 (ads), 3221999 (data), 11495fd (sms), d0e7b41+32a4734 (asana) |

**Phase 1 (parallel -- ads, data, sms):**
- autom8y-ads: ECS (receiver run 22798201434, chain 4m 28s)
- autom8y-data: ECS (receiver run 22798264656, chain 13m 56s -- long pole due to fuller test suite)
- autom8y-sms: Lambda (receiver run 22798199280, chain 4m 19s)

**Phase 2 (gated -- asana, elevated risk):**
- autom8y-asana: ECS + Lambda (receiver run 22798523117, chain 13m 2s)
  - Source changes: lifespan.py, insights_export.py, 6 runbooks, .gitignore, ~1974 .claude/ deletions
  - All 3 test jobs passed (unit, integration, lint+type)
  - ECS deploy 316s + Lambda/Terraform deploy 64s (parallel)

**Issues encountered:**
1. Pre-commit hook failures on asana commit: coverage.json (2082 KB) caught by check-added-large-files, mypy failed due to expired CodeArtifact token. Both resolved.
2. Dependabot advisories on ads (1 high) and data (1 high + 1 moderate) -- pre-existing, informational.

**Notable:**
- Cached topology was stale: sms and ads have broader SDK footprints than prior cached map (7 deps each, not 3)
- autom8y-events 0.1.0 discovered as new SDK (zero consumers)
- autom8y-data working tree had .claude/ deletions despite state map recording dirty_scope: clean (no impact)

**Carry-forward:**
- autom8y-sms: tighten autom8y-ai pin from >=0.1.0 to >=1.1.0

---

### 2026-03-06 -- autom8y-telemetry 0.5.1 deploy wave (trace pipeline fix)

| Detail | Value |
|--------|-------|
| Package | autom8y-telemetry@0.5.1 (2 P0 bug fixes: Lambda force_flush + FastAPI FastAPIInstrumentor) |
| Services | auth, reconcile-spend, slack-alert, pull-payments, auth-mysql-sync, autom8y-data, autom8y-asana, autom8y-ads |
| Complexity | RELEASE (SDK publish + 8 service rebuilds across monorepo + 3 satellites) |
| Outcome | PASS (8/8 green, 0 failures) |
| Duration | ~25 min (execution + verification) |
| Commits | dac2866 (SDK fixes), 65900ac (knossos consolidation) |

**Deployed successfully (8/8):**
- auth: ECS (run 22773081961, 9m 3s)
- reconcile-spend: Lambda (run 22773083260, 4m 59s)
- slack-alert: Lambda (run 22773084645, 4m 1s)
- pull-payments: Lambda (run 22773085935, 5m 0s)
- auth-mysql-sync: Lambda (run 22773090696, 4m 22s)
- autom8y-data: ECS via satellite chain (receiver run 22773480980, 9m 51s)
- autom8y-asana: ECS+Lambda via satellite chain (receiver run 22773483735, 8m 23s)
- autom8y-ads: ECS via satellite chain (receiver run 22773484375, 3m 39s)

**Issues encountered:**
1. GitHub Actions cache service responded with 400 across 5 monorepo runs (non-blocking)
2. Step summary upload aborted on 4 runs (content > 1024k, non-blocking)

**Carry-forward:**
- Trace pipeline verification in Tempo (SRE exit gate -- parked session)

---

### 2026-03-05 -- Service deployment wave (SRE observability)

| Detail | Value |
|--------|-------|
| Services | pull-payments, reconcile-spend, auth-mysql-sync, autom8y-asana, autom8y-data, autom8y-sms |
| Complexity | PLATFORM (5 repos, 6 services, 2 deployment patterns) |
| Outcome | PARTIAL (5/6 green, 1 infra failure) |
| Duration | ~45 min (reconnaissance through verification) |

**Failed (1/6):**
- auth-mysql-sync: Terraform state lock contention (infra_issue, resolved in Mar 6 run)

---

### 2026-03-04 -- 6-SDK batch release (config, log, telemetry, ai, meta, auth)

| Detail | Value |
|--------|-------|
| Packages | autom8y-config@1.2.0, autom8y-log@0.5.6, autom8y-telemetry@0.4.0, autom8y-ai@1.1.0, autom8y-meta@0.2.1, autom8y-auth@1.1.1 |
| Complexity | RELEASE (3 SDKs with dependents, 4-phase topological order) |
| Outcome | PASS |
| Duration | ~45 min (2 attempts -- first blocked by mypy) |

---

### 2026-03-03 -- autom8y-stripe@1.3.1

| Detail | Value |
|--------|-------|
| Package | autom8y-stripe@1.3.1 |
| Complexity | RELEASE (auto-escalated from PATCH) |
| Outcome | PASS |
| Duration | ~25 min (3 attempts) |

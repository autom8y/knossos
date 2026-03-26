---
domain: scar-tissue
generated_at: "2026-03-25T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "429f242"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/scar-tissue.md"
land_hash: "cabe8eaab8f326f9ac8f1a7a1660becb548be0d411371c8c1d7b474626197927"
---
# Codebase Scar Tissue

## Failure Catalog Completeness

This section catalogs every identified past failure mode drawn from git commit history (66 fix/revert commits), inline code markers (SC-CANARY-, IMPLICIT-, DEF-, LOAD-, RISK-, FINDING-), and cross-session experience records.

### ECS Canary Deployment Failures (SC-CANARY series)

**SC-CANARY-03: Listener Rule Pre-Flight Contamination**
- Failure: Starting a new canary when a previous failed canary left the ALB listener rule with two active target groups caused double-routing.
- Fix location: `internal/deploy/ecs_strategy.go:45-73` — `StartCanary` now calls `DescribeListenerRule` and rejects start if `activeTrafficTGs > 1`.

**SC-CANARY-06: Listener Rule Left with 2 Target Groups Post-Lifecycle**
- Failure: After canary lifecycle completed, green TG remained attached. Subsequent canary starts contaminated.
- Fix location: `internal/deploy/ecs_strategy.go:219-228, 253-261, 400-476` — `cleanupListenerRule()` called after both Promote and Rollback.

**SC-CANARY-06-v2: ECS Async Controller Re-Adds Green TG After Cleanup**
- Failure: ECS async controller re-added green TG moments after cleanup. Needed poll-and-verify retry loop.
- Fix location: `internal/deploy/ecs_strategy.go:428-474` — `cleanupMaxRetries=3`, 5s delay.

**SC-CANARY-08: Cleanup Must Happen AFTER Strategy Switched to ROLLING**
- Failure: Cleaning up while still in CANARY strategy caused ECS to re-add green TG.
- Fix location: `internal/deploy/ecs_strategy.go:219, 254` — ordering constraint comments.

**SC-CANARY-10: Cross-Process Deploy Recovery Missing**
- Failure: `a8 deploy promote` after original process exited returned `ErrNoActiveDeployment`.
- Fix location: `cmd/a8/deploy.go:185-188` — falls back to `RecoverAndPromote`.

**SC-CANARY-11: ECS Re-Contaminates Listener Rule During Bake Period**
- Failure: No post-bake cleanup; ECS re-added green TG during bake.
- Fix location: `internal/deploy/controller.go:177-190` — `PostBakeCleanup` interface.

**SC-CANARY-13: TF Provision Leaves Listener Rule in Green=0, Blue=N Steady State**
- Failure: Pre-flight check incorrectly treated TF-provisioned steady state as contamination.
- Fix location: `internal/deploy/ecs_strategy.go:53-73` — only `activeTrafficTGs > 1` (both weights > 0) treated as contamination.

**SC-CANARY-14: TF Apply Creates New Task Def Revision, PRIMARY's Def Becomes INACTIVE**
- Failure: Starting canary off INACTIVE revision would re-deploy stale code.
- Fix location: `internal/deploy/ecs_strategy.go:94-112` — resolves latest ACTIVE task def via family name lookup.

**SC-CANARY-15: Post-Native-Promote Leaves Listener Rule in Green=N, Blue=0 State**
- Failure: Pre-flight blocked because it saw two TGs without identifying single-TG active traffic.
- Fix location: `internal/deploy/ecs_strategy.go:56` — tolerated steady state.

**SC-CANARY-SPOT-SAFETY: FARGATE_SPOT Mixed Strategy During Canary**
- Failure: Spot interruptions contaminated canary metrics.
- Fix location: `internal/deploy/ecs_strategy.go:29-105, 262-303, 443-467`; `internal/deploy/health.go:190-244`; `internal/deploy/spot_detector.go`

**SP-02 (NEW): Terraform Listener Rule Revert on Canary Services**
- Failure: Subsequent `terraform apply` reverted ALB listener rule's forward action to primary TG only while ECS tasks remained in green TG. Produced 503 errors.
- Fix location: `terraform/modules/primitives/alb-target-group/main.tf` (`create_listener_rule` variable); `terraform/modules/stacks/service-stateless/main.tf` (canary-aware listener rule with `ignore_changes = [action]`).

### AWS Client Implicit Contract Failures (IMPLICIT-WS1 series)

**IMPLICIT-WS1-01: ECS DescribeServices Returns Failures Array, Not HTTP Error**
- Failure: Code accessing `Services[0]` without checking `Failures` would panic for missing services.
- Fix location: `internal/aws/ecs_deploy_client.go:27-46`

**IMPLICIT-WS1-02: Lambda GetFunctionConcurrency Returns nil When No Reserved Concurrency**
- Failure: Nil pointer dereference when no reserved concurrency configured.
- Fix location: `internal/aws/lambda_client.go:46-50`

**IMPLICIT-WS1-03: EventBridge Returns Extended State Enum Not in "ENABLED"/"DISABLED"**
- Failure: `"ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS"` caused incorrect DISABLED report.
- Fix location: `internal/aws/eventbridge_client.go:36-41` — `strings.HasPrefix(state, "ENABLED")`.

### Mock State Drift (IMPLICIT-WS2 series)

**IMPLICIT-WS2-01: Mock Mutation Methods Did Not Update State**
- Failure: Mock mutation methods returned stale data; tests passed incorrectly.
- Fix locations: `internal/aws/mock.go:576,609,681,814,824`; `internal/deploy/store_mock.go:7`; `internal/tfstate/mock.go:9,32,43`

**IMPLICIT-WS2-03: Reconcile Apply Must Be Sequential**
- Failure: Concurrent AWS mutations produced non-deterministic ordering and race conditions.
- Fix location: `cmd/a8/reconcile.go:361-362`

### Flag/State Global Leak

**TENSION-006: Package-Level Flag Variables Leaked Between Tests**
- Failure: 72 package-level flag globals caused test pollution.
- Fix location: `cmd/a8/cmd_test.go:61-76` — `resetCobraFlags()` helper.

### Nil-Pointer / Missing Initialization

**DEF-WS2-001: buildClients Omitted Terraform Field**
- Failure: Nil `Terraform` field caused panic on Terraform drift.
- Fix location: `cmd/a8/cmd_test.go:858-898` (regression test); production fix in `4d99274`.

**DEF-009: Terraform Init Must Precede Plan/Apply**
- Failure: `tf plan/apply` without `terraform init` failed on fresh directories.
- Fix location: `cmd/a8/tf.go:155, 170`; `internal/reconcile/executor.go:254`.

**DEF-01: Terraform Module Scanner False Positive on Repos Ending in "a8"**
- Failure: Regex matched repos like `my-extra8.git`.
- Fix location: `internal/tfmod/scanner.go:15-17` — regex anchored to `/a8\.git//`.

### Train / Release Gate Defects

**DEF-002: Lambda Concurrency Query Failure Masked as NONE**
- Failure: `QUERY_FAILED` status masked as `NONE`.
- Fix: Commit `2f4f0b4`.

**DEF-003 / DEF-004: Train Promote Flag Validation**
- Failure: `--check-only` + `--override` silently allowed; `--override` without `--reason` accepted.
- Fix location: `cmd/a8/train.go:169-177`.

### Security Vulnerabilities (L-series)

**L-001: Shell Metacharacter Injection in Manifest Fields**
- Failure: `org.domain` and `org.env_prefix` used in shell commands without sanitization.
- Fix location: `pkg/manifest/` — `ValidateAll()` with `containsDisallowedChars()`.

**L-002: Stale Advisory Lock Files From Crashed Processes**
- Failure: Crashed process left lock file blocking all subsequent operations.
- Fix location: `internal/deploy/store_file.go`; `internal/reconcile/watch_state.go:113-118` — 10-minute stale lock detection.

**L-003: TerraformRunner Creation Errors Swallowed**
- Failure: `newTerraformRunner` returned non-functional runner when binary missing.
- Fix location: `internal/aws/terraform_runner.go:36-55` — returns `(TerraformRunner, error)`.

### Observability Failures

**Grafana JSON Injection (141581a)**: `fmt.Sprintf` used for JSON body; replaced with `json.Marshal`.

**Grafana Mock Race Condition (1424706)**: Missing `sync.Mutex` on `MockGrafanaClient`.

**OBS Search Error Swallowed (39c4c4e)**: Dashboard search errors silently displayed as "no dashboard configured". Three-way branch and `sanitizeQueryValue` added.

### Manifest Schema / YAML Hazards (LOAD series)

**LOAD-002: YAML Boolean Gotcha — Absent `enabled` Field Defaults to true**
- Failure: Raw struct booleans caused disabled services to appear enabled.
- Fix location: `pkg/manifest/types.go:498,545,679,695,740,769` — `*bool` pointer semantics.

**LOAD-007: Manifest Write Must Use Atomic Rename**
- Failure: Direct writes left partially-written manifest on crash.
- Fix location: `pkg/manifest/writer.go:307`; `internal/deploy/store_file.go:279,329`.

### Watch Loop State Failures (IMPLICIT-PIII series)

**IMPLICIT-PIII-01/02**: Multiple inline path constructions for watch state files. Fix: sole path constructors in `internal/reconcile/watch_state.go:64-74`.

**IMPLICIT-PIII-03**: JSON `null` for nil Go slice instead of `[]`. Fix: `internal/reconcile/emitter.go:120`.

**IMPLICIT-PIII-05**: Pretty-printed inner JSON produced different HMAC. Fix: compact JSON before HMAC in `internal/reconcile/watch_state.go:223`.

**IMPLICIT-PIII-06**: Watch lock functions had same names as deploy lock functions. Fix: unique names in `internal/reconcile/watch_state.go:98-129`.

**FINDING-03**: Watch loop only polled after first ticker interval; 5-minute delay. Fix: `internal/reconcile/watcher.go:224-228`.

### Fork Safety Failures (IMPLICIT-FRK series)

**IMPLICIT-FRK-01/02**: Go port used different extension set than Python spec. Fix: `internal/fork/rename.go:14-36`.

**IMPLICIT-FRK-03/04**: Wrong replacement order corrupted already-replaced strings. Fix: `internal/fork/rename.go:66-70, 204, 240`.

### Scaffold/CI Template Failures

**IMPLICIT-WSE-01**: CI templates didn't escape `${{ }}` expressions. Fix: `internal/scaffold/scaffold_test.go:1126, 1601`.

**IMPLICIT-WSE-02**: Template blocks left trailing empty lines. Fix: `internal/scaffold/scaffold_test.go:1437`.

### Devenv / Config Failures

- **Global Config Fallback Missing (7323693)**: Satellite repos had no ecosystem config. Fix: `_a8_resolve_config` shell Tier 3 fallback.
- **AWS Region Silent Default (475cff9)**: Empty region defaulted to us-east-1. Fix: `internal/aws/clients.go:19-38` — let SDK chain handle resolution.
- **pip.conf Pollution (34c0d34)**: CodeArtifact login baked token into global pip.conf. Fix: `cmd/a8/doctor.go` — `checkPipConfPollution()`.
- **ECS Service Name Resolution (8dabba4)**: Manifest key passed directly to ECS API. Fix: `pkg/manifest` `ResolveECSServiceName()`.

### CI / Toolchain Failures

- **golangci-lint v2 Incompatibility (db104cd, 6709f44)**: v6 action used v1 config; 116 violations hidden.
- **GitHub App Token vs PAT (5adba67)**: PAT failed for homebrew-tap private repo.
- **Metric Name Mismatch (cd78ef4)**: Constants diverged from `autom8y-telemetry` SDK.

---

## Category Coverage

| Category | Count | Description |
|----------|-------|-------------|
| ECS Canary Lifecycle | 16 | ALB listener contamination, IAM, task def staleness, cross-process recovery, bake, TF revert |
| AWS Client Implicit Contracts | 3 | Failures array, nil concurrency, extended state enum |
| Mock State Drift | 3 | Mutation state updates, cleanup error propagation |
| Flag / Global State Leakage | 2 | Cobra flag variables, duplicate flag registration |
| Nil-Pointer / Missing Init | 3 | Terraform field nil, terraform init ordering, regex false positive |
| Security / Injection | 3 | Shell metachar, stale lock DoS, TF runner error swallowed |
| Observability Failures | 3 | JSON injection, mock race, search error swallowed |
| Devenv / Config | 4 | Missing fallback, silent region, pip.conf, ECS service name |
| Manifest Schema / YAML | 3 | Boolean pointer, non-atomic writes, trailing doc marker |
| Watch Loop State | 6 | Path constructors, JSON null, HMAC stability, symbol collision, first-poll delay |
| Regex / Pattern Matching | 5 | TF scanner false positive, fork rename divergence |
| CI Toolchain | 4 | golangci-lint v2, PAT vs App token, metric name mismatch |
| Release / Train Gates | 3 | Query failure masking, useless flag combos |
| Scaffold / Templates | 2 | GH Actions expression escaping, empty template lines |

**18 distinct failure mode categories documented.**

---

## Fix-Location Mapping

All primary fix files verified to exist on disk. Notable exceptions:
- SC-CANARY-04: architectural workaround (no single fix file)
- SC-CANARY-05: upstream Terraform provider bug (no in-repo fix)
- DEF-002: commit-level fix only, precise file unknown

See individual scar entries above for file paths and line ranges.

---

## Defensive Pattern Documentation

1. **AWS Error Classification** — Born from IMPLICIT-WS1. `ClassifyAWSError()` + `errors.Join(sentinel, err)`. Location: `internal/aws/errors.go`.

2. **Listener Rule Pre-Flight Check** — Born from SC-CANARY-03/13/15. `DescribeListenerRule` before canary start; reject if `activeTrafficTGs > 1`. Location: `internal/deploy/ecs_strategy.go:44-73`.

3. **Cleanup Non-Fatal Guard** — Born from IMPLICIT-ECS-FIX-03. Cleanup failures log warning but don't fail deploy. Location: `internal/deploy/ecs_strategy.go:222, 256`.

4. **SC-CANARY-08 Ordering Constraint** — ROLLING before cleanup. Location: `internal/deploy/ecs_strategy.go:219, 254`.

5. **FARGATE-Only Canary Pinning** — Born from SC-CANARY-SPOT-SAFETY. Override to FARGATE-only during canary. Location: `internal/deploy/ecs_strategy.go:29-105, 443-467`.

6. **Terraform Listener Rule Non-Revert (SP-02)** — `ignore_changes = [action]` on canary listener rules. Location: `terraform/modules/stacks/service-stateless/main.tf`.

7. **Compile-Time Interface Checks** — Born from DEF-WS2-001. `var _ Interface = (*Impl)(nil)`.

8. **Cobra Flag Reset Helper** — Born from TENSION-006. `resetFlags(t)` in `cmd/a8/cmd_test.go:61-83`.

9. **Mock Mutation State Update Invariant** — Born from IMPLICIT-WS2-01. `"CRITICAL: mock MUST update Snapshot on mutations"`. Location: `internal/deploy/store_mock.go:7`.

10. **Sole Path Constructors** — Born from IMPLICIT-PIII-01/02. `DefaultDeployStatePath()` and `DefaultWatchStatePath()` are sole authorized path constructors.

11. **YAML `*bool` Pointer Convention** — Born from LOAD-002. Absent = nil = defaults to true. `Enabled()` method encapsulates nil-safety.

12. **Atomic Write Pattern** — Born from LOAD-007. Temp file + `os.Rename`. Location: `internal/deploy/store_file.go:279,329`; `pkg/manifest/writer.go:307`.

13. **Stale Lock Detection** — Born from L-002. 10-minute timeout on advisory locks.

14. **Fail-Fast Region Resolution** — Empty region passed to SDK chain; no silent fallback. Location: `internal/aws/clients.go:19-38`.

15. **Shell Metacharacter Validation** — Born from L-001. `containsDisallowedChars()` in `ValidateAll()`.

16. **TerraformRunner Fail-Fast** — Born from L-003. Returns error at construction time if binary missing. Location: `internal/aws/terraform_runner.go:36-55`.

17. **Archetype Exhaustiveness Guard** — Born from IMPLICIT-08. `TestNewDiffer_ExhaustivenessGuard` breaks on new archetype without differ. Location: `internal/reconcile/differ_test.go:9-39`.

18. **Sequential Apply** — Born from IMPLICIT-WS2-03. Apply operations sequential, not concurrent. Location: `cmd/a8/reconcile.go:361-362`.

19. **DriftEvent Empty Array Guard** — Born from IMPLICIT-PIII-03. `DriftingServices` initialized to `[]string{}`. Location: `internal/reconcile/emitter.go:120`.

20. **HMAC Inner Compact JSON** — Born from IMPLICIT-PIII-05. Compact JSON before HMAC computation.

---

## Agent-Relevance Tagging

| Category | Relevant Agent Roles | Why |
|----------|---------------------|-----|
| ECS Canary Lifecycle | release-executor, pipeline-monitor | Must understand ALB listener rule lifecycle and contamination states |
| AWS Client Contracts | release-executor, cartographer | Must check Failures array, handle nil concurrency, normalize state enums |
| Mock State Drift | release-executor (test authoring) | Must update mock state on mutations |
| Sequential Apply | release-executor, release-planner | Concurrent mutations break audit logs and produce non-deterministic ordering |
| Cobra Flag Leak | release-executor (test authoring) | Must call resetFlags in every cmd test |
| Nil-Pointer / Missing Init | release-executor, cartographer | Must wire all fields in client bundles |
| Security | release-executor, release-planner | Shell validation, fail-fast construction, stale lock detection |
| Manifest Schema | cartographer, dependency-resolver | *bool pointer semantics, atomic writes |
| Watch Loop State | release-executor, pipeline-monitor | Sole path constructors, HMAC compact JSON |
| Fork Safety | cartographer | Extension set must match Python spec exactly |
| SP-02 (TF listener revert) | release-executor, release-planner, pipeline-monitor | Terraform must not revert canary listener rules |

---

## Knowledge Gaps

1. **SC-CANARY-07, 09, 12 absent** — Numbering gaps suggest renumbered or externally documented scars not backfilled.
2. **DEF-002 fix location imprecise** — Commit-level attribution only; specific file/line unknown.
3. **SC-CANARY-01 fix in external repo** — IAM role fix in autom8y satellite, not this repo.
4. **SC-CANARY-05 no in-repo guard** — Upstream Terraform provider bug; procedural workaround only.

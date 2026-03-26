---
domain: design-constraints
generated_at: "2026-03-25T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "429f242"
confidence: 0.82
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/initiative-history.md"
land_hash: "c24a2e09584c57137a5f5607e2d38a30d0c2b06533ec9df377740b41406405a2"
---
# Codebase Design Constraints

**Primary language:** Go 1.25 (module `github.com/autom8y/a8`)
**Source scope:** `cmd/a8/`, `internal/`, `pkg/manifest/`

---

## Tension Catalog Completeness

### TENSION-001: `DeployTarget` Name Collision Across Packages

- **Type:** Naming mismatch
- **Location:** `pkg/manifest/types.go` (`DeployTarget` string enum) vs `internal/deploy/types.go` (`DeployTarget` runtime struct)
- **Historical reason:** Both types are valid within their packages but share a name describing different things.
- **Ideal resolution:** Rename manifest type to `ComputeTarget`.
- **Resolution cost:** High (schema-breaking manifest change).

### TENSION-002: `internal/aws` Contains Terraform Orchestration Logic

- **Type:** Package responsibility mismatch
- **Location:** `internal/aws/terraform.go`, `internal/aws/terraform_runner.go`
- **Evidence:** Comment at `cmd/a8/reconcile.go:618`: "not an AWS adapter concern (ADR-051, TENSION-002)."
- **Historical reason:** Added incrementally alongside AWS adapters during early Go CLI sessions.
- **Ideal resolution:** Dedicated `internal/tf` or `internal/infra` package.
- **Resolution cost:** Medium (update all callers).

### TENSION-004: `buildMockClientsFromManifest` Duplicates Archetype Dispatch

- **Type:** Duplicated logic
- **Location:** `cmd/a8/reconcile.go:616` mirrors switch in `internal/reconcile/differ.go:NewDiffer()`
- **Evidence:** Comment: "Switch case order mirrors NewDiffer in internal/reconcile/differ.go."
- **Resolution cost:** Medium (extract shared dispatch).

### TENSION-007: Duplicated Lock/HMAC Patterns Across deploy and reconcile

- **Type:** Duplicated logic (intentional)
- **Location:** `internal/deploy/store_file.go` and `internal/reconcile/watch_state.go`
- **Evidence:** Comment at `watch_state.go:5`: "Replicates (not imports) advisory lock, atomic write, and HMAC patterns from internal/deploy/store_file.go to respect the BC-07 boundary."
- **ADR:** ADR-137
- **Resolution cost:** High (requires shared abstraction package).

### TENSION-008: `ECSCanaryStrategy` Accumulates Mutable Lifecycle State

- **Type:** Structural tension
- **Location:** `internal/deploy/ecs_strategy.go`
- **Evidence:** Mutable fields (`canaryDeploymentID`, `baselineTaskDefARN`, `savedCapacityProviders`) set by `StartCanary`, consumed by `Promote`/`Rollback`.
- **Resolution cost:** High (blocked on design for external state store).

### TENSION-010: `AUTOM8Y_ENV` Legacy Env Var Hardcoded in config.go

- **Type:** Naming mismatch / backward compatibility
- **Location:** `internal/config/config.go:76` — `ResolveOrgEnv("ENV", "AUTOM8Y")`
- **Evidence:** Fork tooling renames string literals but may not catch logic code.
- **Resolution cost:** Low-medium (two-pass manifest discovery refactor).

### TENSION-012: CodeArtifact Env Var Prefix Migration — ACTIVE

- **Type:** Naming mismatch / backward compatibility
- **Location:** `cmd/a8/train_publish.go:67`
- **Evidence:** `// BREAKING: legacy CODEARTIFACT_* names (without A8_ prefix) no longer supported.`
- **ADR:** ADR-122
- **Resolution cost:** Low (documentation/operator migration).

### TENSION-AWS-CLIENT-SPLIT: Reconcile vs Deploy Client Interface Split — ACTIVE BY DESIGN

- **Type:** Dual-system pattern (intentional)
- **Location:** `internal/aws/ecs.go` (`ECSClient`), `internal/aws/ecs_deploy.go` (`ECSDeployClient`)
- **Evidence:** `"Separate interface from the FROZEN ECSClient. [BC-06, FR-M13]"`
- **ADR:** ADR-083, BC-06, FR-M13
- **Current state:** Active by design.

### TENSION-IMPLICIT: ECS DeploymentID Discovery Race — ACTIVE

- **Type:** TOCTOU gap
- **Location:** `internal/deploy/ecs_strategy.go:134-146`
- **Evidence:** After `UpdateServiceDeployment`, re-queries `GetServiceDeployments` with no retry if `canaryDeploymentID` is empty.
- **Resolution cost:** High (blocked on ECS API atomicity).

### NEW: `internal/metrics` Cross-SDK Dependency — ACTIVE

- **Type:** Implicit cross-system contract
- **Location:** `internal/metrics/names.go:1-8`
- **Evidence:** "These constants are verified against the autom8y-telemetry SDK."
- **Resolution cost:** Medium (CI cross-repo verification needed).

### NEW: `AUTOM8Y_ENV` vs `A8_ENV` Env Var Split — ACTIVE

- **Type:** Dual env var consumption
- **Location:** `cmd/a8/mock_guard.go:32` uses `AUTOM8Y_ENV` directly; `internal/config/config.go:70-76` establishes `A8_ENV` as canonical.
- **Resolution cost:** Low (update `mock_guard.go` to use `config.ResolveOrgEnv`).

### TENSION-PENTEST-002: Production Confirmation Gate — ACTIVE BY DESIGN

- **Type:** Security gate
- **Location:** `internal/reconcile/confirm.go:43-54`
- **Evidence:** `--yes` flag silently overridden when `Env == "production"`.
- **Resolution cost:** N/A (intentional constraint).

### TENSION-PATHREGEX: Strict Path Validation Regex — ACTIVE

- **Type:** Input validation boundary
- **Location:** `pkg/manifest/validate.go:14-16`
- **Evidence:** `pathRegex = regexp.MustCompile("^[a-zA-Z0-9._/-]+$")`. Security hardening from commit d83c747.

### Resolved Tensions (preserved for history)

- **TENSION-002 (import direction)**: `internal/aws` → `pkg/manifest` import removed. `BuildMockClients` moved to `cmd/a8`.
- **TENSION-004 (dual validation)**: Private `validate()` removed. Single path via `ValidateAll()`.
- **TENSION-005 (Lambda name resolution)**: Centralized in `Service.ResolveFunctionName()`.
- **TENSION-006 (global flag vars)**: Closure-scoped; `resetCobraFlags()` in tests.
- **TENSION-009 (os.Setenv for TF vars)**: `ExtraEnv` struct field in `TerraformRunner`.

---

## Trade-off Documentation

### TENSION-007: Lock/HMAC Duplication (Trade-off)

- **Chosen:** Three implementations of the same locking pattern.
- **Rejected:** Shared `internal/statestore` package.
- **Why current state persists:** BC-07 prohibits `internal/reconcile` from importing `internal/deploy`. ADR-137 explicitly chose duplication.
- **Side effect:** Changes to lock-stale-age or HMAC behavior must be applied in two places.

### TENSION-AWS-CLIENT-SPLIT: Interface Split (Trade-off)

- **Chosen:** Four AWS interfaces split by operation type (read-only reconcile vs. mutating deploy).
- **Rejected:** Single unified interface per AWS service.
- **Why current state persists:** Intentional (BC-06). Merging would allow reconcile to accidentally depend on deploy mutations.

### LOAD-002: `*bool` Pointer Semantics (Trade-off)

- **Chosen:** `*bool` pointers to distinguish absent from false.
- **Rejected:** `bool` with default value.
- **Why current state persists:** `gopkg.in/yaml.v3` does not support default values natively.

### LOAD-001: Versioned JSON Output Contract (Trade-off)

- **Chosen:** COORDINATED string constants, additive changes only.
- **Rejected:** Free renaming/removal.
- **Why current state persists:** External contract with JSON consumers and CI parsers.
- **ADR:** ADR-120.

### ADR-087: Lambda Strategy Statelessness (Trade-off)

- **Chosen:** `LambdaCanaryStrategy` does NOT implement `StrategyHydrator`.
- **Rejected:** Persisting Lambda alias routing state.
- **Why current state persists:** Lambda alias routing is queryable from AWS; recovery straightforward without local state.

---

## Abstraction Gap Mapping

### GAP-001: No Shared `statestore` Abstraction (3 implementations)

Advisory lock + atomic write + HMAC pattern implemented in: `internal/deploy/store_file.go` and `internal/reconcile/watch_state.go`. Functionally identical. Maintenance burden: lock-stale-age changes must be applied twice.

### GAP-002: `buildDeployStrategy` Factory in cmd Layer

`cmd/a8/deploy.go:678` creates strategies based on archetype. Location commented as intentional (ADR). Cannot be tested without full cmd layer.

### GAP-003: `ecs-fargate-hybrid` Reuses `service-stateless` Stack (Placeholder)

`internal/scaffold/types.go:105` maps hybrid to stateless template. Test asserts TODO exists. No dedicated Terraform module.

### GAP-004: Grafana Logic Split Across Two Packages

Dashboard generation (`internal/dashgen/`) and Grafana API interaction (`internal/grafana/`) — complementary halves without shared abstraction.

### GAP-005: `internal/metrics` Thin Constants Package

Two-file package existing solely to prevent metric name duplication between `internal/dashgen` and `internal/amp`. Functional but borderline over-engineering for current scope.

---

## Load-Bearing Code Identification

### LOAD-BEARING-001: SurfaceName and Status Constants

`internal/reconcile/surface.go` — versioned JSON output contract (LOAD-006). Any rename/removal breaks all JSON consumers.

### LOAD-BEARING-002: DriftEvent Schema (OWD-1)

`internal/reconcile/emitter.go` — `DriftEvent.SchemaVersion = 1` frozen. `DriftEmitter.Emit` interface frozen (OWD-2). One-way door.

### LOAD-BEARING-003: `.a8/drift-history.json` Path (OWD-3)

`internal/reconcile/watch_state.go:10-11` — file path frozen. `DefaultWatchStatePath()` is sole constructor.

### LOAD-BEARING-004: AWS Client Interfaces (BC-06)

`LambdaClient` and `ECSClient` are FROZEN interfaces. Extensions via separate interfaces (`LambdaDeployClient`, `ECSDeployClient`).

### LOAD-BEARING-005: Archetype Constant Values

`pkg/manifest/types.go` — string values YAML-serialized in every `manifest.yaml`. Changing any value silently breaks manifest loading. 7-location coordinated update documented.

### LOAD-BEARING-006: `.a8/deploy-state.json` Path

`internal/deploy/store_file.go` — `DefaultDeployStatePath()` sole constructor (RISK-007). Path frozen once state files exist.

---

## Evolution Constraint Documentation

### CONSTRAINT-001: BC-07 — Deploy Cannot Import Reconcile

`internal/deploy/types.go:5-6`. Any feature needing deploy + reconcile interaction must go through cmd layer.

### CONSTRAINT-002: BC-06 — Frozen Client Interfaces

Adding methods requires creating new interfaces, not extending frozen originals.

### CONSTRAINT-003: pkg/manifest Stable API

Internal packages import from it; it does not import from `internal/`. Adding fields is additive. Removing/renaming is breaking.

### CONSTRAINT-004: Schema Version Contracts

Status and SurfaceName constants versioned. Renaming is breaking; addition is safe.

### CONSTRAINT-005: Archetype Exhaustiveness Enforcement

`TestArchetypeDescriptors_Exhaustiveness` breaks on new archetype without coordinated 7-location update.

### CONSTRAINT-006: AUTOM8Y Prefix Legacy Fallback

`internal/config/config.go` — LOW changeability without two-pass discovery refactor.

### CONSTRAINT-007: `.a8/` State Directory Frozen Paths

Three files frozen: `deploy-state.json`, `drift-history.json`, `watch-events.jsonl`.

---

## Risk Zone Mapping

### RISK-001: `buildMockClientsFromManifest` No Exhaustiveness Guard

`cmd/a8/reconcile.go:640-665` — switch on archetype lacks `default:` case. New archetype silently produces no mock state.

### RISK-002: `ECSCanaryStrategy` Mutable State Without Synchronization

Struct fields set by `StartCanary` and read by `Promote`/`Rollback` have no mutex. Single-goroutine usage pattern assumed but not enforced.

### RISK-003: `AUTOM8Y` Hardcoded for Fork Consumers

After fork, `internal/config/config.go:76` still reads `AUTOM8Y_ENV`. Fork rename tooling may not catch logic code.

### RISK-004: Advisory Lock Stale-Age Not Shared Constant

`watchStaleLockAge = 10 * time.Minute` in watch_state.go and similar in store_file.go — separate constants. Changing one without the other creates inconsistent lock-expiry behavior.

### RISK-005: `sanitizeQueryValue` Limited Character Escaping

`cmd/a8/obs_helpers.go:66` — escapes only backslash and double-quote. LogQL-meaningful characters (`{`, `}`, `|`) pass through unescaped. Input is trusted from manifest.

### RISK-006: Canary Listener Rule Contamination Requires Manual Recovery

`internal/deploy/ecs_strategy.go:74-77` — if canary fails leaving 2 active TGs, next deploy returns error requiring manual `aws elbv2 modify-rule` fix.

### RISK-007: `buildDeployRecorder` Silent Write Failures

`cmd/a8/deploy.go:25` — `A8_AUDIT_LOG` env var points to non-existent directory; audit records silently lost.

---

## Knowledge Gaps

1. **`internal/ci/`** — Not fully read. Role in CI gate commands not captured.
2. **`internal/grafana/names.go`** — Contents not read.
3. **`internal/fork/`** — Load-bearing constraints around fork tooling not fully documented.
4. **`internal/tfstate/` vs `internal/tfmod/` vs `internal/aws/terraform*`** — Boundary between these three locations not captured.
5. **`pkg/manifest/writer.go`** — Write constraints not audited.
6. **ADRs** — Multiple ADR numbers referenced (051, 083, 085, 086, 087, 111, 114, 120, 122, 137) but ADR directory (`.ledge/decisions/`) not inspected.
7. **`internal/deploy/health.go` and `amp_querier.go`** — Risk zones not captured.

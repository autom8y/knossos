---
domain: conventions
generated_at: "2026-03-25T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "429f242"
confidence: 0.91
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "1d5c945d46857fffeeb248674d6adbe89b9b33b94dfd6a946448973c591917b6"
---
# Codebase Conventions

## Error Handling Style

### Error Creation Pattern

The project uses two distinct error creation strategies, applied consistently by domain:

**1. Package-level sentinel errors** — declared in `errors.go` files at the top of each internal package. All sentinels follow the naming prefix `"<package>: <reason>"`.

From `internal/aws/errors.go`:
```go
var ErrPermissionDenied = errors.New("aws: permission denied")
var ErrResourceNotFound = errors.New("aws: resource not found")
var ErrThrottled = errors.New("aws: throttled")
var ErrTimeout = errors.New("aws: timeout")
```

This pattern exists in 7 packages with dedicated `errors.go` files:
- `internal/aws/errors.go` — 4 sentinels
- `internal/aws/cloudwatch_errors.go` — 4 sentinels
- `internal/amp/errors.go` — 6 sentinels
- `internal/grafana/errors.go` — 8 sentinels
- `internal/scaffold/errors.go` — 9 sentinels
- `internal/ci/errors.go` — 1 sentinel
- `internal/dashgen/types.go` — 1 sentinel (in types, not errors.go)

**2. Inline `fmt.Errorf` with `%w` wrapping** — used for contextual errors at call boundaries.

Format: `"<verb> <subject>: <detail>: %w"` — colon-separated action chain:
```go
return fmt.Errorf("load AWS config (region=%s): %w", region, err)
return fmt.Errorf("start canary %s: get alias routing: %w", target.Service, err)
```

### Error Classification Pattern: `errors.Join(sentinel, original)`

This is the defining error pattern of the codebase. Each external-facing client package has a `Classify*Error` function that maps raw errors to sentinels using `errors.Join`:

```go
func ClassifyAWSError(err error) error {
    if errors.Is(err, context.DeadlineExceeded) {
        return errors.Join(ErrTimeout, err)
    }
    // ... smithy API error inspection
    return errors.Join(ErrPermissionDenied, err)
}
```

`errors.Join` preserves the original error in the chain while prepending the sentinel. The pattern is explicitly cross-referenced:
- `internal/amp/errors.go`: `// Follows the internal/aws/errors.go pattern.`
- `internal/grafana/errors.go`: `// Follows the internal/aws/errors.go pattern: errors.Join(sentinel, original).`

Implemented in: `ClassifyAWSError`, `ClassifyAMPError`, `ClassifyGrafanaError`, `ClassifyCloudWatchError`.

### Error Propagation Style

- **Immediate return**: All call sites check `if err != nil { return ..., err }` — no error aggregation patterns.
- **`errors.Is` / `errors.As` at boundaries**: Callers use `errors.Is(err, ErrPermissionDenied)` to distinguish causes at decision points.
- **Never swallowed**: Only `os.Hostname()` / `user.Current()` errors are intentionally discarded (comment explains: "All fields fall back to empty string on error so logging is never fatal").
- **Multi-error aggregation**: `errors.Join(errs...)` used in `internal/reconcile/emitter.go` for MultiEmitter fan-out.

### Error Handling at Boundaries

The CLI boundary in `cmd/a8/main.go` uses `errors.As` to unwrap `*ExitCodeError`:
```go
var exitErr *ExitCodeError
if errors.As(err, &exitErr) {
    fmt.Fprintln(os.Stderr, exitErr.Message)
    os.Exit(exitErr.Code)
}
```

`ExitCodeError` (in `cmd/a8/helpers.go`) carries an integer exit code for fleet commands (e.g., drift count = exit code).

User-facing output uses `internal/cli` helpers: `PrintOK`, `PrintWarn`, `PrintError`, `PrintInfo` with bracketed tags `[OK]`, `[WARN]`, `[ERROR]`, `[INFO]`, `[FAIL]`, `[SKIP]`. Structured logging via `slog` to stderr via `appLogger` in `root.go`. `NO_COLOR` env var respected.

Manifest validation uses a **collector pattern** (`*ValidationError` in `pkg/manifest/validate.go`): collects all validation failures and returns a slice.

---

## File Organization

### Package-Level Split Conventions

Each `internal/` package follows a consistent multi-file layout:

| File name pattern | Contents |
|---|---|
| `types.go` | All exported types, constants, enums. Package doc comment lives here. |
| `errors.go` | Sentinel error variables and `Classify*Error` functions |
| `{noun}.go` | Interface definition (e.g., `ecs.go` defines `ECSClient`) |
| `{noun}_client.go` | Real AWS/HTTP implementation (e.g., `ecs_client.go` defines `RealECSClient`) |
| `mock.go` | Test double implementations of all interfaces (`Mock*Client` structs) |
| `{noun}_test.go` | Tests for the corresponding source file |
| `export_test.go` | Exported wrappers for unexported functions (black-box test access) — found in `internal/release/` |

### cmd/a8 File Organization

Each CLI subcommand tree gets its own file: `reconcile.go`, `deploy.go`, `svc.go`, `train.go`, etc. Multi-file subcommand trees split by child: `train.go` + `train_create.go` + `train_bump.go` + `train_publish.go`. Naming: `{parent}_{child}.go`.

Helper functions shared across commands live in `helpers.go`. Output formatting helpers in `obs_helpers.go`.

All cobra command constructors are unexported: `new{Name}Cmd() *cobra.Command`. All subcommands registered in `root.go` `init()`.

### Constants, Variables, Init

- **Package-level sentinels and constants**: dedicated `errors.go` or `types.go`
- **Compile-time interface checks**: `var _ Interface = (*Impl)(nil)` at end of `_client.go` files (38 occurrences across 15 files)
- **`init()` functions**: only in `cmd/a8/root.go` (cobra command wiring). None in `internal/`
- **Package documentation**: `// Package <name> ...` doc comment at top of the most central file per package

### pkg/manifest Layout

- `types.go` — all manifest types (~800 lines)
- `loader.go` — YAML deserialization and `FindManifest`
- `validate.go` — validation logic
- `writer.go` — YAML serialization
- `node.go` — YAML node manipulation helpers

### Generated Code

No generated code markers (`// Code generated`) found. Templates for Terraform scaffolding stored as Go string constants or embedded files in `internal/scaffold/templates.go`.

---

## Domain-Specific Idioms

### 1. Sentinel + Join Error Pattern

The canonical pattern for external-API error classification:

```go
// 1. Declare sentinel
var ErrThrottled = errors.New("amp: throttled")
// 2. Classify wraps: Join(sentinel, original)
return errors.Join(ErrThrottled, err)
// 3. Caller checks with errors.Is
if errors.Is(err, amp.ErrThrottled) { ... }
```

### 2. Interface / Real / Mock Triad

Every external dependency is abstracted as three Go files:
- `{noun}.go` — interface
- `{noun}_client.go` — real implementation (suffix `Real*`)
- `mock.go` — test doubles (suffix `Mock*`), map-backed

### 3. Archetype Dispatch Table

`archetypeDescriptors` map in `pkg/manifest/types.go` is the single source of truth for all archetype metadata. Adding an archetype requires updating 7 locations (explicitly documented). `NewDiffer` switch in `internal/reconcile/differ.go` is the primary dispatch point.

### 4. Pointer-Bool LOAD-002 Convention

Optional boolean fields use `*bool` pointers to distinguish "absent" (nil, defaults to true) from "explicitly false":
```go
Enabled *bool `yaml:"enabled,omitempty"`
```
All `*bool` fields have `Is*()` accessor methods that handle nil → true.

### 5. C-007: Sorted Map Iteration

Map iteration over manifest maps always uses `sort.Strings` before iterating. Tagged `// C-007` throughout (100+ occurrences in 23+ files):
```go
sort.Strings(names) // C-007
for _, name := range names { ... }
```

### 6. `envOr` Helper

Environment variable fallback in `cmd/a8/train.go`:
```go
func envOr(key, fallback string) string {
    if v := os.Getenv(key); v != "" { return v }
    return fallback
}
```

### 7. `Resolve*` Accessor Pattern

`Org` and `Service` types use `Resolve*()` methods for defaults/fallback logic:
```go
m.Org.ResolveClusterName()       // explicit ECSCluster or "{name}-cluster"
svc.ResolveFunctionName(key)     // explicit FunctionName or serviceKey
```

### 8. Observability Nil-Guard Pattern

Optional observability features use nil-guard: nil `*ObservabilityClients` means "not configured" and degrades to `StatusDeferred`.

### 9. Cross-Process State Recovery via JSON + Interface Hydration

`DeployController` persists state to `.a8/deploy-state.json` and recovers via `StrategyHydrator` interface.

### 10. ADR/FR/SC Reference Tags in Comments

Design decision references embedded as inline tags: `ADR-085`, `FR-M01`, `SC-18`, `BC-07`, `LOAD-002`. Reference entries in `.ledge/decisions/`.

### 11. `Clients` Struct Bundle

Dependencies grouped into bundle structs: `reconcile.Clients`, `reconcile.ObservabilityClients`, `reconcile.EfficiencyClients`. Engine grows new optional bundles backward-compatibly (nil = DEFERRED behavior).

### 12. `ExitCodeError` Pattern

Fleet/batch commands return `&ExitCodeError{Code: n, Message: "N service(s) with drift"}` rather than `fmt.Errorf`. `main()` type switch extracts and calls `os.Exit(exitErr.Code)`.

### 13. COORDINATED Constants (LOAD-001)

`SurfaceName` and `SurfaceID` constants marked `// LOAD-001: COORDINATED` — versioned JSON output contract. Cannot be renamed/removed without schema version bump.

---

## Naming Patterns

### Exported Type Names

- **State enums**: `DeployState`, `Status`, `SurfaceName`, `SurfaceID` — noun-heavy, no `Type` suffix
- **Config structs**: `DeployOpts`, `VerifyConfig`, `CanaryConfig` — `Opts` for optional, `Config` for required
- **Result types**: `DeployResult`, `VerifyResult`, `AffectedSet` — `Result` suffix
- **Client interfaces**: `ECSClient`, `LambdaClient`, `AMPClient` — named by service
- **Real implementations**: `RealECSClient`, `RealLambdaClient` — `Real` prefix
- **Mock implementations**: `MockECSClient`, `MockLambdaClient` — `Mock` prefix

### Constant Naming

Type-name-as-prefix pattern:
```go
type DeployState string
const StatePending DeployState = "PENDING"
const StateCanaryActive DeployState = "CANARY_ACTIVE"
```

SCREAMING_SNAKE_CASE for simple tokens, human-readable strings for display labels.

### Function Naming

- **Constructors**: `New{TypeName}(...)` for structs, `new{Name}Cmd()` (unexported) for cobra commands
- **Builder variants**: `New{Type}With{Feature}(args)` for optional capabilities
- **Resolver methods**: `Resolve{FieldName}()` on structs with default logic
- **Boolean predicates**: `Is{Property}()`, `Has{Property}()` — no `Get` prefix
- **Classification functions**: `Classify{Package}Error()` — one per error-boundary package

### Variable Naming

- **Global CLI logger**: `appLogger`
- **Global CLI flags**: `flag{Name}` — `flagManifest`, `flagOutput`, `flagVerbose`, `flagEnv`
- **Import aliases**: `internal/aws` aliased `iaws` in `cmd/a8` (avoid collision with AWS SDK `aws` package)
- **Build-time variables**: `version`, `commit`, `date`, `goVersion` — lowercase, set via ldflags

### Package Naming

All lowercase single words: `reconcile`, `deploy`, `release`, `scaffold`, `manifest`, `metrics`, `dashgen`. No nested internal packages.

### File Naming

- `_client.go` suffix for real implementations
- `_deploy.go` suffix for deploy-specific extensions
- `kill_chain_{N}_test.go` for multi-phase security test sequences
- `export_test.go` for test export shims

### Acronym Conventions

- `ECS`, `RDS`, `ALB`, `ARN`, `IAM` — all uppercase
- `URL` → `URL` in struct fields, `url` in variable names
- `ID` → `ID` always (`SurfaceID`, `WorkspaceID`)

---

## Knowledge Gaps

1. **`internal/metrics/` package** — metric constant naming convention not fully documented.
2. **`internal/fork/` package** — fork-management idioms not fully explored.
3. **`internal/dashgen/` package** — dashboard generation template conventions not captured.
4. **`internal/scaffold/` package** — template rendering conventions not fully explored.
5. **Test helper patterns** — `cmd/a8` test files use `package main` (not `_test`); exact table-driven patterns not systematically cataloged.

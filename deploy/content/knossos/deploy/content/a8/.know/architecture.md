---
domain: architecture
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
# Codebase Architecture

**Module**: `github.com/autom8y/a8`
**Language**: Go 1.25.8
**Primary dependencies**: cobra (CLI), aws-sdk-go-v2 (AWS), gopkg.in/yaml.v3 (manifest parsing), grafana-foundation-sdk (dashboard generation), charm.land/bubbletea (TUI), golang.org/x/sync (concurrency)

---

## Package Structure

The codebase has two source trees: `cmd/` (CLI entry point) and `internal/` (domain logic), plus `pkg/` (shared foundational packages).

### cmd/a8/ (59 Go files)

The single binary entry point. All files are in `package main`. This package wires cobra commands to internal logic and contains no domain logic itself.

Key files:
- `cmd/a8/main.go` — 19 lines; calls `rootCmd.Execute()` with `ExitCodeError` unwrapping
- `cmd/a8/root.go` — defines `rootCmd`, global flags (`--manifest`, `--output`, `--verbose`, `--env`), `buildLogger()`, and wires 17 subcommands in `init()`
- `cmd/a8/reconcile.go`, `cmd/a8/reconcile_watch.go` — `reconcile plan/apply/watch`
- `cmd/a8/deploy.go` — progressive deployment commands; delegate to `internal/deploy`
- `cmd/a8/rollback.go` — production rollback
- `cmd/a8/observe.go`, `cmd/a8/observe_generate.go`, `cmd/a8/obs_helpers.go` — observability CLI; delegates to `internal/dashgen`, `internal/reconcile`
- `cmd/a8/status.go` — ecosystem dashboard; delegates to `internal/reconcile`
- `cmd/a8/train.go`, `cmd/a8/train_create.go`, `cmd/a8/train_publish.go`, `cmd/a8/train_bump.go` — SDK release train management
- `cmd/a8/svc.go` — service lifecycle management
- `cmd/a8/tf.go`, `cmd/a8/tf_bootstrap.go`, `cmd/a8/tf_upgrade.go` — Terraform operations; delegates to `internal/tfmod`, `internal/tfstate`
- `cmd/a8/fork.go`, `cmd/a8/fork_check.go`, `cmd/a8/fork_codeartifact.go`, etc. — fork workflow commands; delegates to `internal/fork`
- `cmd/a8/scaffold_terraform.go` — Terraform scaffold; delegates to `internal/scaffold`
- `cmd/a8/doctor.go` — environment health checks (includes pip.conf CodeArtifact pollution detection)
- `cmd/a8/ci.go` — CI gate and status; delegates to `internal/ci`
- `cmd/a8/workflow.go` — workflow orchestration commands; delegates to `internal/workflows`
- `cmd/a8/query.go` — manifest query; delegates to `pkg/manifest`
- `cmd/a8/dev.go` — local development stack management
- `cmd/a8/validate.go` — manifest validation
- `cmd/a8/init.go` — interactive fork wizard (`a8 init`)
- `cmd/a8/changeset.go` — release changeset intent; delegates to `internal/release`
- `cmd/a8/completion.go` — shell completion
- `cmd/a8/helpers.go` — shared command-layer helpers
- `cmd/a8/env_helpers.go`, `cmd/a8/obs_helpers.go` — env and observability helpers scoped to cmd layer
- `cmd/a8/mock_guard.go` — mock client injection for dry-run/test modes

### pkg/manifest/ — Foundational Manifest Package (leaf, 9 files)

The single source of truth for typed manifest schema. This is a **leaf package** — it has no internal imports. Every other package imports it.

Key files:
- `pkg/manifest/types.go` — all manifest types (`Manifest`, `Service`, `Org`, `Platform`, `SDK`, `ReleaseTrain`, `DeployConfig`, `Observability`, `Workflow`, etc.)
- `pkg/manifest/loader.go` — YAML loading and path discovery (`FindManifest()`, `Load()`)
- `pkg/manifest/validate.go` — field validation (archetype set, required fields)
- `pkg/manifest/writer.go` — manifest serialization (write back to YAML)
- `pkg/manifest/node.go` — YAML node manipulation for selective field updates

### internal/reconcile/ — Reconciliation Engine (hub, ~15 files)

The core reconcile loop. Most complex package in the codebase.

- `internal/reconcile/engine.go` — `Engine` struct with `PlanService()`, `PlanAll()`, concurrent fleet planning via `errgroup`
- `internal/reconcile/surface.go` — `Surface`, `Status`, `SurfaceName`, `SurfaceID` constants (COORDINATED schema contract LOAD-001)
- `internal/reconcile/differ.go` — `Differ` interface + 5 archetype implementations (ecsFargateRDS, ecsFargateStateless, ecsFargateHybrid, lambdaScheduled, lambdaEventDriven)
- `internal/reconcile/planner.go` — `Plan`, `Operation`, `BuildPlan()`, `deriveOperations()` (DRIFT → Operation mapping)
- `internal/reconcile/executor.go` — applies operations (ECS, Lambda, EventBridge SDK calls)
- `internal/reconcile/differ_obs.go` — observability surface differs (dashboard, alert rule, datasource, metric health)
- `internal/reconcile/differ_module.go` — Terraform module ref version surface
- `internal/reconcile/differ_capacity.go`, `differ_efficiency.go`, `differ_spot.go` — efficiency/capacity surfaces
- `internal/reconcile/workflow.go` — workflow reconciliation (Step Functions state machine surface)
- `internal/reconcile/emitter.go` — JSON output for versioned schema
- `internal/reconcile/watcher.go`, `watch_state.go` — `a8 reconcile watch` live monitoring
- `internal/reconcile/confirm.go` — interactive confirmation UI
- Imports: `pkg/manifest`, `internal/aws`, `internal/grafana`, `internal/amp`, `internal/metrics`, `internal/dashgen`, `internal/tfmod`

### internal/deploy/ — Progressive Deployment (hub, ~11 files)

Canary deployment state machine.

- `internal/deploy/controller.go` — `DeployController` state machine (PENDING → CANARY_ACTIVE → VERIFYING → PROMOTING/ROLLING_BACK → COMPLETE/FAILED)
- `internal/deploy/strategy.go` — `DeployStrategy` interface + `LambdaCanaryStrategy` (alias routing), `ECSCanaryStrategy` (in separate `ecs_strategy.go`)
- `internal/deploy/health.go` — `HealthGate`, `CanaryScorer`, `DifferentialScorer`, `CloudWatchMetricsQuerier`
- `internal/deploy/amp_querier.go` — `AMPMetricsQuerier` (Prometheus-based canary scoring)
- `internal/deploy/types.go` — `DeployState`, `DeployTarget`, `DeployOpts`, `DeployResult`, `DeployStatus`
- `internal/deploy/store.go`, `store_file.go` — `DeployStateStore` interface + file-backed persistence for cross-process promote/rollback
- `internal/deploy/progress.go` — `Progression` type (step weights + verify durations)
- `internal/deploy/spot_detector.go` — FARGATE_SPOT interruption detection
- Layer rule (BC-07): imports `internal/aws`, `pkg/manifest` — does NOT import `internal/reconcile`
- Imports: `pkg/manifest`, `internal/aws`, `internal/amp`, `internal/metrics`

### internal/aws/ — AWS Client Interfaces (leaf, ~28 files)

Interface definitions and real/mock implementations for all AWS services.

| Interface | Purpose |
|---|---|
| `ECSClient` | ECS service queries and mutations |
| `LambdaClient` | Lambda function queries |
| `LambdaDeployClient` | Lambda version publishing and alias routing |
| `ECSDeployClient` | ECS canary deployment (advanced configuration) |
| `ELBv2Client` | ALB listener rule and target group management |
| `EventBridgeClient` | EventBridge rule management |
| `StepFunctionsClient` | Step Functions workflow execution |
| `TerraformRunner` | Terraform plan/apply |
| `CloudWatchMetricsClient` | Metrics queries for health gates |
| `AppAutoscalingClient`, `CostExplorerClient`, `ECSCapacityClient` | Efficiency/operational surfaces |

- `internal/aws/clients.go` — `BuildRealClients()`, `BuildRealDeployClients()` factory functions
- `internal/aws/mock.go` — test mock implementations
- Imports: aws-sdk-go-v2 only; no internal imports

### internal/grafana/ — Grafana Client (leaf, 5 files)

Grafana HTTP client interface and types.

- `internal/grafana/grafana.go` — `GrafanaClient` interface (GetDashboard, PutDashboard, ListAlertRules, GetDatasources, etc.)
- `internal/grafana/grafana_client.go` — HTTP client implementation
- `internal/grafana/types.go` — `Dashboard`, `DashboardFull`, `RuleGroup`, `AlertRule`, `Datasource` types
- Imports: no internal imports

### internal/amp/ — AMP Client (leaf, ~6 files)

Amazon Managed Prometheus client.

- `internal/amp/amp.go` — `AMPClient` interface (Health, Query, QueryRange, Series)
- `internal/amp/amp_client.go` — SigV4-authenticated HTTP client implementation
- Imports: `internal/metrics` (for metric name constants)

### internal/dashgen/ — Dashboard Generator (hub, ~8 files)

Generates Grafana dashboard JSON from manifest service definitions.

- `internal/dashgen/generator.go` — `Generate()` entry point; archetype-dispatch to composer functions
- ECS and Lambda dashboard composers per archetype
- `internal/dashgen/panels.go` — shared panel builders (request rate, error rate, latency, CPU, memory)
- Imports: `pkg/manifest`, `internal/metrics`, `github.com/grafana/grafana-foundation-sdk/go`

### internal/metrics/ — Metric Name Constants (leaf, 2 files)

Canonical Prometheus metric name constants shared across packages. Pure constants, no logic.

### internal/ci/ — CI Gate (leaf, ~5 files)

GitHub Actions CI status querying via `gh` CLI. Concurrent per-repo queries with `errgroup`.

### internal/cli/ — CLI Output Utilities (leaf, 2 files)

Output formatting: `PrintTable`, `PrintJSON`, `ColorStatus`, `ColorHealth`, tag constants.

### internal/config/ — Runtime Configuration (near-leaf, 2 files)

Manifest path discovery and environment variable resolution. Discovery order: `--manifest` flag → `A8_MANIFEST` env → CWD upward walk.

### internal/release/ — Release Orchestration (~9 files)

SDK ecosystem release management. DAG topological sort (Kahn's algorithm), changeset parsing, CodeArtifact publishing.

### internal/scaffold/ — Terraform Scaffold Generator (~5 files)

Generates Terraform HCL files from manifest templates.

### internal/tfmod/ — Terraform Module Ref Scanner (~4 files)

Scans and rewrites `?ref=` pins in Terraform source attributes.

### internal/tfstate/ — Terraform State Bootstrap (~5 files)

Terraform backend state configuration bootstrapping.

### internal/fork/ — Fork Workflow (~7 files)

Operations for forking the `a8` CLI into a new organization.

### internal/workflows/ — Workflow Types (leaf, 2 files)

Runtime contract types for Step Functions + Lambda orchestration.

---

## Layer Boundaries

The import graph follows a strict layered model with no circular dependencies:

```
cmd/a8/ (CLI surface)
  └─> internal/config/         (config resolution)
  └─> internal/reconcile/      (reconcile engine)
  └─> internal/deploy/         (deployment lifecycle)
  └─> internal/scaffold/       (terraform scaffolding)
  └─> internal/release/        (SDK release orchestration)
  └─> internal/tfmod/          (terraform module upgrade)
  └─> internal/tfstate/        (terraform state bootstrap)
  └─> internal/fork/           (fork workflow)
  └─> internal/ci/             (CI gate)
  └─> internal/dashgen/        (dashboard generation)
  └─> internal/workflows/      (workflow types)
  └─> internal/cli/            (output formatting)
  └─> pkg/manifest/            (manifest types)

internal/reconcile/
  └─> pkg/manifest/
  └─> internal/aws/
  └─> internal/grafana/
  └─> internal/amp/
  └─> internal/metrics/
  └─> internal/dashgen/
  └─> internal/tfmod/

internal/deploy/
  └─> pkg/manifest/
  └─> internal/aws/
  └─> internal/amp/
  └─> internal/metrics/

internal/dashgen/
  └─> pkg/manifest/
  └─> internal/metrics/
  └─> grafana-foundation-sdk (external)

internal/amp/
  └─> internal/metrics/

internal/config/
  └─> pkg/manifest/

internal/tfstate/
  └─> internal/aws/

internal/scaffold/, internal/fork/, internal/tfmod/, internal/release/
  └─> pkg/manifest/

internal/aws/, internal/grafana/, internal/metrics/,
internal/ci/, internal/workflows/, internal/cli/
  └─> external dependencies only (leaf packages)
```

**Hub packages** (import many siblings):
- `internal/reconcile/` — imports aws, grafana, amp, metrics, dashgen, tfmod, manifest
- `internal/deploy/` — imports aws, amp, metrics, manifest
- `cmd/a8/` — imports all internal packages

**Leaf packages** (no internal imports or single import):
- `pkg/manifest/` — imported by everyone, imports nothing internal
- `internal/aws/` — imported by reconcile, deploy, amp, tfstate; imports nothing internal
- `internal/grafana/` — leaf
- `internal/metrics/` — pure constants, imports nothing
- `internal/workflows/` — pure types, imports nothing
- `internal/cli/` — imports only fatih/color
- `internal/ci/` — imports only golang.org/x/sync
- `internal/tfmod/` — imports nothing internal
- `internal/release/` — imports nothing internal

**Layer boundary enforcement patterns**:
- `internal/deploy/types.go` documents its layer rule: "this package imports `internal/aws` and `pkg/manifest`. It does NOT import `internal/reconcile` or `internal/amp`." (BC-07)
- The `infrastructure` archetype guard (`IMPLICIT-08`) is enforced at `reconcile.NewDiffer()` — infrastructure services return an error before any diff computation
- The `ExitCodeError` type in `cmd/a8/` is the only cmd-layer type; all other types flow upward from `internal/`
- `PROGRESSING` status bypass (BC-04): when an active canary deployment exists, the reconciler emits `StatusProgressing` rather than treating in-flight drift as actionable

---

## Entry Points and API Surface

### CLI Entry Point

`cmd/a8/main.go`: `main()` calls `rootCmd.Execute()`. Error handling:
1. Unwraps `*ExitCodeError` for structured exit codes
2. Falls back to `os.Exit(1)` for unstructured errors

### Root Command and Global Flags

Defined in `cmd/a8/root.go`. Persistent flags:
- `--manifest string` — override manifest path
- `--output string` — `table|json` (default: `table`)
- `--verbose/-v bool` — enable debug logging
- `--env string` — override `A8_ENV`

Logger initialized in `PersistentPreRunE` after flags parse. All log output goes to `os.Stderr`; user-facing output uses `os.Stdout`.

### CLI Command Tree

| Command | Subcommands | Purpose |
|---|---|---|
| `a8 version` | — | Print build version, commit, date, Go version |
| `a8 reconcile` | `plan [svc]`, `apply [svc]`, `watch` | Diff and converge manifest vs. AWS state |
| `a8 validate` | — | Validate manifest.yaml |
| `a8 svc` | `enable/disable/pause/resume <svc>`, `status [svc]`, `list` | Manage service control state in manifest |
| `a8 train` | `status`, `create`, `lock`, `promote`, `test`, `bump`, `publish` | Release train lifecycle management |
| `a8 tf` | `init/plan/apply <svc>`, `bootstrap`, `upgrade-refs` | Terraform operations with manifest-injected env vars |
| `a8 doctor` | — | Ecosystem health checks |
| `a8 dev` | `up/down/logs/status` | Local dev environment (docker compose) |
| `a8 query` | `services`, `repos`, `sdks`, `service-archetype`, ... , `events` | Read manifest fields as structured data |
| `a8 status` | — | Ecosystem dashboard (all services) |
| `a8 observe` | `health`, `metrics`, `logs`, `traces`, `dashboard`, `generate dashboards` | Observability queries (Grafana/AMP) |
| `a8 rollback` | `<svc>` | ECS service rollback to previous task definition |
| `a8 workflow` | `list`, `status <wf>`, `run <wf>`, `history <wf>` | Step Functions workflow management |
| `a8 changeset` | `add`, `status`, `validate` | Release changeset intent files |
| `a8 deploy` | `service <svc>`, `status <svc>`, `promote <svc>`, `rollback <svc>` | Progressive canary deployment |
| `a8 ci` | `gate`, `status` | CI pipeline gate evaluation |
| `a8 fork` | `check`, `rename`, `update-codeartifact`, `rewrite-module`, `verify` | Fork/rebrand operations |
| `a8 init` | — | Interactive fork wizard (huh TUI) |
| `a8 scaffold terraform` | — | Generate Terraform scaffold files from manifest |
| `a8 completion` | `bash/fish/powershell/zsh` | Shell completion generation |

### Key Exported Interfaces (contracts between packages)

| Interface | Package | Consumers |
|---|---|---|
| `ECSClient` | `internal/aws` | `internal/reconcile`, `internal/deploy`, `cmd/a8` |
| `LambdaClient` | `internal/aws` | `internal/reconcile`, `cmd/a8` |
| `EventBridgeClient` | `internal/aws` | `internal/reconcile`, `cmd/a8` |
| `TerraformRunner` | `internal/aws` | `internal/reconcile`, `cmd/a8` |
| `Differ` | `internal/reconcile` | `internal/reconcile` (internal dispatch) |
| `AMPClient` | `internal/amp` | `internal/deploy`, `internal/reconcile`, `cmd/a8` |
| `GrafanaClient` | `internal/grafana` | `internal/reconcile`, `cmd/a8` |
| `DeployStrategy` | `internal/deploy` | `internal/deploy` (controller composition) |
| `CanaryScorer` | `internal/deploy` | `internal/deploy` (health gate) |
| `DeployStateStore` | `internal/deploy` | `internal/deploy` (file-backed persistence) |
| `StrategyHydrator` | `internal/deploy` | Cross-process state recovery injection |
| `BootstrapClient` | `internal/tfstate` | `cmd/a8` |
| `CommandRunner` | `internal/release` | `internal/release` (subprocess injection) |

---

## Key Abstractions

### 1. `manifest.Manifest` — Root Document (`pkg/manifest/types.go`)

The single source of truth for all ecosystem metadata. Parsed from `manifest.yaml`. Contains:
- `Org` — AWS account, region, cluster name, CodeArtifact config
- `Services map[string]Service` — all service descriptors keyed by manifest name
- `SDKs map[string]SDK` — Python SDK package definitions
- `ReleaseTrains []ReleaseTrain` — coordinated SDK version sets
- `Observability *Observability` — optional Grafana/AMP config
- `Workflows map[string]Workflow` — Step Functions state machines
- `EventBus *EventBus` — custom EventBridge bus

### 2. `manifest.Archetype` — Service Type Enum (`pkg/manifest/types.go`)

Typed string with 7 values: `ecs-fargate-rds`, `ecs-fargate-stateless`, `ecs-fargate-hybrid`, `ecs-fargate-worker`, `lambda-scheduled`, `lambda-event-driven`, `infrastructure`. The `archetypeDescriptors` map provides metadata for each. Adding a new archetype requires updates in 7 places (documented in types.go).

### 3. `reconcile.Surface` — Diff Result Unit (`internal/reconcile/surface.go`)

Represents the diff state for one reconciliation surface. Fields: `Name SurfaceName`, `Desired string`, `Actual string`, `Status Status`. Surface names are COORDINATED constants (LOAD-001) versioned via `schema_version` in JSON output.

### 4. `reconcile.Engine` — Reconcile Orchestrator (`internal/reconcile/engine.go`)

Holds `Manifest`, `Clients`, `Log`, optional `ObsClients`, `EfficiencyClients`. Key methods: `PlanService()`, `PlanAll()` (concurrent with errgroup). The Engine is read-only — it never writes to manifest.yaml. Infrastructure archetype guard enforced at `NewDiffer()`.

### 5. `deploy.DeployController` — Progressive Deployment State Machine (`internal/deploy/controller.go`)

State machine: PENDING → CANARY_ACTIVE → VERIFYING → PROMOTING|ROLLING_BACK → COMPLETE|FAILED. Key methods: `Deploy()` (full lifecycle), `Promote()`, `Rollback()`, `RecoverAndPromote()`, `RecoverAndRollback()` (cross-process recovery). Composes `DeployStrategy` + `HealthGate`. Three-tier state lookup: in-memory → state file → platform API.

### 6. `deploy.HealthGate` — Canary Health Evaluation (`internal/deploy/health.go`)

Evaluates canary health via differential metric comparison (canary vs baseline population). Uses `MetricsQuerier` (CloudWatch or AMP backends) + `CanaryScorer` (differential or statistical). `VerifyCanary()` runs repeated `ScoreCanary()` calls at `EvalInterval` for `duration`. `ConsecutiveFails` threshold triggers early rollback. NODATA during `WarmupDuration`.

### 7. `manifest.Service` — Service Descriptor (`pkg/manifest/types.go`)

All archetype-specific fields in one struct. Key resolve methods: `Enabled()`, `ResolveFunctionName()`, `ResolveRuleName()`, `ResolveSchedulesEnabled()`. `ServiceControl` uses `*bool` pointer semantics to distinguish absent (default `true`) from explicit `false` (LOAD-002 pattern).

### 8. `Org.Resolve*()` — Default-Computing Accessors

9 `Resolve*()` methods on `Org` derive values from explicit fields or computed defaults: `ResolveClusterName()`, `ResolveMetricsPrefix()`, `ResolveCodeArtifactDomain()`, `ResolveEnvPrefix()`, `ResolveStateBucket()`, `ResolveCPUArchitecture()`, etc. Centralizes default-resolution logic.

### 9. `loadManifestFromFlags()` — Universal Command Preamble (`cmd/a8/helpers.go`)

Called at the start of virtually every command's `RunE`. Chains `config.Resolve()` → `manifest.FindManifest()` → `manifest.LoadFromFile()`.

### Design Patterns

- **Polymorphic dispatch on Archetype**: Both `NewDiffer()` and `archetypeDescriptors` use switch/map dispatch
- **Interface injection for testability**: Every AWS service has an interface; factory vars enable test injection
- **COORDINATED string values (LOAD-001)**: `SurfaceName` and `Status` constants form a versioned JSON contract
- **Error wrapping convention**: `fmt.Errorf("context: %w", err)` with sentinel errors via `errors.Join`

---

## Data Flow

### Manifest Load Pipeline

```
disk: manifest.yaml
  -> os.ReadFile()
  -> yaml.Unmarshal() into *Manifest
  -> ValidateAll() (validArchetypes, field presence, URL format)
  -> *Manifest returned to cmd layer
     -> flags (--manifest, --env) -> config.Resolve() -> Config{ManifestPath, Env, Verbose}
     -> env: A8_MANIFEST overrides auto-discovery
     -> env: A8_ENV / AUTOM8Y_ENV sets environment
```

### Reconcile Pipeline

```
User: a8 reconcile plan [--with-terraform]
  │
  ├─ internal/config.Resolve() ── flag > A8_MANIFEST env > CWD upward walk
  │    └─> pkg/manifest.Load() ── parse manifest.yaml into manifest.Manifest
  │
  ├─ internal/aws.BuildRealClients() ── AWS SDK credential chain → ECS/Lambda/EB/SF clients
  │
  ├─ reconcile.NewEngine(manifest, clients, logger) [+ optional ObsClients]
  │
  └─ engine.PlanAllWithOpts(ctx, PlanOptions{WithTerraform, Concurrency, DesiredModuleRef})
       │
       ├─ [concurrent via errgroup] for each service:
       │    ├─ reconcile.NewDiffer(archetype, clients) → Differ implementation
       │    ├─ differ.Diff(ctx, serviceName, svc, manifest) → []Surface
       │    └─ [optional] TerraformRunner.Plan(), diffModuleVersion(), OBS diffs
       │
       └─ BuildPlan(service, surfaces, planOpts) → Plan{Surfaces, Operations}
```

### Deploy Pipeline

```
User: a8 deploy --service auth --image-tag v1.2.3
  │
  ├─ manifest.Load() + DeployTarget construction
  ├─ NewDeployStrategy(archetype) → LambdaCanaryStrategy | ECSCanaryStrategy
  ├─ NewHealthGate(MetricsQuerier, Scorer, log)
  │
  └─ DeployController.Deploy(ctx, target, opts):
       ├─ PENDING: register active deployment
       ├─ CANARY_ACTIVE: strategy.StartCanary(weight=5%)
       ├─ VERIFYING: HealthGate.VerifyCanary(duration=5m)
       ├─ [on pass] PROMOTING: strategy.Promote()
       └─ [on fail] ROLLING_BACK: strategy.Rollback()
```

### Release Pipeline

```
User: a8 train publish [--version v1.2.0]
  │
  ├─ manifest.Load() → SDK dependency graph
  ├─ release.BuildDAG(sdks) → topological sort (Kahn's algorithm)
  ├─ release.ParseChangesets(.changeset/*.yaml) → []Changeset
  └─ for each SDK in topological order:
       ├─ publish to CodeArtifact
       └─ release.VerifyCodeArtifact() with exponential backoff
```

### Config Cascade

```
Env resolution order:
  1. --env flag (explicit CLI override)
  2. A8_ENV env var (canonical)
  3. {ORG}_ENV env var (legacy fallback, e.g., AUTOM8Y_ENV)
  4. manifest default

Manifest path resolution:
  1. --manifest flag
  2. A8_MANIFEST env var
  3. CWD upward walk looking for manifest.yaml
```

### Dashboard Generation Flow

```
a8 observe generate-dashboards [--service auth]
  └─ for each service:
       ├─ dashgen.Generate(serviceName, svc, manifest, datasourceUID)
       │    ├─ composerFor(archetype) → composerFunc
       │    └─ marshalDashboard() → deterministic JSON
       └─ write to output file
```

### Manifest Write Pipeline (svc mutations)

```
a8 svc enable/disable/pause/resume <svc>
  -> manifest.LoadAndParse() (no validation — preserve YAML structure)
  -> manifest.Writer.SetField(node, field, value) — yaml.Node tree manipulation
  -> os.WriteFile(manifestPath, updatedYAML)
```

---

## Knowledge Gaps

1. **`internal/deploy` `LambdaCanaryStrategy`** — only `ECSCanaryStrategy` and `DeployController` were read in detail. Lambda canary implementation details (ADR-087) are undocumented here.

2. **`pkg/manifest/types.go` remaining `Service` fields** — Lambda-specific fields, SDK fields, and observability fields beyond line 550 were not fully read.

3. **`internal/reconcile` efficiency/capacity differs** — `differ_capacity.go`, `differ_efficiency.go`, `differ_spot.go` were not fully traced.

4. **`cmd/a8/dev.go`** — not read in detail. Purpose inferred from file name.

5. **`pkg/manifest/validate.go`** — validation rules not documented in detail.

6. **`internal/aws` real client implementations** — only interfaces examined; `_client.go` files not read.

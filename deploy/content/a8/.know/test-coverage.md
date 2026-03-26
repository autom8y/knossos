---
domain: test-coverage
generated_at: "2026-03-25T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "429f242"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "1d5c945d46857fffeeb248674d6adbe89b9b33b94dfd6a946448973c591917b6"
---
# Codebase Test Coverage

## Coverage Gaps

### Coverage Infrastructure

The test command is `CGO_ENABLED=0 go test -count=1 ./...` (from `Justfile`). No coverage measurement (`-coverprofile`) is collected during standard test runs or CI. No coverage badge or threshold enforcement is present in the repository.

### Package Coverage Distribution

All 18 source packages have at least one test file. Total: 1,549 test functions across 88 test files.

| Package | Test Functions | Test Files |
|---|---|---|
| `cmd/a8` | 458 | 23 |
| `internal/reconcile` | 213 | 13 |
| `pkg/manifest` | 169 | 4 |
| `internal/deploy` | 165 | 6 |
| `internal/release` | 117 | 10 |
| `internal/fork` | 74 | 6 |
| `internal/scaffold` | 70 | 3 |
| `internal/dashgen` | 66 | 7 |
| `internal/aws` | 56 | 6 |
| `internal/amp` | 40 | 1 |
| `internal/grafana` | 27 | 1 |
| `internal/tfmod` | 25 | 2 |
| `internal/cli` | 21 | 1 |
| `internal/ci` | 20 | 1 |
| `internal/config` | 15 | 1 |
| `internal/tfstate` | 8 | 1 |
| `internal/metrics` | 3 | 1 |
| `internal/workflows` | 2 | 1 |

### Source Files Without Paired Test Files

**cmd/a8 — untested file list** (no `{name}_test.go` counterpart):

| File | Notes |
|---|---|
| `cmd/a8/env_helpers.go` | Environment variable utilities |
| `cmd/a8/helpers.go` | General CLI helpers |
| `cmd/a8/obs_helpers.go` | Observability client construction |
| `cmd/a8/reconcile_watch.go` | `watch` subcommand registration + emitter builder |
| `cmd/a8/root.go` | Root cobra command setup |
| `cmd/a8/scaffold_terraform.go` | Terraform scaffold command wiring |
| `cmd/a8/tf_bootstrap.go` | Bootstrap command handler |
| `cmd/a8/tf_upgrade.go` | Module upgrade command wiring |
| `cmd/a8/train_bump.go`, `train_create.go`, `train_publish.go` | Train subcommands |
| `cmd/a8/fork_*.go` (5 files) | Fork subcommand wiring; logic in `internal/fork` (tested) |
| `cmd/a8/validate.go` | Tested via `TestValidate_*` in `cmd_test.go` |

**internal/aws — largest gap by ratio** (23 of 28 files untested):

All `_client.go` files wrap AWS SDK calls — behavior tested indirectly via mocks. Direct coverage exists only for: `clients.go`, `cloudwatch.go`, `errors.go`, `lambda_deploy.go`, `mock.go`, `terraform_runner.go` (security tests).

**internal/reconcile — untested source files:**
- `differ_capacity.go`, `differ_efficiency.go`, `differ_spot.go` — no direct test files
- `emitter.go` — tested via `watch_emitter_test.go` (cross-file)
- `planner.go` — tested via `engine_test.go` (cross-file)

**internal/deploy — untested source files:**
- `progress.go` — deployment progression builder, no direct tests
- `spot_detector.go` — FARGATE_SPOT detection, no direct tests
- `store_mock.go` — test double, no behavioral tests

**Other gaps:**
- `pkg/manifest/node.go` — no test file; all other manifest files tested
- `internal/workflows/types_test.go` — only 2 tests

### Critical Path Coverage

| Critical Path | Coverage Status | Evidence |
|---|---|---|
| CLI command handlers | Well covered | 458 tests across 23 test files |
| Reconcile pipeline | Comprehensive | 213 tests; engine, executor, adversarial/edge tests |
| Deploy pipeline | Comprehensive | 165 tests; controller (1,248 lines), ECS strategy, health |
| Hook handlers (watcher) | Well covered | `watcher_test.go` (18 tests), `watch_state_test.go` |
| Terraform security | Dedicated file | `terraform_runner_security_test.go` — DEBT-001/006 |
| Kill chain (security) | Dedicated files | `kill_chain_1_test.go`, `kill_chain_4_test.go` |
| OBS diff pipeline | Exhaustive | `differ_obs_test.go` (2,200+ lines), adversarial + edge |
| Release pipeline | Integration + unit | `integration_test.go` + 10 unit test files |
| Scaffold golden files | Golden file testing | `internal/scaffold/testdata/golden/` |

### Prioritized Gap List

1. **`internal/aws/`** — 23 of 28 files untested; AWS SDK boundary wrappers. Highest risk by ratio.
2. **`internal/deploy/progress.go`, `spot_detector.go`** — deployment helpers, no tests
3. **`cmd/a8/reconcile_watch.go`** — watch-mode command handler, complex signal/loop logic
4. **`pkg/manifest/node.go`** — single untested file in well-tested package
5. **`internal/reconcile/differ_capacity.go`, `differ_efficiency.go`, `differ_spot.go`** — may be exercised through engine_test but no direct tests

---

## Testing Conventions

### Test Function Naming

Dominant pattern: `TestSubject_Condition_ExpectedResult`:
```
TestECSFargateStateless_EnabledCountMatches
TestDiffDashboard_DashboardNotFound_Drift
TestExecutor_ThirdOperationFails_RestSkipped
```

Security/kill-chain tests use longer descriptive names:
```
TestKillChain1_TFPathTraversal_RejectedByValidation
```

Exhaustiveness guards: `TestType_Purpose`:
```
TestBuildMockClients_ExhaustivenessGuard
TestNewDiffer_ExhaustivenessGuard
TestSurfaceID_AllSurfacesHaveID
```

### Subtest Patterns

`t.Run()` used 76 times. Table-driven subtests in 59 instances:
```go
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```

### Assertion Patterns

**Primary: stdlib only.** Zero testify imports.
- `t.Fatalf` / `t.Errorf` for all assertions
- `github.com/google/go-cmp/cmp` for structural diffs (primarily `internal/reconcile`)
- Standard `if got != want { t.Errorf(...) }` pattern dominates

### Test Helper Patterns

`t.Helper()` called 69 times. Canonical helpers:
- `cmd/a8/cmd_test.go`: `copyManifestFixture(t)`, `executeCommand(args...)`, `resetCobraFlags(cmd)`, `resetFlags(t)`
- `internal/reconcile/engine_test.go`: `discardLogger()`, `ptr(b)`, `intPtr(i)`, `buildManifest(services)`, `defaultClients()`
- `internal/deploy/controller_test.go`: `newTestController(strategy, health, recorder)`
- `cmd/a8/mock_guard_test.go`: `setEnvForTest(t, vars)`, `unsetEnvForTest(t, keys)`

### Mock Patterns

Two-tier mock strategy:
1. **Shared mock types in `internal/aws/mock.go`** — all Mock* structs verified by `mock_test.go` to satisfy interfaces
2. **Local mock structs in test files** — `mockStrategy`, `mockRecorder`, `mockMetricsQuerier`

Mock pattern: structs with `Calls []callRecord` fields to verify invocations without external frameworks.

### Test Skip Patterns

- `t.Skip("gh not on PATH")` — 9 occurrences in `cmd/a8/ci_test.go`
- `testing.Short()` — in `internal/fork/rewrite_module_test.go`
- Platform-conditional skips in `internal/aws/terraform_runner_security_test.go` (5 occurrences)

### Testdata Directory Contents

| Directory | Contents | Purpose |
|---|---|---|
| `cmd/a8/testdata/` | `manifest_workflows.yaml` | Workflow-enabled manifest fixture |
| `pkg/manifest/testdata/` | 3 YAML manifest fixtures | Primary manifest fixtures |
| `internal/release/testdata/` | Changeset YAML files, pyproject TOML fixtures | Release pipeline inputs |
| `internal/scaffold/testdata/golden/` | 5 archetype golden outputs | Golden file testing |
| `internal/tfmod/testdata/` | 6 Terraform HCL fixture directories | TF module scanning edge cases |

### Test Fixture Patterns

1. **Golden file pattern**: `internal/scaffold/scaffold_test.go` uses `flag.Bool("update", false, "update golden files")`
2. **Testdata YAML fixtures**: `internal/release` and `pkg/manifest` use static YAML files
3. **In-memory construction**: Most packages construct test fixtures programmatically using builder helpers

### Test Environment Management

- **`TestMain` in `cmd/a8/cmd_test.go`**: Sets `AUTOM8Y_ENV=development` for all cmd/a8 tests
- **`t.TempDir()`**: Universally used (268 calls) for filesystem isolation
- **`t.Setenv`**: Used for environment variable mutation with automatic cleanup
- **`t.Cleanup`**: Used for flag state restoration via `resetFlags(t)`
- **Flag leakage prevention**: `resetCobraFlags(rootCmd)` walks cobra tree (TENSION-006 pattern)

### Export Test Pattern

`internal/release/export_test.go` uses `var BuildSDKExported = buildSDK` for black-box test access to internal logic.

---

## Test Structure Summary

### High-Level Distribution

- **Total test files** (main tree): 88
- **Total `Test` function declarations**: 1,549
- **Packages with tests**: 18 of 18 (100%)
- **Test command**: `CGO_ENABLED=0 go test -count=1 ./...`

### Most Heavily Tested Areas

1. **`cmd/a8`** — 458 tests, 23 files. Includes security kill-chain tests, audit log tests, per-command tests.
2. **`internal/reconcile`** — 213 tests, 13 files. Adversarial dashboard drift detection (2,200+ line test file).
3. **`pkg/manifest`** — 169 tests, 4 files. Loader, validator, type marshaling, writer.
4. **`internal/deploy`** — 165 tests, 6 files. Controller, ECS/Lambda strategies, health gates.
5. **`internal/release`** — 117 tests, 10 files including `integration_test.go`.

### Test Package Naming Patterns

| Pattern | Examples |
|---|---|
| `package main` (internal, cmd/a8) | `cmd_test.go`, `deploy_test.go` |
| `package {pkg}_test` (external, black-box) | `reconcile_test`, `release_test`, `fork_test`, `manifest_test` |
| `package {pkg}` (internal, white-box) | `dashgen`, `deploy`, `aws`, `reconcile` |

Mixed: some packages use both internal and external test packages.

### Integration vs Unit Tests

No formal build-tag separation. Integration tests identified by naming:
- `internal/release/integration_test.go` — "Cross-domain integration tests"
- `cmd/a8/kill_chain_1_test.go`, `kill_chain_4_test.go` — security integration
- `internal/fork/rewrite_module_test.go` — gated by `testing.Short()`
- `test/workstation/workstation-recipes.bats` — shell integration

~95% unit tests with mocked dependencies; ~5% integration tests.

### CI Configuration

| Workflow | Trigger | Test Command | Gates |
|---|---|---|---|
| `go-ci.yml` | push/PR to main (Go paths) | `go test -count=1 -coverprofile=coverage.out ./...` | 55% coverage floor, lint (golangci-lint v2.11.2), build |
| `workstation-ci.yml` | workstation config changes | bats smoke tests | Exit code check |
| `e2e-distribution.yml` | on release publish | Homebrew install + binary smoke test | Binary executes cleanly |

### Local Test Commands

From `Justfile`:
- `just test` → `CGO_ENABLED=0 go test -count=1 ./...`
- `just test-verbose` → same with `-v`

No `-race` or `-cover` target in Justfile.

---

## Knowledge Gaps

1. **Actual coverage percentage unknown** — No cached `coverage.out` file found. The 55% CI floor is documented but current actual percentage unverifiable without running tests.

2. **`internal/workflows` depth** — Only 2 tests. Coverage path through workflow reconcile tests unclear without profiling data.

3. **Race condition coverage** — No `-race` test target; concurrent patterns in watcher and deploy controller may have untested races.

4. **`pkg/manifest/node.go` content** — No test. Role relative to `loader.go` and `types.go` not fully characterized.

5. **`internal/fork/verify.go`** — Single `Verify` function with no test file.

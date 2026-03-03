---
domain: test-coverage
generated_at: "2026-03-03T19:45:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "1599813"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
---

# Codebase Test Coverage

**Language**: Go 1.22+ (confirmed by `go.mod`)
**Test runner**: `CGO_ENABLED=0 go test ./...`
**Total test files**: 175
**Total test functions**: 2,695
**Total packages (internal)**: 58 (49 with tests, 9 without)

## Coverage Gaps

### Package-Level Gap Inventory

49 of 58 internal packages have test files (84.5%). 9 packages have zero test files:

| Package | Files | Lines (approx) | Criticality |
|---|---|---|---|
| `internal/assets` | `assets.go` | 42 | Low — thin embed wrapper |
| `internal/cmd/common` | `annotations.go`, `context.go`, `embedded.go` | 160 | Medium — shared CLI context wiring used by all commands |
| `internal/cmd/root` | `root.go` | 224 | Medium — Cobra root setup, version injection |
| `internal/cmd/artifact` | 5 files | 600 | Medium — `ari artifact` commands: list, query, rebuild, register |
| `internal/cmd/inscription` | 6 files | 729 | High — `ari inscription sync/rollback/validate/backups/diff` CLI layer |
| `internal/cmd/manifest` | 5 files | 572 | Medium — `ari manifest diff/merge/show/validate` CLI layer |
| `internal/cmd/naxos` | 2 files | 138 | Low — thin CLI wrapper around `internal/naxos` (which has tests) |
| `internal/cmd/tribute` | 2 files | 218 | Low — thin CLI wrapper |
| `internal/cmd/provenance` | `provenance.go` | 272 | Low — thin CLI wrapper |

**Pattern**: Untested packages are overwhelmingly CLI command wiring (`internal/cmd/*`). The underlying business logic packages (`internal/inscription`, `internal/materialize`, etc.) are tested. The gap is in the command-layer handlers that parse flags, call business logic, and format output.

### Within-Package Gaps (files without dedicated tests)

Inside otherwise-tested packages, several specific source files lack direct test counterparts:

- `internal/cmd/session/migrate.go` (267 lines) — schema migration logic, no test file
- `internal/cmd/session/park.go` (148 lines) — session park command
- `internal/cmd/session/resume.go` (118 lines) — session resume command
- `internal/cmd/session/audit.go` (135 lines) — audit command
- `internal/cmd/session/transition.go` (217 lines) — shared transition logic

Note: `park.go` and `resume.go` functionality may be partially covered by integration tests in `session_test.go` and `integration_test.go` within the same package — they are white-box tests (same package declaration `package session`).

### Critical Path Coverage Assessment

| Critical Path | Coverage | Assessment |
|---|---|---|
| **Sync pipeline** (`internal/inscription`) | Strong — 7 test files, 908-line integration test | Well covered. `pipeline_test.go`, `merger_test.go`, `generator_test.go`, `integration_test.go` all present. |
| **Materialization** (`internal/materialize`) | Strong — 26 test files | Most heavily tested area. Includes scar regression tests and provenance integration tests. |
| **Hook handlers** (`internal/cmd/hook`) | Strong — 12 test files, 11,167 total lines | `writeguard`, `agentguard`, `clew`, `validate`, `autopark`, `precompact`, `subagent` all tested. |
| **Hook contracts** (`internal/hook/clewcontract`) | Strong — 9 test files | Typed events, writer, triggers, lifecycle, orchestrator all tested. |
| **Session management** (`internal/cmd/session`) | Strong — 17 test files | Park/resume/migrate/transition lack dedicated test files but broader session logic is heavily tested. |
| **Agent scaffolding** (`internal/agent`) | Good — 7 test files | Validate, frontmatter, scaffold, MCP, integration all tested. |
| **Inscription CLI** (`internal/cmd/inscription`) | Absent | 729 lines of CLI command logic, 0 test files. The underlying `internal/inscription` package is tested; this layer is not. |
| **Know system** (`internal/know`) | Strong — 4 test files | `astdiff`, `know`, `manifest`, `validate` all tested. 1,445-line `astdiff_test.go`. |
| **Sails** (`internal/sails`) | Strong — 7 test files including integration | Thresholds, contracts, generator, proofs, integration all covered. |

### Prioritized Gap List (highest-risk untested areas)

1. **`internal/cmd/inscription`** (HIGH): 729 lines, zero tests. The `sync`, `rollback`, `validate`, `backups`, and `diff` subcommands orchestrate the inscription pipeline. Errors here corrupt `CLAUDE.md`. Business logic is tested in `internal/inscription`, but CLI flag parsing, error formatting, and output are not.
2. **`internal/cmd/session/migrate.go`** (HIGH): 267 lines of schema migration logic — handles version upgrades of `SESSION_CONTEXT.md`. Migrations that fail silently could corrupt session state.
3. **`internal/cmd/common`** (MEDIUM): 160 lines of shared CLI scaffolding — annotations, context injection, embedded asset wiring. Defects propagate to all commands. Small target, high blast radius.
4. **`internal/cmd/artifact`** (MEDIUM): 600 lines — `register` (205 lines) in particular contains non-trivial logic. The underlying `internal/artifact` domain package is tested.
5. **`internal/cmd/manifest`** (MEDIUM): 572 lines — `merge` (163 lines) and `diff` (102 lines) are non-trivial. `internal/manifest` domain is tested.
6. **`internal/cmd/session/transition.go`** (MEDIUM): 217 lines of shared state transition logic. May be exercised by integration tests but has no dedicated test.
7. **`internal/cmd/root`** (LOW-MEDIUM): 224 lines — root Cobra setup. Errors affect startup. Low complexity but zero coverage.
8. **`internal/cmd/tribute`** (LOW): 218 lines, thin wrapper over tested domain package.
9. **`internal/cmd/naxos`** (LOW): 138 lines, thin wrapper over tested domain package.
10. **`internal/assets`** (LOW): 42-line embed wrapper, trivially correct.

### Error Path Coverage

- 101 of 175 test files (57.7%) test for error conditions (grep for `errors`, `wantErr`, `ErrInvalid`, `require.Error`, `assert.Error`).
- 15 test files use explicit `wantErr` table-driven error patterns.
- Negative tests are present but not uniform — lower-level packages have them; CLI command packages mostly lack them (and lack tests entirely).

### Coverage Measurement Infrastructure

No CI coverage pipeline found. No `go test -coverprofile` invocations in CI config or Makefiles. Coverage is not instrumented or gated.

## Testing Conventions

### Test Function Naming

Primary convention: `Test{TypeOrSubject}_{Scenario}`.

Examples observed:
- `TestFSM_CanTransition` (`internal/session/fsm_test.go:7`)
- `TestMaterializeWorkflow_WritesFile` (`internal/materialize/workflow_test.go:15`)
- `TestMaterializeWorkflow_NoWorkflowFile` (`internal/materialize/workflow_test.go:42`)
- `TestSCAR002_StagedMaterializeAbsent` (`internal/materialize/scar_regression_test.go`) — scar regression tests use `TestSCARNNN_` prefix

Alternative (simpler) naming also present:
- `TestAllConceptsLoaded`, `TestSortedNamesCorrect` (`internal/cmd/explain`) — flat descriptive names
- `TestCollectClaude_Exists`, `TestCollectClaude_NotExists` (`internal/cmd/status`) — `_Exists`/`_NotExists` suffix for presence checks

### Subtest Patterns (`t.Run`)

79 of 175 test files (45.1%) use `t.Run` for subtests.

Naming conventions within `t.Run`:
- State transition notation: `"ACTIVE->PARKED"` (`internal/session/fsm_test.go:32`)
- Struct field `name`: table-driven pattern with `tests []struct{ name string; ... }` then `t.Run(tt.name, ...)` — dominant pattern throughout
- Concept names as subtest keys: `t.Run(name, ...)` where name is a string from a slice (`internal/cmd/explain`)

### Assertion Patterns

Two patterns coexist:

1. **stdlib testing** — `t.Errorf`, `t.Fatalf`, `if got != want { t.Errorf(...) }`. Used in approximately 50% of test files. Dominant in older or simpler packages.
2. **testify** (`github.com/stretchr/testify`) — `assert.*` (23 files import it) and `require.*` (21 files import it). Used in newer, higher-complexity tests. `require.NoError(t, err)` is the dominant error guard pattern in testify-using tests.

Mixed usage: both patterns appear in the same codebase, sometimes the same package. No single enforced convention.

### Test Helper Patterns

`t.Helper()` is called in 121 of 175 test files, indicating widespread use of helper function patterns. Common helper patterns:

- `setupProject(t *testing.T) string` — creates temp directory with minimal project structure
- `setupTestSession(...)` — creates a full session fixture with `SESSION_CONTEXT.md`
- `createTestProject(t *testing.T) string` — variant naming
- `t.TempDir()` used universally for temp directories (auto-cleaned via `t.Cleanup`)

Dedicated test utility packages:
- `test/hooks/testutil/golden.go` — `GoldenFile` type with `Assert`, `AssertJSON`, `AssertJSONString`. Updated via `UPDATE_GOLDEN=1` env var.
- `test/hooks/testutil/env.go` — environment helpers for hook testing
- `test/worktree/testutil/worktree.go` — worktree test utilities
- `internal/materialize/userscope/helpers_test.go` — inline helpers

### Skip Patterns

`t.Skip` appears in 10 files. Common pattern: skip when testdata path is missing.

```go
if _, err := os.Stat(absPath); os.IsNotExist(err) {
    t.Skipf("testdata not found at %s", absPath)
}
```

Also: `testing.Short()` used in 0 files — no short-mode test gating.

### Testdata Directories

No `testdata/` directories exist at the package level. The golden file utility (`test/hooks/testutil/golden.go`) writes to `testdata/golden/` relative to the test. Only 2 test files actively reference `testdata`: `context_loader_test.go` and `workflow_test.go` in `internal/rite`.

### Integration Test Identification

No `//go:build integration` tags present in test files (only 1 file uses the old `// +build integration` tag: `internal/cmd/session/status_integration_test.go`).

Integration tests are identified by:
- File naming: `*_integration_test.go` — 7 files: `sails/integration_test.go`, `inscription/integration_test.go`, `materialize/provenance_integration_test.go`, `materialize/rite_switch_integration_test.go`, `agent/integration_test.go`, `cmd/session/moirai_integration_test.go`, `cmd/session/status_integration_test.go`
- Comment header: `// Integration tests for the inscription pipeline...`
- No build tag separating them from unit tests — they run with `go test ./...`

### Test Environment Management

- `t.TempDir()` is the universal isolation mechanism
- `os.MkdirAll` and `os.WriteFile` used to construct fixture directory trees in-test
- No database fixtures, no docker-compose, no external service mocking
- Hook tests use stdin injection by passing `bytes.NewReader(json)` directly

## Test Structure Summary

### Overall Distribution

| Area | Test Files | Est. Test Functions | Characterization |
|---|---|---|---|
| `internal/materialize` (all sub-pkgs) | 36 | ~600 | Most heavily tested. Drives confidence in the sync pipeline. |
| `internal/cmd/session` | 17 | ~450 | Deep session lifecycle coverage. |
| `internal/inscription` | 7 | ~200 | End-to-end pipeline including 908-line integration test. |
| `internal/cmd/hook` | 12 | ~350 | Hook handler logic thoroughly tested. |
| `internal/hook/clewcontract` | 9 | ~250 | Event type and contract coverage. |
| `internal/session` | 12 | ~300 | FSM, lifecycle, rotation, snapshot. |
| `internal/agent` | 7 | ~150 | Scaffold, frontmatter, MCP, integration. |
| `internal/sails` | 7 | ~200 | Contract, generator, thresholds, integration. |
| `internal/know` | 4 | ~180 | AST diff, manifest, validate, core know logic. |
| Others | ~64 | ~215 | Smaller focused packages. |

### Most Heavily Tested Areas

1. **Materialization** (`internal/materialize`): 26 test files in the core package alone, plus 10 more across sub-packages. Coverage includes scar regression testing (prevents re-introduction of known-bad patterns), archetype tests, soft-switch tests, cross-rite agent tests.
2. **Session management** (`internal/cmd/session` + `internal/session`): Combined 29 test files. The session lifecycle FSM, create/wrap/park/resume/gc/lock/log/field/snapshot all have dedicated tests.
3. **Hook system** (`internal/cmd/hook` + `internal/hook` + `internal/hook/clewcontract`): Combined 23 test files. The hook pipeline is the most safety-critical path (controls Claude Code behavior) and has correspondingly dense test coverage.

### Test Package Naming

- **White-box testing dominant**: 170 of 175 test files use `package {pkgname}` (same package as the code under test). This gives tests access to unexported identifiers.
- **Black-box testing rare**: Only 5 files use `package {pkgname}_test`: 4 in `internal/manifest` (`manifest_test`) and 1 in `internal/sync` (`sync_test`).

### Integration vs Unit Test Boundary

Integration tests run with `go test ./...` — not separated by build tags. The absence of build tag gating means CI runs all tests including integration-style filesystem tests on every invocation. This is intentional — tests use `t.TempDir()` for isolation and are fast enough to run unconditionally.

### TestMain Patterns

No `TestMain` functions found in any test file. Test lifecycle is managed entirely by individual test functions using `t.TempDir()` and `t.Cleanup()`.

### How Tests Are Run

```
CGO_ENABLED=0 go test ./...
```

No special flags, no build constraints required. No coverage profiling in CI.

## Knowledge Gaps

- **Actual runtime coverage percentages** are unknown — no `go test -cover` output was available, and no CI coverage artifacts exist.
- **`internal/cmd/session/park.go`, `resume.go`, `transition.go`** may be indirectly covered by the 17 session test files (all in `package session` white-box). Their actual test coverage cannot be confirmed without running `go test -coverprofile`.
- **`.github/workflows/`** contents were not read — CI configuration could contain coverage steps not visible from Makefile/shell script search.
- **`test/worktree/testutil/worktree.go`** and **`test/hooks/testutil/`** — these are utility packages used by tests that themselves have 0 test files.
- The golden file mechanism (`UPDATE_GOLDEN=1`) indicates snapshot tests exist but which tests currently use golden files vs in-line assertions could not be enumerated without running the test suite.

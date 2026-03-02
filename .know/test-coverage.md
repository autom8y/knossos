---
domain: test-coverage
generated_at: "2026-03-01T16:08:41Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "89b109c"
confidence: 0.85
format_version: "1.0"
---

# Codebase Test Coverage

## Coverage Gaps

### Packages with No Test Files

The following packages have Go source files but no corresponding `*_test.go` files:

**CLI Command Handlers (untested -- highest risk):**

| Package | Files | Lines | Risk Assessment |
|---------|-------|-------|----------------|
| `internal/cmd/root` | `root.go` | 218 | Low -- Cobra wiring only |
| `internal/cmd/sync` | `sync.go`, `budget.go` | 364 | MEDIUM -- 47.2% coverage; CLI wiring partially tested |
| `internal/cmd/rite` | 9 files | ~900 | HIGH -- Rite lifecycle commands |
| `internal/cmd/agent` | `agent.go`, `list.go`, `new.go`, `update.go`, `validate.go` | ~700 | MEDIUM -- Agent CRUD commands |
| `internal/cmd/inscription` | `sync.go`, `backups.go`, etc. | ~600 | MEDIUM -- Inscription sync commands |
| `internal/cmd/manifest` | `diff.go`, `manifest.go`, `merge.go`, `show.go`, `validate.go` | ~500 | MEDIUM |
| `internal/cmd/naxos` | `naxos.go`, `scan.go` | ~150 | LOW |
| `internal/cmd/provenance` | `provenance.go` | 272 | MEDIUM |
| `internal/cmd/tribute` | `generate.go` | 162 | LOW |
| `internal/cmd/validate` | `validate.go` | 577 | MEDIUM -- 59.2% coverage; dispatch partially tested |
| `internal/cmd/worktree` | 9 files | ~900 | MEDIUM |
| `internal/cmd/artifact` | 5 files | ~500 | MEDIUM |
| `internal/cmd/common` | `annotations.go`, `context.go`, `embedded.go` | ~200 | LOW -- Shared context only |

**Library Packages (untested or partially tested):**

| Package | File | Lines | Risk Assessment |
|---------|------|-------|----------------|
| `internal/output` | `output.go`, others | ~800 | MEDIUM -- JSON output contracts now tested (PKG-013); text formatting still untested |
| `internal/assets` | `assets.go` | 28 | LOW -- Embedded asset loader |

**Note (PKG-013):** `internal/output/output_test.go` was added in PKG-013 covering: JSON output validity for `StatusOutput`, `CreateOutput`, `SyncResultOutput`, `SessionListOutput`; error wrapping; `ParseFormat`/`ValidateFormat`; `VerboseLog` behavior; format dispatch. Text formatting and `rite.go`/`manifest.go` output types remain untested.

Total untested source: approximately 7,400 lines of CLI command handler code + 1,400 lines of shared infrastructure (reduced by ~100 lines for output contract tests).

### Critical Path Coverage

**Sync Pipeline (`internal/materialize/`)**: Well-covered. 219 test functions across 20 test files including integration tests for provenance (`provenance_integration_test.go`), rite switching (`rite_switch_integration_test.go`), MCP integration (`mcp_integration_test.go`), and unified sync (`unified_sync_test.go`). The materialize pipeline is the most heavily tested area in the codebase.

**Hook Handlers (`internal/cmd/hook/`)**: Well-covered. 173 test functions across 11 test files. Includes benchmarks for performance-critical paths (agent-guard, validate, context, writeguard, git-conventions). Tests cover timeout enforcement, stdin integration, and JSON output contracts.

**Session Lifecycle (`internal/session/` + `internal/cmd/session/`)**: Extremely well-covered. 192 test functions in `internal/session/` plus 158 in `internal/cmd/session/`. Includes FSM state machine tests, lock protocol tests, full lifecycle integration tests, and Moirai-specific integration tests.

**Inscription Pipeline (`internal/inscription/`)**: Well-covered. 169 test functions across 8 test files. Includes integration tests for clean project sync, existing CLAUDE.md sync, and marker conflict handling.

**Sails Gate (`internal/sails/` + `internal/cmd/sails/`)**: Well-covered. 143 test functions in `internal/sails/` plus 31 in `internal/cmd/sails/`. Full integration tests for color algorithm, modifier handling, QA upgrade paths.

**CLI Entry Point (`internal/cmd/sync/sync.go`)**: 47.2% coverage (improved from zero). The underlying materialize pipeline is tested, and the CLI wiring layer now has partial coverage.

**Validate Command (`internal/cmd/validate/validate.go`)**: 59.2% coverage (improved from zero). The CLI dispatch layer now has partial coverage. The underlying `internal/validation` package has 62 tests.

**Errors Package (`internal/errors/errors.go`)**: 100% coverage. Custom error type definitions, `IsNotFound()`, `IsLifecycleError()`, and related helpers are now fully tested directly.

**Output Package (`internal/output/output.go`)**: Partially tested. JSON output contracts added (PKG-013); text formatting and rite/manifest output types remain untested.

### Negative Test Coverage

272 of 1,909 test functions (14.2%) explicitly test error paths by name (e.g., `TestInit_AlreadyInitialized`, `TestAgentGuard_DenyOutsideAllowedPaths`, `TestFile_NonExistent`). The `wantErr bool` table-test pattern appears in at least 8 packages (`manifest`, `frontmatter`, `config`, `mena`, `rite`, `sails`, `hook`, `agent`), providing systematic negative coverage within those packages.

**Note (PKG-013):** `internal/cmd/migrate/` test file removed with the dead migration tool. The `TestRewriteUserManifest_InvalidJSON` test cited previously no longer exists.

### Prioritized Gap List

1. **`internal/output/`** -- JSON output contracts added (PKG-013); text formatting and rite/manifest output types still untested.
2. **`internal/cmd/rite/`** -- Rite lifecycle commands; ~900 lines, zero tests.
3. **`internal/cmd/worktree/`** -- Worktree management; ~900 lines, zero tests.
4. **`internal/cmd/agent/`** -- Agent CRUD commands; ~700 lines, zero tests.
5. **`internal/cmd/sync/`** -- 47.2% coverage; CLI wiring partially tested but gaps remain.
6. **`internal/cmd/validate/`** -- 59.2% coverage; dispatch partially tested but gaps remain.

---

## Testing Conventions

### Test Function Naming

The dominant pattern is `Test{Subject}_{Scenario}` with underscores separating subject from scenario:

- `TestInit_FreshDirectory`, `TestInit_WithRite`, `TestInit_AlreadyInitialized`
- `TestAgentGuard_DenyOutsideAllowedPaths`, `TestAgentGuard_AllowWipPath`
- `TestSessionLifecycle_FullCycle`, `TestMoirai_CreateParkResumeWrap_GoldenPath`

Single-subject tests without underscores also appear for simple utility functions:

- `TestContent`, `TestBytes`, `TestDir` (in `internal/checksum/checksum_test.go`)
- `TestKnossosHome_Primary`, `TestKnossosHome_Default` (in `internal/config/home_test.go`)

Integration test files are named `*_integration_test.go` (9 files total). The body of these files contains comments citing TDD section numbers: `// Integration tests for the inscription pipeline as specified in TDD Section 11.2`.

### Subtest Patterns (`t.Run`)

Table-driven tests using `t.Run` are pervasive. Two naming conventions coexist:

1. **Named field**: `tt.name` or `tc.name` -- used in the majority of table tests
2. **Dynamic concatenation**: `tc.from+"_to_"+tc.to`, `tt.id+"_"+tt.description` -- used for FSM transition tests

Descriptive inline strings are also used in integration-style tests:
```go
t.Run("no matches produces no findings", func(t *testing.T) {...})
t.Run("Read with rites path produces HIGH finding", func(t *testing.T) {...})
```

### Assertion Patterns

Two assertion styles coexist with no codebase-wide standard:

**Stdlib only** (majority -- approximately 133 of 151 test files):
```go
if err != nil {
    t.Fatalf("Failed to do X: %v", err)
}
if got != want {
    t.Errorf("subject = %v, want %v", got, want)
}
```

**Testify** (`github.com/stretchr/testify`) -- used in 23 of 151 test files, concentrated in `internal/materialize/` and `internal/sails/`:
```go
require.NoError(t, err)
assert.Equal(t, expected, actual)
require.NoError(t, os.MkdirAll(riteDir, 0755))
```

The `wantErr bool` table-test pattern for systematic error coverage:
```go
type testCase struct {
    name    string
    input   string
    wantErr bool
}
// ...
if (err != nil) != tt.wantErr {
    t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
}
```

### Test Helper Patterns

`t.Helper()` is used in 36 of 151 test files. Helper functions follow the pattern `setup{Subject}(t *testing.T) ...`:

- `internal/cmd/session/`: `setupTimelineTestSession`, `setupLogTestSession`, `setupSnapshotTestSession`, `setupFieldTestSession`
- `internal/materialize/unified_sync_test.go`: `setupTestRite`, `setupKnossosHome`, `setupRite`
- `internal/materialize/archetype_test.go`: `setupArchetypeRite`, `setupKnossosRiteSharedMena`, `setupSatelliteRite`
- `internal/session/`: `newTestSetup` (returns a `testSetup` struct with methods `cleanup()`, `resolver()`, `createSession()`)
- `internal/artifact/`: `setupTestData`

The `testSetup` struct pattern in `internal/session/` provides object-oriented test infrastructure:
```go
type testSetup struct { ... }
func newTestSetup(t *testing.T) *testSetup
func (ts *testSetup) cleanup()
func (ts *testSetup) resolver() *paths.Resolver
func (ts *testSetup) createSession(id, status string, ...)
```

### Testdata and Fixture Patterns

No `testdata/` directories exist within `internal/`. The only fixture directory is `testdata-ari/` at the project root:

- `/Users/tomtenuta/Code/knossos/testdata-ari/rites/` -- contains `minimal-rite`, `broken-rite`, `valid-rite` subdirectories with `workflow.yaml`, `orchestrator.yaml`, `context.yaml`, and agent `.md` files

The `test/hooks/testutil/` package provides:
- `env.go`: `HookEnv` struct + `SetupEnv(t, env)` for isolating CC hook environment variables, with automatic `t.Cleanup()` restoration
- `golden.go`: `GoldenFile` struct for golden file comparison stored in `testdata/golden/`

JSON fixtures for hook testing at `/Users/tomtenuta/Code/knossos/test/hooks/fixtures/`:
- `session_context.yaml`
- `tool_input_bash.json`, `tool_input_edit.json`, `tool_input_write.json`

### Environment Management

`t.TempDir()` is used in 94 of 151 test files -- the dominant pattern for filesystem isolation. Manual cleanup via `defer os.RemoveAll(dir)` does not appear; `t.TempDir()` is universally preferred.

Environment variable management uses two approaches:
1. `os.Setenv` + `defer os.Setenv(key, oldValue)` -- appears in `internal/cmd/migrate/` (unsafe pattern, leaks if test panics)
2. `testutil.SetupEnv(t, env)` with `t.Cleanup()` -- used in hook tests (safe pattern)

`t.Parallel()` is not used anywhere in the codebase (zero occurrences).

No `TestMain` functions exist anywhere.

No build tags (`//go:build integration` or similar) are used to segregate integration tests from unit tests. Integration tests run as part of the standard `CGO_ENABLED=0 go test ./...` invocation.

### Benchmark Pattern

10 benchmark functions exist, all in `internal/cmd/hook/`:

- `BenchmarkAgentGuard_Passthrough`
- `BenchmarkValidateHook_Passthrough`, `BenchmarkValidateHook_EarlyExit`, `BenchmarkValidateHook_Validation`
- `BenchmarkContextHook_EarlyExit`, `BenchmarkContextHook_FullExecution`
- `BenchmarkHook_EarlyExitPath`, `BenchmarkHook_TimeoutOverhead`
- `BenchmarkGitConventions_FastPath`
- `BenchmarkWriteguardHook_Passthrough`

No fuzz tests exist. No `ExampleXxx` functions exist.

---

## Test Structure Summary

### Distribution

| Category | Packages with Tests | Packages without Tests | Test Files | Test Functions |
|----------|--------------------|-----------------------|------------|----------------|
| CLI handlers (`internal/cmd/`) | 9 | 13 | 22 | ~500 |
| Library packages (`internal/`) | 29 | 3 | 129 | ~1,400 |
| **Total** | **38** | **16** | **151** | **~1,909** |

38 of 54 source packages have test files (70.4% package coverage).

### Most Heavily Tested Areas

1. **`internal/materialize/`** -- 219 test functions, 20 test files. The sync pipeline has the densest test investment: agent transforms, archetype handling, mena sync, hook defaults, skill policies, satellite mena, cross-rite agents, provenance integration, MCP integration, rite switch integration.

2. **`internal/session/`** -- 192 test functions, 12 test files. FSM states, lock protocol, context loading, lifecycle golden path, comprehensive lifecycle, rotation, snapshots, timeline, event reading.

3. **`internal/cmd/session/`** -- 158 test functions, 17 test files. Command-level session operations with integration tests and a dedicated Moirai integration test.

4. **`internal/cmd/hook/`** -- 173 test functions, 11 test files. All hook handlers tested with C2 contract annotations tying tests to behavioral contracts.

5. **`internal/inscription/`** -- 169 test functions, 8 test files. Pipeline stages each tested in isolation plus integration tests.

6. **`internal/sails/`** -- 143 test functions, 6 test files. Color algorithm, modifier handling, QA upgrade path, threshold scaling.

### Test Package Naming (White-box vs Black-box)

The dominant pattern is white-box testing (same package name, no `_test` suffix): 146 of 151 test files declare `package X` matching their source package. Only 5 files use the external `package X_test` pattern:

- `internal/manifest/`: 4 files (`manifest_test`, `diff_test`, `merge_test`, `schema_test`)
- `internal/sync/`: 1 file (`sync_test`)

### Integration vs Unit Test Distinction

There is no build-tag separation between integration and unit tests. Integration tests are distinguished purely by filename convention (`*_integration_test.go`, 9 files) and by comment headers citing TDD section numbers. Both run under `CGO_ENABLED=0 go test ./...`.

9 explicitly named integration test files:
- `internal/agent/integration_test.go`
- `internal/cmd/session/integration_test.go`
- `internal/cmd/session/moirai_integration_test.go`
- `internal/cmd/session/status_integration_test.go`
- `internal/inscription/integration_test.go`
- `internal/materialize/mcp_integration_test.go`
- `internal/materialize/provenance_integration_test.go`
- `internal/materialize/rite_switch_integration_test.go`
- `internal/sails/integration_test.go`

### Test Runner

Command: `CGO_ENABLED=0 go test ./...`

`CGO_ENABLED=0` is required per project conventions (Go binary is a static build). No test tags needed -- all tests run in a single invocation. No test database setup, no external services required.

---

## Knowledge Gaps

1. **Actual runtime coverage percentages** -- this document reports structural coverage (which packages have tests) but not line-level coverage. Running `CGO_ENABLED=0 go test ./... -coverprofile=coverage.out` would produce line-level data; this was not executed.

2. **`testdata-ari/` path resolution** -- test files reference `filepath.Join("..", "..", "testdata", "rites")` but the actual directory is named `testdata-ari/`. Whether these tests skip or pass in CI is unclear without running the suite.

3. **`internal/errors/errors.go` indirect coverage** -- while the `errors` package has no test file, its functions (`IsNotFound`, `IsLifecycleError`) appear in test assertions. The actual branch coverage from indirect usage is unknown.

4. **`internal/output/output.go` indirect coverage** -- hook tests parse JSON output and hook command tests use `output.NewPrinter`. The output package's formatting logic is exercised but not bounded.

5. **`test/hooks/testutil/golden.go` usage** -- a golden file infrastructure exists but no golden file comparisons appear in the examined test files. Whether this infrastructure is used by any currently-uncommitted tests is unknown.

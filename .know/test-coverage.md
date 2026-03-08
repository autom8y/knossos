---
domain: test-coverage
generated_at: "2026-03-08T21:08:37Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "dbf81b8"
confidence: 0.90
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "602e86d06024078a2882f921cbc613ed4bb78196e981232db79d664f342225c8"
---

# Codebase Test Coverage

## Coverage Gaps

### Overall Distribution

The codebase has **197 test files** containing **2,977 test functions** across **54 packages**. Out of **65 packages** with Go source files, **11 packages lack any test files** (83% package coverage rate).

### Untested Packages and Criticality Assessment

All 11 untested packages are concentrated in the `internal/cmd/` layer — the Cobra CLI command wiring layer. These packages are thin command routers that delegate to tested domain packages.

| Package | Go Files | Criticality | Notes |
|---|---|---|---|
| `internal/cmd/inscription` | 6 | Medium | CLI for sync/rollback/diff of CLAUDE.md inscriptions. Core domain logic tested in `internal/inscription/` (which has dense coverage: backup, merger, marker, generator, pipeline, manifest, integration). |
| `internal/cmd/common` | 6 | Medium | Shared command context types (BaseContext, SessionContext, error printer). All consumers tested; this package itself is pure struct/type plumbing. |
| `internal/cmd/manifest` | 5 | Low | CLI for manifest diff/merge/show/validate. Domain logic tested in `internal/manifest/` (diff, merge, manifest, schema tests). |
| `internal/cmd/artifact` | 5 | Low | CLI for artifact register/list/rebuild/query. Domain logic tested in `internal/artifact/` (4 test files). |
| `internal/cmd/land` | 2 | Low | CLI for session land/synthesize — thin wrappers. |
| `internal/cmd/ledge` | 3 | Low | CLI for ledge promote/list. Domain logic tested in `internal/ledge/` (promote, auto_promote tests). |
| `internal/cmd/naxos` | 2 | Low | CLI for naxos scan. Domain logic tested in `internal/naxos/` (scanner, report tests). |
| `internal/cmd/tribute` | 2 | Low | CLI for tribute generate. Domain logic tested in `internal/tribute/` (generator, renderer, extractor tests). |
| `internal/cmd/provenance` | 1 | Low | Single CLI command wrapper. Domain tested in `internal/provenance/`. |
| `internal/cmd/root` | 1 | Very Low | Cobra root command setup only. |
| `internal/assets` | 1 | Very Low | Embed declaration only. |

### Untested Files Within Otherwise-Tested Packages

Several files inside well-tested packages lack corresponding test files:

**`internal/cmd/hook/`** — heavily tested overall (195 test functions) but 4 files have no tests:
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/cheapo_revert.go` (95 lines) — hook for reverting materialization changes
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/worktreeseed.go` (165 lines) — worktree seeding logic, moderate complexity
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/worktreeremove.go` (96 lines) — worktree removal hook
- `/Users/tomtenuta/Code/knossos/internal/cmd/hook/wiring.go` (28 lines) — shared materializer factory helper

**`internal/search/`** — 5 of 6 source files tested; `entry.go` has no corresponding test. This file is an index entry type.

**`internal/suggest/`** — `generator.go` tested but `suggestion.go` (53 lines) has no test. This file defines the `Suggestion` struct and `Kind` constants — pure type definitions, low risk.

### Critical Path Coverage

| Critical Path | Coverage Assessment |
|---|---|
| CLI command handlers | All `cmd/hook` handlers tested (13 of 17 files). `cmd/session` heavily tested (181 test functions). `cmd/sync`, `cmd/agent`, `cmd/rite`, `cmd/validate`, `cmd/worktree`, `cmd/explain`, `cmd/ask`, `cmd/knows`, `cmd/sails`, `cmd/lint`, `cmd/org`, `cmd/handoff`, `cmd/initialize`, `cmd/status`, `cmd/tour` all have tests. |
| Sync/materialization pipeline | Strongest coverage area. `internal/materialize/` has 42 test files and 339 test functions. All major sub-packages tested: `hooks/`, `mena/`, `userscope/`, `orgscope/`, `source/`. |
| Hook handlers | `internal/hook/` (169 test functions) and `internal/hook/clewcontract/` (all 9 sub-files tested with event, lifecycle, orchestrator, writer, record, typed_event, handoff, event_types, triggers). |
| Session FSM | `internal/session/` has 13 test files with 207 test functions. FSM transitions exhaustively tested in `lifecycle_comprehensive_test.go`. |
| Agent validation | `internal/agent/` has 8 test files: validate, frontmatter, sections, scaffold, regenerate, mcp_validate, integration, fuzz. |
| Inscription sync | `internal/inscription/` has 8 test files: backup, generator, integration, manifest, marker, merger, pipeline tests. |

### Test Blind Spots

1. **Worktree hook operations** (`cheapo_revert.go`, `worktreeseed.go`, `worktreeremove.go`) — these files contain materialization side effects (filesystem writes) that are not directly tested. The underlying materialize logic is tested, but the hook invocation paths are not.

2. **CLI command wiring** (`cmd/inscription`, `cmd/manifest`, `cmd/artifact`) — cobra command setup, flag parsing, error path dispatch are untested at the CLI layer. Integration failures here would surface only through manual testing.

3. **Suggest/suggestion.go type definitions** — pure type file, negligible risk.

4. **Search entry.go** — index entry type, low risk.

### Negative Test Coverage

Negative tests (error cases, invalid inputs) are present but modest. A count of test function names containing `Error/Fail/Invalid/Bad/Wrong/Missing/Empty/Nil/Negative` yields approximately 10 functions. Most error testing happens through table-driven test cases with error fields rather than named negative test functions.

### Coverage Measurement Infrastructure

No automated coverage reporting is configured. The CI workflow (`ariadne-tests.yml`) runs `CGO_ENABLED=0 go test -v ./...` and optionally `go test -race ./...` but does not generate coverage profiles or enforce coverage thresholds. No `go test -cover` or `-coverprofile` target exists in the Makefile or justfile.

### Prioritized Gap List

1. **Priority 1 (Medium risk)**: `internal/cmd/inscription/` — 6 files wrapping inscription sync/rollback/diff/validate. The CLI layer converts errors into exit codes; failures are invisible without tests.
2. **Priority 2 (Medium risk)**: `internal/cmd/hook/cheapo_revert.go`, `worktreeseed.go`, `worktreeremove.go` — hook command files with side effects in an otherwise well-tested package.
3. **Priority 3 (Low risk)**: `internal/cmd/common/` — error presentation and command context types used by every command.
4. **Priority 4 (Low risk)**: `internal/cmd/manifest/`, `internal/cmd/artifact/`, `internal/cmd/naxos/` — thin CLI wrappers.
5. **Not a priority**: `internal/assets/`, `internal/cmd/root/`, `internal/cmd/provenance/`, `internal/suggest/suggestion.go` — trivial files.

---

## Testing Conventions

### Test Function Naming

The dominant convention is `Test{Noun}_{Condition}` or `Test{Noun}_{Method}_{Condition}`. Examples observed:

- `TestAgentGuard_DenyOutsideAllowedPaths` — action + condition
- `TestFSM_AllTransitionPairs` — type + scenario
- `TestSCAR002_StagedMaterializeAbsent` — SCAR regression tag + assertion
- `TestAskOutputTextNoResults` — type + method + state
- `TestMaterializeSettingsWithManifest_NoMCPServers` — function + configuration

SCAR regression tests use the convention `TestSCAR{NNN}_{Description}` (14 tests across `internal/materialize/scar_regression_test.go` and `internal/materialize/source/source_test.go`). These are tests that enforce permanent absence of dangerous patterns.

### Subtest Patterns

`t.Run()` is used in 383 locations across 91 test files. The primary uses are:

1. **Table-driven tests**: iterate over `tests []struct{name, input, want}` slices
2. **FSM exhaustive testing**: nested loops generating subtests for every state pair
3. **Scenario grouping**: named scenarios within a single Test function

Standard table-driven pattern (from `internal/cmd/lint/lint_test.go`):
```
tests := []struct{ name string; input string; wantN int }{...}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

### Assertion Patterns

Two assertion libraries coexist:

- **Standard library** (`t.Errorf`, `t.Fatalf`): 193 of 197 test files use stdlib `testing`. This is the baseline for most tests.
- **testify** (`assert.*`, `require.*`): 29 files use `github.com/stretchr/testify`. Concentrated in `materialize/`, `cmd/ask/`, `cmd/tour/`, `cmd/sails/`, `cmd/explain/`, and `cmd/sync/`. New tests (recent commits) tend to use testify.

The split follows package-level consistency: once testify is used in a package, all tests in that package use it.

### Test Helper Patterns

- `t.Helper()` called in 90 locations, signaling well-established helper function discipline.
- Helper functions named `setup{Noun}`, `make{Noun}Ctx`, `run{Noun}Test` pattern (e.g., `setupTestRite`, `makeAgentGuardCtx`, `runAgentGuardTest`).
- Package `test/hooks/testutil` provides `HookEnv` and `SetupEnv` for hook tests.
- Package `test/worktree/testutil` provides worktree test utilities.
- Most helpers use `t.TempDir()` (913 occurrences) for filesystem isolation — no shared testdata directories inside `internal/`.

### Test Skip Patterns

`t.Skip` is used in 13 locations. Skip conditions observed:
- Git repository not present: `t.Skip("Skipping test: not in a git repository")`
- Testdata not found: `t.Skipf("testdata not found at %s", absPath)`
- Root user execution: `t.Skip("Skipping test when running as root")`
- Timestamp precision: `t.Skip("Test requires timestamp resolution...")`
- No agent files: `t.Skip("no agent files found")`

### Fixture Patterns

No `testdata/` directories exist inside `internal/`. Two fixture strategies:

1. **In-memory construction**: test helpers create structs, write to `t.TempDir()`, and pass paths. Dominant pattern in `materialize/`.
2. **External testutil packages**: `test/hooks/testutil/` for hook environment setup; `test/worktree/testutil/` for worktree state.
3. **Real rite data**: `testdata-ari/rites/` contains real rite fixtures consumed by integration tests in `internal/rite/`.

### Test Environment Management

- All tests are hermetic via `t.TempDir()` (auto-cleanup).
- Environment variables set in tests use `t.Setenv()` pattern (inferred from testutil usage).
- No `TestMain` exists anywhere in the codebase — no global test setup/teardown.
- `CGO_ENABLED=0` is required for all test runs (macOS dyld compatibility issue documented in justfile).

### Fuzz Tests

Three fuzz targets exist:
- `/Users/tomtenuta/Code/knossos/internal/frontmatter/fuzz_test.go`: `FuzzParse`
- `/Users/tomtenuta/Code/knossos/internal/agent/fuzz_test.go`: `FuzzParseAgentFrontmatter`
- `/Users/tomtenuta/Code/knossos/internal/know/fuzz_test.go`: `FuzzComputeFileDiff`

These are not run in CI (no `go test -fuzz` in workflows) — they are available for local fuzzing only.

### Package Declaration Style

The overwhelming majority (193 of 197 test files) use white-box testing with the same package name as the source file. Only 5 files use the `_test` external package suffix, all in `internal/manifest/`.

---

## Test Structure Summary

### Overall Distribution

| Area | Test Files | Test Functions | Notes |
|---|---|---|---|
| `internal/materialize/` | 42 | 339 | Dominant coverage area |
| `internal/cmd/session/` | 19 | 181 | Dense session command coverage |
| `internal/cmd/hook/` | ~18 | 195 | Hook command coverage |
| `internal/session/` | 13 | 207 | Session FSM, lifecycle, timeline |
| `internal/hook/` (core) | 11 | 169 | Hook input, clewcontract (9 sub-files) |
| `internal/inscription/` | 8 | ~150 | CLAUDE.md sync |
| `internal/agent/` | 8 | ~108 | Agent validation, frontmatter |
| `internal/sails/` | 7 | ~126 | Ship-readiness system |
| `internal/search/` | 5 | ~134 | Search index, collectors, score |
| `internal/rite/` | 6 | ~60 | Rite state, workflow, context |
| Others | ~60 | ~1,309 | Distributed across 30+ packages |
| **Total** | **197** | **2,977** | |

### Most Heavily Tested Areas

1. **`internal/materialize/`**: The historically hottest area (corroborating experiential knowledge from `.sos/land/workflow-patterns.md`). Contains dedicated scar regression tests, integration tests, and sub-package coverage across hooks, mena, userscope, orgscope, and source sub-packages. 339 test functions in 42 files.

2. **`internal/session/`** + **`internal/cmd/session/`**: Combined 388 test functions across 32 files. Session lifecycle is comprehensively tested with an exhaustive FSM test covering all state transition pairs (`lifecycle_comprehensive_test.go`).

3. **`internal/hook/` + `internal/cmd/hook/`**: 364 combined test functions. The clewcontract sub-package has 9 distinct test files covering event system, lifecycle, orchestrator, writer, record, typed events, handoff, event_types, and triggers.

### Integration vs Unit Tests

The codebase does not use build tags to separate integration from unit tests. Integration tests are identified by file name convention only:

- **Integration test files** (named `*integration*_test.go`): 9 files
  - `/Users/tomtenuta/Code/knossos/internal/agent/integration_test.go`
  - `/Users/tomtenuta/Code/knossos/internal/cmd/session/integration_test.go`
  - `/Users/tomtenuta/Code/knossos/internal/cmd/session/moirai_integration_test.go`
  - `/Users/tomtenuta/Code/knossos/internal/cmd/session/status_integration_test.go`
  - `/Users/tomtenuta/Code/knossos/internal/inscription/integration_test.go`
  - `/Users/tomtenuta/Code/knossos/internal/materialize/mcp_integration_test.go`
  - `/Users/tomtenuta/Code/knossos/internal/materialize/provenance_integration_test.go`
  - `/Users/tomtenuta/Code/knossos/internal/materialize/rite_switch_integration_test.go`
  - `/Users/tomtenuta/Code/knossos/internal/sails/integration_test.go`

All 9 integration tests run with the same `CGO_ENABLED=0 go test ./...` command — no build tags, no separate CI job.

### Test Package Naming

- 193/197 files use white-box (same-package) testing.
- 5/197 files use black-box (`_test` suffix) in `internal/manifest/`.

### Scar Regression Tests

A distinct test category specific to this codebase: **SCAR regression tests** enforce permanent absence of previously-dangerous patterns. 14 functions named `TestSCAR{NNN}_{Description}` exist, concentrated in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go` (12 functions) and `/Users/tomtenuta/Code/knossos/internal/materialize/source/source_test.go` (2 functions). One additional scar test in `/Users/tomtenuta/Code/knossos/internal/materialize/mena/content_rewrite_test.go`.

These tests use reflection (`reflect.TypeFor`) and filesystem inspection to assert structural invariants that must never regress.

### How Tests Are Run

- **Local**: `CGO_ENABLED=0 go test ./...` (justfile `test` recipe)
- **Verbose local**: `CGO_ENABLED=0 go test -v ./...`
- **CI**: Same as verbose local, plus optional race detector (`go test -race -v ./...`, `continue-on-error: true`)
- **No coverage measurement**: No `-cover` or `-coverprofile` flags in any workflow or justfile recipe.

### Mental Model for Writing New Tests

1. Place test in the same package (white-box) unless testing `internal/manifest/`-style sealed APIs.
2. Use `t.TempDir()` for all filesystem fixtures — no global testdata directories in `internal/`.
3. Use testify (`assert.*`/`require.*`) if the package already uses it; use stdlib `t.Errorf`/`t.Fatalf` otherwise.
4. Name tests `Test{Noun}_{Condition}` for unit cases, `TestSCAR{NNN}_{Description}` for regression guards.
5. Use `t.Run()` for table-driven tests and exhaustive state-pair testing.
6. Declare helpers with `t.Helper()` and name them `setup{Noun}`, `make{Noun}Ctx`, or `run{Noun}Test`.
7. Fuzz targets go in `*_fuzz_test.go` files — not run in CI.

---

## Knowledge Gaps

- **Actual line coverage percentage**: No `go test -cover` data collected during this audit. Line coverage per package is unknown.
- **Integration test scope**: The 9 integration tests are identified by naming convention only. No formal separation from unit tests exists; their actual external dependencies (real filesystem, live sessions) were not verified.
- **`suggest/suggestion.go` and `search/entry.go`**: Untested type-definition files — confirmed low-risk but not deeply read.
- **`cmd/common` behavior under error paths**: The shared command context error presentation logic is untested; behavior under malformed input is unverified.


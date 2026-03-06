---
domain: test-coverage
generated_at: "2026-03-06T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "3847e28"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "f6d34ca03d89c8f6e949f8b89f44463c9ce80aa2cb884aa1d31c8abbd3f054ba"
---

# Codebase Test Coverage

**Language**: Go 1.23+ (confirmed by `go.mod`)
**Test runner**: `CGO_ENABLED=0 go test ./...`
**Total test files**: 189
**Total test functions**: 2,817 (`Test*`) + 17 benchmarks + 3 fuzz targets
**Subtests (t.Run)**: 370

## Coverage Gaps

### Untested Packages (12 of 63 packages have no test files)

81% of packages (51/63) have test coverage. The 12 untested packages:

| Package | LOC (approx) | Nature | Priority |
|---------|-------------|--------|----------|
| `internal/cmd/inscription/` | 729 | Sync pipeline CLI wiring: sync, rollback, backups, diff, validate | **HIGH** |
| `internal/cmd/land/` | 398 | Land synthesize command (362-line synthesize.go) | **HIGH** |
| `internal/cmd/manifest/` | 569 | Manifest diff, merge, show, validate CLI | MEDIUM |
| `internal/cmd/provenance/` | 273 | Provenance CLI handler | MEDIUM |
| `internal/cmd/tribute/` | 219 | Tribute generate CLI | MEDIUM |
| `internal/cmd/naxos/` | ~80 | Naxos scan CLI wiring | LOW |
| `internal/cmd/ledge/` | ~100 | Ledge CLI wiring | LOW |
| `internal/cmd/artifact/` | ~160 | Artifact CLI wiring | LOW |
| `internal/cmd/root/` | 230 | Root cobra command (CLI wiring only) | LOW |
| `internal/cmd/common/` | 160 | Shared CLI utilities (annotations, context, embedded) | LOW |
| `internal/assets/` | 42 | Asset embed declarations | NONE |
| `cmd/ari/` | ~20 | main.go entry point | NONE |

### Files Without Tests Within Otherwise-Tested Packages

Within `internal/cmd/hook/` (otherwise well-tested at 14 test files):
- `internal/cmd/hook/cheapo_revert.go` — 99 lines, hook command with JSON output
- `internal/cmd/hook/worktreeremove.go` — 90 lines
- `internal/cmd/hook/worktreeseed.go` — 160 lines

Within `internal/inscription/` (otherwise tested):
- `internal/inscription/sync.go` — 154 lines, core `SyncCLAUDEmd` function (critical sync path)

### Prioritized Gap List

1. **`internal/inscription/sync.go`** (HIGH): `SyncCLAUDEmd` is the core CLAUDE.md sync operation. 154 lines with no direct test. The surrounding package has 7 test files but this file's function is only exercised through integration paths.
2. **`internal/cmd/inscription/`** (HIGH): 729-line sync pipeline CLI layer (sync, rollback, backups, diff, validate). Business logic is tested in `internal/inscription/` but CLI orchestration layer is not.
3. **`internal/cmd/land/synthesize.go`** (HIGH): 362 lines handling cross-session synthesis for multiple domains. No tests.
4. **`internal/cmd/hook/cheapo_revert.go`** (MEDIUM): Hook handler in a well-tested package, but this file's behavior is not covered.
5. **`internal/cmd/hook/worktreeseed.go`** and **`worktreeremove.go`** (MEDIUM): 250 combined lines of worktree hook handlers without tests.

### Error Path Coverage

- 101 of 189 test files (53.4%) test for error conditions (grep for `errors`, `wantErr`, `ErrInvalid`, `require.Error`, `assert.Error`).
- 15 test files use explicit `wantErr` table-driven error patterns.
- Negative tests are present but not uniform — lower-level packages have them; CLI command packages mostly lack them.

### Coverage Measurement Infrastructure

No CI coverage pipeline found. No `go test -coverprofile` invocations in CI config. Coverage is not instrumented or gated.

## Testing Conventions

### Test Function Naming

Primary convention: `Test{TypeOrSubject}_{Scenario}`.

Examples:
- `TestFSM_AllTransitionPairs` (`internal/session/fsm_test.go`)
- `TestMaterializeWorkflow_WritesFile` (`internal/materialize/workflow_test.go`)
- `TestNewContext_Defaults` (`internal/session/context_test.go`)

SCAR regression tests use a specific convention: `TestSCAR{NNN}_{Description}`:
- `TestSCAR002_StagedMaterializeAbsent`
- `TestSCAR004_CorruptProvenanceManifest_PropagatesError`
- `TestSCAR008_BudgetHook_MustNotBeAsync`
- `TestSCAR018_KnowDromenon_NoContextFork`
- `TestSCAR020_SessionDromena_ExplicitSessionIDPassing`
- `TestSCAR027_SharedMena_NoSessionArtifacts`

### Subtest Patterns (`t.Run`)

370 `t.Run(...)` calls are present. Table-driven tests are the dominant pattern: 255 table literal slice initializations (`tests := []struct{...}`) with 288 corresponding range loops.

### Assertion Patterns

The codebase uses a **hybrid** approach:
- **testify** (`github.com/stretchr/testify v1.11.1`) is the primary assertion library: 1,082 `assert.` / `require.` call occurrences
- **stdlib** `t.Errorf` / `t.Fatal` / `t.Error`: 8,403 occurrences — stdlib assertions are still heavily used

Both patterns coexist without a strict rule. `require.*` is used for preconditions that must halt the test; `assert.*` for non-fatal checks.

### Test Helper Patterns

- `t.Helper()` is used in 46 test files — proper helper declaration is common
- `t.TempDir()` is the dominant temporary directory pattern (124 test files), creating isolated filesystem state per test
- One dedicated helper file: `internal/materialize/userscope/helpers_test.go`
- No shared testutil package; helpers are local to each package's `_test.go` files

### Testdata Directories

No `testdata/` directories found at the package level. Fixture data is constructed inline within tests using `t.TempDir()` + explicit file writes. Some tests reference production rite directories directly.

### Test Environment Management

- `t.TempDir()` for filesystem isolation — auto-cleaned after test
- `CGO_ENABLED=0` is a required build constraint for macOS compatibility
- Integration tests create full `.sos/sessions/` directory trees in `t.TempDir()`

### Fuzz Tests

3 fuzz targets:
- `internal/frontmatter/fuzz_test.go`
- `internal/agent/fuzz_test.go`
- `internal/know/fuzz_test.go`

All are in parsing-critical packages.

### Integration vs Unit Tests

No build tags distinguish integration from unit tests. Integration tests are identified by file naming convention (`*_integration_test.go`):
- `internal/agent/integration_test.go`
- `internal/inscription/integration_test.go`
- `internal/materialize/mcp_integration_test.go`
- `internal/materialize/provenance_integration_test.go`
- `internal/materialize/rite_switch_integration_test.go`
- `internal/cmd/session/integration_test.go`
- `internal/cmd/session/moirai_integration_test.go`
- `internal/cmd/session/status_integration_test.go`
- `internal/sails/integration_test.go`

## Test Structure Summary

### Overall Distribution

| Package Area | Test Files | Notes |
|---|---|---|
| `internal/materialize/` (+sub) | 38 | Most heavily tested area |
| `internal/cmd/session/` | 19 | Session command tests, 3 integration files |
| `internal/cmd/hook/` | 14 | All major hook handlers tested; 4 files without tests |
| `internal/session/` | 13 | Deep FSM, lifecycle, rotation coverage |
| `internal/hook/clewcontract/` | 9 | Event contract tests |
| `internal/agent/` | 8 | Validation, scaffold, fuzz, integration |
| `internal/inscription/` | 7 | Pipeline tested; sync.go unguarded |
| `internal/sails/` | 7 | Includes integration test |
| `internal/rite/` | 6 | Budget, discovery, manifest, state, workflow |
| `internal/know/` | 6 | AST diff, discover, fuzz, manifest, validate |

### Most Heavily Tested Areas

1. **Materialization pipeline** — `internal/materialize/` and its subdirectories hold 38 test files covering archetypes, agent transforms, mena engine, hook defaults, unified sync, provenance integration, SCAR regressions, rite switches, and worktree behavior.

2. **Session management** — `internal/session/` (13 files) and `internal/cmd/session/` (19 files) together form 32 test files. Coverage includes FSM transitions, lifecycle state machines, event reading, snapshot, rotation, moirai integration, and archive boundary.

3. **Hook handlers** — `internal/cmd/hook/` has 14 test files covering agentguard, attributionguard, autopark, budget, clew, context, git conventions, precompact, session end, subagent, validate, and writeguard.

### Test Package Naming

- **White-box testing dominant**: 190 of 195 test files use `package {pkgname}` (same package as the code under test).
- **Black-box testing rare**: Only 5 files use `package {pkgname}_test`: 4 in `internal/manifest/` and 1 in `internal/sync/`.

### SCAR Regression Test Pattern

A dedicated SCAR naming convention: `TestSCAR{NNN}_{Description}`. 10 SCAR-named functions across:
- `internal/materialize/scar_regression_test.go` (SCARs 002, 004, 021, 023, 027)
- `internal/cmd/hook/budget_test.go` (SCAR 008)
- `internal/cmd/status/status_test.go` (SCAR 015, 016)
- `internal/cmd/knows/knows_test.go` (SCAR 018)
- `internal/cmd/session/session_test.go` (SCAR 020)

### How Tests Are Run

```
CGO_ENABLED=0 go test ./...
```

`CGO_ENABLED=0` is mandatory on macOS — omitting it causes test binary aborts. No special flags or build constraints required.

## Knowledge Gaps

1. **Actual line coverage percentages are unknown** — `go test -cover` was not run; the above analysis is structural (file/package presence), not instrumented coverage.
2. **`internal/inscription/sync.go` test exposure is uncertain** — may be exercised indirectly through integration tests, but direct unit test coverage is not documented.
3. **`internal/cmd/land/synthesize.go` domain logic** — 362 lines with no direct tests.
4. **E2E test suite** — a `Makefile` with `e2e-linux` target and `Dockerfile.e2e` exists, but e2e test contents were not evaluated.

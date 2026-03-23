---
domain: test-coverage
generated_at: "2026-03-23T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "78abb186"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "9b8f28035904e1dcd19d584717ac3629753bc808309497ce41f27180caabe7ac"
---

# Codebase Test Coverage

## Coverage Gaps

### Package-Level Coverage

76 packages contain non-test Go source files. 61 of those have at least one test file. **15 packages have no test files (80.3% package coverage).**

**Untested packages and their criticality:**

| Package | Files / Lines | Criticality Assessment |
|---|---|---|
| `cmd/ari` | `main.go` (entry point only) | Low — single-line `main()` that delegates to root cmd |
| `internal/assets` | `assets.go` | Low — likely embedded file access only |
| `internal/cmd/root` | `root.go` (255 lines) | Medium — Cobra root command wiring; substantial CLI setup logic |
| `internal/cmd/common` | 6 files, ~258 lines | Medium — shared args, context, and error helpers used across all cmd subpackages |
| `internal/cmd/artifact` | 5 files, ~595 lines | Medium — `artifact list/query/rebuild/register` CLI handlers |
| `internal/cmd/inscription` | 6 files, ~730 lines | Medium — `inscription sync/rollback/diff/validate` CLI handlers (core write pathway) |
| `internal/cmd/land` | `land.go` + `synthesize.go` (399 lines) | Medium — `synthesize.go` is 362 lines of session land synthesis logic |
| `internal/cmd/ledge` | 3 files, ~204 lines | Low-Medium — ledge promote/list handlers |
| `internal/cmd/manifest` | 5 files, ~561 lines | Medium — manifest diff/merge/show/validate CLI handlers |
| `internal/cmd/naxos` | 3 files, ~230 lines | Low — scan/triage wrappers; core logic in `internal/naxos` (tested) |
| `internal/cmd/provenance` | `provenance.go` (274 lines) | Medium — provenance query CLI handler |
| `internal/cmd/tribute` | 2 files, ~220 lines | Low — tribute generate wrapper; core in `internal/tribute` (tested) |
| `test/hooks/testutil` | `env.go`, `golden.go` | Not production code — test infrastructure |
| `test/worktree/testutil` | `worktree.go` | Not production code — test infrastructure |

**The two `test/*/testutil` packages are test helpers, not production code.** Excluding them, 13 production packages lack tests.

### Critical Path Coverage

**CLI command handlers:** Heavily tested. 22 of 35 `internal/cmd/` sub-packages have dedicated test files. The untested CLI sub-packages are mostly thin Cobra wrappers that delegate to tested library packages (e.g., `cmd/naxos` → `internal/naxos` has 4 test files; `cmd/tribute` → `internal/tribute` has 3 test files). Exceptions: `cmd/inscription/synthesize.go` (362 lines) and `cmd/common` (6 shared helper files) represent meaningful untested logic.

**Sync pipeline:** Well covered. `internal/materialize` has the most concentrated test investment in the codebase — 54 test files covering unified sync, channel routing, MCP, mena transforms, procession rendering, and userscope operations.

**Hook handlers:** Comprehensively covered. All 14 hook command handlers in `internal/cmd/hook/` have dedicated test files (agentguard, attributionguard, autopark, budget, clew, context, driftdetect, gitconventions, hook, precompact, sessionend, subagent, suggest, validate, writeguard).

### Test Blind Spots

1. **`os.Setenv` without cleanup (parallelization blocker):** ~13 test files use `os.Setenv` without `t.Setenv` or deferred restore. These tests cannot safely run in parallel, blocking `-parallel` flag usage.

2. **No coverage measurement infrastructure:** No `.coverprofile` files, no `codecov.yml`, no `-coverprofile` flags in CI (`ariadne-tests.yml` runs `CGO_ENABLED=0 go test -v ./...` without coverage flags). There is no way to see numeric coverage percentages.

3. **Golden file infrastructure unused:** `test/hooks/testutil/golden.go` defines a full golden file framework (GoldenFile type, UPDATE_GOLDEN env var, JSON/string comparison), but zero test files import or use it currently.

4. **Negative tests and error path coverage:** 108 `wantErr`/`errExpected` patterns found across 17 files; 164 files assert `err != nil`. Error path coverage is present but not uniform — lighter packages like `cmd/common`, `cmd/root`, and `cmd/inscription` have no tests at all, so their error paths are uncovered.

### Prioritized Gap List

1. **HIGH: `internal/cmd/common`** — Shared CLI helpers (`args.go`, `context.go`, `errors.go`) used by all cmd handlers. A bug here affects every command.
2. **HIGH: `internal/cmd/inscription`** — 730 lines of inscription sync/rollback/validate logic with zero tests; this is a core write pathway.
3. **MEDIUM: `internal/cmd/root`** — 255 lines of Cobra wiring. CLI startup failures cannot be caught.
4. **MEDIUM: `internal/cmd/land/synthesize.go`** — 362 lines of session synthesis logic without coverage.
5. **MEDIUM: `internal/cmd/manifest`** — 561 lines across 5 files; manifest is a frequently-used command group.
6. **LOW: `os.Setenv` cleanup** — ~13 test files use `os.Setenv` without cleanup. Replace with `t.Setenv` to enable parallel execution.
7. **LOW: Golden file adoption** — The infrastructure exists in `test/hooks/testutil/golden.go` but is unused.

---

## Testing Conventions

### Test Function Naming

All test functions follow the Go convention: `TestFunctionName` (or `TestTypeName_MethodName`). 3,299 `func Test` declarations found across 223 files. No deviations observed from standard naming.

### Subtest Patterns

446 `t.Run(` calls found across 103 files. Subtests are widely used for scenario grouping. Dominant patterns:

- **Table-driven subtests:** 114 files use `tests := []struct{...}` or equivalent, iterating with `for _, tt := range tests { t.Run(...) }`. Example: `internal/cmd/lint/lint_preferential_test.go`.
- **Named scenario subtests:** Single-file scenario grouping with `t.Run("description", func(t *testing.T) {...})`. Prevalent in hook tests and session tests.

### Assertion Patterns

The codebase uses **stdlib `testing` only** for assertions — no testify or gomock. All 226 test files import `"testing"`. Only 32 files use `require.` or `assert.` (from testify), and inspection shows these are used in a minority of tests. The dominant assertion style is:

```go
if got != want {
    t.Errorf("description: got %v, want %v", got, want)
}
```

For fatal conditions: `t.Fatal`, `t.Fatalf` (found across 195 files with 9,350 occurrences total).

### Test Helper Patterns

116 `t.Helper()` calls found across 58 files. Helper functions are defined locally within test files (e.g., `filterByRule` in `internal/cmd/lint/lint_preferential_test.go`). No shared test helper packages are imported by production test files — `test/hooks/testutil` and `test/worktree/testutil` exist but are not currently consumed.

### Skip Patterns

11 `t.Skip(` calls across 8 files. Skip reasons include:
- Environment-dependent tests (e.g., `internal/agent/integration_test.go`: skips when testdata is unavailable)
- Platform or tooling dependency checks
- `internal/cmd/worktree/worktree_test.go`: explicit skip with note

### Testdata Directories

Only one `testdata/` directory found: `internal/cmd/complaint/testdata/`. It contains 3 YAML complaint fixture files:
- `COMPLAINT-20260310-180045-integration-engineer.yaml`
- `COMPLAINT-20260311-091500-pythia.yaml`
- `COMPLAINT-20260311-143022-drift-detect.yaml`

No other `testdata/` directories exist. Most tests create ephemeral fixtures inline using `t.TempDir()`.

### Test Fixture Patterns

1. **`t.TempDir()` (primary pattern):** 142 files use `t.TempDir()` with 1,020 total calls. Tests write fixtures programmatically into temp directories. This is the dominant approach.
2. **Inline string/byte fixtures:** Tests embed YAML, JSON, and markdown strings directly in test code.
3. **Golden files:** Infrastructure defined (`test/hooks/testutil/golden.go`, UPDATE_GOLDEN env var pattern), but unused in any test file.
4. **YAML complaint fixtures:** Only in `internal/cmd/complaint/testdata/`.

### Test Environment Management

- `t.TempDir()`: 142 files — idiomatic, auto-cleaned
- `t.Setenv()`: 6 files — correct pattern, cleaned up automatically
- `os.Setenv()` without cleanup: ~13 files — problematic for parallel execution (confirmed by session memory: "t.Setenv friction: 2 sessions targeted t.Setenv elimination")
- No `TestMain` functions found — no global setup/teardown

### Fuzz Tests

3 fuzz test files:
- `internal/know/fuzz_test.go` — `FuzzComputeFileDiff`
- `internal/frontmatter/fuzz_test.go` — `FuzzParse`
- `internal/agent/fuzz_test.go` — `FuzzParseAgentFrontmatter`

All target parsing/diff functions — appropriate targets for fuzzing (parser correctness and crash safety).

---

## Test Structure Summary

### Overall Distribution

| Metric | Count |
|---|---|
| Total test files | 226 |
| Total `func Test` declarations | 3,299 |
| Packages with test files | 61 of 76 (80.3%) |
| Packages without test files | 15 (19.7%) |
| Production packages without tests | 13 |
| Test helper packages without tests | 2 |

### Most Heavily Tested Areas

By file count:

1. **`internal/materialize/`** — 54 test files (largest concentration). Covers the materialization pipeline comprehensively: sync, routing, MCP integration, mena transforms, hook defaults, procession, userscope, worktree.
2. **`internal/cmd/session/`** — 23 test files. Covers session lifecycle commands: create, wrap, fray, gc, lock, log, query, recover, snapshot, status, suggest_next, timeline, moirai integration, archive boundary.
3. **`internal/cmd/hook/`** — 16 test files. Covers every hook handler.
4. **`internal/session/`** — 14 test files. Core session domain logic.
5. **`internal/hook/clewcontract/`** — 11 test files. Hook event contract layer.
6. **`internal/inscription/`** — 8 test files (library layer tested; cmd layer untested).

### Test Package Naming

**Dominant pattern: `package <name>` (same package, white-box testing).** 217 of 226 test files use white-box package declarations. Only 9 files use the `package <name>_test` external test pattern:

- `internal/manifest/diff_test.go`
- `internal/manifest/merge_test.go`
- `internal/manifest/manifest_test.go`
- `internal/manifest/schema_test.go`
- `internal/paths/channel_test.go`
- `internal/materialize/compiler/compiler_test.go`
- `internal/materialize/channel_routing_test.go`
- `internal/channel/tools_test.go`
- `internal/sync/state_test.go`

### Integration vs Unit Test Files

There is no formal separation. Files named `*_integration_test.go` (12 files) are mixed with unit test files in the same directories:

- `internal/agent/integration_test.go`
- `internal/inscription/integration_test.go`
- `internal/cmd/session/integration_test.go`
- `internal/cmd/session/status_integration_test.go`
- `internal/cmd/session/moirai_integration_test.go`
- `internal/materialize/rite_switch_integration_test.go`
- `internal/materialize/mcp_integration_test.go`
- `internal/materialize/provenance_integration_test.go`
- `internal/sails/integration_test.go`

No build tags separate them from unit tests — all run under `CGO_ENABLED=0 go test ./...`.

### TestMain Patterns

No `TestMain` functions found anywhere in the codebase. There is no global test setup or teardown.

### Standard Test Command

```
CGO_ENABLED=0 go test ./...
```

This is the canonical invocation (confirmed in `.github/workflows/ariadne-tests.yml`, referenced in `CLAUDE.md` and MEMORY.md). The CI workflow additionally attempts a race detector run:

```
go test -race -v ./... || echo "Race detector tests skipped (CGO required)"
```

The race detector run is non-blocking (allowed to fail), reflecting that some packages require CGO.

---

## Knowledge Gaps

1. **Numeric coverage percentages:** No coverage measurement infrastructure (no `-coverprofile` in CI). This document cannot state "package X is 72% covered" — only structural coverage (file/package presence) is observable.
2. **Error path saturation in tested packages:** While many packages have tests, whether every error branch is exercised cannot be determined without instrumented coverage reports.
3. **`internal/cmd/inscription/synthesize.go` internals:** This 362-line file has no tests and synthesizes session data from land files. Its internal logic paths are unknown without reading it.
4. **Race conditions in `os.Setenv` tests:** The full list of tests affected by `os.Setenv` without cleanup may be larger than the 13 files identified — tests that call helpers which internally call `os.Setenv` would not appear in this search.
5. **`testutil` adoption timeline:** It is unclear whether `test/hooks/testutil/golden.go` was pre-built for future use or is residual from an abandoned migration.

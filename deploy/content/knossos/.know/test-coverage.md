---
domain: test-coverage
generated_at: "2026-03-27T19:57:42Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "5501b0aa"
confidence: 0.85
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "1f2c9d187ac50eb67dffec49dddc3dd9217e4be0cb56e62cda1bd1d52dd7c00f"
---

# Codebase Test Coverage

> Generated: 2026-03-27 | Language: Go | Test runner: `CGO_ENABLED=0 go test ./...`

## Coverage Gaps

### Overall Distribution

The codebase contains approximately 108 source packages and 87 packages with test files, yielding **80.6% package-level coverage**. There are 290 test files containing approximately 4,029 test functions.

### Untested Packages

**Low criticality (thin wiring, minimal logic):**

| Package | Assessment |
|---|---|
| `cmd/ari` | Binary entrypoint only -- thin wiring, not directly testable |
| `internal/assets` | Embed registry with setters -- no logic to test |
| `internal/cmd/root` | Cobra wiring only, delegates to tested sub-packages |
| `internal/cmd/common` | Shared context structs and annotations -- low business logic |
| `internal/cmd/tribute` | Thin CLI wrapper -- delegates to `internal/tribute` which IS tested |
| `internal/cmd/land` | Thin CLI wrapper |
| `internal/cmd/naxos` | Thin CLI wrapper -- delegates to `internal/naxos` which IS tested |
| `internal/cmd/manifest` | Thin CLI wrappers |
| `internal/cmd/ledge` | Thin CLI wrappers |
| `internal/cmd/inscription` | Thin CLI wrappers |

**Medium criticality (business logic present):**

| Package | Assessment |
|---|---|
| `internal/cmd/artifact` | Registry artifact commands -- no CLI-layer tests |
| `internal/cmd/provenance` | Substantial CLI orchestration (274 lines); domain package IS tested |
| `internal/concept` | Concept registry parser with embedding -- no unit tests for parse logic |

### Critical Path Assessment

| Critical Path | Coverage Status | Notes |
|---|---|---|
| Materialization pipeline | Strong -- 31 test files in `internal/materialize` | Most tested area |
| Hook handlers (`internal/cmd/hook`) | Good -- 16 of 21 source files tested | `call.go`, `cheapo_revert.go`, `worktreeseed.go` untested |
| Session lifecycle | Good -- `internal/session` 14 test files; `internal/cmd/session` 20 test files | Includes integration tests |
| Agent management | Good -- `internal/agent` fully covered (8 test files) | |
| Search / Clew | Good -- `internal/search` and all sub-packages tested | |
| Serve / HTTP layer | Partial -- middleware and webhook tested; `server.go`, `config.go` untested | Clew growth area |
| Inscription (CLAUDE.md merger) | Good -- generator, marker, merger, pipeline, backup all tested | |
| Sails (quality gates) | Strong -- full suite of contract, gate, generator, proofs, thresholds tests | |

### Blind Spots

1. **Error paths in CLI commands**: Most `internal/cmd/*` thin wrappers have no tests.
2. **`internal/concept`**: Parser with no unit tests despite non-trivial logic.
3. **Hook signing wrapper** (`call.go`): HMAC signing wrapper untested.
4. **Slack Clew packages**: `config.go`, `streaming.go`, `summarizer.go`, `citations.go` -- active development, limited tests.
5. **Integration tests not isolated**: No build tags; run alongside unit tests.

### Prioritized Gap List (Highest Risk First)

1. **`internal/cmd/hook/cheapo_revert.go` + `worktreeseed.go`**: Hook logic (106+172 lines) with no dedicated tests
2. **`internal/concept/concept.go`**: Concept registry parsing -- used by `ari explain`; parse failures are silent
3. **`internal/cmd/provenance` (274 lines)**: Provenance command orchestration -- substantial logic, no CLI tests
4. **Error-path coverage broadly**: Only ~16% of test files include explicit negative-path testing
5. **`cmd/ari/main.go`**: No binary-level smoke test

---

## Testing Conventions

### Test Function Naming

Dominant pattern: `TestSubject_BehaviorOrScenario(t *testing.T)`:

- `TestWriteIfChanged_SkipsIdentical`
- `TestCreate_BasicCreation`
- `TestSCAR002_StagedMaterializeAbsent` -- named after scar ticket

### Subtest Patterns

`t.Run` used in 137 of 290 test files (47%). Table-driven tests in 147 files (51%):

```go
tests := []struct {
    name  string
    input ...
    want  ...
}{...}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

### Assertion Patterns

Two styles coexist:
- **Primary: testify** -- `require` in 56 files (fatal), `assert` in 65 files (non-fatal)
- **Secondary: stdlib** -- `t.Fatalf` / `t.Errorf` in older tests

Convention: `require.*` for setup/precondition assertions, `assert.*` for behavioral claims.

### Test Helper Patterns

- `t.Helper()` not widely used; helper behavior achieved via standalone packages
- `test/hooks/testutil/env.go` -- `SetupEnv(t, env)` pipes hook JSON to stdin
- `test/hooks/testutil/golden.go` -- `GoldenFile` type for output comparison
- `test/worktree/testutil/worktree.go` -- `SetupWorktreeTestFixture(t)` creates real git repository

### Test Skip Patterns

`t.Skip` / `t.Skipf` in 13 test files:
- Environment guards: not in a git repository
- Platform guards: macOS root chmod 000 behavior
- Testdata availability checks
- Environment variable gates (`AGENT_STRICT`)

No `go:build` build tags. No `TestMain` in the main codebase.

### Test Data and Fixture Patterns

**`testdata/` directories:** Only 2:
- `internal/cmd/complaint/testdata/` -- YAML complaint fixtures
- `internal/reason/testdata/` -- Go source fixture for parsing tests

**Golden files:** Infrastructure exists (`test/hooks/testutil/golden.go`) but zero `.golden` files in the project.

**Dominant pattern:** In-test fixture construction with `t.TempDir()` (154 of 290 files) and inline content.

### Test Environment Management

- `t.TempDir()` -- 154 of 290 test files (53%) for filesystem isolation
- `t.Parallel()` -- 84 of 290 test files (29%) for parallelism
- `os.Setenv` / deferred restore for environment variables

### Fuzz Tests

Three fuzz test files: `internal/agent/fuzz_test.go`, `internal/know/fuzz_test.go`, `internal/frontmatter/fuzz_test.go`. Target parser-critical paths.

### Error Path Testing

`wantErr`/`expectErr` pattern in 47 of 290 files (~16%). Error-path coverage is lighter than happy-path coverage.

---

## Test Structure Summary

### Distribution

- **108 source packages**, **87 with test files** -- 80.6% package coverage
- **290 test files**, **~4,029 test functions**
- **13 packages with no tests** -- primarily CLI thin wrappers and Clew/Slack development area

### Most Heavily Tested Areas

1. **`internal/materialize`** -- 31 test files (workflow, agent defaults, userscope, mena, hooks, scar regressions)
2. **`internal/cmd/session`** -- 20 test files (create, list, lock, log, query, fray, gc, status, integration)
3. **`internal/cmd/hook`** -- 16 test files (one per major hook implementation)
4. **`internal/session`** -- 14 test files (lifecycle, complexity, state management)
5. **`internal/hook/clewcontract`** -- 10 test files (events, channels, handlers, lifecycle)
6. **`internal/agent`** -- 8 test files (scaffold, sections, MCP, regenerate, integration)
7. **`internal/inscription`** -- 8 test files (integration, equivalence, backup, merger, marker)
8. **`internal/know`** -- 7 test files (AST diff, discover, manifest, validate)
9. **`internal/sails`** -- 7 test files (color, contract, gate, generator, integration, proofs, thresholds)

### Package Naming Patterns

- **Internal tests** (`package foo`) -- 280 of 290 files, dominant convention
- **External tests** (`package foo_test`) -- 10 files, concentrated in `internal/manifest/`

### Integration vs Unit Tests

No formal separation. Integration tests identified by filename:
- `internal/agent/integration_test.go`
- `internal/cmd/session/integration_test.go`
- `internal/cmd/session/status_integration_test.go`
- `internal/inscription/integration_test.go`
- `internal/materialize/mcp_integration_test.go`
- `internal/materialize/provenance_integration_test.go`
- `internal/materialize/rite_switch_integration_test.go`
- `internal/reason/integration_test.go`
- `internal/sails/integration_test.go`
- `internal/serve/health/integration_test.go`

All run with `CGO_ENABLED=0 go test ./...`. No coverage flags in CI.

### Test Parallelism

`t.Parallel()` in 84 files (997 call sites). Heavy use in materialize and hook tests. Session and search tests generally skip `t.Parallel()`.

---

## Knowledge Gaps

1. **Actual line coverage percentages**: No `go test -cover` in CI. Line-level coverage unknown.
2. **`internal/cmd/hook/cheapo_revert.go`**: File exists without tests; criticality unclear.
3. **`t.Parallel` completeness in session package**: Whether omission is by design (shared state) or oversight undetermined.
4. **Error path coverage within tested packages**: `wantErr` signal gives partial visibility; line-level error branch coverage unknown.
5. **Search sub-packages test depth**: All have tests but individual depth not verified.

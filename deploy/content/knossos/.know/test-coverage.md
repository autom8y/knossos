---
domain: test-coverage
generated_at: "2026-03-26T17:14:25Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "a73d68a6"
confidence: 0.88
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/workflow-patterns.md"
land_hash: "1f2c9d187ac50eb67dffec49dddc3dd9217e4be0cb56e62cda1bd1d52dd7c00f"
---

# Codebase Test Coverage

> Generated: 2026-03-26 | Language: Go | Test runner: `CGO_ENABLED=0 go test ./...`

## Coverage Gaps

### Overall Distribution

The codebase contains approximately 102 source packages and 86 packages with test files, yielding **84.3% package-level coverage**. There are 280 test files containing approximately 3,860 test functions.

### Untested Packages

**Internal cmd packages (CLI thin wrappers):**
- `internal/cmd/artifact`, `internal/cmd/common`, `internal/cmd/inscription`, `internal/cmd/land`, `internal/cmd/ledge`, `internal/cmd/manifest`, `internal/cmd/naxos`, `internal/cmd/provenance`, `internal/cmd/root`, `internal/cmd/tribute`

**Domain packages:**
- `internal/assets` — embedded file assets
- `internal/concept` — concept registry with Levenshtein fuzzy matching (`LookupConcept`, `parseConcept`, `AllConcepts`) — **medium criticality**: used by `cmd/explain` and `internal/search`

### Critical Path Assessment

| Critical Path | Coverage Status | Notes |
|---|---|---|
| Materialization pipeline | Strong — 31 test files in `internal/materialize` | Most tested area |
| Hook handlers (`internal/cmd/hook`) | Good — 14 of 18 source files tested | `call.go` (HMAC signing wrapper) untested |
| Session lifecycle | Good — `internal/session` 12 of 13 tested; `internal/cmd/session` 15 of 20 tested | `park.go`, `resume.go`, `audit.go` untested |
| Agent management | Good — `internal/agent` fully covered (6 test files) | |
| Search / Clew | Good — `internal/search` and all sub-packages tested | |
| Serve / HTTP layer | Partial — `middleware.go` and `webhook/verify.go` tested; `server.go`, `config.go`, `challenge.go` untested | Clew growth area |
| Inscription (CLAUDE.md merger) | Good — generator, marker, merger, pipeline, manifest, backup all tested | |
| Sails (quality gates) | Strong — full suite of contract, gate, generator, proofs, thresholds tests | |

### Blind Spots

1. **Error paths in CLI commands**: `internal/cmd/artifact`, `internal/cmd/land`, `internal/cmd/naxos`, `internal/cmd/tribute` have no tests.
2. **`internal/concept`**: `LookupConcept`, `parseConcept`, and fuzzy suggestion have no tests despite non-trivial logic.
3. **Hook signing wrapper** (`call.go`): HMAC signing wrapper untested — a bug would silently bypass signature validation.
4. **Harness adapters** (`adapter_claude.go`, `adapter_gemini.go`): Multi-harness translation layer untested.
5. **Slack Clew packages**: `config.go`, `streaming.go`, `summarizer.go`, `citations.go` — active development, none tested.
6. **Session park/resume**: `park.go` and `resume.go` in `internal/cmd/session` are lifecycle-critical and untested.

### Prioritized Gap List (Highest Risk First)

1. **`internal/cmd/hook/call.go`** — HMAC signing wrapper; a bug bypasses security entirely
2. **`internal/concept/concept.go`** — non-trivial fuzzy matching logic, no tests
3. **`internal/serve/server.go`** and **`serve/webhook/challenge.go`** — Clew HTTP entrypoint
4. **`internal/cmd/session/park.go`**, **`resume.go`** — session lifecycle operations
5. **Harness adapters** (`adapter_claude.go`, `adapter_gemini.go`) — multi-harness correctness
6. **`internal/slack/streaming/citations.go`**, **`conversation/summarizer.go`** — active development

---

## Testing Conventions

### Test Function Naming

Dominant pattern: `TestFunctionName_Scenario` using underscores:
```
TestArchetypeDefaults_IncludeCCNativeFields
TestNewSessionWrap_Success
TestMerge_Conflict
```

### Subtest Patterns

`t.Run` used in 128 files (504 call sites). Standard table-driven testing:
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
- **Primary: stdlib** (`t.Errorf`, `t.Fatalf`) — 10,032 call sites across 229 files
- **Secondary: `require`/`assert` packages** — 2,608 call sites across 59 files (search, reason, trust, Slack domains)

### Test Helper Patterns

- `t.Helper()` in 63 files (126 call sites)
- Naming: `newTestValidator(t *testing.T)`, `setupTestProject(t)`

### Test Skip Patterns

`t.Skip` in 12 files. No build tags (`//go:build integration`). Integration tests distinguished by filename suffix (`_integration_test.go`) or `t.Skip` guards.

### Test Environment Management

- `t.TempDir()` in 153 files (1,090 call sites) — dominant filesystem isolation pattern
- No manual `os.MkdirTemp` + `defer os.RemoveAll` as primary approach

### Testdata Directories

Only 2 testdata directories:
- `internal/cmd/complaint/testdata/` — YAML complaint fixtures
- `internal/reason/testdata/` — Go source fixture for AST parsing tests

No golden file patterns. Fixtures are inline-constructed in tests.

### Fuzz Tests

Three fuzz test files: `internal/agent/fuzz_test.go`, `internal/know/fuzz_test.go`, `internal/frontmatter/fuzz_test.go`. Target parser-critical paths.

### Error Path Testing

`wantErr`/`expectErr` pattern in 21 files. Error-path coverage is lighter than happy-path coverage.

---

## Test Structure Summary

### Distribution

- **102 source packages**, **86 with test files** — 84.3% package coverage
- **280 test files**, **~3,860 test functions**
- **16 packages with no tests** — primarily CLI thin wrappers and Clew/Slack development area

### Most Heavily Tested Areas

1. **`internal/materialize`** — 26 source files, 31+ test files (workflow, agent defaults, userscope, mena, hooks, routing, satellite, rite-switching, provenance)
2. **`internal/cmd/session`** — 20 source files, 15 tested (wrap, claim, create, fray, gc, lock, log, query, recover, status, timeline)
3. **`internal/session`** — 12 of 13 source files tested (FSM, events, lifecycle, resolve, rotation, snapshot)
4. **`internal/cmd/hook`** — 14 of 18 source files tested
5. **`internal/search`** and all sub-packages — BM25, fusion, content, knowledge stores
6. **`internal/hook/clewcontract`** — 11 test files (events, channels, handlers, lifecycle, orchestrator)
7. **`internal/sails`** — complete coverage (contract, gate, generator, proofs, thresholds, color)

### Package Naming Patterns

- **Internal tests** (`package foo`) — 250 files, dominant convention
- **External tests** (`package foo_test`) — 10 files, used selectively for black-box behavioral tests

### Integration vs Unit Tests

No build tags for separation. Integration tests distinguished by:
1. **Filename suffix** — `*_integration_test.go` (e.g., `internal/agent/integration_test.go`, `internal/serve/health/integration_test.go`)
2. **`t.Skip` guards** — 12 files with environment/service requirements

All tests run under `CGO_ENABLED=0 go test ./...`. No `--tags=integration` separation.

### Test Parallelism

`t.Parallel()` in 83 files (997 call sites). Heavy use in materialize and hook tests. Not universally applied — session and search tests generally skip `t.Parallel()`.

---

## Knowledge Gaps

1. **Actual line coverage percentages**: No `go test -cover` output available. Line-level coverage unknown.
2. **`internal/cmd/hook/cheapo_revert.go`**: File exists without tests; criticality unclear.
3. **`t.Parallel` completeness in session package**: Whether omission is by design (shared state) or oversight is undetermined.
4. **Error path coverage within tested packages**: `wantErr` signal gives partial visibility; line-level error branch coverage unknown without instrumentation.

---
domain: test-coverage
generated_at: "2026-03-13T10:04:06Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "59a0de2"
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

### Completely Untested Packages

**cmd sub-packages with zero test files (10 packages, 34 source files):**

| Package | Source Files | Notes |
|---------|-------------|-------|
| `internal/cmd/artifact/` | 5 | CLI command wrappers: `artifact.go`, `list.go`, `query_cmd.go`, `rebuild.go`, `register.go` |
| `internal/cmd/common/` | 6 | Shared CLI utilities: `annotations.go`, `args.go`, `context.go`, `embedded.go`, `errors.go`, `group.go` |
| `internal/cmd/inscription/` | 6 | Inscription commands: `backups.go`, `diff.go`, `inscription.go`, `rollback.go`, `sync.go`, `validate.go` |
| `internal/cmd/land/` | 2 | `land.go`, `synthesize.go` |
| `internal/cmd/ledge/` | 3 | `ledge.go`, `list.go`, `promote.go` |
| `internal/cmd/manifest/` | 5 | `diff.go`, `manifest.go`, `merge.go`, `show.go`, `validate.go` |
| `internal/cmd/naxos/` | 3 | `naxos.go`, `scan.go`, `triage.go` |
| `internal/cmd/provenance/` | 1 | `provenance.go` |
| `internal/cmd/root/` | 1 | `root.go` (CLI entry wiring) |
| `internal/cmd/tribute/` | 2 | `generate.go`, `tribute.go` |

**Other packages with zero test files (1 package):**

| Package | Source Files | Notes |
|---------|-------------|-------|
| `internal/assets/` | 1 | `assets.go` — embedded asset loading |

### Partially Tested Critical Packages

**`internal/materialize/` (hottest path per workflow-patterns: 87 changes, 5 sessions):**

- 23 source files in the root package, but 16 lack direct `*_test.go` counterparts. The untested source files include: `collision.go`, `frontmatter.go`, `hooks.go`, `materialize.go`, `materialize_agents.go`, `materialize_gitignore.go`, `materialize_inscription.go`, `materialize_mena.go`, `materialize_rules.go`, `materialize_settings.go`, `mena.go`, `org_scope.go`, `source.go`, `sync_types.go`, `syncer.go`, `user_scope.go`.
- Mitigation: 29 integration-style test files cover behavior across these — `unified_sync_test.go`, `workflow_test.go`, `write_test.go`, `rite_switch_integration_test.go`, `provenance_integration_test.go`, etc.

**`internal/hook/` (8 source files):**

- `adapter_claude.go`, `adapter_gemini.go`, and `output.go` have no test files. Platform-specific adapters are not tested.

**`internal/cmd/hook/` (20 source files, 16 test files):**

- Untested: `call.go`, `cheapo_revert.go`, `wiring.go`, `worktreeremove.go`, `worktreeseed.go`

**`internal/rite/` (10 source files, 6 test files):**

- `invoker.go`, `syncer.go`, `validate.go`, `context.go` have no direct test counterparts.

**`internal/perspective/` (6 source files, 1 test file):**

- Only `perspective_test.go` exists. `assemble.go`, `audit.go`, `context.go`, `simulate.go`, and `types.go` are uncovered.

**`internal/worktree/` (6 source files, 2 test files):**

- `git.go`, `metadata.go`, `session_integration.go`, `worktree.go` have no direct test counterparts.

### Test Blind Spots

1. **CLI command handler logic** (`cmd/artifact`, `cmd/inscription`, `cmd/manifest`, `cmd/land`, `cmd/ledge`, `cmd/naxos`, `cmd/tribute`) — complete absence of tests.
2. **Hook adapter platform differences** — `adapter_claude.go` and `adapter_gemini.go` are untested.
3. **`internal/cmd/common/`** — shared CLI utilities used by almost every command handler have no tests; wide blast radius.
4. **`internal/perspective/`** — 5 of 6 source files untested.

### Negative Test Coverage

- 126 occurrences of `want.*error`, `wantErr`, `expectErr`, `shouldErr` patterns across test files.
- 3,163 `err != nil` / `errors.Is` checks in test files.
- Table-driven negative cases present in 116 files using `[]struct` patterns.

### Coverage Measurement Infrastructure

- No `coverage.out` or `.coverprofile` files in the repository.
- No `go test -cover` or `-coverprofile` flags in `justfile` or `.github/workflows/ariadne-tests.yml`.
- Coverage percentage is unknown and not tracked.

### Prioritized Gap List

1. **HIGH** — `internal/cmd/common/` (6 files, no tests): shared utilities, wide blast radius.
2. **HIGH** — `internal/hook/adapter_claude.go`, `adapter_gemini.go`: platform-specific hook dispatch paths.
3. **MEDIUM** — `internal/cmd/inscription/` (6 files, no tests): inscription is a core materialization surface.
4. **MEDIUM** — `internal/materialize/syncer.go`, `materialize.go`: entry points covered indirectly but no unit isolation.
5. **MEDIUM** — `internal/perspective/` (5 of 6 files uncovered).
6. **LOW** — `internal/cmd/artifact/`, `cmd/naxos/`, `cmd/tribute/`, `cmd/manifest/`: less critical CLI wrappers.
7. **LOW** — `internal/assets/assets.go`: single-file embedded asset loader.

---

## Testing Conventions

### Test Function Naming

The dominant pattern is `Test{Subject}_{Scenario}`:

```
TestAutoPromoteSession_AllPromotable
TestUnifiedSync_CollisionDetection
TestMaterializeWorkflow_WritesFile
TestWriteIfChanged_SkipsIdentical
TestValidateCmd_ShortDescription
```

Both underscore-separated (~80%) and PascalCase-only (~20%) variants appear.

### Subtest Patterns (`t.Run`)

446 `t.Run(` calls across 103 files — widespread adoption. Heavy use in table-driven tests:

```go
for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```

### Assertion Patterns

Two distinct patterns coexist:

- **Standard library** (`t.Errorf`, `t.Fatalf`): 193 of 223 test files. Baseline for most tests.
- **testify** (`require.*`/`assert.*`): 30 files (1,592 occurrences). Concentrated in `internal/materialize/`, `internal/cmd/`, `internal/search/`, `internal/sails/`.

Package-level consistency: once testify is used in a package, all tests in that package use it.

### Test Helper Patterns

- `t.Helper()` used in 140 files — consistent and broad adoption.
- Helper functions named `setup{Noun}`, `make{Noun}Ctx`, `run{Noun}Test`.
- `test/hooks/testutil/golden.go` — golden file utility with `UPDATE_GOLDEN=1` env var support.
- `test/hooks/testutil/env.go` — hook test environment helpers.
- Most helpers use `t.TempDir()` (913+ occurrences) for filesystem isolation.

### Skip Patterns

`t.Skip` in 11 locations across 8 files. Skip conditions: git repo not present, testdata not found, root user execution, timestamp precision, no agent files.

No build-tag-gated integration tests (no `//go:build integration` patterns).

### Fixture Patterns

- **Temp dirs**: `t.TempDir()` used in 140 test files — standard for filesystem isolation.
- **testdata directory**: `internal/cmd/complaint/testdata/` contains YAML complaint fixtures.
- **Golden files**: `test/hooks/fixtures/` with testutil golden infrastructure.
- **Embedded rites**: `testdata-ari/rites/` contains `broken-rite/`, `minimal-rite/`, `valid-rite/`.

### Test Environment Management

- No `TestMain` functions anywhere — no global setup/teardown.
- `CGO_ENABLED=0` required for all test runs (macOS dyld compatibility issue).
- Race detector optionally run via `go test -race -v ./...` in CI (`continue-on-error: true`).

### Fuzz Tests

Three fuzz targets:
- `internal/frontmatter/fuzz_test.go`: `FuzzParse`
- `internal/agent/fuzz_test.go`: `FuzzParseAgentFrontmatter`
- `internal/know/fuzz_test.go`: `FuzzComputeFileDiff`

Not run in CI — available for local fuzzing only.

### SCAR Regression Tests

14+ functions named `TestSCAR{NNN}_{Description}` concentrated in `internal/materialize/scar_regression_test.go` and `internal/materialize/source/source_test.go`. These use reflection (`reflect.TypeFor`) and filesystem inspection to assert structural invariants that must never regress.

---

## Test Structure Summary

### Overall Distribution

- **223 test files** across the `internal/` package tree.
- **335 source files** in `internal/` (excluding test files).
- **Test-to-source ratio**: 0.67 (223/335).
- **3,249 total test functions** (`func Test*`) across 220 files.
- **446 subtest calls** (`t.Run`) across 103 files.
- **3 fuzz test functions**.

### Most Heavily Tested Areas

| Package | Test Files | Test Functions | Notes |
|---------|-----------|----------------|-------|
| `internal/cmd/session/` | 20 | 300+ | Session lifecycle most tested cmd area |
| `internal/session/` | 14 | 215+ | Core session model — near-complete coverage |
| `internal/materialize/` | 29 | 200+ | Integration-style, behavior-over-unit |
| `internal/cmd/hook/` | 16 | 180+ | Hook handlers heavily tested |
| `internal/inscription/` | 7 | 180+ | Inscription pipeline well covered |
| `internal/hook/clewcontract/` | 10 | 170+ | Clew contract event system |
| `internal/sails/` | 7 | 150+ | Sails health-check subsystem |
| `internal/search/` | 5 | 139+ | Search index and scoring |
| `internal/know/` | 6 | 169+ | Knowledge management |
| `internal/cmd/rite/` | 1 | 100 | Single test file but 100 test functions |

### Test Package Naming Patterns

White-box testing (same package name) dominates: 214 of 223 test files. Only `internal/manifest/` uses external test packages (`package manifest_test` — 4 files).

### Integration Tests vs Unit Tests

No build-tag separation. "Integration" signaled by file naming convention only:
- 9 integration-style files: `rite_switch_integration_test.go`, `mcp_integration_test.go`, `provenance_integration_test.go`, `sails/integration_test.go`, `cmd/session/integration_test.go`, `cmd/session/status_integration_test.go`, `cmd/session/moirai_integration_test.go`, `agent/integration_test.go`, `inscription/integration_test.go`

All run with same `CGO_ENABLED=0 go test ./...` command.

### How Tests Are Run

```bash
# Standard (required: CGO disabled)
CGO_ENABLED=0 go test ./...

# Verbose
CGO_ENABLED=0 go test -v ./...

# Specific package
CGO_ENABLED=0 go test -v ./internal/sails/...

# CI race detector (optional, best-effort)
go test -race -v ./...
```

CI triggers on changes to `cmd/**`, `internal/**`, `test/**`, `go.mod`, `go.sum`.

### Mental Model for Writing New Tests

1. Place test in the same package (white-box) unless testing `internal/manifest/`-style sealed APIs.
2. Use `t.TempDir()` for all filesystem fixtures.
3. Use testify if the package already uses it; use stdlib otherwise.
4. Name tests `Test{Noun}_{Condition}` for unit cases, `TestSCAR{NNN}_{Description}` for regression guards.
5. Use `t.Run()` for table-driven tests.
6. Declare helpers with `t.Helper()` and name them `setup{Noun}`, `make{Noun}Ctx`, or `run{Noun}Test`.
7. Fuzz targets go in `*_fuzz_test.go` files — not run in CI.

---

## Knowledge Gaps

- **Actual line coverage percentage**: No `go test -cover` data collected. Line coverage per package is unknown.
- **Integration test scope**: The 9 integration tests are identified by naming convention only. Their actual external dependencies were not verified.
- **`cmd/common` behavior under error paths**: Shared command context error presentation logic is untested.
- **Test execution time**: No timing data observed; slow tests are unidentified.

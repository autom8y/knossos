---
domain: scar-tissue
generated_at: "2026-03-26T17:14:25Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "a73d68a6"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/scar-tissue.md"
land_hash: "91df3a763b10021ddd9c7dd6dafb17e4bbc463258383c169d4a7a9067985bc51"
---

# Codebase Scar Tissue

## Failure Catalog Completeness

**Catalog total**: 37 entries — 33 numbered SCAR entries (SCAR-001 through SCAR-033), 3 unnumbered GO entries (GO-001/002/003), and 1 unnumbered writeguard path traversal entry. Sources: `internal/materialize/scar_regression_test.go`, `internal/materialize/mena/content_rewrite_test.go`, `internal/materialize/source/source_test.go`, `internal/agent/regenerate_test.go`, git history, `.sos/land/scar-tissue.md`.

### SCAR-001: TOCTOU Race in Stale Lock Reclamation
**Category**: race_condition | **Fix**: `internal/lock/lock.go:117-134` — atomic flock-on-existing-fd | **Tests**: `TestManager_StaleLockReclamation`, `TestManager_StaleLockReclamation_Concurrent`

### SCAR-002: StagedMaterialize Freeze in Claude Code
**Category**: integration_failure | **Fix**: Method removed; per-file atomic `writeIfChanged()` in `internal/fileutil/fileutil.go`. All `.claude/` touches carry `HA-FS` annotation | **Tests**: `TestSCAR002_StagedMaterializeAbsent`, `TestSCAR002_MaterializeWithOptions_NoClaudeRename`

### SCAR-003: Idempotency Guard Bypass on Mutation Flags
**Category**: idempotency_failure | **Fix**: `internal/materialize/materialize.go:194` — mutation flags imply `--force` | **Tests**: `TestRematerializeMena_RepopulatesAfterWipe`

### SCAR-004: Silent Error Discard at Provenance Load Sites
**Category**: data_corruption | **Fix**: `internal/materialize/materialize.go` load sites (~242-258, ~388-403) — WARN log on non-file-not-found; parse errors abort | **Tests**: `TestSCAR004_CorruptProvenanceManifest_PropagatesError`, `TestSCAR004_InvalidSchemaProvenanceManifest_PropagatesError`

### SCAR-005: Destructive os.RemoveAll on Agents/Commands/Skills
**Category**: data_corruption | **Fix**: `internal/materialize/materialize_agents.go:30` — selective write from managed-set | **Tests**: 6 tests in `internal/materialize/selective_write_test.go`

### SCAR-006: Shared Mena Drop for Satellite-Local Rites
**Category**: integration_failure | **Fix**: `internal/materialize/materialize_mena.go:85-99` — `sharedRitesBase` from `KnossosHome()` | **Tests**: 4 tests in `internal/materialize/satellite_mena_test.go`

### SCAR-007: Mixed Dro/Lego Directories Block Skill Resolution
**Category**: schema_evolution | **Fix**: Mena directories split into separate dirs. `ari lint` enforces separation at `internal/cmd/lint/lint.go:32-33` | **Tests**: Lint enforcement

### SCAR-008: Budget Hook Async Log Spam
**Category**: performance_cliff | **Fix**: Sub-100ms hooks must be synchronous in `config/hooks.yaml` | **Tests**: `TestSCAR008_BudgetHook_MustNotBeAsync`

### SCAR-009: Hooks Generated in Wrong Nested-Matcher Format
**Category**: integration_failure | **Fix**: `internal/materialize/hooks.go` — CC nested-matcher format. Guard: `internal/hook/output.go:27` | **Tests**: `internal/materialize/hooks_test.go`

### SCAR-010: Missing Timeout on Hook and Git Subprocess Commands
**Category**: performance_cliff | **Fix**: `internal/cmd/hook/budget.go:65` (timeout), `internal/worktree/git.go:17-21` (30s CommandContext) | **Tests**: `internal/cmd/hook/hook_test.go`

### SCAR-011: Session Writeguard Used .current-session
**Category**: configuration_drift | **Fix**: `internal/cmd/hook/writeguard.go` — priority chain resolution | **Tests**: `TestWriteguard_ParkedSession_*`

### SCAR-012: Archived Session Writeguard Gap
**Category**: integration_failure | **Fix**: `internal/cmd/hook/writeguard.go:201-204` — `isSessionArchived()` check | **Tests**: `TestWriteguard_ArchivedSession_DeniesWithClearMessage`

### SCAR-013: Ghost Dirs and Already-Archived Session Wrap
**Category**: integration_failure | **Fix**: `internal/cmd/session/wrap.go` — guards for all three edge cases | **Tests**: `TestWrapAlreadyArchived`, `TestWrapNoGhostDirectory`

### SCAR-014: Phantom Status Values in Session FSM
**Category**: schema_evolution | **Fix**: `internal/session/status.go:44-51` — `NormalizeStatus()` applied at every read site | **Tests**: `TestParseContext_NormalizesPhantomStatus`, `TestParseContext_NormalizesComplete`

### SCAR-015: stdout Pollution from Shell Log Functions
**Category**: data_corruption | **Fix**: All shell log functions redirect to stderr with `>&2` | **Tests**: `TestSCAR015_ShellScripts_StderrLogging`

### SCAR-016: Bash Arithmetic Increment with set -e
**Category**: schema_evolution | **Fix**: `((var++)) || true` suffix in affected `.sh` files | **Tests**: `TestSCAR016_ShellScripts_NoUnprotectedArithmeticIncrement`

### SCAR-017: @skill-name Anti-Pattern in Agent Prompts
**Category**: configuration_drift | **Fix**: `internal/cmd/lint/lint.go:909-937` — lint rule `skill-at-syntax` (HIGH severity) | **Tests**: 14 tests in `internal/cmd/lint/lint_test.go`

### SCAR-018: context: fork Blocks Task Tool Access
**Category**: integration_failure | **Fix**: `mena/know/INDEX.dro.md` (absence of `context: fork`). Lint: `internal/cmd/lint/lint.go:668-674` | **Tests**: Lint tests

### SCAR-019: Invalid Agent Colors in CC Palette
**Category**: configuration_drift | **Fix**: `internal/cmd/lint/lint.go:381-385` — 8-value palette validation | **Tests**: Lint tests

### SCAR-020: Session ID Not Passed to CLI Subprocesses
**Category**: integration_failure | **Fix**: Explicit `-s` flag instruction in session dromena | **Tests**: `TestSCAR020_SessionDromena_ExplicitSessionIDPassing`

### SCAR-021: Cross-Rite Agents Materialized to Project Scope
**Category**: integration_failure | **Fix**: `materializeCrossRiteAgents()` removed from project scope. User-scope only. | **Tests**: `TestSCAR021_CrossRiteAgents_ProjectScopeExclusion`

### SCAR-022: Provenance Schema Rejects Abbreviated SHA256
**Category**: data_corruption | **Fix**: Test fixtures updated to full 64-char sha256 with `sha256:` prefix

### SCAR-023: Template Path Resolution Fails for Self-Hosting
**Category**: configuration_drift | **Fix**: `internal/materialize/source/resolver.go:273-280` — knossos-home-relative resolution | **Tests**: `TestSCAR023_*`

### SCAR-024: Stale Throughline IDs After Rite Switch
**Category**: integration_failure | **Fix**: `internal/materialize/materialize.go:791` — `cleanupThroughlineIDs()` | **Tests**: 5 tests in `internal/materialize/throughline_cleanup_test.go`

### SCAR-025: Deleted Files in User Scope Sync
**Category**: data_corruption | **Fix**: `internal/materialize/userscope/sync.go` — `os.Stat` missing-file handling | **Tests**: `internal/materialize/userscope/sync_test.go`

### SCAR-026: Revert — Moirai Delegation in /start and /sprint
**Category**: historical_boundary | **Fix**: `additionalContext` removed from writeguard output. Pure allow/deny.

### SCAR-027: Ephemeral Skill in Shared Mena
**Category**: historical_boundary | **Fix**: `internal/cmd/lint/lint.go:847-896` — lint rule `session-artifact-in-shared-mena` | **Tests**: `TestSCAR027_SharedMena_NoSessionArtifacts`

### SCAR-028: MCP Servers Written to Wrong CC Config File
**Category**: integration_failure | **Fix**: `internal/materialize/materialize_settings.go:22,40,72` — guard comment + `materializeMcpJson()` to `.mcp.json` | **Tests**: `TestSCAR028_MCPServers_NotInSettingsLocalJson`

### SCAR-029 (Candidate): Moirai Template Uses ${SESSION_ID}
**Category**: configuration_drift | Status: No formal SCAR number. No regression test.

### SCAR-030: Go 1.23 Lacks t.Chdir — Use DI
**Category**: testing | **Fix**: Add `workDir string` field (dependency injection). 38 remaining `t.Setenv` usages.

### SCAR-031: Materialize Tests Are I/O-Bound
**Category**: performance | Root cause: `t.TempDir()` I/O dominates. ~30% parallelism expected, not a bug.

### SCAR-032: TestInit_WithRite Global State
**Category**: global_state | **Fix**: `internal/cmd/initialize/init_test.go:73` — comment "Not parallel: mutates global embedded assets"

### SCAR-033: Config Canonicalization Timing
**Category**: integration_failure | **Fix**: Operational pattern: build + install binary BEFORE committing config changes.

### Unnumbered: Writeguard Path Traversal Check
**Category**: integration_failure | **Fix**: `internal/cmd/hook/writeguard.go` — removal of overly broad `filepath.IsAbs` check.

### GO-001/002/003: Three Content Rewriting Bypass Paths
**Category**: data_corruption | **Fix**: `internal/materialize/mena/engine.go:153`, `mena/walker.go:86`, `userscope/sync_mena.go:323,378`, `mena/content_rewrite.go:50` | **Tests**: `TestSCAR_ContentRewriteNotBypassed`

---

## Category Coverage

| Category | Count | SCAR IDs |
|----------|-------|----------|
| integration_failure | 13 | 002, 006, 009, 012, 013, 018, 020, 021, 024, 028, 033, Writeguard, Sync Orphan |
| data_corruption | 7 | 004, 005, 015, 022, 025, GO-001/002/003 |
| configuration_drift | 6 | 008, 011, 017, 019, 023, 029 |
| schema_evolution | 3 | 007, 014, 016 |
| performance_cliff | 2 | 008, 010 |
| historical_boundary | 2 | 026, 027 |
| testing_constraints | 2 | 030, 032 |
| race_condition | 1 | 001 |
| idempotency_failure | 1 | 003 |
| performance_measurement | 1 | 031 |

**Categories searched but not found**: memory leak, SQL/database error, authentication failure, network timeout, infinite loop.

---

## Fix-Location Mapping

All primary fix locations verified against HEAD (`a73d68a6`). Key files spot-checked: `internal/lock/lock.go`, `internal/materialize/materialize.go`, `internal/cmd/hook/writeguard.go`, `internal/materialize/materialize_settings.go`, `internal/hook/output.go`, `internal/cmd/lint/lint.go`, `internal/session/status.go` — all confirmed to exist. No broken links found.

Compound fixes fully mapped: SCAR-009, SCAR-010, SCAR-012, GO-001/002/003 (4+ files each).

---

## Defensive Pattern Documentation

| SCAR | Defensive Pattern | Guard Type | Regression Test |
|------|------------------|-----------|-----------------|
| 001 | Atomic flock-on-existing-fd | Behavioral | `TestManager_StaleLockReclamation*` |
| 002 | Never rename `.claude/`; writeIfChanged; HA-FS markers | Behavioral + Structural | `TestSCAR002_*` |
| 003 | Mutation flags imply Force | Behavioral | `TestRematerializeMena_*` |
| 004 | Never `_`-discard provenance errors | Behavioral | `TestSCAR004_*` |
| 005 | Selective write from managed-set | Behavioral | `selective_write_test.go` |
| 006 | `sharedRitesBase` from KnossosHome() | Behavioral | `satellite_mena_test.go` |
| 007 | `ari lint` dro/lego separation | Lint | Lint enforcement |
| 008 | No async on sub-100ms hooks | Structural | `TestSCAR008_*` |
| 009 | Nested-matcher wire format; comment guard | Behavioral + Comment | `hooks_test.go` |
| 010 | withTimeout() + CommandContext | Behavioral | `hook_test.go` |
| 011 | Session resolution via priority chain | Behavioral | `TestWriteguard_ParkedSession_*` |
| 012 | `isSessionArchived()` check | Behavioral | `TestWriteguard_ArchivedSession_*` |
| 014 | `NormalizeStatus()` at all read sites | Behavioral | `TestParseContext_Normalizes*` |
| 017 | skill-at-syntax lint rule (HIGH) | Lint | Lint tests + `regenerate_test.go` |
| 018 | No `context: fork` on Task-needing dromena | Structural + Lint | Lint tests |
| 020 | Explicit -s flag in session dromena | Structural | `TestSCAR020_*` |
| 021 | Cross-rite agents user-scope only | Behavioral | `TestSCAR021_*` |
| 027 | `session-artifact-in-shared-mena` lint | Lint | `TestSCAR027_*` |
| 028 | `materializeMcpJson()` to `.mcp.json` | Behavioral + Comment | `TestSCAR028_*` |
| GO-001/002/003 | `RewriteMenaContentPaths` in ALL copy paths | Behavioral | `TestSCAR_ContentRewriteNotBypassed` |

---

## Agent-Relevance Tagging

| SCAR | Agent Role(s) | Why |
|------|--------------|-----|
| 001 | integration-engineer | Lock code must use atomic flock — Remove+reopen is TOCTOU |
| 002 | integration-engineer, context-architect | Never rename `.claude/`. Use writeIfChanged. HA-FS annotation required |
| 003 | integration-engineer | Mutation flags must imply Force |
| 004 | integration-engineer | Never `_`-discard provenance errors |
| 005 | integration-engineer, context-architect | RemoveAll destroys user content — use selective write |
| 006 | integration-engineer, compatibility-tester | Shared mena must come from KnossosHome() |
| 007 | context-architect | Mixed dro/lego blocks skill resolution |
| 008 | context-architect | Sub-100ms hooks must be synchronous |
| 009 | integration-engineer | CC hook format is nested-matcher — flat format silently rejected |
| 010 | integration-engineer | ALL hooks need timeout; ALL git needs CommandContext |
| 011 | integration-engineer | `.current-session` deprecated — use priority chain |
| 012 | integration-engineer, context-architect | Archived sessions need distinct denial — generic denial causes retry loops |
| 014 | integration-engineer | All status reads through NormalizeStatus() |
| 015 | integration-engineer | Shell log functions must redirect to stderr |
| 016 | platform-wide | Bash `((var++))` returns 1 under `set -e` |
| 017 | context-architect | `@skill-name` not a CC primitive |
| 018 | context-architect | `context: fork` removes Task tool access |
| 019 | platform-wide | Use CC's 8-value color palette |
| 020 | integration-engineer | Always pass session ID via `-s` flag |
| 021 | integration-engineer, context-architect | Cross-rite agents user-scope only |
| 022 | platform-wide | Full 64-char SHA256 with `sha256:` prefix |
| 028 | integration-engineer, context-architect | MCP to `.mcp.json` — CC ignores `settings.local.json` for MCP |
| 030 | platform-wide (test authors) | Go 1.23 lacks `t.Chdir()` — use DI workDir |
| 032 | platform-wide (test authors) | `common.SetEmbeddedAssets()` mutates global state |
| 033 | integration-engineer | Build binary before committing config changes |
| GO-001/002/003 | integration-engineer | RewriteMenaContentPaths in ALL copy paths |

---

## Knowledge Gaps

1. **GO-001/002/003 unnumbered**: Not formally assigned SCAR numbers.
2. **SCAR-029 candidate unconfirmed**: No SCAR number, no regression test.
3. **Writeguard path traversal unnumbered**: Commit `5dd1901` — no SCAR entry.
4. **Binary PATH mismatch trap**: `go build ./cmd/ari` writes `./ari` but CC uses PATH binary. No SCAR number.
5. **GITHUB_TOKEN/GoReleaser CI scar**: In land synthesis but no SCAR number.
6. **Hook stdin-only transport**: Critical constraint — no SCAR number, no regression test.
7. **SCAR-015 shell coverage**: Regression test scans 3 dirs only.
8. **Sync orphan bug (summon: prefix)**: Commit `4268b74b` — no SCAR assigned.
9. **Clew nil-deref and goroutine leak**: Commit `fd516aeb` — no SCAR assigned.
10. **12 entries lack defensive pattern table rows**: SCAR-013, 016, 019, 022-026, 029-031, 033.

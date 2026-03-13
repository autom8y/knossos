---
domain: scar-tissue
generated_at: "2026-03-13T10:04:06Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "59a0de2"
confidence: 0.92
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/scar-tissue.md"
land_hash: "c5cdd76ae3802cff6d1fc703b8d5fbe2f8f09253542e7b3972775badb05797fa"
---

# Codebase Scar Tissue

## Failure Catalog

33 SCAR entries cataloged. Sources: `internal/materialize/scar_regression_test.go`, `internal/materialize/source/source_test.go`, `internal/agent/regenerate_test.go`, `internal/materialize/mena/content_rewrite_test.go`, `internal/materialize/mcp_integration_test.go`, git history through commit `59a0de2`, and `.sos/land/scar-tissue.md`.

### SCAR-001: TOCTOU Race in Stale Lock Reclamation

**Category**: race_condition
**What failed**: Stale lock reclamation used `os.Remove(lockfile)` then `os.OpenFile(lockfile)`. A competing process could acquire the lock between Remove and reopen.
**Fix location**: `internal/lock/lock.go:117-134`
**Defensive pattern**: `syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)` on the already-open fd; never Remove+reopen.
**Regression tests**: `TestManager_StaleLockReclamation`, `TestManager_StaleLockReclamation_Concurrent` in `internal/lock/lock_test.go`

### SCAR-002: StagedMaterialize Freeze in Claude Code

**Category**: integration_failure
**What failed**: `StagedMaterialize` renamed `.claude/` to `.claude.bak/`. CC's file watcher lost track, causing hard freezes. Commit `95cf0bc` removed the method entirely.
**Fix location**: Method removed from `internal/materialize/materialize.go`; all writes use `writeIfChanged()` in `internal/fileutil/fileutil.go`.
**Defensive pattern**: Never rename `.claude/`. Per-file atomic `writeIfChanged()` prevents watcher triggers.
**Regression tests**: `TestSCAR002_StagedMaterializeAbsent`, `TestSCAR002_MaterializeWithOptions_NoClaudeRename` in `internal/materialize/scar_regression_test.go`

### SCAR-003: Idempotency Guard Bypass on Mutation Flags

**Category**: idempotency_failure
**What failed**: `--remove-all` and `--promote-all` did not set `Force: true`. Pipeline short-circuited on matching `ACTIVE_RITE`.
**Fix location**: `internal/materialize/materialize.go:194`; `internal/cmd/sync/materialize.go`
**Defensive pattern**: Mutation flags imply `--force`.
**Regression tests**: `TestRematerializeMena_RepopulatesAfterWipe` in `internal/materialize/routing_test.go`

### SCAR-004: Silent Error Discard at Provenance Load Sites

**Category**: data_corruption (silent failure)
**What failed**: `LoadOrBootstrap` errors discarded via `_`. Corrupted manifests masked; pipeline bootstrapped from empty state. Commits `e5655a9`, `e7277cf`.
**Fix location**: `internal/materialize/materialize.go` (load sites at lines 242-258 and 388-403)
**Defensive pattern**: WARN log on non-file-not-found errors; parse/validation errors abort pipeline. Never `_` discard provenance errors.
**Regression tests**: `TestSCAR004_CorruptProvenanceManifest_PropagatesError`, `TestSCAR004_InvalidSchemaProvenanceManifest_PropagatesError` in `internal/materialize/scar_regression_test.go`

### SCAR-005: Destructive os.RemoveAll on Agents/Commands/Skills Directories

**Category**: data_corruption
**What failed**: `os.RemoveAll` on `agents/`, `commands/`, `skills/` destroyed user content on every sync.
**Fix location**: `internal/materialize/materialize_agents.go:25` (comment: "selective -- do NOT RemoveAll")
**Defensive pattern**: Build managed-set from provenance manifest; remove only managed files.
**Regression tests**: 6 tests in `internal/materialize/selective_write_test.go`

### SCAR-006: Shared Mena Drop for Satellite-Local Rites

**Category**: integration_failure (silent)
**What failed**: `materializeMena()` derived `ritesBase` from `filepath.Dir(resolved.RitePath)`, missing `$KNOSSOS_HOME/rites/`. Commit `89b109c`.
**Fix location**: `internal/materialize/materialize_mena.go:73-87`
**Defensive pattern**: `sharedRitesBase` computed from `resolver.KnossosHome()`.
**Regression tests**: 4 tests in `internal/materialize/satellite_mena_test.go`

### SCAR-007: Mixed Dro/Lego Directories Block Skill Resolution

**Category**: schema_evolution
**What failed**: Directories with both `INDEX.dro.md` and `*.lego.md` routed entirely as dromena. Legomena invisible.
**Fix location**: Mena directories split into separate `*-ref` and `*-catalog` directories.
**Defensive pattern**: `ari lint` enforces separation.

### SCAR-008: Budget Hook Async Log Spam

**Category**: performance_cliff
**What failed**: Budget hook with `async: true` flooded logs. Commit `85d66d5`.
**Fix location**: `config/hooks.yaml` (absence of `async: true` on budget entry)
**Defensive pattern**: Sub-100ms hooks must be synchronous.
**Regression tests**: `TestSCAR008_BudgetHook_MustNotBeAsync` in `internal/materialize/scar_regression_test.go`

### SCAR-009: Hooks Generated in Wrong Nested-Matcher Format

**Category**: integration_failure
**What failed**: Flat `{command, matcher}` format silently rejected by CC. Commit `bb1666a`.
**Fix location**: `internal/materialize/hooks.go`; guard comment at `internal/hook/output.go:27`
**Defensive pattern**: Hook generation always uses CC's nested-matcher format.
**Regression tests**: `internal/materialize/hooks_test.go`

### SCAR-010: Missing Timeout on Hook and Git Subprocess Commands

**Category**: performance_cliff
**What failed**: Budget hook stdin read blocked indefinitely; git calls without timeout. Full sweep commit `6304dc5`.
**Fix location**: `internal/cmd/hook/budget.go:62`; `internal/worktree/git.go:17-21` (`gitCmdCtx` with 30s timeout); 5 additional files
**Defensive pattern**: ALL hooks use `withTimeout()`. ALL git subprocesses use `exec.CommandContext`.
**Regression tests**: `internal/cmd/hook/hook_test.go`

### SCAR-011: Session Writeguard Used .current-session (Last Production Caller)

**Category**: configuration_drift
**What failed**: `writeguard.go:isMoiraiLockHeld()` was last caller of deprecated `.current-session`.
**Fix location**: `internal/cmd/hook/writeguard.go:360`
**Defensive pattern**: Session resolution via standard priority chain.
**Regression tests**: `TestWriteguard_ParkedSession_*` in `internal/cmd/hook/writeguard_test.go`

### SCAR-012: Archived Session Writeguard Gap

**Category**: integration_failure
**What failed**: Archived sessions received generic "Use Moirai" denial; CC agents entered infinite retry loops.
**Fix location**: `internal/cmd/hook/writeguard.go:192`; `internal/session/resolve.go:96`
**Defensive pattern**: Check `isSessionArchived()` after `isMoiraiLockHeld()` fails.
**Regression tests**: `TestWriteguard_ArchivedSession_DeniesWithClearMessage` in `internal/cmd/hook/writeguard_test.go`

### SCAR-013: Ghost Dirs and Already-Archived Session Wrap

**Category**: integration_failure (edge cases)
**What failed**: Three unguarded wrap edge cases: already-archived, ghost dirs, existing archive targets.
**Fix location**: `internal/cmd/session/wrap.go`
**Regression tests**: `TestWrapAlreadyArchived`, `TestWrapNoGhostDirectory` in `internal/cmd/session/wrap_test.go`

### SCAR-014: Phantom Status Values in Session FSM

**Category**: schema_evolution
**What failed**: `COMPLETE` and `COMPLETED` written to SESSION_CONTEXT.md but not registered in FSM.
**Fix location**: `internal/session/status.go:48-51`
**Defensive pattern**: `NormalizeStatus()` with alias map.
**Regression tests**: `TestParseContext_NormalizesPhantomStatus`, `TestParseContext_NormalizesComplete` in `internal/session/context_test.go`

### SCAR-015: stdout Pollution from Shell Log Functions Corrupting Manifests

**Category**: data_corruption
**What failed**: Shell log functions used bare `echo` without `>&2`. Commit `c8b551a`.
**Fix location**: Shell sync scripts (4 files)
**Defensive pattern**: ALL shell log functions redirect with `>&2`.
**Regression tests**: `TestSCAR015_ShellScripts_StderrLogging` in `internal/materialize/scar_regression_test.go`

### SCAR-016: Bash Arithmetic Increment from Zero with set -e

**Category**: schema_evolution (shell behavior)
**What failed**: `((var++))` returns exit code 1 when var is zero under `set -euo pipefail`. Commit `1641792`.
**Defensive pattern**: `((var++)) || true` suffix.
**Regression tests**: `TestSCAR016_ShellScripts_NoUnprotectedArithmeticIncrement` in `internal/materialize/scar_regression_test.go`

### SCAR-017: @skill-name Anti-Pattern in Agent Prompts

**Category**: configuration_drift
**What failed**: 195+ agent files used `@skill-name` syntax. CC has no `@mention` mechanism. Commit `57f3601`.
**Fix location**: `internal/cmd/lint/lint.go:394` (rule `skill-at-syntax`, HIGH severity)
**Regression tests**: 14 tests in `internal/cmd/lint/lint_test.go`; `internal/agent/regenerate_test.go:376`

### SCAR-018: context: fork Blocks Task Tool Access

**Category**: integration_failure
**What failed**: `/know` dromenon had `context: fork`. Forked slash commands cannot use Task tool. Commit `4d92db4`.
**Fix location**: `mena/know/INDEX.dro.md` (absence of `context: fork`); `internal/cmd/lint/lint.go:625`
**Defensive pattern**: Dromena needing Task tool must NOT have `context: fork`.

### SCAR-019: Invalid Agent Colors in CC Palette

**Category**: configuration_drift
**What failed**: 13+ agents used colors outside CC's 8-value palette.
**Fix location**: `internal/cmd/lint/lint.go:381-385`
**Regression tests**: Lint tests in `internal/cmd/lint/lint_test.go`

### SCAR-020: Session ID Not Passed to CLI Subprocesses from Dromena

**Category**: integration_failure
**What failed**: Session dromena failed to pass session ID via `-s` flag. Commit `6f35325`.
**Fix location**: `mena/session/fray/INDEX.dro.md`; `mena/session/sos/INDEX.dro.md`
**Regression tests**: `TestSCAR020_SessionDromena_ExplicitSessionIDPassing` in `internal/materialize/scar_regression_test.go`

### SCAR-021: Cross-Rite Agents Materialized to Project Scope

**Category**: integration_failure (scope misrouting)
**What failed**: Global agents written to project `.claude/agents/` instead of user scope. Commit `7ef0213`.
**Fix location**: `materializeCrossRiteAgents()` removed from `internal/materialize/materialize.go`
**Regression tests**: `TestSCAR021_CrossRiteAgents_ProjectScopeExclusion` in `internal/materialize/scar_regression_test.go`

### SCAR-022: Provenance Schema Rejects Abbreviated SHA256

**Category**: data_corruption (test fixture)
**What failed**: Test fixtures used abbreviated SHA256. Schema requires full 64-char hex with `sha256:` prefix. Commit `dacc620`.
**Fix location**: `internal/materialize/mena/mena_test.go` (test fixtures)

### SCAR-023: Template Path Resolution Fails for Knossos Self-Hosting

**Category**: configuration_drift
**What failed**: Template resolution only checked `$PROJECT/templates/sections/`. Commit `bff1293`.
**Fix location**: `internal/materialize/source/resolver.go:273-280`
**Regression tests**: `TestSCAR023_*` in `internal/materialize/source/source_test.go`

### SCAR-024: Stale Throughline IDs After Rite Switch

**Category**: integration_failure (scope misrouting)
**What failed**: `.throughline-ids.json` retained agent IDs from previous rite.
**Fix location**: `internal/materialize/materialize.go:660` (`cleanupThroughlineIDs()`)
**Regression tests**: 5 tests in `internal/materialize/throughline_cleanup_test.go`

### SCAR-025: Deleted Files in User Scope Sync Not Handled

**Category**: data_corruption (silent failure)
**What failed**: Stale manifest records caused sync to permanently skip deleted files.
**Fix location**: `internal/materialize/mena.go` (three `os.Stat` call sites)
**Regression tests**: `internal/materialize/mena_test.go`

### SCAR-026: Revert -- Moirai Delegation in /start and /sprint

**Category**: historical_boundary
**What failed**: Writeguard `additionalContext` field created coupling with session orchestration. Reverted commit `104bb12`.
**Fix location**: `internal/cmd/hook/writeguard.go:outputBlock()` (absence of `additionalContext`)
**Defensive pattern**: Writeguard must not contain delegation hints.

### SCAR-027: Ephemeral Skill in Shared Mena

**Category**: historical_boundary
**What failed**: Session artifact added to `rites/shared/mena/`. Reverted commit `f0971e4`.
**Fix location**: `internal/cmd/lint/lint.go:621` (rule `session-artifact-in-shared-mena`)
**Regression tests**: `TestSCAR027_SharedMena_NoSessionArtifacts` in `internal/materialize/scar_regression_test.go`

### SCAR-028: MCP Servers Written to Wrong CC Config File (RESOLVED)

**Category**: integration_failure (silent)
**What failed**: MCP servers written to `settings.local.json` instead of `.mcp.json`. CC doesn't read MCPs from `settings.local.json`.
**Fix location**: `internal/materialize/materialize_settings.go` — `materializeMcpJson()` now writes to `.mcp.json`. Comment at line 21: "MCP servers are NOT written here -- see materializeMcpJson (SCAR-028)."
**Regression tests**: `TestSCAR028_MCPServers_NotInSettingsLocalJson` in `internal/materialize/mcp_integration_test.go:264`
**Status**: Fixed post `dbf81b8`.

### SCAR-029 (Candidate): Moirai Template Uses ${SESSION_ID} Shell Variable Syntax

**Category**: configuration_drift
**What failed**: `agents/moirai.md` used `${SESSION_ID}` (shell syntax) instead of `{session-id}` placeholder. Fixed commit `6f897d1`.
**Fix location**: `agents/moirai.md:187`
**Defensive pattern**: Agent prompts use `{kebab-case}` placeholder syntax, never `${SHELL_VAR}`.

### SCAR-030: Go 1.23 Lacks t.Chdir -- Use DI Parameter Injection

**Category**: testing
**What failed**: `os.Chdir` in tests mutates process-global CWD, blocking `t.Parallel()`.
**Fix pattern**: Add `workDir string` field to struct or function parameter. Commit `5ec4a18`.

### SCAR-031: Materialize Tests Are I/O-Bound -- t.Parallel Yields ~30% Not 50%+

**Category**: performance
**Root cause**: Tests perform filesystem materialization to `t.TempDir()`. I/O wait dominates, not CPU.

### SCAR-032: TestInit_WithRite Global State via common.SetEmbeddedAssets

**Category**: global_state
**What failed**: `common.SetEmbeddedAssets()` mutates package-level global state.
**Guard**: Test deliberately excluded from `t.Parallel()`. Commit `5ec4a18`.

### SCAR-033: Config Canonicalization Timing

**Category**: integration_failure
**What failed**: `hooks.yaml` updated to canonical names BEFORE translation-aware binary was rebuilt. CC rejected invalid keys. Fix commit `6d22a96`.
**Defensive pattern**: Build + install binary BEFORE committing config format changes.

### Unnumbered: Writeguard Overly Broad Path Traversal Check

**Category**: integration_failure (false positive)
**What failed**: `parseFilePath` blocked all absolute paths. CC requires absolute paths for Write/Edit tools. Fix commit `5dd1901`.
**Fix location**: `internal/cmd/hook/writeguard.go` — removed `filepath.IsAbs` / `HasPrefix ".."` check.

### GO-001/002/003: Three Content Rewriting Bypass Paths

**Category**: data_corruption (pipeline bypass)
**What failed**: `RewriteMenaContentPaths` omitted from three code paths. Commits `3e82fa6`, `137d4ca`, `3614eec`.
**Fix locations**: `internal/materialize/mena/engine.go:144`; `internal/materialize/mena/walker.go:83`; `internal/materialize/userscope/sync_mena.go:303,353`; `internal/materialize/mena/content_rewrite.go:50` (export rename)
**Regression tests**: `TestSCAR_ContentRewriteNotBypassed` in `internal/materialize/mena/content_rewrite_test.go:288`

---

## Category Coverage

| Category | Count | SCAR IDs |
|----------|-------|----------|
| Integration failure | 12 | SCAR-002, 006, 009, 012, 013, 018, 020, 021, 024, 028, 033, Writeguard (unnumbered) |
| Data corruption | 6 | SCAR-004, 005, 015, 022, 025, GO-001/002/003 |
| Configuration drift | 6 | SCAR-008, 011, 017, 019, 023, 029 (candidate) |
| Schema evolution | 3 | SCAR-007, 014, 016 |
| Performance cliff | 2 | SCAR-008, 010 |
| Historical boundary | 2 | SCAR-026, 027 |
| Testing constraints | 2 | SCAR-030, 032 |
| Race condition | 1 | SCAR-001 |
| Idempotency failure | 1 | SCAR-003 |
| Performance (measurement) | 1 | SCAR-031 |

Categories searched but not found: memory leak, nil pointer panic, network timeout, infinite loop.

---

## Fix-Location Mapping

All fix locations verified against HEAD (`59a0de2`). See individual SCAR entries above for full paths. Key compound fixes:
- SCAR-004 (2 load sites), SCAR-009 (hooks.go + output.go), SCAR-010 (6 files), SCAR-012 (2 files), SCAR-015 (4 shell scripts), SCAR-018 (mena + lint), SCAR-020 (2 dromena), GO-001/002/003 (4 locations)

---

## Defensive Pattern Documentation

| SCAR | Pattern | Guard Type | Test |
|------|---------|-----------|------|
| SCAR-001 | Atomic flock-on-existing-fd | Behavioral | `TestManager_StaleLockReclamation*` |
| SCAR-002 | Never rename `.claude/`; writeIfChanged | Behavioral + Structural | `TestSCAR002_*` |
| SCAR-003 | Mutation flags imply Force | Behavioral | `TestRematerializeMena_RepopulatesAfterWipe` |
| SCAR-004 | Never `_` discard provenance errors | Behavioral | `TestSCAR004_*` |
| SCAR-005 | Selective write from managed-set | Behavioral | 6 tests in `selective_write_test.go` |
| SCAR-006 | `sharedRitesBase` from KnossosHome() | Behavioral | 4 tests in `satellite_mena_test.go` |
| SCAR-007 | `ari lint` enforces dro/lego separation | Lint | `ari lint` |
| SCAR-008 | No async on sub-100ms hooks | Structural | `TestSCAR008_*` |
| SCAR-009 | Nested-matcher format; wire format comment guard | Behavioral + Comment | `hooks_test.go` |
| SCAR-010 | withTimeout() + CommandContext everywhere | Behavioral | `hook_test.go` |
| SCAR-014 | NormalizeStatus() at all read sites | Behavioral | `TestParseContext_Normalizes*` |
| SCAR-017 | skill-at-syntax lint rule (HIGH) | Lint + Behavioral | 14 lint tests |
| SCAR-018 | No context:fork on Task-needing dromena | Structural + Lint | Lint tests |
| SCAR-020 | Explicit -s flag in session dromena | Structural | `TestSCAR020_*` |
| SCAR-021 | Cross-rite agents user-scope only | Behavioral | `TestSCAR021_*` |
| SCAR-027 | session-artifact-in-shared-mena lint rule | Lint | `TestSCAR027_*` |
| SCAR-028 | materializeMcpJson() to .mcp.json | Behavioral + Comment | `TestSCAR028_*` |
| GO-001/002/003 | RewriteMenaContentPaths in ALL paths | Behavioral + Compile | `TestSCAR_ContentRewriteNotBypassed` |

---

## Agent-Relevance Tagging

| SCAR | Agent Role(s) | Why |
|------|--------------|-----|
| SCAR-001 | principal-engineer | Lock code must use atomic flock-on-existing-fd |
| SCAR-002 | principal-engineer, context-architect | Never rename `.claude/`; use writeIfChanged |
| SCAR-003 | principal-engineer | Mutation flags must imply Force |
| SCAR-004 | principal-engineer | Never `_` discard provenance errors |
| SCAR-005 | principal-engineer, context-architect | RemoveAll on user dirs destroys user content |
| SCAR-006 | principal-engineer, qa-adversary | Satellite paths differ from core; test both |
| SCAR-007 | context-architect | Mixed dro/lego dirs break CC skill resolution |
| SCAR-008 | context-architect | Sub-100ms hooks must be sync |
| SCAR-009 | principal-engineer | CC hook format is nested; flat silently rejected |
| SCAR-010 | principal-engineer | ALL hooks need timeout; ALL git need CommandContext |
| SCAR-011 | principal-engineer | .current-session deprecated; use priority chain |
| SCAR-012 | principal-engineer, context-architect | Archived sessions need distinct denial path |
| SCAR-013 | principal-engineer, qa-adversary | Session wrap has three edge case guards |
| SCAR-014 | principal-engineer | All status reads through NormalizeStatus() |
| SCAR-017 | context-architect | @skill-name not a CC primitive |
| SCAR-018 | context-architect | context:fork removes Task tool access |
| SCAR-020 | principal-engineer | LLM context not accessible to bash; pass -s flag |
| SCAR-021 | principal-engineer, context-architect | Cross-rite agents in project scope cause shadowing |
| SCAR-023 | principal-engineer | Self-hosted knossos uses knossos/templates/ |
| SCAR-024 | principal-engineer | Rite switch must clean .throughline-ids.json |
| SCAR-027 | context-architect | Shared mena = permanent; session artifacts in .sos/wip/ |
| SCAR-028 | principal-engineer, context-architect | MCP servers to .mcp.json, not settings.local.json |
| SCAR-033 | principal-engineer, context-architect | Build binary before committing config changes |
| GO-001/002/003 | principal-engineer | RewriteMenaContentPaths in ALL copy paths |

---

## Knowledge Gaps

1. **GO-001/002/003 numbering collision**: Prior `.know/scar-tissue.md` recommended SCAR-028/029/030 but SCAR-028 was used for MCP servers. Recommend SCAR-034/035/036.

2. **SCAR-029 candidate** (moirai template) has no regression test and no formal SCAR number.

3. **Writeguard path traversal removal** (commit `5dd1901`) has no SCAR catalog entry.

4. **SCAR-010 binary mismatch trap**: No automated guard for build-vs-install divergence. MEMORY.md documents this but no SCAR entry exists.

5. **CI infrastructure scar (GITHUB_TOKEN/GoReleaser)**: Documented in `.sos/land/scar-tissue.md` but not numbered.

6. **Hook stdin-only transport**: Documented in MEMORY.md but no SCAR entry guards against re-introduction of env var paths.

7. **Shell-era session locking scars** (LOCK-001, LOCK-003, STATE-001): Documented in git history but superseded by Go implementations.

8. **TestSCAR018 not found in current test files**: Protection is through lint rule at `lint.go:625` instead.

---
domain: scar-tissue
generated_at: "2026-03-23T18:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "78abb186"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/scar-tissue.md"
land_hash: "c5cdd76ae3802cff6d1fc703b8d5fbe2f8f09253542e7b3972775badb05797fa"
---

# Codebase Scar Tissue

## Failure Catalog Completeness

**Catalog total**: 33 numbered SCAR entries (SCAR-001 through SCAR-033) plus 4 unnumbered entries (Writeguard path traversal, GO-001/002/003 content rewriting bypass). Sources: `internal/materialize/scar_regression_test.go`, `internal/materialize/mena/content_rewrite_test.go`, `internal/materialize/source/source_test.go`, `internal/agent/regenerate_test.go`, git history through `78abb186` (HEAD), and `.sos/land/scar-tissue.md`.

---

### SCAR-001: TOCTOU Race in Stale Lock Reclamation

**Category**: race_condition
**What failed**: Stale lock reclamation used `os.Remove(lockfile)` then `os.OpenFile(lockfile)`. A competing process could acquire the lock between Remove and reopen.
**Fix location**: `internal/lock/lock.go:117-134`
**Defensive pattern**: `syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)` on the already-open fd. Never Remove+reopen.
**Regression tests**: `TestManager_StaleLockReclamation`, `TestManager_StaleLockReclamation_Concurrent` in `internal/lock/lock_test.go` (lines 424, 478)

---

### SCAR-002: StagedMaterialize Freeze in Claude Code

**Category**: integration_failure
**What failed**: `StagedMaterialize` renamed `.claude/` to `.claude.bak/` during materialization. CC's file watcher lost track of its config directory, causing hard freezes. Commit `95cf0bc` removed the method entirely.
**Fix location**: Method absent from `internal/materialize/materialize.go` (verified via reflect in test). Per-file atomic writes use `writeIfChanged()` in `internal/fileutil/fileutil.go`.
**Defensive pattern**: Never rename `.claude/`. Per-file atomic `writeIfChanged()` prevents watcher triggers. All code touching the channel dir carries `HA-FS` annotation.
**Regression tests**: `TestSCAR002_StagedMaterializeAbsent`, `TestSCAR002_MaterializeWithOptions_NoClaudeRename` in `internal/materialize/scar_regression_test.go:31,50`

---

### SCAR-003: Idempotency Guard Bypass on Mutation Flags

**Category**: idempotency_failure
**What failed**: `--remove-all` and `--promote-all` did not set `Force: true`. Pipeline short-circuited on matching `ACTIVE_RITE`, skipping the intended mutation.
**Fix location**: `internal/materialize/materialize.go:194`; `internal/cmd/sync/materialize.go`
**Defensive pattern**: Mutation flags imply `--force`.
**Regression tests**: `TestRematerializeMena_RepopulatesAfterWipe` in `internal/materialize/routing_test.go`

---

### SCAR-004: Silent Error Discard at Provenance Load Sites

**Category**: data_corruption (silent failure)
**What failed**: `LoadOrBootstrap` errors discarded via blank identifier `_`. Corrupted manifests masked; pipeline bootstrapped from empty state.
**Fix location**: `internal/materialize/materialize.go` (load sites at lines 242-258 and 388-403)
**Defensive pattern**: WARN log on non-file-not-found errors; parse/validation errors abort pipeline. Never `_` discard provenance errors.
**Regression tests**: `TestSCAR004_CorruptProvenanceManifest_PropagatesError`, `TestSCAR004_InvalidSchemaProvenanceManifest_PropagatesError` in `internal/materialize/scar_regression_test.go:100,253`

---

### SCAR-005: Destructive os.RemoveAll on Agents/Commands/Skills Directories

**Category**: data_corruption
**What failed**: `os.RemoveAll` on `agents/`, `commands/`, `skills/` destroyed user content on every sync.
**Fix location**: `internal/materialize/materialize_agents.go:30` (comment: "selective — do NOT RemoveAll")
**Defensive pattern**: Build managed-set from provenance manifest; remove only managed files. User content in satellite regions is never destroyed.
**Regression tests**: 6 tests in `internal/materialize/selective_write_test.go`

---

### SCAR-006: Shared Mena Drop for Satellite-Local Rites

**Category**: integration_failure (silent)
**What failed**: `materializeMena()` derived `ritesBase` from `filepath.Dir(resolved.RitePath)`, missing `$KNOSSOS_HOME/rites/`. Satellite-local rites silently dropped shared mena.
**Fix location**: `internal/materialize/materialize_mena.go:85-99` — `sharedRitesBase` computed from `resolver.KnossosHome()`.
**Defensive pattern**: Always resolve shared/dependency mena via `KnossosHome()` not local rite dir.
**Regression tests**: 4 tests in `internal/materialize/satellite_mena_test.go`

---

### SCAR-007: Mixed Dro/Lego Directories Block Skill Resolution

**Category**: schema_evolution
**What failed**: Directories with both `INDEX.dro.md` and `*.lego.md` routed entirely as dromena. Legomena invisible to CC — 195+ skills silently uncallable.
**Fix location**: Mena directories split into separate `*-ref` (legomena) and `*-catalog` (dromena) directories. 4 rites required splitting.
**Defensive pattern**: `ari lint` enforces dro/lego separation at `internal/cmd/lint/lint.go`.
**Regression tests**: None dedicated; covered by lint enforcement.

---

### SCAR-008: Budget Hook Async Log Spam

**Category**: performance_cliff
**What failed**: Budget hook configured with `async: true`. Running async created "Async hook completed" notification for every single tool call.
**Fix location**: `config/hooks.yaml` (absence of `async: true` on the budget hook entry)
**Defensive pattern**: Sub-100ms hooks must be synchronous.
**Regression tests**: `TestSCAR008_BudgetHook_MustNotBeAsync` in `internal/materialize/scar_regression_test.go:148`

---

### SCAR-009: Hooks Generated in Wrong Nested-Matcher Format

**Category**: integration_failure
**What failed**: Flat `{command, matcher}` format silently rejected by CC. Hooks generated but never fired.
**Fix location**: `internal/materialize/hooks.go`; guard comment at `internal/hook/output.go:27` — "CC wire format (SCAR-009) -- do not change to canonical name"
**Defensive pattern**: Hook generation always uses CC's nested-matcher format; wire format comment guards the exact key name.
**Regression tests**: `internal/materialize/hooks_test.go`

---

### SCAR-010: Missing Timeout on Hook and Git Subprocess Commands

**Category**: performance_cliff
**What failed**: Budget hook stdin read blocked indefinitely; git calls without timeout.
**Fix location**: `internal/cmd/hook/budget.go:65` (`ctx.withTimeout()`); `internal/worktree/git.go:17-21` (`gitCmdCtx` with 30s timeout context); 5 additional files in hook package.
**Defensive pattern**: ALL hooks use `withTimeout()`. ALL git subprocesses use `exec.CommandContext`.
**Regression tests**: `internal/cmd/hook/hook_test.go`

---

### SCAR-011: Session Writeguard Used .current-session (Last Production Caller)

**Category**: configuration_drift
**What failed**: `writeguard.go:isMoiraiLockHeld()` was the last caller of deprecated `.current-session` file. Session resolution broke silently.
**Fix location**: `internal/cmd/hook/writeguard.go:360` — migrated to priority chain resolution.
**Defensive pattern**: Session resolution via standard priority chain.
**Regression tests**: `TestWriteguard_ParkedSession_*` in `internal/cmd/hook/writeguard_test.go`

---

### SCAR-012: Archived Session Writeguard Gap

**Category**: integration_failure
**What failed**: Archived sessions received generic "Use Moirai" denial. CC agents entered infinite retry loops.
**Fix location**: `internal/cmd/hook/writeguard.go:201-204` (`isSessionArchived()` check); `internal/session/resolve.go:96`
**Defensive pattern**: Check `isSessionArchived()` after `isMoiraiLockHeld()` fails.
**Regression tests**: `TestWriteguard_ArchivedSession_DeniesWithClearMessage` in `internal/cmd/hook/writeguard_test.go`

---

### SCAR-013: Ghost Dirs and Already-Archived Session Wrap

**Category**: integration_failure (edge cases)
**What failed**: Three unguarded wrap edge cases: (1) wrapping already-archived session, (2) ghost directory cleanup, (3) existing archive target collision.
**Fix location**: `internal/cmd/session/wrap.go` — guards for all three cases.
**Regression tests**: `TestWrapAlreadyArchived` (line 933), `TestWrapNoGhostDirectory` (line 998) in `internal/cmd/session/wrap_test.go`

---

### SCAR-014: Phantom Status Values in Session FSM

**Category**: schema_evolution
**What failed**: `COMPLETE` and `COMPLETED` written by older versions but not registered in FSM.
**Fix location**: `internal/session/status.go:44-51` — alias map; `NormalizeStatus()` called at all read sites.
**Defensive pattern**: `NormalizeStatus()` applied at every status read site.
**Regression tests**: `TestParseContext_NormalizesPhantomStatus`, `TestParseContext_NormalizesComplete` in `internal/cmd/session/context_test.go`

---

### SCAR-015: stdout Pollution from Shell Log Functions Corrupting Manifests

**Category**: data_corruption
**What failed**: Shell log functions used bare `echo` without `>&2`. Log messages embedded in manifest JSON keys.
**Fix location**: 4 shell sync scripts — all log functions redirect to stderr with `>&2`.
**Defensive pattern**: ALL shell log functions redirect with `>&2`.
**Regression tests**: `TestSCAR015_ShellScripts_StderrLogging` in `internal/materialize/scar_regression_test.go:309`

---

### SCAR-016: Bash Arithmetic Increment from Zero with set -e

**Category**: schema_evolution (shell behavior)
**What failed**: `((var++))` returns exit code 1 when var is zero under `set -euo pipefail`. Script silently exits.
**Fix location**: All affected `.sh` files — `((var++)) || true` suffix.
**Regression tests**: `TestSCAR016_ShellScripts_NoUnprotectedArithmeticIncrement` in `internal/materialize/scar_regression_test.go:397`

---

### SCAR-017: @skill-name Anti-Pattern in Agent Prompts

**Category**: configuration_drift
**What failed**: 195+ agent files used `@skill-name` syntax. CC has no `@mention` mechanism. Skills silently ignored.
**Fix location**: `internal/cmd/lint/lint.go:909-937` (rule `skill-at-syntax`, HIGH severity).
**Regression tests**: 14 tests in `internal/cmd/lint/lint_test.go`; negative assertion in `internal/agent/regenerate_test.go:376`

---

### SCAR-018: context: fork Blocks Task Tool Access

**Category**: integration_failure
**What failed**: `/know` dromenon had `context: fork`. Forked slash commands run as subagents and CANNOT use the Task tool.
**Fix location**: `mena/know/INDEX.dro.md` (absence of `context: fork`); `internal/cmd/lint/lint.go:668-674` (SCAR-018 lint rule).
**Regression tests**: Lint tests in `internal/cmd/lint/lint_test.go`.

---

### SCAR-019: Invalid Agent Colors in CC Palette

**Category**: configuration_drift
**What failed**: 13+ agents used colors outside CC's 8-value palette. Invalid colors silently accepted but displayed incorrectly.
**Fix location**: `internal/cmd/lint/lint.go:381-385` (color validation against 8-value palette).
**Regression tests**: Lint tests in `internal/cmd/lint/lint_test.go`.

---

### SCAR-020: Session ID Not Passed to CLI Subprocesses from Dromena

**Category**: integration_failure
**What failed**: Session lifecycle dromena did not instruct the LLM to pass session ID via `-s` flag. Bash subprocesses cannot access LLM context.
**Fix location**: `mena/session/fray/INDEX.dro.md`; `mena/session/sos/INDEX.dro.md`.
**Regression tests**: `TestSCAR020_SessionDromena_ExplicitSessionIDPassing` in `internal/materialize/scar_regression_test.go:460`

---

### SCAR-021: Cross-Rite Agents Materialized to Project Scope

**Category**: integration_failure (scope misrouting)
**What failed**: Global agents materialized into project `.claude/agents/` alongside rite-specific agents, causing shadowing and orphan accumulation.
**Fix location**: `materializeCrossRiteAgents()` removed from `internal/materialize/materialize.go`. Cross-rite agents now exclusively use user-scope.
**Regression tests**: `TestSCAR021_CrossRiteAgents_ProjectScopeExclusion` in `internal/materialize/scar_regression_test.go:208`

---

### SCAR-022: Provenance Schema Rejects Abbreviated SHA256

**Category**: data_corruption (test fixture)
**What failed**: Test fixtures used abbreviated SHA256 checksums. Schema requires full 64-character hex with `sha256:` prefix.
**Fix location**: `internal/materialize/mena/mena_test.go` — test fixtures updated to full 64-char sha256 with prefix.

---

### SCAR-023: Template Path Resolution Fails for Knossos Self-Hosting

**Category**: configuration_drift
**What failed**: Template resolution only checked `$PROJECT/templates/sections/`. Knossos self-hosting requires `knossos/templates/`.
**Fix location**: `internal/materialize/source/resolver.go:273-280` — added knossos-home-relative resolution.
**Regression tests**: `TestSCAR023_*` in `internal/materialize/source/source_test.go:587`

---

### SCAR-024: Stale Throughline IDs After Rite Switch

**Category**: integration_failure (scope misrouting)
**What failed**: `.throughline-ids.json` retained agent IDs from previous rite after rite switch.
**Fix location**: `internal/materialize/materialize.go:791` — `cleanupThroughlineIDs()` called during rite switch.
**Regression tests**: 5 tests in `internal/materialize/throughline_cleanup_test.go`

---

### SCAR-025: Deleted Files in User Scope Sync Not Handled

**Category**: data_corruption (silent failure)
**What failed**: Stale manifest records caused sync to permanently skip re-projection of deleted source files.
**Fix location**: `internal/materialize/userscope/sync.go` — three `os.Stat` call sites with missing-file handling.
**Regression tests**: `internal/materialize/userscope/sync_test.go`

---

### SCAR-026: Revert — Moirai Delegation in /start and /sprint

**Category**: historical_boundary
**What failed**: `additionalContext` field in writeguard output created coupling between writeguard and Moirai session orchestration. Reverted.
**Fix location**: `internal/cmd/hook/writeguard.go` — absence of `additionalContext` field.
**Defensive pattern**: Writeguard output is pure allow/deny with reason. No orchestration hints.

---

### SCAR-027: Ephemeral Skill in Shared Mena

**Category**: historical_boundary
**What failed**: ARCH-REVIEW session artifact accidentally added to `rites/shared/mena/` as a legomenon. Shared mena is permanent platform knowledge.
**Fix location**: `internal/cmd/lint/lint.go:847-896` (rule `session-artifact-in-shared-mena`)
**Regression tests**: `TestSCAR027_SharedMena_NoSessionArtifacts` in `internal/materialize/scar_regression_test.go:520`

---

### SCAR-028: MCP Servers Written to Wrong CC Config File

**Category**: integration_failure (silent)
**What failed**: MCP server declarations written to `settings.local.json` instead of `.mcp.json`. CC reads MCP servers from `.mcp.json` only.
**Fix location**: `internal/materialize/materialize_settings.go:22` — guard comment; `materializeMcpJson()` writes to `.mcp.json`; line 40-42 adds SCAR-028 cleanup.
**Regression tests**: `TestSCAR028_MCPServers_NotInSettingsLocalJson` in `internal/materialize/mcp_integration_test.go:39`

---

### SCAR-029 (Candidate): Moirai Template Uses ${SESSION_ID} Shell Variable Syntax

**Category**: configuration_drift
**What failed**: `agents/moirai.md` used `${SESSION_ID}` instead of `{session-id}` placeholder syntax. Substitution never occurred.
**Fix location**: `agents/moirai.md:187`
**Status**: No formal SCAR number assigned. No dedicated regression test.

---

### SCAR-030: Go 1.23 Lacks t.Chdir — Use DI Parameter Injection

**Category**: testing
**What failed**: `os.Chdir` in tests mutates process-global CWD, blocking `t.Parallel()`. Go 1.23 does not have `t.Chdir()`.
**Fix pattern**: Add `workDir string` field to struct or function parameter (dependency injection).
**Current status**: 38 remaining `t.Setenv` usages. `os.Chdir` eliminated from tests (0 occurrences).

---

### SCAR-031: Materialize Tests Are I/O-Bound — t.Parallel Yields ~30% Not 50%+

**Category**: performance (measurement)
**Root cause**: Tests perform filesystem materialization to `t.TempDir()`. I/O wait dominates, not CPU.
**Fix pattern**: No code fix — calibration of expectations only.

---

### SCAR-032: TestInit_WithRite Global State via common.SetEmbeddedAssets

**Category**: global_state
**What failed**: `common.SetEmbeddedAssets()` mutates package-level global state. Tests using this path cannot run in parallel.
**Fix location**: `internal/cmd/initialize/init_test.go:73` — comment "Not parallel: mutates global embedded assets".
**Defensive pattern**: Tests that mutate package globals must not call `t.Parallel()`.

---

### SCAR-033: Config Canonicalization Timing

**Category**: integration_failure
**What failed**: `hooks.yaml` updated to canonical key names BEFORE the translation-aware binary was rebuilt. CC received unrecognized keys.
**Fix location**: Operational pattern — build + install binary BEFORE committing config format changes.
**Defensive pattern**: Deployment sequence: (1) build new binary, (2) install to PATH, (3) commit config.

---

### Unnumbered: Writeguard Overly Broad Path Traversal Check

**Category**: integration_failure (false positive)
**What failed**: `parseFilePath` blocked all absolute paths. CC requires absolute paths for Write/Edit tool calls.
**Fix location**: `internal/cmd/hook/writeguard.go` — removed `filepath.IsAbs` / `HasPrefix ".."` check.

---

### GO-001/002/003: Three Content Rewriting Bypass Paths

**Category**: data_corruption (pipeline bypass)
**What failed**: `RewriteMenaContentPaths` implemented in primary path but absent from three bypass paths: (1) engine.go standalone, (2) userscope walker, (3) userscope standalone sync.
**Fix locations**:
  - `internal/materialize/mena/engine.go:153`
  - `internal/materialize/mena/walker.go:86`
  - `internal/materialize/userscope/sync_mena.go:323,378`
  - Export rename: `internal/materialize/mena/content_rewrite.go:50` (`RewriteMenaContentPaths` exported)
**Regression tests**: `TestSCAR_ContentRewriteNotBypassed` in `internal/materialize/mena/content_rewrite_test.go:292`

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

Categories searched but not found: memory leak, nil pointer panic, network timeout, infinite loop, SQL/database error, authentication failure.

---

## Fix-Location Mapping

All primary fix locations verified against HEAD (`78abb186`). Key compound fixes (affecting multiple files):

| SCAR | Primary file(s) | Secondary files |
|------|----------------|-----------------|
| SCAR-004 | `internal/materialize/materialize.go` (2 load sites) | — |
| SCAR-009 | `internal/materialize/hooks.go` | `internal/hook/output.go:27` |
| SCAR-010 | `internal/cmd/hook/budget.go:65` | `internal/worktree/git.go:17-21` + 5 files |
| SCAR-012 | `internal/cmd/hook/writeguard.go:201-204` | `internal/session/resolve.go:96` |
| SCAR-015 | 4 shell scripts in rites/, scripts/sync/, scripts/hooks/ | — |
| SCAR-018 | `mena/know/INDEX.dro.md` | `internal/cmd/lint/lint.go:668-674` |
| SCAR-020 | `mena/session/fray/INDEX.dro.md` | `mena/session/sos/INDEX.dro.md` |
| GO-001/002/003 | `internal/materialize/mena/engine.go:153` | `mena/walker.go:86`, `userscope/sync_mena.go:323,378`, `mena/content_rewrite.go:50` |

---

## Defensive Pattern Documentation

| SCAR | Defensive Pattern | Guard Type | Regression Test |
|------|------------------|-----------|-----------------|
| SCAR-001 | Atomic flock-on-existing-fd | Behavioral | `TestManager_StaleLockReclamation*` |
| SCAR-002 | Never rename `.claude/`; writeIfChanged; HA-FS markers | Behavioral + Structural | `TestSCAR002_*` |
| SCAR-003 | Mutation flags imply Force | Behavioral | `TestRematerializeMena_RepopulatesAfterWipe` |
| SCAR-004 | Never `_`-discard provenance errors | Behavioral | `TestSCAR004_*` |
| SCAR-005 | Selective write from managed-set | Behavioral | 6 tests in `selective_write_test.go` |
| SCAR-006 | `sharedRitesBase` from KnossosHome() | Behavioral | 4 tests in `satellite_mena_test.go` |
| SCAR-007 | `ari lint` dro/lego separation enforcement | Lint | `ari lint` |
| SCAR-008 | No async on sub-100ms hooks | Structural | `TestSCAR008_BudgetHook_MustNotBeAsync` |
| SCAR-009 | Nested-matcher wire format; comment guard | Behavioral + Comment | `hooks_test.go` |
| SCAR-010 | withTimeout() everywhere; CommandContext for all git | Behavioral | `hook_test.go` |
| SCAR-012 | `isSessionArchived()` checked after lock fails | Behavioral | `TestWriteguard_ArchivedSession_*` |
| SCAR-014 | `NormalizeStatus()` at all read sites | Behavioral | `TestParseContext_Normalizes*` |
| SCAR-017 | skill-at-syntax lint rule (HIGH severity) | Lint | 14 lint tests; `regenerate_test.go:376` |
| SCAR-018 | No `context: fork` on Task-needing dromena | Structural + Lint | Lint tests |
| SCAR-020 | Explicit -s flag instruction in session dromena | Structural | `TestSCAR020_*` |
| SCAR-021 | Cross-rite agents user-scope only | Behavioral | `TestSCAR021_*` |
| SCAR-027 | `session-artifact-in-shared-mena` lint rule | Lint | `TestSCAR027_*` |
| SCAR-028 | `materializeMcpJson()` to `.mcp.json`; guard comment | Behavioral + Comment | `TestSCAR028_*` |
| SCAR-032 | No `t.Parallel()` on global-mutating tests; comment | Structural | None (guard is absence) |
| GO-001/002/003 | `RewriteMenaContentPaths` in ALL copy paths | Behavioral | `TestSCAR_ContentRewriteNotBypassed` |

---

## Agent-Relevance Tagging

| SCAR | Agent Role(s) | Why This Matters |
|------|--------------|-----------------|
| SCAR-001 | integration-engineer | Lock code must use atomic flock-on-existing-fd — Remove+reopen is a TOCTOU trap |
| SCAR-002 | integration-engineer, context-architect | Never rename `.claude/`. Use writeIfChanged. All `.claude/` path references carry HA-FS annotation |
| SCAR-003 | integration-engineer | Mutation flags must imply Force — pipeline short-circuits otherwise |
| SCAR-004 | integration-engineer | Never `_`-discard provenance errors — masks data corruption |
| SCAR-005 | integration-engineer, context-architect | RemoveAll on user dirs destroys satellite content — use selective write from managed-set |
| SCAR-006 | integration-engineer, compatibility-tester | Satellite-local rites resolve differently — shared mena must come from KnossosHome() |
| SCAR-007 | context-architect | Mixed dro/lego dirs block CC skill resolution — always split into separate directories |
| SCAR-008 | context-architect | Sub-100ms hooks must be synchronous — async creates per-tool-call notification flood |
| SCAR-009 | integration-engineer | CC hook format is nested-matcher — flat format is silently rejected |
| SCAR-010 | integration-engineer | ALL hooks need timeout; ALL git needs CommandContext — no indefinite blocking |
| SCAR-011 | integration-engineer | `.current-session` is deprecated — use priority chain for session resolution |
| SCAR-012 | integration-engineer, context-architect | Archived sessions need distinct denial path — generic denial causes agent retry loops |
| SCAR-013 | integration-engineer, compatibility-tester | Session wrap has three edge case guards |
| SCAR-014 | integration-engineer | All status reads through NormalizeStatus() — phantom aliases exist in wild |
| SCAR-017 | context-architect | `@skill-name` is not a CC primitive — use plain `skills:` frontmatter |
| SCAR-018 | context-architect | `context: fork` removes Task tool access — dromena needing Task must not fork |
| SCAR-020 | integration-engineer | LLM context not accessible to bash subprocesses — always pass session ID via `-s` flag |
| SCAR-021 | integration-engineer, context-architect | Cross-rite agents in project scope cause shadowing — user scope only |
| SCAR-023 | integration-engineer | Self-hosted knossos uses `knossos/templates/` not project `templates/` |
| SCAR-024 | integration-engineer | Rite switch must clean `.throughline-ids.json` — stale IDs misroute agents |
| SCAR-027 | context-architect | Shared mena is permanent platform knowledge — session artifacts to `.sos/wip/` |
| SCAR-028 | integration-engineer, context-architect | MCP servers to `.mcp.json` — CC ignores `settings.local.json` for MCP |
| SCAR-033 | integration-engineer, documentation-engineer | Build binary before committing config changes |
| GO-001/002/003 | integration-engineer | RewriteMenaContentPaths must be wired in ALL copy paths |

Platform-wide scars (no specific agent role):
- SCAR-016: anyone writing bash scripts with `set -euo pipefail`
- SCAR-019: anyone adding agent definitions
- SCAR-022: anyone writing provenance test fixtures
- SCAR-030: all test authors using environment-dependent logic
- SCAR-031: performance calibration — informational
- SCAR-032: test authors using `common.SetEmbeddedAssets`

---

## Knowledge Gaps

1. **GO-001/002/003 numbering**: These have not been formally assigned SCAR numbers. If formalized, they should be SCAR-034, SCAR-035, SCAR-036.
2. **SCAR-029 candidate is unconfirmed**: No SCAR number assigned and no regression test exists.
3. **Unnumbered writeguard path traversal scar**: Commit `5dd1901` — no SCAR entry, no regression test.
4. **SCAR-010 binary mismatch trap**: The operational trap (build writes `./ari` but CC uses PATH binary) has no SCAR number and no automated guard.
5. **GITHUB_TOKEN/GoReleaser CI scar**: Documented in `.sos/land/scar-tissue.md` but no SCAR number (CI infrastructure concern).
6. **Hook stdin-only transport**: Critical integration constraint with no SCAR number and no regression test.
7. **SCAR-015 shell script coverage**: Regression test scans limited directories. New scripts elsewhere are not covered.

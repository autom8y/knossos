---
domain: scar-tissue
generated_at: "2026-03-06T12:00:00Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "3847e28"
confidence: 0.93
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/scar-tissue.md"
land_hash: "95d778e0ce285309732fdad30462168d53be155f7a631e5811b59d7ebe6d1421"
---

# Codebase Scar Tissue

## Failure Catalog

27 SCAR entries cataloged with regression tests, defensive guards, and cross-validated against `internal/materialize/scar_regression_test.go`, `internal/materialize/source/source_test.go`, and git history through commit `3847e28`. Three additional post-catalog items (GO-001/002/003 content rewriting bypasses) documented from land source.

### SCAR-001: TOCTOU Race in Stale Lock Reclamation

**Category**: race_condition
**What failed**: Stale lock reclamation used `os.Remove(lockfile)` then `os.OpenFile(lockfile)`. A competing process could acquire the lock between Remove and reopen.
**Fix location**: `internal/lock/lock.go:117-134`
**Defensive pattern**: `syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)` on the already-open fd; never Remove+reopen.
**Regression tests**: `TestManager_StaleLockReclamation`, `TestManager_StaleLockReclamation_Concurrent` in `internal/lock/lock_test.go`

### SCAR-002: StagedMaterialize Freeze in Claude Code

**Category**: integration_failure
**What failed**: `StagedMaterialize` renamed `.claude/` to `.claude.bak/`. CC's file watcher lost track, causing hard freezes.
**Fix location**: Method removed from `internal/materialize/materialize.go`; all writes use `writeIfChanged()`.
**Defensive pattern**: Never rename `.claude/`. Per-file atomic `writeIfChanged()` writes prevent watcher triggers.
**Regression tests**: `TestSCAR002_StagedMaterializeAbsent` (reflection-based), `TestSCAR002_MaterializeWithOptions_NoClaudeRename` in `internal/materialize/scar_regression_test.go`

### SCAR-003: Idempotency Guard Bypass on Mutation Flags

**Category**: idempotency_failure
**What failed**: `--remove-all` and `--promote-all` did not set `Force: true`. Pipeline short-circuited on matching `ACTIVE_RITE`, leaving directories empty.
**Fix location**: `internal/materialize/materialize.go:194`; `internal/cmd/sync/materialize.go`
**Defensive pattern**: Mutation flags imply `--force`. `writeIfChanged()` prevents spurious CC watcher triggers.
**Regression tests**: `TestRematerializeMena_RepopulatesAfterWipe` in `internal/materialize/routing_test.go`

### SCAR-004: Silent Error Discard at Provenance Load Sites

**Category**: data_corruption (silent failure)
**What failed**: `LoadOrBootstrap` and `DetectDivergence` errors discarded via `_`. Filesystem permission errors masked; pipeline bootstrapped from empty state, overwriting user content.
**Fix location**: `internal/materialize/materialize.go` (load sites at lines 240-241, 327-328)
**Defensive pattern**: WARN log on non-file-not-found errors; parse/validation errors abort pipeline. Never `_` discard provenance errors.
**Regression tests**: `TestSCAR004_CorruptProvenanceManifest_PropagatesError`, `TestSCAR004_InvalidSchemaProvenanceManifest_PropagatesError` in `internal/materialize/scar_regression_test.go`

### SCAR-005: Destructive os.RemoveAll on Agents/Commands/Skills Directories

**Category**: data_corruption
**What failed**: Materialization called `os.RemoveAll` on `agents/`, `commands/`, `skills/` before rewriting. User-created content destroyed.
**Fix location**: `internal/materialize/materialize_agents.go:25` (comment: "selective — do NOT RemoveAll")
**Defensive pattern**: Build managed-set from manifest; remove only managed files; preserve everything outside manifest.
**Regression tests**: 6 tests in `internal/materialize/selective_write_test.go`

### SCAR-006: Shared Mena Drop for Satellite-Local Rites

**Category**: integration_failure (silent)
**What failed**: `materializeMena()` derived `ritesBase` from `filepath.Dir(resolved.RitePath)`. For satellite-local rites, this missed `$KNOSSOS_HOME/rites/`. Shared mena silently dropped.
**Fix location**: `internal/materialize/materialize_mena.go:73-87`
**Defensive pattern**: `sharedRitesBase` computed from `resolver.KnossosHome()`; never infer knossos home from rite path alone.
**Regression tests**: 4 tests in `internal/materialize/satellite_mena_test.go`

### SCAR-007: Mixed Dro/Lego Directories Block Skill Resolution

**Category**: schema_evolution
**What failed**: Directories with both `INDEX.dro.md` and `*.lego.md` routed entirely as dromena. Legomena invisible to `skills:` preloading.
**Fix location**: Split into separate `*-ref` (dromenon) and `*-catalog` (legomenon) directories.
**Defensive pattern**: `ari lint` enforces separation. Never place `.lego.md` siblings inside a dromenon directory.
**Regression tests**: `ari lint` validation

### SCAR-008: Budget Hook Async Log Spam

**Category**: performance_cliff
**What failed**: Budget hook configured `async: true`. Fires on every `PostToolUse` and completes in <5ms. Async created "Async hook completed" notification for every tool call.
**Fix location**: `config/hooks.yaml` (absence of `async: true` on budget entry)
**Defensive pattern**: Sub-100ms hooks must be synchronous.
**Regression tests**: `TestSCAR008_BudgetHook_MustNotBeAsync` in `internal/materialize/scar_regression_test.go`

### SCAR-009: Hooks Generated in Wrong Nested-Matcher Format

**Category**: integration_failure
**What failed**: `settings.local.json` hooks used flat `{command, matcher}` format. CC expects nested `{matcher, hooks: [{type, command}]}`. Hooks silently rejected.
**Fix location**: `internal/materialize/hooks.go`
**Regression tests**: `internal/materialize/hooks_test.go`

### SCAR-010: Budget and Autopark Hooks Missing Timeout / Context-Cancel

**Category**: performance_cliff
**What failed**: `budget` RunE read stdin without timeout (blocked indefinitely after hot rite switch). `getGitStatusQuick()` used `exec.Command` without context (orphaned processes).
**Fix location**: `internal/cmd/hook/budget.go:62`; `internal/cmd/hook/autopark.go:156`
**Defensive pattern**: ALL hook RunE use `withTimeout()`. All git subprocesses use `exec.CommandContext` with deadline.
**Regression tests**: `internal/cmd/hook/hook_test.go`

### SCAR-011: Session Writeguard Used .current-session (Last Production Caller)

**Category**: configuration_drift
**What failed**: `writeguard.go:isMoiraiLockHeld()` was the last caller of deprecated `.current-session` file. Failed silently, denying all Moirai writes.
**Fix location**: `internal/cmd/hook/writeguard.go:360`
**Defensive pattern**: Session resolution via standard priority chain. Never read `.current-session` directly.
**Regression tests**: `TestWriteguard_ParkedSession_MoiraiLockAllow`, `TestWriteguard_ParkedSession_NoLock`, `TestWriteguard_ParkedSession_StaleLock` in `internal/cmd/hook/writeguard_test.go`

### SCAR-012: Archived Session Writeguard Gap

**Category**: integration_failure
**What failed**: Archived sessions got generic "Use Moirai" denial instead of archived explanation. CC agents entered infinite Moirai retry loops.
**Fix location**: `internal/cmd/hook/writeguard.go:192`; `internal/session/resolve.go:96`
**Defensive pattern**: Check `isSessionArchived()` after `isMoiraiLockHeld()` fails. Remove `.moirai-lock` before archive.
**Regression tests**: `TestWriteguard_ArchivedSession_DeniesWithClearMessage` in `internal/cmd/hook/writeguard_test.go`

### SCAR-013: Ghost Dirs and Already-Archived Session Wrap

**Category**: integration_failure (edge cases)
**What failed**: Three unguarded wrap edge cases: wrapping already-archived session, ghost live directories after cross-device rename, existing archive targets.
**Fix location**: `internal/cmd/session/wrap.go`
**Defensive pattern**: Pre-lock archived guard, post-rename ghost removal, graceful archive target handling.
**Regression tests**: `TestWrapAlreadyArchived`, `TestWrapNoGhostDirectory` in `internal/cmd/session/wrap_test.go`

### SCAR-014: Phantom Status Values in Session FSM

**Category**: schema_evolution
**What failed**: Values `COMPLETE` and `COMPLETED` written to `SESSION_CONTEXT.md` but not in FSM. Sessions became undetectable.
**Fix location**: `internal/session/status.go:48-51`
**Defensive pattern**: `NormalizeStatus()` with alias map wired into `ParseContext()` and `readStatusFromFrontmatter()`.
**Regression tests**: `TestParseContext_NormalizesPhantomStatus`, `TestParseContext_NormalizesComplete` in `internal/session/context_test.go`

### SCAR-015: stdout Pollution from Shell Log Functions Corrupting Manifests

**Category**: data_corruption
**What failed**: Shell log functions used `echo` without `>&2`, embedding log messages in manifest JSON keys via `$()` substitution.
**Fix location**: Shell sync scripts
**Defensive pattern**: ALL shell log functions redirect with `>&2`.
**Regression tests**: `TestSCAR015_ShellScripts_StderrLogging` (structural scan) in `internal/materialize/scar_regression_test.go`

### SCAR-016: Bash Arithmetic Increment from Zero with set -e

**Category**: schema_evolution (shell behavior)
**What failed**: `((var++))` returns exit 1 when var is zero. With `set -euo pipefail`, scripts silently exited.
**Fix location**: All `.sh` files
**Defensive pattern**: `((var++)) || true` suffix on all arithmetic increments.
**Regression tests**: `TestSCAR016_ShellScripts_NoUnprotectedArithmeticIncrement` (structural scan) in `internal/materialize/scar_regression_test.go`

### SCAR-017: @skill-name Anti-Pattern in Agent Prompts

**Category**: configuration_drift
**What failed**: 195+ agent files used `@skill-name` syntax. CC has no `@mention` resolution; skills silently ignored.
**Fix location**: `internal/cmd/lint/lint.go:394` (rule `skill-at-syntax`, HIGH severity)
**Defensive pattern**: `ari lint` reports HIGH on any `@skill-name` in body content.
**Regression tests**: 14 tests in `internal/cmd/lint/lint_test.go`

### SCAR-018: context: fork Blocks Task Tool Access

**Category**: integration_failure
**What failed**: `/know` dromenon had `context: fork`. Forked commands run as subagents without Task tool. The Argus Pattern silently degraded.
**Fix location**: `mena/know/INDEX.dro.md` (absence of `context: fork`)
**Defensive pattern**: Dromena needing Task tool must NOT have `context: fork`.
**Regression tests**: `TestSCAR018_KnowDromenon_NoContextFork` in `internal/materialize/scar_regression_test.go`

### SCAR-019: Invalid Agent Colors in CC Palette

**Category**: configuration_drift
**What failed**: 13+ agents used colors outside CC's supported 8-value palette. CC silently ignored invalid colors.
**Fix location**: `internal/cmd/lint/lint.go:381-385` (rule `agent-invalid-color`)
**Defensive pattern**: `ari lint` enforces 8-value CC color palette.
**Regression tests**: Lint tests in `internal/cmd/lint/lint_test.go`

### SCAR-020: Session ID Not Passed to CLI Subprocesses from Dromena

**Category**: integration_failure
**What failed**: ALL session lifecycle dromena failed to pass session ID to CLI subprocesses. Bash cannot access LLM context.
**Fix location**: Session dromena: `mena/session/fray/INDEX.dro.md`, `mena/session/sos/INDEX.dro.md`
**Defensive pattern**: Session dromena must explicitly instruct session ID extraction and `-s` flag passing.
**Regression tests**: `TestSCAR020_SessionDromena_ExplicitSessionIDPassing` in `internal/materialize/scar_regression_test.go`

### SCAR-021: Cross-Rite Agents Materialized to Project Scope

**Category**: integration_failure (scope misrouting)
**What failed**: Global agents written to project `.claude/agents/`. CC's shadowing model made stale project copies mask user-scope updates.
**Fix location**: `materializeCrossRiteAgents()` removed from `internal/materialize/materialize.go:988`
**Defensive pattern**: Cross-rite agents exclusively use user-scope (`~/.claude/agents/`). Never project level.
**Regression tests**: `TestSCAR021_CrossRiteAgents_ProjectScopeExclusion` in `internal/materialize/scar_regression_test.go`

### SCAR-022: Provenance Schema Rejects Abbreviated SHA256

**Category**: data_corruption (test fixture)
**What failed**: Test fixtures used abbreviated SHA256 values. Schema validation requires full 64-character hex with `sha256:` prefix.
**Fix location**: `internal/materialize/mena/mena_test.go` (test fixtures)
**Defensive pattern**: All provenance test fixtures use full zero-padded 64-character checksums.

### SCAR-023: Template Path Resolution Fails for Knossos Self-Hosting

**Category**: configuration_drift
**What failed**: `SourceProject` template resolution only checked `templates/sections/`. Self-hosted knossos uses `knossos/templates/`.
**Fix location**: `internal/materialize/source/resolver.go:273-280` (fallback logic)
**Defensive pattern**: Fallback to `knossos/templates/sections/` when `templates/sections/` absent.
**Regression tests**: `TestSCAR023_TemplatePathResolution_SelfHosting`, `TestSCAR023_TemplatePathResolution_StandardProject` in `internal/materialize/source/source_test.go`

### SCAR-024: Stale Throughline IDs After Rite Switch

**Category**: integration_failure (scope misrouting)
**What failed**: On rite switch, `.throughline-ids.json` retained agent IDs from previous rite. Pythia wasted turns on invalid resume attempts.
**Fix location**: `internal/materialize/materialize.go:660` (`cleanupThroughlineIDs()`)
**Defensive pattern**: On rite switch, remove stale `.throughline-ids.json` from all session directories.
**Regression tests**: 5 tests in `internal/materialize/throughline_cleanup_test.go`

### SCAR-025: Deleted Files in User Scope Sync Not Handled

**Category**: data_corruption (silent failure)
**What failed**: User-owned entries with stale manifest records caused sync to skip deleted files forever.
**Fix location**: `internal/materialize/mena.go` (three sync function bodies now call `os.Stat` before comparing)
**Defensive pattern**: Always `os.Stat` before checksum comparison. Missing files with stale entries get cleared.
**Regression tests**: `internal/materialize/mena_test.go`

### SCAR-026: Revert — Moirai Delegation in /start and /sprint

**Category**: historical_boundary
**What failed**: Adding Moirai delegation required writeguard `outputBlock` to include `additionalContext`, creating coupling.
**Fix location**: `internal/cmd/hook/writeguard.go:outputBlock()` (absence of `additionalContext`)
**Defensive pattern**: Writeguard output block has no `additionalContext` field.
**Regression tests**: Writeguard test suite

### SCAR-027: Ephemeral Skill in Shared Mena

**Category**: historical_boundary
**What failed**: Session artifact added to `rites/shared/mena/` as legomenon. Shared mena is permanent platform knowledge; session artifacts become permanent.
**Fix location**: `internal/cmd/lint/lint.go:624-662` (rule `session-artifact-in-shared-mena`)
**Defensive pattern**: Shared mena is for permanent platform knowledge only. Session artifacts belong in `.sos/wip/`.
**Regression tests**: `TestSCAR027_SharedMena_NoSessionArtifacts` in `internal/materialize/scar_regression_test.go`

### GO-001/002/003: Three Content Rewriting Bypass Paths

**Category**: data_corruption (pipeline bypass)
**What failed**: `RewriteMenaContentPaths` was implemented but omitted from three code paths in `mena/engine.go`, `userscope/sync_mena.go` standalone copy, and `userscope/sync_mena.go` shared fallback. Stale `.lego.md`/`.dro.md` link targets survived into materialized output.
**Fix locations**: `internal/materialize/mena/engine.go:144`, `internal/materialize/mena/walker.go:83`, `internal/materialize/userscope/sync_mena.go:303,353`, `internal/materialize/mena/content_rewrite.go:50`
**Defensive pattern**: `TestSCAR_ContentRewriteNotBypassed` (compile-time + runtime guard) in `internal/materialize/mena/content_rewrite_test.go:279`.

## Category Coverage

| Category | Count | Scar IDs |
|----------|-------|----------|
| Race condition | 1 | SCAR-001 |
| Integration failure | 8 | SCAR-002, SCAR-006, SCAR-009, SCAR-012, SCAR-013, SCAR-018, SCAR-020, SCAR-021 |
| Data corruption | 6 | SCAR-004, SCAR-005, SCAR-015, SCAR-022, SCAR-025, GO-001/002/003 |
| Configuration drift | 5 | SCAR-011, SCAR-017, SCAR-019, SCAR-023, SCAR-008 |
| Idempotency failure | 1 | SCAR-003 |
| Schema evolution | 3 | SCAR-007, SCAR-014, SCAR-016 |
| Performance cliff | 2 | SCAR-008, SCAR-010 |
| Scope misrouting | 2 | SCAR-021, SCAR-024 |
| Historical boundary | 2 | SCAR-026, SCAR-027 |

Categories searched but not found: memory leak, nil pointer panic, network timeout (not applicable — no network I/O in core).

## Fix-Location Mapping

All 28 fix locations verified against current filesystem. See individual SCAR entries above for full path, function, and verification status.

**Compound fixes (multiple locations)**: SCAR-004 (2 locations), SCAR-012 (2 locations), SCAR-015 (4+ shell scripts), SCAR-016 (all .sh files), SCAR-020 (dromena files), GO-001/002/003 (4 locations).

**Note**: Prior version incorrectly listed SCAR-023 fix location as `source_test.go`. The actual production fix is at `internal/materialize/source/resolver.go:273-280`.

## Defensive Pattern Documentation

| Scar | Defensive Pattern | Regression Test |
|------|------------------|-----------------|
| SCAR-001 | Atomic flock-on-existing-fd; never remove+reopen | `TestManager_StaleLockReclamation*` |
| SCAR-002 | Never rename `.claude/`; per-file atomic writes | `TestSCAR002_*` |
| SCAR-003 | Mutation flags imply Force; `writeIfChanged` | `TestRematerializeMena_RepopulatesAfterWipe` |
| SCAR-004 | WARN log on non-expected provenance errors; never `_` discard | `TestSCAR004_*` |
| SCAR-005 | Selective write from managed-set; never `RemoveAll` | 6 tests in `selective_write_test.go` |
| SCAR-006 | `sharedRitesBase` from `KnossosHome()` | 4 tests in `satellite_mena_test.go` |
| SCAR-008 | No `async: true` on sub-100ms hooks | `TestSCAR008_*` |
| SCAR-010 | ALL hook RunE use `withTimeout()`; git subprocess use `exec.CommandContext` | `hook_test.go` |
| SCAR-011 | Session resolution via priority chain; never `.current-session` | `TestWriteguard_ParkedSession_*` |
| SCAR-014 | `NormalizeStatus()` at all session status read sites | `TestParseContext_Normalizes*` |
| SCAR-017 | `skill-at-syntax` lint rule at HIGH | 14 lint tests |
| SCAR-018 | Commands needing Task tool must NOT have `context: fork` | `TestSCAR018_*` |
| SCAR-020 | Dromena must explicitly instruct session ID + `-s` flag | `TestSCAR020_*` |
| SCAR-021 | Cross-rite agents user-scope only; never project `.claude/agents/` | `TestSCAR021_*` |
| SCAR-024 | On rite switch, clean `.throughline-ids.json` | 5 tests in `throughline_cleanup_test.go` |
| SCAR-027 | `session-artifact-in-shared-mena` lint rule | `TestSCAR027_*` |
| GO-001/002/003 | `RewriteMenaContentPaths` in ALL mena copy paths | `TestSCAR_ContentRewriteNotBypassed` |

## Agent-Relevance Tagging

**Heat map**: `integration-engineer` appears in 27 of 28 scars; `context-architect` in 9; `compatibility-tester` in 2.

| Scar | Agent Role(s) | Why |
|------|--------------|-----|
| SCAR-001 | integration-engineer | Lock code must use atomic flock |
| SCAR-002 | integration-engineer | Never rename `.claude/` |
| SCAR-003 | integration-engineer | Mutation flags must imply Force |
| SCAR-004 | integration-engineer | Never `_` discard provenance errors |
| SCAR-005 | integration-engineer, context-architect | `RemoveAll` on user dirs destroys content |
| SCAR-006 | integration-engineer, compatibility-tester | Satellite paths differ from core |
| SCAR-007 | integration-engineer, context-architect | Mixed dro/lego dirs break skill resolution |
| SCAR-009 | integration-engineer | CC hook format is nested; flat is silently rejected |
| SCAR-010 | integration-engineer | ALL hooks need timeout; ALL git subprocess need context |
| SCAR-017 | context-architect, integration-engineer | `@skill-name` is not a CC primitive |
| SCAR-018 | context-architect, integration-engineer | `context: fork` removes Task tool access |
| SCAR-020 | integration-engineer | LLM context not accessible to bash; must pass `-s` flag |
| SCAR-021 | integration-engineer, context-architect | Cross-rite agents in project scope cause shadowing |
| SCAR-024 | integration-engineer | Rite switch must clean stale throughline IDs |
| SCAR-027 | integration-engineer, context-architect | Shared mena = permanent; session artifacts in `.sos/wip/` |

## Knowledge Gaps

1. **GO-001/002/003 are not formally numbered** in the SCAR catalog. Recommend formal assignment as SCAR-028/029/030.
2. **9 of 27 SCARs lack dedicated behavioral regression tests** — they have structural scans, lint enforcement, or integration tests instead.
3. **SCAR-010 binary mismatch trap**: No automated guard for build-vs-install binary divergence (`go build ./cmd/ari` writes `./ari` locally; CC uses PATH binary).
4. **CI infrastructure scar (GITHUB_TOKEN/GoReleaser)**: Documented in `.sos/land/scar-tissue.md` but not numbered in SCAR catalog.

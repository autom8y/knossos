---
domain: scar-tissue
generated_at: "2026-03-08T21:08:37Z"
expires_after: "7d"
source_scope:
  - "./cmd/**/*.go"
  - "./internal/**/*.go"
  - "./go.mod"
generator: theoros
source_hash: "dbf81b8"
confidence: 0.95
format_version: "1.0"
update_mode: "full"
incremental_cycle: 0
max_incremental_cycles: 3
land_sources:
  - ".sos/land/scar-tissue.md"
land_hash: "a360816ba21cf78956942c0c663b046fa7fe9b1c84ca51202a1143da710d45c3"
---

# Codebase Scar Tissue

## Failure Catalog

29 SCAR entries cataloged with regression tests, defensive guards, and cross-validated against `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`, `/Users/tomtenuta/Code/knossos/internal/materialize/source/source_test.go`, `/Users/tomtenuta/Code/knossos/internal/agent/regenerate_test.go`, `/Users/tomtenuta/Code/knossos/internal/materialize/mena/content_rewrite_test.go`, and git history through commit `dbf81b8`. Three additional post-catalog items (GO-001/002/003 content rewriting bypasses) documented but not yet formally numbered. Three CI/shell scars (LOCK-001, LOCK-003, STATE-001) documented from legacy shell era.

### SCAR-001: TOCTOU Race in Stale Lock Reclamation

**Category**: race_condition
**What failed**: Stale lock reclamation used `os.Remove(lockfile)` then `os.OpenFile(lockfile)`. A competing process could acquire the lock between Remove and reopen.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/lock/lock.go:117-134`
**Defensive pattern**: `syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)` on the already-open fd; never Remove+reopen.
**Regression tests**: `TestManager_StaleLockReclamation`, `TestManager_StaleLockReclamation_Concurrent` in `/Users/tomtenuta/Code/knossos/internal/lock/lock_test.go`

### SCAR-002: StagedMaterialize Freeze in Claude Code

**Category**: integration_failure
**What failed**: `StagedMaterialize` renamed `.claude/` to `.claude.bak/`. CC's file watcher lost track, causing hard freezes during materialization. Fix: commit `95cf0bc` removed the method entirely.
**Fix location**: Method removed from `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`; all writes use `writeIfChanged()` in `/Users/tomtenuta/Code/knossos/internal/fileutil/fileutil.go`.
**Defensive pattern**: Never rename `.claude/`. Per-file atomic `writeIfChanged()` prevents watcher triggers. `AtomicWriteFile` (temp-file-then-rename) for all state file writes.
**Regression tests**: `TestSCAR002_StagedMaterializeAbsent` (reflection-based absence guard), `TestSCAR002_MaterializeWithOptions_NoClaudeRename` in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-003: Idempotency Guard Bypass on Mutation Flags

**Category**: idempotency_failure
**What failed**: `--remove-all` and `--promote-all` did not set `Force: true`. Pipeline short-circuited on matching `ACTIVE_RITE`, leaving directories empty after mena operations.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:194`; `/Users/tomtenuta/Code/knossos/internal/cmd/sync/materialize.go`
**Defensive pattern**: Mutation flags imply `--force`. `writeIfChanged()` prevents spurious CC watcher triggers.
**Regression tests**: `TestRematerializeMena_RepopulatesAfterWipe` in `/Users/tomtenuta/Code/knossos/internal/materialize/routing_test.go`

### SCAR-004: Silent Error Discard at Provenance Load Sites

**Category**: data_corruption (silent failure)
**What failed**: `LoadOrBootstrap` and `DetectDivergence` errors were discarded via blank identifier `_` in `MaterializeMinimal` and `MaterializeWithOptions`. Filesystem permission errors and corrupted manifest conditions were completely masked. Pipeline bootstrapped from empty state, potentially overwriting user content. Commits `e5655a9`, `e7277cf`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` (load sites at lines 242-258 and 388-403)
**Defensive pattern**: WARN log on non-file-not-found errors; parse/validation errors abort pipeline. Never `_` discard provenance errors.
**Regression tests**: `TestSCAR004_CorruptProvenanceManifest_PropagatesError`, `TestSCAR004_InvalidSchemaProvenanceManifest_PropagatesError` in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-005: Destructive os.RemoveAll on Agents/Commands/Skills Directories

**Category**: data_corruption
**What failed**: Materialization called `os.RemoveAll` on `agents/`, `commands/`, `skills/` before rewriting. User-created content was destroyed on every sync.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize_agents.go:25` (comment: "selective — do NOT RemoveAll")
**Defensive pattern**: Build managed-set from provenance manifest; remove only managed files; preserve everything outside manifest scope.
**Regression tests**: 6 tests in `/Users/tomtenuta/Code/knossos/internal/materialize/selective_write_test.go`

### SCAR-006: Shared Mena Drop for Satellite-Local Rites

**Category**: integration_failure (silent)
**What failed**: `materializeMena()` derived `ritesBase` from `filepath.Dir(resolved.RitePath)`. For satellite-local rites, this missed `$KNOSSOS_HOME/rites/`. Shared mena was silently dropped. Commit `89b109c`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize_mena.go:73-87`
**Defensive pattern**: `sharedRitesBase` computed from `resolver.KnossosHome()`; never infer knossos home from rite path alone.
**Regression tests**: 4 tests in `/Users/tomtenuta/Code/knossos/internal/materialize/satellite_mena_test.go`

### SCAR-007: Mixed Dro/Lego Directories Block Skill Resolution

**Category**: schema_evolution
**What failed**: Directories containing both `INDEX.dro.md` and `*.lego.md` were routed entirely as dromena. Legomena siblings were invisible to `skills:` frontmatter preloading.
**Fix location**: Affected mena directories split into separate `*-ref` (dromenon) and `*-catalog` (legomenon) directories. Four rites fixed during Cross-Rite Rollout.
**Defensive pattern**: `ari lint` enforces separation. Never place `.lego.md` siblings inside a dromenon directory.
**Regression tests**: `ari lint` validation

### SCAR-008: Budget Hook Async Log Spam

**Category**: performance_cliff
**What failed**: Budget hook was configured with `async: true`. It fires on every `PostToolUse` event and completes in <5ms. Running async created an "Async hook completed" notification for every single tool call, flooding logs. Commit `85d66d5`.
**Fix location**: `/Users/tomtenuta/Code/knossos/config/hooks.yaml` (absence of `async: true` on `ari hook budget` entry)
**Defensive pattern**: Sub-100ms hooks must be synchronous. `async: true` reserved for hooks with meaningful latency.
**Regression tests**: `TestSCAR008_BudgetHook_MustNotBeAsync` in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-009: Hooks Generated in Wrong Nested-Matcher Format

**Category**: integration_failure
**What failed**: `settings.local.json` hooks used flat `{command, matcher}` format. CC expects nested `{matcher, hooks: [{type, command}]}`. Hooks were silently rejected. Commit `bb1666a`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/hooks.go`
**Defensive pattern**: Hook generation always uses CC's nested-matcher format.
**Regression tests**: `/Users/tomtenuta/Code/knossos/internal/materialize/hooks_test.go`

### SCAR-010: Missing Timeout on Hook and Git Subprocess Commands

**Category**: performance_cliff
**What failed**: (1) `budget` hook `RunE` read stdin without timeout — blocked indefinitely after hot rite switch. (2) `getGitStatusQuick()` used `exec.Command` without context — orphaned processes. (3) All 20+ git calls in `worktree/` package used bare `exec.Command`. (4) Hook subprocesses in `worktreeseed.go`, `worktreeremove.go`, `session/create.go`, `session/status.go` all lacked timeout context. Full sweep commit `6304dc5`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/budget.go:62`; `/Users/tomtenuta/Code/knossos/internal/cmd/hook/autopark.go:156`; `/Users/tomtenuta/Code/knossos/internal/worktree/git.go:17-21` (`gitCmdCtx` helper with 30s timeout); additional sites in `internal/cmd/hook/worktreeseed.go`, `worktreeremove.go`, `internal/cmd/session/create.go`, `internal/cmd/session/status.go`
**Defensive pattern**: ALL hook `RunE` use `withTimeout()`. ALL git subprocesses use `exec.CommandContext` with deadline. Zero bare `exec.Command` calls permitted in production code (verified: grep confirms none remain).
**Regression tests**: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/hook_test.go`

### SCAR-011: Session Writeguard Used .current-session (Last Production Caller)

**Category**: configuration_drift
**What failed**: `writeguard.go:isMoiraiLockHeld()` was the last caller of the deprecated `.current-session` file. When `.current-session` was removed from the session model, this failed silently, denying all Moirai writes.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go:360`
**Defensive pattern**: Session resolution via standard priority chain. Never read `.current-session` directly.
**Regression tests**: `TestWriteguard_ParkedSession_MoiraiLockAllow`, `TestWriteguard_ParkedSession_NoLock`, `TestWriteguard_ParkedSession_StaleLock` in `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard_test.go`

### SCAR-012: Archived Session Writeguard Gap

**Category**: integration_failure
**What failed**: Archived sessions received a generic "Use Moirai" denial instead of an archived-session explanation. CC agents entered infinite Moirai retry loops with no resolution path.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go:192`; `/Users/tomtenuta/Code/knossos/internal/session/resolve.go:96`
**Defensive pattern**: Check `isSessionArchived()` after `isMoiraiLockHeld()` fails. Remove `.moirai-lock` before archiving.
**Regression tests**: `TestWriteguard_ArchivedSession_DeniesWithClearMessage` in `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard_test.go`

### SCAR-013: Ghost Dirs and Already-Archived Session Wrap

**Category**: integration_failure (edge cases)
**What failed**: Three unguarded wrap edge cases: (1) wrapping an already-archived session, (2) ghost live directories after cross-device rename, (3) existing archive targets.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/cmd/session/wrap.go`
**Defensive pattern**: Pre-lock archived guard, post-rename ghost removal, graceful archive target handling.
**Regression tests**: `TestWrapAlreadyArchived`, `TestWrapNoGhostDirectory` in `/Users/tomtenuta/Code/knossos/internal/cmd/session/wrap_test.go`

### SCAR-014: Phantom Status Values in Session FSM

**Category**: schema_evolution
**What failed**: Values `COMPLETE` and `COMPLETED` were written to `SESSION_CONTEXT.md` but not registered in the FSM. Sessions with these statuses became undetectable by the session management system.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/session/status.go:48-51`
**Defensive pattern**: `NormalizeStatus()` with alias map wired into `ParseContext()` and `readStatusFromFrontmatter()`.
**Regression tests**: `TestParseContext_NormalizesPhantomStatus`, `TestParseContext_NormalizesComplete` in `/Users/tomtenuta/Code/knossos/internal/session/context_test.go`

### SCAR-015: stdout Pollution from Shell Log Functions Corrupting Manifests

**Category**: data_corruption
**What failed**: Shell log functions (`log()`, `log_success()`, `log_info()`) in sync scripts used bare `echo` without `>&2`. When called inside data-returning functions, log messages were captured by `$()` substitution and embedded in manifest JSON keys. Commit `c8b551a`.
**Fix location**: Shell sync scripts (`sync-user-agents.sh`, `sync-user-commands.sh`, `sync-user-hooks.sh`, `sync-user-skills.sh`)
**Defensive pattern**: ALL shell log functions must redirect with `>&2`. Comment in code: "IMPORTANT: All log functions MUST output to stderr to avoid polluting captured stdout."
**Regression tests**: `TestSCAR015_ShellScripts_StderrLogging` (structural scan of sync script dirs) in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-016: Bash Arithmetic Increment from Zero with set -e

**Category**: schema_evolution (shell behavior)
**What failed**: `((var++))` returns exit code 1 when `var` is zero. With `set -euo pipefail` enabled, this caused immediate silent script exit. Scripts appeared to hang in `COMMITTING` phase but were actually exiting at collision detection. Commit `1641792`.
**Fix location**: All `.sh` files — `swap-team.sh`, `lib/sync/sync-core.sh`, `lib/sync/sync-manifest.sh`, `lib/sync/merge/merge-docs.sh`
**Defensive pattern**: `((var++)) || true` suffix on all arithmetic increments.
**Regression tests**: `TestSCAR016_ShellScripts_NoUnprotectedArithmeticIncrement` (structural scan of all `.sh` files) in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-017: @skill-name Anti-Pattern in Agent Prompts

**Category**: configuration_drift
**What failed**: 195+ agent files used `@skill-name` syntax (e.g., `@standards`, `@file-verification`). CC has no `@mention` resolution mechanism; skills were silently ignored. Also present in Skills Reference sections of 12 pythia files fixed in commit `57f3601`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/cmd/lint/lint.go:394` (rule `skill-at-syntax`, HIGH severity)
**Defensive pattern**: `ari lint` reports HIGH on any `@skill-name` in agent body content. Correct form: use skill names from `skills:` frontmatter.
**Regression tests**: 14 tests in `/Users/tomtenuta/Code/knossos/internal/cmd/lint/lint_test.go`; inline check `t.Errorf("output must not contain @ prefix (SCAR-017 anti-pattern)")` in `/Users/tomtenuta/Code/knossos/internal/agent/regenerate_test.go:376`

### SCAR-018: context: fork Blocks Task Tool Access

**Category**: integration_failure
**What failed**: `/know` dromenon had `context: fork` set. Forked slash commands run as subagents which cannot use the Task tool — only the main thread has Task tool access. The Argus Pattern silently degraded from agent dispatch to in-context observation (30-50 Read calls per run). Commit `4d92db4`; spike documented at `/Users/tomtenuta/Code/knossos/docs/spikes/SPIKE-know-fork-dispatch-impossibility.md`.
**Fix location**: `/Users/tomtenuta/Code/knossos/mena/know/INDEX.dro.md` (absence of `context: fork`)
**Defensive pattern**: Dromena needing Task tool must NOT have `context: fork`. Verified by secondary scan of all platform and shared dromena.
**Regression tests**: `TestSCAR018_KnowDromenon_NoContextFork` in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-019: Invalid Agent Colors in CC Palette

**Category**: configuration_drift
**What failed**: 13+ agents used colors outside CC's supported 8-value palette. CC silently ignored invalid colors.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/cmd/lint/lint.go:381-385` (rule `agent-invalid-color`)
**Defensive pattern**: `ari lint` enforces 8-value CC color palette.
**Regression tests**: Lint tests in `/Users/tomtenuta/Code/knossos/internal/cmd/lint/lint_test.go`

### SCAR-020: Session ID Not Passed to CLI Subprocesses from Dromena

**Category**: integration_failure
**What failed**: ALL session lifecycle dromena (`/continue`, `/fray`, `/park`, `/wrap`, `/start`) failed to pass session ID to CLI subprocess commands. Root cause: `SessionStart` hook injects session ID into LLM context as a markdown table, but Bash subprocesses cannot access LLM context. `GetSessionID()` fell back to `FindActiveSession()` which only found sessions with `status: ACTIVE` in `SESSION_CONTEXT.md`. Commit `6f35325`.
**Fix location**: Session dromena: `/Users/tomtenuta/Code/knossos/mena/session/fray/INDEX.dro.md`, `/Users/tomtenuta/Code/knossos/mena/session/sos/INDEX.dro.md`
**Defensive pattern**: Session dromena must explicitly instruct the LLM to extract session ID from hook-injected context table and pass it via `-s` flag to CLI subprocesses.
**Regression tests**: `TestSCAR020_SessionDromena_ExplicitSessionIDPassing` in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-021: Cross-Rite Agents Materialized to Project Scope

**Category**: integration_failure (scope misrouting)
**What failed**: Global agents (`pythia`, `moirai`, `context-engineer`, `theoros`) were written to project `.claude/agents/` alongside rite agents. CC's shadowing model treats project-level as overrides — stale project copies masked user-scope updates after upgrades. Orphan accumulation across rite switches. Commit `7ef0213`.
**Fix location**: `materializeCrossRiteAgents()` removed from `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go`; cross-rite agents now exclusively routed to user scope (`~/.claude/agents/`).
**Defensive pattern**: Cross-rite agents exclusively use user-scope. Never project `.claude/agents/`.
**Regression tests**: `TestSCAR021_CrossRiteAgents_ProjectScopeExclusion` in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-022: Provenance Schema Rejects Abbreviated SHA256

**Category**: data_corruption (test fixture)
**What failed**: Test fixtures used abbreviated SHA256 values. Schema validation requires full 64-character hex with `sha256:` prefix. Tests failed unexpectedly when fixtures were reused. Commit `dacc620`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/mena/mena_test.go` (test fixtures)
**Defensive pattern**: All provenance test fixtures use full zero-padded 64-character checksums with `sha256:` prefix.

### SCAR-023: Template Path Resolution Fails for Knossos Self-Hosting

**Category**: configuration_drift
**What failed**: `SourceProject` template resolution only checked `$PROJECT/templates/sections/`. In the knossos self-hosting case (where knossos is the project), templates live at `$PROJECT/knossos/templates/`. No fallback existed; template resolution silently failed. Commit `bff1293`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/source/resolver.go:273-280` (fallback: when `templates/sections/` absent, check `knossos/templates/sections/`)
**Defensive pattern**: Template resolver always checks `knossos/templates/sections/` as fallback.
**Regression tests**: `TestSCAR023_TemplatePathResolution_SelfHosting`, `TestSCAR023_TemplatePathResolution_StandardProject` in `/Users/tomtenuta/Code/knossos/internal/materialize/source/source_test.go`

### SCAR-024: Stale Throughline IDs After Rite Switch

**Category**: integration_failure (scope misrouting)
**What failed**: On rite switch, `.throughline-ids.json` retained agent IDs from the previous rite. Pythia wasted turns attempting to resume invalid agents from the prior context.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go:660` (`cleanupThroughlineIDs()`)
**Defensive pattern**: On rite switch, remove stale `.throughline-ids.json` from all session directories.
**Regression tests**: 5 tests in `/Users/tomtenuta/Code/knossos/internal/materialize/throughline_cleanup_test.go`

### SCAR-025: Deleted Files in User Scope Sync Not Handled

**Category**: data_corruption (silent failure)
**What failed**: User-owned entries with stale manifest records caused sync to permanently skip re-syncing deleted files. Files deleted from source remained absent from destination forever without error.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/mena.go` (three sync function bodies call `os.Stat` before checksum comparison)
**Defensive pattern**: Always `os.Stat` before checksum comparison. Missing files with stale manifest entries get cleared and re-synced.
**Regression tests**: `/Users/tomtenuta/Code/knossos/internal/materialize/mena_test.go`

### SCAR-026: Revert — Moirai Delegation in /start and /sprint

**Category**: historical_boundary
**What failed**: Adding Moirai delegation required `writeguard.outputBlock()` to include `additionalContext` field with a specific `Task(moirai, "...")` invocation pattern, creating coupling between writeguard and session orchestration. Reverted (commit `104bb12`) because the coupling was unsound.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard.go:outputBlock()` (absence of `additionalContext` field)
**Defensive pattern**: Writeguard output block must not contain `additionalContext`. Delegation hints belong in agent prompts, not infrastructure responses.
**Regression tests**: Writeguard test suite in `/Users/tomtenuta/Code/knossos/internal/cmd/hook/writeguard_test.go`

### SCAR-027: Ephemeral Skill in Shared Mena

**Category**: historical_boundary
**What failed**: An ARCH-REVIEW session artifact (`materialize-review` skill) was accidentally added to `rites/shared/mena/` as a legomenon. Shared mena is permanent platform knowledge; session artifacts become permanent once merged. Reverted via commit `f0971e4`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/cmd/lint/lint.go:621` (rule `session-artifact-in-shared-mena`, checks frontmatter for session-specific markers)
**Defensive pattern**: Shared mena is for permanent platform knowledge only. Session artifacts belong in `.sos/wip/`. MEMORY.md Golden Rule: "NEVER add session artifacts to rites/shared/mena/."
**Regression tests**: `TestSCAR027_SharedMena_NoSessionArtifacts` in `/Users/tomtenuta/Code/knossos/internal/materialize/scar_regression_test.go`

### SCAR-028: MCP Servers Written to Wrong CC Config File

**Category**: integration_failure (silent)
**What failed**: `materializeSettingsWithManifest()` writes `mcp_servers` from rite manifests into `.claude/settings.local.json` under the `mcpServers` key. CC does NOT read MCP server definitions from this file. CC reads MCPs from three sources: `~/.claude.json` (user-global), `~/.mcp.json` (user-global), and `.mcp.json` (project root). MCP servers declared in rite manifests are materialized correctly but into a dead-letter location — they are never connected.
**Evidence**: `go-semantic` appeared connected despite being in `settings.local.json`, but was actually loaded from `~/.mcp.json`. `duckdb` and `github` in `settings.local.json` with no entry in `~/.mcp.json` or `.mcp.json` — neither connects. Confirmed via `ListMcpResourcesTool` returning "Server 'duckdb' not found".
**Affected rites**: All rites with `mcp_servers` in manifest.yaml: `ecosystem` (go-semantic, terraform), `10x-dev` (github), `data-analyst` (duckdb). Only work if the user ALSO manually configures the server in `~/.mcp.json` or `~/.claude.json`.
**Fix location**: `/Users/tomtenuta/Code/knossos/internal/materialize/materialize_settings.go` — `mergeMCPServers` must also (or instead) write to `.mcp.json` at project root.
**Defensive pattern**: MCP server materialization target must be `.mcp.json` (project root), not `.claude/settings.local.json`. Union merge semantics preserved. Provenance tracking for `.mcp.json`.
**Regression tests**: NEEDED — `TestSCAR028_MCPServers_WrittenToMcpJson` verifying `.mcp.json` is created/updated and CC can discover the tools.
**CC config file reference**:
- `~/.claude.json` → user-global MCPs (read by CC)
- `~/.mcp.json` → user-global MCPs (read by CC)
- `.mcp.json` → project-level MCPs (read by CC)
- `.claude/settings.json` → team settings, hooks, permissions (NOT MCPs)
- `.claude/settings.local.json` → personal settings overrides (NOT MCPs)
- `~/.claude/settings.local.json` → user-global permissions, `enableAllProjectMcpServers`, `enabledMcpjsonServers`

### GO-001/002/003: Three Content Rewriting Bypass Paths (Unnumbered)

**Category**: data_corruption (pipeline bypass)
**What failed**: `RewriteMenaContentPaths` was implemented but omitted from three code paths: (1) `mena/engine.go` standalone file write path, (2) `userscope/sync_mena.go` `walkMenaEntryFS` function, (3) `userscope/sync_mena.go` `syncMenaStandalone` function. Stale `.lego.md`/`.dro.md` link targets and backtick references survived into materialized output. Also: `RewriteMenaContentPaths` was unexported (`rewriteMenaContentPaths`), preventing external wiring. Commits `3e82fa6`, `137d4ca`, `3614eec`.
**Fix locations**: `/Users/tomtenuta/Code/knossos/internal/materialize/mena/engine.go:144`; `/Users/tomtenuta/Code/knossos/internal/materialize/mena/walker.go:83`; `/Users/tomtenuta/Code/knossos/internal/materialize/userscope/sync_mena.go:303,353`; `/Users/tomtenuta/Code/knossos/internal/materialize/mena/content_rewrite.go:50` (export rename)
**Defensive pattern**: `TestSCAR_ContentRewriteNotBypassed` in `/Users/tomtenuta/Code/knossos/internal/materialize/mena/content_rewrite_test.go:288` — compile-time reference guard. The function name appearing in a test ensures deletion is caught at build time.
**Note**: Not yet formally numbered. Recommend SCAR-028/029/030.

---

## Category Coverage

| Category | Count | SCAR IDs |
|----------|-------|----------|
| Integration failure | 10 | SCAR-002, SCAR-006, SCAR-009, SCAR-012, SCAR-013, SCAR-018, SCAR-020, SCAR-021, SCAR-024, SCAR-028 |
| Data corruption | 6 | SCAR-004, SCAR-005, SCAR-015, SCAR-022, SCAR-025, GO-001/002/003 |
| Configuration drift | 5 | SCAR-008, SCAR-011, SCAR-017, SCAR-019, SCAR-023 |
| Schema evolution | 3 | SCAR-007, SCAR-014, SCAR-016 |
| Performance cliff | 2 | SCAR-008, SCAR-010 |
| Scope misrouting | 2 | SCAR-021, SCAR-024 |
| Historical boundary | 2 | SCAR-026, SCAR-027 |
| Race condition | 1 | SCAR-001 |
| Idempotency failure | 1 | SCAR-003 |

Categories searched but not found: memory leak, nil pointer panic, network timeout (not applicable — no network I/O in core pipeline), infinite loop.

Note: SCAR-008 spans both configuration_drift and performance_cliff categories.

---

## Fix-Location Mapping

All 28 fix locations verified against current filesystem at HEAD (`dbf81b8`).

| SCAR | Fix Location | Status |
|------|-------------|--------|
| SCAR-001 | `internal/lock/lock.go:117-134` | Verified |
| SCAR-002 | Method absent from `internal/materialize/materialize.go` | Verified (grep confirms no StagedMaterialize) |
| SCAR-003 | `internal/materialize/materialize.go:194` | Verified |
| SCAR-004 | `internal/materialize/materialize.go:242-258, 388-403` | Verified |
| SCAR-005 | `internal/materialize/materialize_agents.go:25` | Verified |
| SCAR-006 | `internal/materialize/materialize_mena.go:73-87` | Verified |
| SCAR-007 | Mena directory structure (split applied) | Structural |
| SCAR-008 | `config/hooks.yaml` (async absent on budget entry) | Verified |
| SCAR-009 | `internal/materialize/hooks.go` | Verified |
| SCAR-010 | `internal/worktree/git.go:17-21` + 5 additional hook/session files | Verified (zero bare exec.Command in production) |
| SCAR-011 | `internal/cmd/hook/writeguard.go:360` | Verified |
| SCAR-012 | `internal/cmd/hook/writeguard.go:192`, `internal/session/resolve.go:96` | Verified |
| SCAR-013 | `internal/cmd/session/wrap.go` | Verified |
| SCAR-014 | `internal/session/status.go:48-51` | Verified |
| SCAR-015 | Shell sync scripts (4 files) | Structural scan |
| SCAR-016 | All `.sh` files (|| true pattern) | Structural scan |
| SCAR-017 | `internal/cmd/lint/lint.go:394` | Verified |
| SCAR-018 | `mena/know/INDEX.dro.md` (context: fork absent) | Verified |
| SCAR-019 | `internal/cmd/lint/lint.go:381-385` | Verified |
| SCAR-020 | `mena/session/fray/INDEX.dro.md`, `mena/session/sos/INDEX.dro.md` | Verified |
| SCAR-021 | `internal/materialize/materialize.go` (materializeCrossRiteAgents absent) | Verified |
| SCAR-022 | `internal/materialize/mena/mena_test.go` (test fixtures) | Verified |
| SCAR-023 | `internal/materialize/source/resolver.go:273-280` | Verified |
| SCAR-024 | `internal/materialize/materialize.go:660` | Verified |
| SCAR-025 | `internal/materialize/mena.go` (3 os.Stat call sites) | Verified |
| SCAR-026 | `internal/cmd/hook/writeguard.go:outputBlock()` (additionalContext absent) | Verified |
| SCAR-027 | `internal/cmd/lint/lint.go:621` | Verified |
| SCAR-028 | `internal/materialize/materialize_settings.go` (writes to `settings.local.json` instead of `.mcp.json`) | **OPEN** |
| GO-001/002/003 | `mena/engine.go:144`, `mena/walker.go:83`, `userscope/sync_mena.go:303,353`, `mena/content_rewrite.go:50` | Verified |

**Compound fixes (multiple locations)**: SCAR-004 (2 load sites), SCAR-010 (6 files), SCAR-012 (2 files), SCAR-015 (4 shell scripts), SCAR-016 (all .sh files), SCAR-020 (2 dromena files), GO-001/002/003 (4 locations).

---

## Defensive Pattern Documentation

| SCAR | Pattern | Guard Type | Test |
|------|---------|-----------|------|
| SCAR-001 | Atomic flock-on-existing-fd; never Remove+reopen | Behavioral | `TestManager_StaleLockReclamation*` |
| SCAR-002 | Never rename `.claude/`; per-file `writeIfChanged` | Behavioral + Structural | `TestSCAR002_*` |
| SCAR-003 | Mutation flags imply `Force: true` | Behavioral | `TestRematerializeMena_RepopulatesAfterWipe` |
| SCAR-004 | WARN log on non-expected provenance errors; never `_` discard | Behavioral | `TestSCAR004_*` |
| SCAR-005 | Selective write from managed-set; never `RemoveAll` | Behavioral | 6 tests in `selective_write_test.go` |
| SCAR-006 | `sharedRitesBase` from `KnossosHome()` only | Behavioral | 4 tests in `satellite_mena_test.go` |
| SCAR-007 | `ari lint` enforces dro/lego directory separation | Lint | `ari lint` |
| SCAR-008 | No `async: true` on sub-100ms hooks | Structural | `TestSCAR008_BudgetHook_MustNotBeAsync` |
| SCAR-009 | Nested-matcher format in hook generation | Behavioral | `hooks_test.go` |
| SCAR-010 | ALL hooks use `withTimeout()`; ALL git subprocess use `CommandContext` | Behavioral | `hook_test.go` |
| SCAR-011 | Session resolution via priority chain; never `.current-session` | Behavioral | `TestWriteguard_ParkedSession_*` |
| SCAR-014 | `NormalizeStatus()` at all status read sites | Behavioral | `TestParseContext_Normalizes*` |
| SCAR-015 | All shell log functions use `>&2` | Structural scan | `TestSCAR015_*` |
| SCAR-016 | `((var++)) || true` on all arithmetic increments | Structural scan | `TestSCAR016_*` |
| SCAR-017 | `skill-at-syntax` lint rule at HIGH severity | Lint + Behavioral | 14 lint tests; `regenerate_test.go` |
| SCAR-018 | Dromena needing Task tool must NOT have `context: fork` | Structural | `TestSCAR018_*` |
| SCAR-019 | `agent-invalid-color` lint rule | Lint | Lint tests |
| SCAR-020 | Dromena must explicitly instruct session ID + `-s` flag | Structural | `TestSCAR020_*` |
| SCAR-021 | Cross-rite agents user-scope only; never project `.claude/agents/` | Behavioral | `TestSCAR021_*` |
| SCAR-023 | Template resolver falls back to `knossos/templates/sections/` | Behavioral | `TestSCAR023_*` |
| SCAR-024 | Rite switch cleans `.throughline-ids.json` | Behavioral | 5 tests in `throughline_cleanup_test.go` |
| SCAR-027 | `session-artifact-in-shared-mena` lint rule | Lint | `TestSCAR027_*` |
| GO-001/002/003 | `RewriteMenaContentPaths` wired in ALL mena copy paths; compile-time function reference guard | Behavioral + Compile | `TestSCAR_ContentRewriteNotBypassed` |

---

## Agent-Relevance Tagging

**Heat map**: `integration-engineer` and `principal-engineer` appear in 27 of 28 scars (universal). `context-architect` in 9. `qa-adversary` in 3.

| SCAR | Agent Role(s) | Why |
|------|--------------|-----|
| SCAR-001 | principal-engineer | Lock code must use atomic flock-on-existing-fd |
| SCAR-002 | principal-engineer, context-architect | Never rename `.claude/`; use writeIfChanged |
| SCAR-003 | principal-engineer | Mutation flags must imply Force |
| SCAR-004 | principal-engineer | Never `_` discard provenance errors |
| SCAR-005 | principal-engineer, context-architect | `RemoveAll` on user dirs destroys user content |
| SCAR-006 | principal-engineer, qa-adversary | Satellite paths differ from core; test both paths |
| SCAR-007 | context-architect, principal-engineer | Mixed dro/lego dirs break CC skill resolution |
| SCAR-008 | context-architect | Sub-100ms hooks must be sync; async creates log spam |
| SCAR-009 | principal-engineer | CC hook format is nested; flat is silently rejected |
| SCAR-010 | principal-engineer | ALL hooks need timeout; ALL git subprocess need CommandContext |
| SCAR-011 | principal-engineer | `.current-session` is deprecated; use priority chain |
| SCAR-012 | principal-engineer, context-architect | Archived sessions need distinct denial path |
| SCAR-013 | principal-engineer, qa-adversary | Session wrap has three edge cases requiring explicit guards |
| SCAR-014 | principal-engineer | All session status reads must go through `NormalizeStatus()` |
| SCAR-015 | principal-engineer | Shell log functions must use `>&2` |
| SCAR-016 | principal-engineer | `((var++))` needs `|| true` guard under `set -euo pipefail` |
| SCAR-017 | context-architect, principal-engineer | `@skill-name` is not a CC primitive; use `skills:` frontmatter |
| SCAR-018 | context-architect | `context: fork` removes Task tool access — breaks Argus Pattern |
| SCAR-019 | context-architect | CC color palette has 8 values; others are silently ignored |
| SCAR-020 | principal-engineer | LLM context not accessible to bash; must pass `-s` flag explicitly |
| SCAR-021 | principal-engineer, context-architect | Cross-rite agents in project scope cause CC shadowing issues |
| SCAR-022 | principal-engineer | Provenance test fixtures require full 64-char SHA256 with `sha256:` prefix |
| SCAR-023 | principal-engineer | Self-hosted knossos uses `knossos/templates/` not `templates/` |
| SCAR-024 | principal-engineer | Rite switch must clean `.throughline-ids.json` |
| SCAR-025 | principal-engineer | Always `os.Stat` before checksum comparison in sync |
| SCAR-026 | context-architect | Writeguard must not include delegation hints — belongs in agent prompts |
| SCAR-027 | context-architect, principal-engineer | Shared mena = permanent knowledge; session artifacts in `.sos/wip/` |
| SCAR-028 | principal-engineer, context-architect | MCP servers must write to `.mcp.json`, not `settings.local.json` |
| GO-001/002/003 | principal-engineer | `RewriteMenaContentPaths` must be called in ALL mena copy paths |

---

## Knowledge Gaps

1. **GO-001/002/003 are not formally numbered** in the SCAR catalog. SCAR-028/029/030 assignment is recommended.

2. **9 of 27 SCARs lack dedicated behavioral regression tests** — SCAR-001, SCAR-003, SCAR-005, SCAR-006, SCAR-009, SCAR-011, SCAR-012, SCAR-013, SCAR-022. They have structural scans, lint enforcement, or integration tests instead. Unguarded regression risk is present.

3. **SCAR-010 binary mismatch trap** (MEMORY.md critical lesson): No automated guard for build-vs-install binary divergence. `go build ./cmd/ari` writes `./ari` locally; CC uses PATH binary (`which ari`). After rebuilding, must `cp ./ari $(which ari)`. No SCAR entry, no regression test.

4. **CI infrastructure scar (GITHUB_TOKEN/GoReleaser)**: GoReleaser uses `GITHUB_TOKEN`, which suppresses downstream workflow triggers. Fix: switch to `HOMEBREW_TAP_TOKEN` PAT (commit `d65bd4c`). Documented in `.sos/land/scar-tissue.md` but not numbered in SCAR catalog.

5. **Hook stdin-only transport**: `d92ff13` removed env var fallback from `ParseEnv` in `internal/hook/env.go`. This is a behavioral constraint documented in MEMORY.md and `design-constraints.md` as SETTLED, but has no SCAR entry and no regression test guarding against re-introduction of env var paths.

6. **Shell-era session locking scars** (LOCK-001, LOCK-003, STATE-001): Three shell session locking bugs from the bash era are documented in git history (`3dea170`) but not numbered in the SCAR catalog. The shell code they fixed has since been superseded by Go implementations.

### SCAR-030: Go 1.23 Lacks t.Chdir — Use DI Parameter Injection

**Category**: testing
**What failed**: `os.Chdir` in tests mutates process-global CWD, blocking `t.Parallel()`. Go 1.24 adds `t.Chdir()` but Go 1.23 does not have it.
**Fix pattern**: Add `workDir string` field to the struct under test or function parameter. Production caller uses `os.Getwd()`; tests inject `t.TempDir()`. Do not use `t.Cleanup` chdir-restore patterns — they are not parallel-safe.
**Guard**: `cmd/initialize` tests verify this pattern (commit `5ec4a18`). Grep for `os.Chdir` in test files to detect regression.

### SCAR-031: Materialize Tests Are I/O-Bound — t.Parallel Yields ~30% Not 50%+

**Category**: performance
**What failed**: Expected `t.Parallel()` to cut materialize root test time from 37.6s to <20s. Actual: 26.4s (-29.7%).
**Root cause**: Tests perform filesystem materialization to `t.TempDir()`. I/O wait dominates, not CPU. Parallelism reduces wall-clock time but cannot eliminate I/O bottleneck.
**Guard**: Baseline measurement captured. Further speedup requires in-memory filesystem mocks or test architecture changes — separate initiative.

### SCAR-032: TestInit_WithRite Global State via common.SetEmbeddedAssets

**Category**: global_state
**What failed**: `TestInit_WithRite` in `cmd/initialize` calls `common.SetEmbeddedAssets()` which mutates package-level global state. Cannot be parallelized.
**Fix pattern**: Refactor `SetEmbeddedAssets` to use DI (struct field or function parameter) when `common` package gets DI treatment.
**Guard**: Test deliberately excluded from `t.Parallel()` adoption (commit `5ec4a18`). Documented exception.

### SCAR-033: Config Canonicalization Timing

**Category**: integration_failure
**Discovered**: 2026-03-12 (harness-agnosticism initiative, Sprint I1)
**Fix commit**: 6d22a96

**What failed**: hooks.yaml was updated to use canonical event names (snake_case) BEFORE the translation-aware ari binary was rebuilt and installed. The running (old) binary read the new config but lacked the `CanonicalToWire()` translation layer, producing `settings.local.json` with canonical names (`pre_tool`, `session_start`) instead of CC wire names (`PreToolUse`, `SessionStart`). Claude Code rejected the invalid keys, breaking all hooks across projects.

**Defensive pattern**: When canonicalizing config file formats (new field names, new value formats), the translation-aware binary MUST be built and installed BEFORE the canonicalized config is committed. Config changes and translation code are an atomic pair — deploy the code first, then the config.

**Guard**: Include a `CGO_ENABLED=0 go build ./cmd/ari && cp ./ari $(which ari)` step before committing config format changes. Alternatively, add a version check in the config parser that rejects unrecognized event names with a clear "rebuild ari" message.

---

## Assessment Metadata

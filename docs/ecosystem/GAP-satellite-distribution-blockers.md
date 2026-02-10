# Gap Analysis: Satellite Distribution Blockers

Three blocking issues prevent reliable knossos distribution across the 4 satellite repos (autom8y_platform, autom8_asana, autom8_data, autom8). Each blocker is traced to specific file:line, with reproduction steps and measurable success criteria.

---

## Summary Table

| # | Blocker | Root Cause Location | Severity | Complexity |
|---|---------|---------------------|----------|------------|
| B1 | SessionStart hook error on satellite startup | Satellite `.claude/settings.local.json` (legacy bash hooks) + `internal/materialize/hooks.go:147` (merge preserves non-ari hooks) | P0 | MODULE |
| B2 | Stale `ari sync` references cause hallucination | `internal/cmd/rite/pantheon.go:46`, `internal/cmd/provenance/provenance.go:132`, `README.md:17,23,93`, 170+ doc references | P1 | PATCH |
| B3 | CC file watcher hang on mid-session rite switch | `internal/materialize/materialize.go:350-395` (MaterializeWithOptions writes agents/, mena/, CLAUDE.md, settings) | P1 | MODULE |

**Overall Complexity: SYSTEM** -- Blockers span the materialize pipeline, the hook merge system, the CLI help strings, and CC behavioral constraints. Fixes require coordinated changes across multiple files and packages, plus a deployment push to 4 satellites.

---

## Blocker 1: SessionStart Hook Error on Satellite Startup

### Root Cause

Satellite `settings.local.json` files contain **two classes of hook entries** for most events:

1. **ari hooks** (new, functional): e.g., `ari hook context --output json`
2. **Legacy bash hooks** (old, broken): e.g., `$CLAUDE_PROJECT_DIR/.claude/hooks/context-injection/session-context.sh` and `orchestrated-mode.sh`

The legacy hooks reference `.sh` files that either do not exist in the satellite (e.g., `orchestrated-mode.sh` missing in autom8y_platform, autom8_asana) or have been removed from the hook pipeline.

**Why they persist**: The hook merge logic in `internal/materialize/hooks.go:147-205` (`mergeHooksSettings()`) is designed to **preserve user-defined matcher groups** -- any hook group where the command does NOT start with `"ari hook"`. Legacy bash hooks (commands starting with `$CLAUDE_PROJECT_DIR/.claude/hooks/...`) are classified as "user" hooks by `isAriManagedGroup()` at `hooks.go:210-247` and are therefore preserved across every sync.

The `isAriManagedGroup()` function at `hooks.go:210-247` checks:
- New format: all hooks in the group must have command starting with `"ari hook"` (line 219)
- Old flat format: top-level command must start with `"ari hook"` (line 242)

Legacy bash hooks fail both checks, so they are retained as "user" hooks indefinitely.

### Satellite Audit

| Satellite | Legacy Hooks | Ari Hooks | Missing .sh Files |
|-----------|-------------|-----------|-------------------|
| autom8y_platform | 12 | 10 | 1 (orchestrated-mode.sh) |
| autom8_asana | 12 | 10 | 1 (orchestrated-mode.sh) |
| autom8_data | 10 | 0 | 1 (delegation-check.sh) |
| autom8 | 10 | 0 | 0 |

**Key observation**: autom8_data and autom8 have ZERO ari hooks -- they have never been synced with the new hooks.yaml pipeline. autom8y_platform and autom8_asana have both old and new hooks coexisting.

### Reproduction

1. Clone any satellite with legacy hooks (e.g., autom8y_platform)
2. Ensure `ari` is installed and on PATH
3. Start a CC session: `claude`
4. Observe: "SessionStart:startup hook error" in CC output
5. Root cause: CC attempts to execute `$CLAUDE_PROJECT_DIR/.claude/hooks/context-injection/orchestrated-mode.sh`, which does not exist

### Success Criteria

- `ari sync` on any satellite removes ALL legacy bash hook entries from `settings.local.json`
- Only `ari hook *` entries remain after sync
- CC session startup completes without hook errors on all 4 satellites
- `isAriManagedGroup()` or equivalent mechanism identifies legacy bash hooks for removal
- Zero `.sh` file references remain in any satellite `settings.local.json` after sync

### Affected Files

| File | Line(s) | Issue |
|------|---------|-------|
| `internal/materialize/hooks.go` | 147-205 | `mergeHooksSettings()` preserves legacy bash hooks as "user" hooks |
| `internal/materialize/hooks.go` | 210-247 | `isAriManagedGroup()` only checks for `ari hook` prefix, not for known-stale patterns |
| Satellite `settings.local.json` (all 4) | hooks section | Contains mixed legacy + ari entries |

---

## Blocker 2: Stale `ari sync` References Cause Hallucination

### Root Cause

The `materialize` subcommand was removed in ADR-0026 Phase 4b. The correct command is now `ari sync [--rite=NAME]`. However, **177 references** to the old `ari sync` form remain across the codebase, concentrated in:

**Go source (2 files, actively executed):**
- `internal/cmd/rite/pantheon.go:46` -- error message tells users to run `ari sync --rite <name>`
- `internal/cmd/provenance/provenance.go:132` -- help text says `Run 'ari sync'`

**README.md (3 references, high-visibility):**
- Line 17: Command table shows `ari sync --rite <name>`
- Line 23: Architecture note references `ari sync`
- Line 93: Rite mena description references `ari sync`

**INTERVIEW_SYNTHESIS.md (1 reference):**
- Line 173: Command table shows `ari sync --rite <name>`

**docs/ directory (170+ references across 20+ files):**
- Strategy docs, audit reports, QA reports, design docs, doctrine, guides, spikes, ADRs, CLI reference

**Mena source (1 reference, legacy-compat note only):**
- `mena/cem/sync.dro.md:48` -- correctly documents as legacy: "Use `ari sync` instead"

**Rite-switching dromena (0 references -- all correct):**
- All 10 rite-switching dromena correctly use `ari sync --rite <name>`

### Hallucination Mechanism

When a haiku-class model executes `/rite-switching:10x` (the `/10x` dromena), it:
1. Reads the dromena source which correctly says `ari sync --rite 10x-dev`
2. But also has in context: CLAUDE.md, rules, and other materialized content
3. CC loads `README.md` content into the project instructions context
4. The model sees `ari sync --rite <name>` in README.md and the Go error messages
5. The haiku model, being smaller, is more susceptible to context pollution and follows the stale pattern

The Go source references are particularly problematic because they appear in CLI error output that the model then reads and follows.

### Reproduction

1. In knossos repo, activate any rite: `ari sync --rite 10x-dev`
2. Start CC session
3. Run `/10x` (the rite-switching dromena)
4. With haiku model: observe it may call `ari sync --rite 10x-dev` instead of `ari sync --rite 10x-dev`
5. The `materialize` subcommand no longer exists, so the command fails with "unknown command"

### Success Criteria

- `grep -r "ari sync" internal/` returns zero results
- `grep -r "ari sync" README.md` returns zero results
- `grep -r "ari sync" INTERVIEW_SYNTHESIS.md` returns zero results
- `ari rite pantheon` error message references `ari sync --rite <name>` (not `ari sync`)
- `ari provenance show` help text references `ari sync` (not `ari sync`)
- Haiku model executing `/10x` calls `ari sync --rite 10x-dev` (not `ari sync`)
- Docs cleanup: all references in docs/ updated or annotated as historical

### Affected Files

**Critical (Go source -- active execution paths):**

| File | Line | Current | Correct |
|------|------|---------|---------|
| `internal/cmd/rite/pantheon.go` | 46 | `use 'ari sync --rite <name>'` | `use 'ari sync --rite <name>'` |
| `internal/cmd/provenance/provenance.go` | 132 | `Run 'ari sync' or 'ari sync user all'` | `Run 'ari sync'` |

**High (user-facing, top-level docs):**

| File | Lines | Count |
|------|-------|-------|
| `README.md` | 17, 23, 93 | 3 |
| `INTERVIEW_SYNTHESIS.md` | 173 | 1 |

**Medium (internal docs, 170+ references across 20+ files):**
- `docs/strategy/` (2 files, 2 refs)
- `docs/briefs/` (1 file, 3 refs)
- `docs/STAKEHOLDER-PREFERENCES-distribution-readiness.md` (2 refs)
- `docs/requirements/` (1 ref)
- `docs/audits/` (7 files, 40+ refs)
- `docs/releases/` (1 file, 4 refs)
- `docs/guides/` (2 files, 15+ refs)
- `docs/design/` (4 files, 30+ refs)
- `docs/doctrine/` (8 files, 25+ refs)
- `docs/spikes/` (1 file, 12 refs)
- `docs/decisions/` (4 files, 6 refs)
- `docs/prd/` (1 file, 1 ref)
- `docs/bugs/` (1 file, 6 refs)
- `docs/hygiene/` (1 file, 1 ref)

---

## Blocker 3: CC File Watcher Hang on Mid-Session Rite Switch

### Root Cause

When `ari sync --rite <name>` is called from within a CC session (via Bash tool), the materialization pipeline in `internal/materialize/materialize.go:264-397` (`MaterializeWithOptions()`) modifies multiple files that CC actively watches:

| Pipeline Stage | File(s) Modified | CC Behavior |
|----------------|-----------------|-------------|
| Step 4: `materializeAgents()` (line 350) | `.claude/agents/*.md` | Agents read on-demand -- **safe** |
| Step 5: `materializeMena()` (line 355) | `.claude/commands/`, `.claude/skills/` | Cached at startup -- **need restart** |
| Step 7: `materializeCLAUDEmd()` (line 365) | `.claude/CLAUDE.md` | Re-read mid-session -- **safe but triggers watcher** |
| Step 8: `materializeSettingsWithManifest()` (line 372) | `.claude/settings.local.json` | **Hooks snapshotted at startup** -- changes are no-ops |
| Step 10: `writeActiveRite()` (line 387) | `.claude/ACTIVE_RITE` | Not watched by CC |

The `writeIfChanged()` function at `internal/fileutil/fileutil.go:66-72` prevents writes when content is identical, but a genuine rite switch changes content for almost every file. The atomic write pattern (temp file + rename at `fileutil.go:60`) minimizes partial-read risk but does not prevent the CC file watcher from detecting the changes.

**CC behavioral constraints** (confirmed via claude-code-guide):
- **Hooks**: Snapshotted at startup. Mid-session changes to `settings.local.json` are silently ignored.
- **Commands/Skills**: Cached at startup. New/changed commands require CC restart to take effect.
- **Agents**: Read from disk on-demand when Task tool invokes them. Safe to modify mid-session.
- **CLAUDE.md**: Re-read by CC mid-session. Safe for dynamic context updates, but rapid multi-file changes can cause the watcher to hang.

**Previous bug**: `docs/bugs/materialization-crash-from-claude-code.md` documents the same fundamental issue -- running `ari sync` from within CC causes crash or hang when CC detects changes to `.claude/CLAUDE.md` and other watched files.

The hang occurs because CC's file watcher detects multiple rapid file changes (agents, commands, skills, CLAUDE.md, settings) and enters a state where it is waiting on filesystem events while the Bash tool that triggered the changes is waiting for CC to process. This creates a deadlock between the Bash tool execution and the file watcher.

### Reproduction

1. Start CC session in a satellite project with an active rite
2. Run via Bash tool: `ari sync --rite 10x-dev`
3. Observe: command hangs at "Running PreToolUse hook... Running..."
4. Alternative: command completes but CC becomes unresponsive
5. Must kill and restart CC session

### Stakeholder Decision: Soft Switch

The stakeholder has decided on a **soft switch** approach:
- Rite switch mid-session should only update **CLAUDE.md + agents/** (both CC-safe for mid-session changes)
- Document that **commands/skills** require CC restart to take effect
- Full sync remains available from external terminal

### Success Criteria

- Mid-session `ari sync --rite <name>` completes without hanging when called from CC Bash tool
- CLAUDE.md and agents/ are updated to reflect the new rite
- Commands/skills changes are deferred or documented as "restart required"
- `settings.local.json` hook changes are documented as "restart required"
- `ari sync` exits 0 within 10 seconds when called from CC Bash tool
- No CC crash or hang after rite switch

### Affected Files

| File | Line(s) | Issue |
|------|---------|-------|
| `internal/materialize/materialize.go` | 264-397 | `MaterializeWithOptions()` writes all file types in single pass |
| `internal/materialize/materialize.go` | 350 | `materializeAgents()` -- safe for mid-session |
| `internal/materialize/materialize.go` | 355 | `materializeMena()` -- triggers CC cache invalidation |
| `internal/materialize/materialize.go` | 365 | `materializeCLAUDEmd()` -- triggers file watcher |
| `internal/materialize/materialize.go` | 372 | `materializeSettingsWithManifest()` -- no-op mid-session but writes file |
| `internal/fileutil/fileutil.go` | 66-72 | `WriteIfChanged()` -- prevents spurious writes but not genuine changes |
| `internal/fileutil/fileutil.go` | 13-61 | `AtomicWriteFile()` -- temp+rename is atomic but still triggers watcher |
| `docs/bugs/materialization-crash-from-claude-code.md` | all | Documents previous manifestation of same issue |

---

## Cross-Cutting Concerns

### Hook Provenance Gap

The current provenance system tracks files in `.claude/` (agents, commands, skills, rules, CLAUDE.md). However, **hook entries within settings.local.json are not individually tracked**. The `materializeSettingsWithManifest()` function at `materialize.go:1386-1431` records a single provenance entry for the entire `settings.local.json` file (line 1420), not per-hook-entry.

This means:
- Provenance cannot distinguish "knossos wrote this hook entry" from "user wrote this hook entry"
- The `isAriManagedGroup()` heuristic (command prefix check) is the only ownership signal
- Legacy bash hooks that predate the provenance system have no ownership metadata

**Stakeholder decision**: Extend provenance to track hook entries (provenance-aware merge strategy).

### Deployment Coordination

All fixes must be deployed to 4 satellites in sequence:
1. Fix knossos pipeline code
2. Rebuild `ari` binary
3. Run `ari sync` in each satellite
4. Verify CC session startup in each satellite

Satellite-specific concerns:
- **autom8_data** and **autom8**: Have never been synced with new hooks pipeline (0 ari hooks). First sync will be a full migration.
- **autom8y_platform** and **autom8_asana**: Have mixed hooks. Sync must cleanly remove legacy while preserving ari hooks.

---

## Test Satellite Matrix

| Satellite | Hook State | Rite | Verification Focus |
|-----------|-----------|------|-------------------|
| autom8y_platform | 12 legacy + 10 ari + 1 missing | ecosystem | Mixed hook cleanup, missing .sh error gone |
| autom8_asana | 12 legacy + 10 ari + 1 missing | ecosystem | Same as above, second satellite confirms consistency |
| autom8_data | 10 legacy + 0 ari + 1 missing | (none?) | Full migration from legacy-only to ari-only |
| autom8 | 10 legacy + 0 ari + 0 missing | (none?) | Clean legacy removal, no missing file errors |

### Verification Checklist (per satellite)

- [ ] `ari sync` exits 0
- [ ] `settings.local.json` contains zero legacy bash hook references
- [ ] `settings.local.json` contains expected ari hook entries (10 per hooks.yaml)
- [ ] CC session starts without hook errors
- [ ] `/10x` dromena calls `ari sync --rite 10x-dev` (not `ari sync`)
- [ ] Mid-session rite switch does not hang CC (soft switch)

---

## Handoff Notes for Context Architect

### Design Decisions Needed

1. **B1 -- Legacy hook removal strategy**: How should `mergeHooksSettings()` identify and remove legacy bash hooks? Options include:
   - Nuke all non-ari hooks (stakeholder preference: "nuke all, with responsible validation")
   - Pattern-match known legacy patterns (`$CLAUDE_PROJECT_DIR/.claude/hooks/`)
   - Provenance-aware: track individual hook entries in provenance manifest

2. **B2 -- Stale reference cleanup scope**: The Go source fixes (2 files) are unambiguous. The 170+ doc references need a policy decision: bulk find-replace vs. annotate-as-historical vs. delete stale docs.

3. **B3 -- Soft switch architecture**: MaterializeWithOptions() currently runs all stages unconditionally. Design options for CC-aware mode:
   - New `--soft` flag that limits pipeline to CLAUDE.md + agents only
   - CC detection (env var check) that auto-selects soft mode
   - Separate `SoftSwitch()` method alongside `MaterializeWithOptions()`
   - Document-only approach: warn users and suggest external terminal

### Constraints

- CC hooks are snapshotted at startup -- settings.local.json changes require restart
- CC commands/skills are cached at startup -- mena changes require restart
- CC agents are read on-demand -- safe for mid-session modification
- CC CLAUDE.md is re-read mid-session -- safe but triggers file watcher on rapid changes
- `writeIfChanged()` prevents writes only when content is byte-identical
- `AtomicWriteFile()` uses temp+rename but still triggers inotify/FSEvents

### Dependency Order

B2 (stale refs) is independent and can be fixed first.
B1 (legacy hooks) requires a design decision on hook ownership tracking.
B3 (CC file watcher) requires B1 to be resolved first (settings.local.json writes during sync are part of the hang trigger).

Recommended execution order: **B2 -> B1 -> B3**.

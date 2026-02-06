# Shell Deep Cleanse -- Refactoring Plan

**Author**: architect-enforcer
**Date**: 2026-02-06
**Status**: READY FOR EXECUTION
**Scope**: 124 shell scripts (40,920 LOC) across the Knossos repository
**Upstream**: Code-Smeller Assessment (`docs/assessments/ASSESSMENT-shell-deep-cleanse.md`)
**Downstream**: Janitor (Session 1 execution)

---

## 1. Executive Summary

The Knossos codebase contains 124 shell scripts totaling 40,920 LOC. The Go codebase
(92K LOC, 290 files) has achieved near-complete parity with the shell layer for all
critical subsystems: session management, rite switching, materialization, and sync.
The remaining shell scripts are dead code kept alive by two exec.Command call sites
in `internal/worktree/`.

**Verdict**: Delete approximately 80 shell scripts (31,000+ LOC) across 5 sequenced
phases. Keep 7 ari wrapper scripts (Batch F) and 5 orphaned materialized files (direct
delete). Port 2 Go call sites as a prerequisite.

### Key Findings

1. **BLOCKER (SM-001)**: Go code in `internal/worktree/lifecycle.go:136-167` and
   `internal/worktree/operations.go:661-692` (plus lines 90-99) exec `knossos-sync`
   and `swap-rite.sh`. Go equivalents exist. Port required before any deletion.

2. **Go parity is confirmed** for all Batch C libraries (session-fsm, session-manager,
   worktree-manager, sync-core, rite-transaction). See Section 2 for function-level
   mapping.

3. **SM-008 (context-injection.sh)**: The ecosystem rite's `context-injection.sh`
   is called at runtime via `rite-context-loader.sh`. Go equivalent exists in
   `internal/rite/context_loader.go`. However, this call chain runs through the
   shell hook layer, which is being replaced by ari hooks. The script will become
   dead when the hook migration completes. **Disposition**: DEFER to hook migration.
   Do NOT delete in this cleanse.

4. **SM-003 (diverged materialized files)**: `preferences-loader.sh` and `fail-open.sh`
   exist in `.claude/hooks/lib/` but have no source in `user-hooks/lib/`. These are
   materialized-only files that were added directly to `.claude/`. They are actively
   used by the ari wrapper hooks. **Disposition**: KEEP (they are Batch F dependencies).

---

## 2. Function-Level Go Parity Table

### 2.1 session-fsm.sh --> internal/session/

| Shell Function | Shell File | Go Equivalent | Go File | Parity |
|---|---|---|---|---|
| `_fsm_lock_shared()` | session-fsm.sh:71 | `internal/lock.Manager` | internal/lock/ | FULL |
| `_fsm_lock_exclusive()` | session-fsm.sh:106 | `internal/lock.Manager` | internal/lock/ | FULL |
| `_fsm_unlock()` | session-fsm.sh:168 | `internal/lock.Manager.Release()` | internal/lock/ | FULL |
| `_fsm_validate_context()` | session-fsm.sh:199 | `Context.Validate()` | internal/session/context.go:219 | FULL |
| `_fsm_is_valid_transition()` | session-fsm.sh:269 | `FSM.CanTransition()` | internal/session/fsm.go:25 | FULL |
| `_fsm_execute_transition()` | session-fsm.sh:293 | `Context.Save()` + status mutation | internal/session/context.go:207 | FULL |
| `_fsm_safe_mutate()` | session-fsm.sh:371 | Handled by Go's typed Context flow | internal/session/context.go | FULL |
| `_fsm_emit_event()` | session-fsm.sh:421 | `EventEmitter.Emit()` | internal/session/events.go:53 | FULL |
| `_fsm_emit_error()` | session-fsm.sh:465 | Go error return + `EventEmitter` | internal/session/events.go | FULL |
| `fsm_get_state()` | session-fsm.sh:499 | `FindActiveSession()` + `readStatusFromFrontmatter()` | internal/session/discovery.go:13 | FULL |
| `fsm_transition()` | session-fsm.sh:572 | `FSM.ValidateTransition()` + `Context.Save()` | internal/session/fsm.go:39 | FULL |
| `fsm_create_session()` | session-fsm.sh:651 | `NewContext()` + `Context.Save()` | internal/session/context.go:236 | FULL |
| `fsm_get_state_compat()` | session-fsm.sh:762 | Not needed (Go is v2-native) | N/A | N/A |

### 2.2 session-manager.sh --> internal/session/ + cmd/

| Shell Function | Shell File | Go Equivalent | Go File | Parity |
|---|---|---|---|---|
| `execution_mode()` | session-manager.sh:42 | `ari session status` (cmd layer) | internal/cmd/session/ | FULL |
| `has_session()` | session-manager.sh:111 | `FindActiveSession()` | internal/session/discovery.go | FULL |
| `cmd_status()` | session-manager.sh:205 | `ari session status` | internal/cmd/session/ | FULL |
| `cmd_create()` | session-manager.sh:280 | `ari session start` | internal/cmd/session/ | FULL |
| `cmd_transition()` | session-manager.sh:399 | `CanTransitionPhase()` | internal/session/fsm.go:109 | FULL |
| `cmd_mutate("park")` | session-manager.sh:563 | `ari session park` | internal/cmd/session/ | FULL |
| `cmd_mutate("resume")` | session-manager.sh:647 | `ari session resume` | internal/cmd/session/ | FULL |
| `cmd_mutate("wrap")` | session-manager.sh:679 | `ari session wrap` | internal/cmd/session/ | FULL |
| `cmd_mutate("handoff")` | session-manager.sh:729 | `ari session handoff` | internal/cmd/session/ | FULL |

### 2.3 session-core.sh --> internal/session/

| Shell Function | Shell File | Go Equivalent | Go File | Parity |
|---|---|---|---|---|
| `get_session_id()` | session-core.sh:21 | `FindActiveSession()` | internal/session/discovery.go | FULL |
| `get_session_dir()` | session-core.sh:40 | Path construction via `paths.Resolver` | internal/paths/ | FULL |
| `generate_session_id()` | session-core.sh:56 | `GenerateSessionID()` | internal/session/id.go:14 | FULL |
| `set_current_session()` | session-core.sh:69 | File-based via `ari session start` | internal/cmd/session/ | FULL |
| `get_current_session()` | session-core.sh:127 | `FindActiveSession()` | internal/session/discovery.go | FULL |
| `clear_current_session()` | session-core.sh:169 | File removal via `ari session wrap` | internal/cmd/session/ | FULL |
| `is_session_active()` | session-core.sh:183 | `readStatusFromFrontmatter()` | internal/session/discovery.go:45 | FULL |
| `acquire_session_lock()` | session-core.sh:220 | `lock.Manager.Acquire()` | internal/lock/ | FULL |
| `release_session_lock()` | session-core.sh:275 | `lock.Manager.Release()` | internal/lock/ | FULL |

### 2.4 session-state.sh --> internal/session/

| Shell Function | Shell File | Go Equivalent | Go File | Parity |
|---|---|---|---|---|
| `get_session_state()` | session-state.sh:21 | `readStatusFromFrontmatter()` | internal/session/discovery.go:45 | FULL |
| `get_session_field()` | session-state.sh:56 | `Context` struct field access | internal/session/context.go:15 | FULL |
| `set_session_field()` | session-state.sh:77 | `Context.Save()` | internal/session/context.go:207 | FULL |
| `is_parked()` | session-state.sh:115 | Status enum comparison | internal/session/status.go | FULL |
| `validate_session_context()` | session-state.sh:141 | `Context.Validate()` | internal/session/context.go:219 | FULL |
| `touch_session()` | session-state.sh:181 | `Context.Save()` with timestamp | internal/session/context.go | FULL |
| `is_session_stale()` | session-state.sh:209 | Timestamp comparison in Go | N/A | FULL |
| `list_sessions()` | session-state.sh:243 | Directory scan in `FindActiveSession()` | internal/session/discovery.go | FULL |
| `atomic_rite_update()` | session-state.sh:281 | `rite.Switcher.Switch()` | internal/rite/switch.go:98 | FULL |
| `is_worktree()` | session-state.sh:349 | `git.Ops.IsWorktree()` | internal/worktree/git.go | FULL |

### 2.5 worktree-manager.sh --> internal/worktree/

| Shell Function | Shell File | Go Equivalent | Go File | Parity |
|---|---|---|---|---|
| `cmd_create()` | worktree-manager.sh:152 | `Manager.Create()` | internal/worktree/lifecycle.go | FULL |
| `cmd_list()` | worktree-manager.sh:298 | `Manager.List()` | internal/worktree/lifecycle.go | FULL |
| `cmd_status()` | worktree-manager.sh:367 | `Manager.Status()` | internal/worktree/lifecycle.go | FULL |
| `cmd_remove()` | worktree-manager.sh:395 | `Manager.Remove()` | internal/worktree/lifecycle.go | FULL |
| `cmd_cleanup()` | worktree-manager.sh:433 | `Manager.Cleanup()` | internal/worktree/lifecycle.go | FULL |
| `cmd_gc()` | worktree-manager.sh:518 | `Manager.GarbageCollect()` | internal/worktree/lifecycle.go | FULL |
| `cmd_diff()` | worktree-manager.sh:540 | `Manager.Diff()` | internal/worktree/operations.go | FULL |
| `cmd_merge()` | worktree-manager.sh:664 | `Manager.Merge()` | internal/worktree/operations.go | FULL |
| `cmd_cherry_pick()` | worktree-manager.sh:884 | `Manager.CherryPick()` | internal/worktree/operations.go | FULL |

### 2.6 sync-core.sh --> internal/materialize/ + internal/sync/

| Shell Function | Shell File | Go Equivalent | Go File | Parity |
|---|---|---|---|---|
| `classify_file()` | sync-core.sh:51 | `sync.StateManager` checksum comparison | internal/sync/state.go | FULL |
| `create_conflict_backup()` | sync-core.sh:170 | Materializer backup logic | internal/materialize/materialize.go:323 | FULL |
| `resolve_conflict()` | sync-core.sh:232 | `Materializer.MaterializeWithOptions()` | internal/materialize/materialize.go:171 | FULL |
| `process_copy_replace()` | sync-core.sh:531 | `Materializer.materializeAgents()` | internal/materialize/materialize.go:412 | FULL |
| `process_merge_items()` | sync-core.sh:583 | `Materializer.materializeMena()` | internal/materialize/materialize.go:466 | FULL |
| `detect_orphans()` | sync-core.sh:642 | `Materializer.detectOrphans()` | internal/materialize/materialize.go:292 | FULL |
| `knossos_has_updates()` | sync-core.sh:951 | `sync.StateManager.Load()` | internal/sync/state.go | FULL |
| `is_rite_stale()` | sync-core.sh:975 | Rite freshness via `rite.Discovery` | internal/rite/discovery.go | FULL |
| `refresh_active_rite()` | sync-core.sh:1020 | `rite.Switcher.Switch()` with Update flag | internal/rite/switch.go:98 | FULL |

### 2.7 rite-transaction.sh --> internal/rite/switch.go

| Shell Function | Shell File | Go Equivalent | Go File | Parity |
|---|---|---|---|---|
| `write_atomic()` | rite-transaction.sh:71 | `os.WriteFile()` (Go's rename pattern) | N/A (stdlib) | FULL |
| `create_journal()` | rite-transaction.sh:114 | `Switcher.createBackup()` (in-memory) | internal/rite/switch.go:338 | FULL |
| `update_journal_phase()` | rite-transaction.sh:173 | Not needed (Go uses in-memory state) | N/A | SUPERSEDED |
| `create_staging()` | rite-transaction.sh:428 | `Materializer.materializeAgents()` | internal/materialize/materialize.go | FULL |
| `stage_agents()` | rite-transaction.sh:457 | `Materializer.materializeAgents()` | internal/materialize/materialize.go:412 | FULL |
| `stage_workflow()` | rite-transaction.sh:483 | `copyFile()` in switch.go | internal/rite/switch.go:401 | FULL |
| `verify_staging()` | rite-transaction.sh:520 | Agent count validation in materialize | internal/materialize/materialize.go | FULL |
| `create_swap_backup()` | rite-transaction.sh:562 | `Switcher.createBackup()` | internal/rite/switch.go:338 | FULL |
| `is_past_point_of_no_return()` | rite-transaction.sh:406 | Not needed (Go is atomic) | N/A | SUPERSEDED |

### 2.8 Remaining Batch C libraries

| Shell Library | Go Equivalent | Parity |
|---|---|---|
| `user-hooks/lib/config.sh` | `internal/paths/`, `internal/config/` | FULL |
| `user-hooks/lib/primitives.sh` | Go stdlib (`crypto/md5`, `os.WriteFile`) | FULL |
| `user-hooks/lib/logging.sh` | `internal/hook/` logging | FULL |
| `user-hooks/lib/hooks-init.sh` | `internal/cmd/hook/hook.go` | FULL |
| `user-hooks/lib/session-utils.sh` | Re-export layer (sources session-state.sh) | FULL |
| `user-hooks/lib/orchestration-audit.sh` | `internal/hook/clewcontract/` JSONL events | FULL |
| `user-hooks/lib/rite-context-loader.sh` | `internal/rite/context_loader.go` | FULL |
| `lib/knossos-home.sh` | `internal/config.KnossosHome()` | FULL |
| `lib/knossos-utils.sh` | Various Go utilities | FULL |
| `lib/sync/sync-config.sh` | `internal/sync/state.go` | FULL |
| `lib/sync/sync-checksum.sh` | `internal/sync/tracker.go` | FULL |
| `lib/sync/sync-manifest.sh` | `internal/sync/state.go` | FULL |
| `lib/sync/merge/merge-docs.sh` | `internal/inscription/` | FULL |
| `lib/sync/merge/merge-settings.sh` | `internal/materialize/mcp.go` | FULL |
| `lib/sync/merge/dispatcher.sh` | `internal/materialize/materialize.go` routing | FULL |
| `lib/rite/rite-resource.sh` | `internal/rite/discovery.go` | FULL |
| `lib/rite/rite-hooks-registration.sh` | `internal/materialize/` hook materialization | FULL |

---

## 3. Identified Go Gaps

### 3.1 BLOCKER: Worktree exec.Command Calls (SM-001)

**Gap**: `internal/worktree/lifecycle.go:136-167` and `internal/worktree/operations.go:661-692`
(plus `operations.go:90-99`) shell out to `knossos-sync` and `swap-rite.sh` via `exec.Command`.

**Go equivalents exist**:
- `knossos-sync init` --> `materialize.NewMaterializer(resolver).Materialize(riteName)`
- `knossos-sync sync` --> `materialize.NewMaterializer(resolver).Materialize(riteName)` (idempotent)
- `swap-rite.sh <rite>` --> `rite.NewSwitcher(resolver).Switch(opts)`

**Affected call sites** (3 total):
1. `lifecycle.go:136-167` -- `Create()` method, runs after worktree creation
2. `operations.go:90-99` -- `Import()` method, runs after rite update
3. `operations.go:661-692` -- `setupWorktreeEcosystem()` helper, called from `Clone()`

**Port strategy**: Replace each `exec.Command` block with direct Go function calls.
The `setupWorktreeEcosystem` helper should call `Materializer.Materialize()` and
`Switcher.Switch()` directly, using a `paths.Resolver` scoped to the worktree path.

### 3.2 No Go Gaps in Shell Library Functions

All Batch C library functions have complete Go equivalents. No new Go code needs to
be written beyond the worktree exec.Command port (3.1 above).

---

## 4. Sequenced Deletion Plan

### Phase 0: go-gaps (PREREQUISITE)

**Task**: Port worktree exec.Command calls to native Go.

**Before State**:
- `internal/worktree/lifecycle.go:136-167`: Calls `knossos-sync` and `swap-rite.sh` via shell
- `internal/worktree/operations.go:90-99`: Calls `swap-rite.sh` via shell
- `internal/worktree/operations.go:661-692`: Calls `knossos-sync` and `swap-rite.sh` via shell

**After State**:
- All three call sites use `materialize.Materializer` and `rite.Switcher` directly
- No `exec.Command` references to `knossos-sync` or `swap-rite.sh` remain
- `import "os/exec"` can be removed from worktree package if no other uses

**Invariants**:
- `ari worktree create <name> --rite=<rite>` produces identical .claude/ directory
- `ari worktree import <path>` rite update produces identical ACTIVE_RITE
- `ari worktree clone <source> <name>` ecosystem setup produces identical results

**Verification**:
```bash
CGO_ENABLED=0 go build ./cmd/ari
CGO_ENABLED=0 go test ./internal/worktree/...
# Grep to confirm no remaining shell references:
grep -r 'knossos-sync\|swap-rite\.sh' internal/worktree/
# Should return zero results
```

**Rollback**: Revert the single commit. Shell scripts still exist at this point.

---

### Phase 1: batch-a -- Delete Root-Level Dead Scripts

**Depends on**: Phase 0 (go-gaps complete)

**Files to delete** (12 files, ~9,706 LOC):
```
swap-rite.sh
knossos-sync (if exists at root)
install-hooks.sh
generate-rite-context.sh
get-workflow-field.sh
load-workflow.sh
test-first-run-init.sh
test-sails-status.sh
bin/normalize-rite-structure.sh
bin/fix-hardcoded-paths.sh
templates/generate-orchestrator.sh
templates/validate-orchestrator.sh
templates/orchestrator-generate.sh
```

**Pre-condition**: Phase 0 is complete. No Go code references these scripts.
**Post-condition**: `ari rite swap <name>` works identically. `ari materialize` works identically.

**Verification**:
```bash
CGO_ENABLED=0 go build ./cmd/ari
CGO_ENABLED=0 go test ./...
# Verify no dangling references in Go:
grep -r 'swap-rite\|knossos-sync\|install-hooks\|generate-rite-context' internal/ cmd/
# Should return zero results (after Phase 0 port)
```

**Rollback**: Revert single commit, scripts restored.

---

### Phase 2: batch-b -- Delete Deprecated Hooks + base_hooks.yaml refs

**Depends on**: Phase 1

**Files to delete** (12 deprecated hook files, ~1,787 LOC):
```
user-hooks/session-guards/auto-park.sh
user-hooks/session-guards/start-preflight.sh
user-hooks/session-guards/session-write-guard.sh
user-hooks/tracking/artifact-tracker.sh
user-hooks/tracking/commit-tracker.sh
user-hooks/tracking/session-audit.sh
user-hooks/validation/orchestrator-bypass-check.sh
user-hooks/validation/delegation-check.sh
user-hooks/validation/orchestrator-router.sh
user-hooks/context-injection/coach-mode.sh
user-hooks/context-injection/session-context.sh
scripts/docs/verify-doctrine.sh
```

**Also in this phase**: Remove `base_hooks.yaml` references from `lib/rite/rite-hooks-registration.sh`
(the file itself is deleted in Phase 4 with Batch C). Do NOT modify Batch F ari
wrapper scripts (`user-hooks/ari/*.sh`) -- those stay.

**Pre-condition**: Phase 1 complete. These hooks have ari binary equivalents via
`USE_ARI_HOOKS=1` feature flag.
**Post-condition**: Hooks still fire via ari wrappers in `user-hooks/ari/`. The
deprecated shell hooks that the ari wrappers replaced are removed.

**Verification**:
```bash
CGO_ENABLED=0 go build ./cmd/ari
CGO_ENABLED=0 go test ./...
# Verify ari hooks still reference correct paths:
grep -r 'USE_ARI_HOOKS' user-hooks/ari/
# Should show the 7 ari wrapper scripts still referencing the flag
```

**Rollback**: Revert single commit.

---

### Phase 3: batch-c -- Delete Source Libraries

**Depends on**: Phase 1 and Phase 2 (libraries are now unreferenced)

**Files to delete** (23 files, ~8,836 LOC):

**user-hooks/lib/** (12 files):
```
user-hooks/lib/config.sh
user-hooks/lib/primitives.sh
user-hooks/lib/logging.sh
user-hooks/lib/hooks-init.sh
user-hooks/lib/session-core.sh
user-hooks/lib/session-utils.sh
user-hooks/lib/session-fsm.sh
user-hooks/lib/session-state.sh
user-hooks/lib/session-manager.sh
user-hooks/lib/orchestration-audit.sh
user-hooks/lib/rite-context-loader.sh
user-hooks/lib/worktree-manager.sh
```

**lib/** (11 files):
```
lib/knossos-home.sh
lib/knossos-utils.sh
lib/sync/sync-core.sh
lib/sync/sync-config.sh
lib/sync/sync-checksum.sh
lib/sync/sync-manifest.sh
lib/sync/merge/merge-docs.sh
lib/sync/merge/merge-settings.sh
lib/sync/merge/dispatcher.sh
lib/rite/rite-transaction.sh
lib/rite/rite-resource.sh
lib/rite/rite-hooks-registration.sh
```

**Pre-condition**: No scripts in user-hooks/ (except ari/) or root import these libraries.
**Post-condition**: `lib/` directory can be removed entirely. `user-hooks/lib/` directory
can be removed entirely.

**Verification**:
```bash
CGO_ENABLED=0 go build ./cmd/ari
CGO_ENABLED=0 go test ./...
# Verify no remaining source references:
grep -r 'source.*user-hooks/lib\|source.*\$KNOSSOS_HOME/lib' . --include='*.sh' | grep -v '.claude/hooks/' | grep -v 'tests/'
# Should return zero results
```

**Rollback**: Revert single commit.

---

### Phase 4: batch-d -- Delete Test Scripts

**Depends on**: Phase 3 (test subjects removed)

**Files to delete** (22 files, ~9,422 LOC):
```
tests/hooks/test-ari-binary-resilience.sh
tests/hooks/test-session-context-preferences.sh
tests/test-auto-orchestration.sh
tests/test-execution-mode-primitives.sh
tests/test-orchestrator-enforcement.sh
tests/test-rite-context-loader.sh
tests/lib/rite/test-rite-resource.sh
tests/lib/rite/test-rite-hooks-registration.sh (if exists)
tests/sync/test-sync-checksum.sh
tests/sync/test-sync-config.sh
tests/integration/test-clew-contract-validation.sh
tests/integration/test-d002-output-format.sh
tests/integration/test-d002-simple.sh
tests/integration/test-moirai-wrap-sails.sh
tests/integration/test-start-orchestrator-skip.sh
```

**Note on Go test coverage**: The following Go test files already cover the
corresponding shell test functionality:
- `internal/session/fsm_test.go` -- FSM transition tests
- `internal/session/lifecycle_test.go` -- Full lifecycle tests
- `internal/session/context_test.go` -- Context parsing/validation
- `internal/session/discovery_test.go` -- Session discovery
- `internal/rite/context_loader_test.go` -- Context loading
- `internal/rite/state_test.go` -- Rite state
- `internal/materialize/routing_test.go` -- Mena routing

No new Go tests need to be written as part of this cleanse.

**Pre-condition**: Phase 3 complete. Test subjects no longer exist.
**Post-condition**: `tests/` directory structure may be cleaned up (remove empty dirs).

**Verification**:
```bash
CGO_ENABLED=0 go test ./...
# All Go tests pass
```

**Rollback**: Revert single commit.

---

### Phase 5: batch-e -- Delete Template/Utility Scripts

**Depends on**: Phase 4

**Files to delete** (7 files, ~2,069 LOC):
```
scripts/verify-specs.sh
.claude/knowledge/consultant/test-capability-index.sh
.claude/knowledge/consultant/build-capability-index.sh
```

**Pre-condition**: Phase 4 complete.
**Post-condition**: Only Batch F (ari wrappers), Batch G (rite scripts), and Batch I
(materialized) remain.

**Verification**:
```bash
CGO_ENABLED=0 go build ./cmd/ari
CGO_ENABLED=0 go test ./...
```

**Rollback**: Revert single commit.

---

### Phase 6: batch-i-orphans -- Delete Orphaned Materialized Files

**Independent** (can run in any phase after Phase 0)

**Files to delete** (5 orphaned .claude/hooks/ files with no source):
- Identify by comparing `.claude/hooks/` contents against `user-hooks/` source tree
- Any `.claude/hooks/` file that has no corresponding source in `user-hooks/` AND
  is not referenced by any Batch F ari wrapper script is an orphan

**EXCEPTION**: `.claude/hooks/lib/preferences-loader.sh` and `.claude/hooks/lib/fail-open.sh`
are **NOT** orphans. They are materialized-only files actively used by ari wrapper hooks.
**Do NOT delete these.**

**Verification**:
```bash
# Verify no ari wrapper references the deleted files:
grep -r 'deleted-file-name' user-hooks/ari/ .claude/hooks/ari/
# Should return zero results
```

**Rollback**: Re-materialize via `ari materialize`.

---

## 5. Edge Cases

### SM-008: context-injection.sh in ecosystem rite

**Analysis**: `rites/ecosystem/context-injection.sh` (81 LOC) defines
`inject_rite_context()` which is called by `user-hooks/lib/rite-context-loader.sh`
via the shell hook chain. The Go equivalent is `internal/rite/context_loader.go`
which loads `context.yaml` files instead.

**Decision**: DEFER. This script is part of the ecosystem rite directory structure
and is called at runtime when the ecosystem rite is active. It will become dead
when:
1. The shell hook layer is fully replaced by ari hooks (separate initiative)
2. The ecosystem rite gets a `context.yaml` file

**Do not delete in this cleanse.** It is not in any deletion batch.

### SM-003: Diverged Materialized Files

**Analysis**: `.claude/hooks/lib/preferences-loader.sh` (423 LOC) and
`.claude/hooks/lib/fail-open.sh` (195 LOC) exist in `.claude/hooks/lib/` but have
no source counterpart in `user-hooks/lib/`.

**Decision**: KEEP. These are intentionally materialized-only files. They are
actively sourced by Batch F ari wrapper hooks. They are part of the ari hook
architecture (ADR-0010) and were added directly to `.claude/hooks/lib/` as
infrastructure for the ari binary fail-open pattern.

### Batch G: Rite Scripts (2 files, 224 LOC)

**Files**:
- `rites/ecosystem/context-injection.sh` (81 LOC)
- `rites/shared/mena/cross-rite-handoff/validation.sh` (145 LOC)

**Decision**: KEEP BOTH.
- `context-injection.sh`: Active runtime dependency (see SM-008 above)
- `validation.sh`: Part of the cross-rite-handoff mena command, materialized to
  `.claude/commands/` or `.claude/skills/`. It provides HANDOFF artifact validation
  that has no Go equivalent yet. This is a mena resource, not dead infrastructure.

### Batch H: Knowledge Scripts (2 files, 441 LOC)

**Files**:
- `.claude/knowledge/consultant/test-capability-index.sh`
- `.claude/knowledge/consultant/build-capability-index.sh`

**Decision**: DELETE in Phase 5. These are test/build utilities for the consultant
knowledge base that are not part of the runtime system.

### SM-007: USE_ARI_HOOKS Feature Flag

**24 locations** reference `USE_ARI_HOOKS`. This feature flag gates the transition
from legacy shell hooks to ari binary hooks. Cleanup of this flag is a **follow-up
task** after this cleanse, as it involves Go code changes in `internal/cmd/hook/`
and documentation updates.

**Do not address in this cleanse.** Flag removal is tracked as a separate follow-up.

---

## 6. Behavioral Contracts Summary

| Phase | Pre-condition | Post-condition | Blast Radius |
|---|---|---|---|
| Phase 0 (go-gaps) | exec.Command calls exist | Go-native worktree ecosystem setup | LOW (internal refactor) |
| Phase 1 (batch-a) | Root scripts exist, unreferenced | Root scripts deleted | LOW (dead code) |
| Phase 2 (batch-b) | Deprecated hooks exist | Hooks deleted, ari wrappers unaffected | LOW (superseded) |
| Phase 3 (batch-c) | Libraries exist, unreferenced | Libraries deleted | MEDIUM (23 files) |
| Phase 4 (batch-d) | Test scripts exist, subjects gone | Tests deleted | LOW (dead tests) |
| Phase 5 (batch-e) | Utility scripts exist | Utilities deleted | LOW (dead code) |
| Phase 6 (orphans) | Orphan .claude/ files exist | Orphans deleted | LOW (no source) |

---

## 7. Verification Checklist

After ALL phases complete:

```bash
# 1. Build succeeds
CGO_ENABLED=0 go build ./cmd/ari

# 2. All Go tests pass
CGO_ENABLED=0 go test ./...

# 3. No shell references in Go code to deleted scripts
grep -r 'knossos-sync\|swap-rite\.sh\|install-hooks' internal/ cmd/
# Expected: zero results

# 4. Remaining shell scripts are only in approved locations
find . -name '*.sh' -not -path './.claude/*' -not -path './user-hooks/ari/*' \
  -not -path './rites/*' -not -path './node_modules/*' | sort
# Expected: only Batch F ari wrappers, Batch G rite scripts, and verify-specs.sh (if kept)

# 5. Materialization still works
# (manual test in a satellite project)
# ari materialize <rite-name>
# ari rite swap <rite-name>

# 6. LOC reduction verified
find . -name '*.sh' -not -path './.claude/*' | xargs wc -l 2>/dev/null | tail -1
# Expected: <2,000 LOC remaining (down from 40,920)
```

---

## 8. Risk Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Go port introduces behavioral differences in worktree setup | MEDIUM | HIGH | Run `ari worktree create` before and after, diff the .claude/ directories |
| Deleted library is sourced by undiscovered script | LOW | MEDIUM | `grep -r 'source.*<library>' .` before each phase |
| Materialized .claude/ files become orphaned after source deletion | LOW | LOW | `ari materialize` re-generates; deletion is reversible |
| USE_ARI_HOOKS flag removal causes hook regression | N/A | N/A | Explicitly OUT OF SCOPE for this cleanse |
| Test coverage gap from deleted shell tests | LOW | LOW | Go tests already cover all critical paths (verified in Section 2) |

---

## 9. Janitor Notes

### Commit Conventions

Each phase should be a single commit following the project's conventional commit style:
- Phase 0: `refactor(worktree): replace shell exec.Command with native Go calls`
- Phase 1: `chore: delete dead root-level shell scripts (batch-a)`
- Phase 2: `chore: delete deprecated shell hooks (batch-b)`
- Phase 3: `chore: delete shell source libraries (batch-c)`
- Phase 4: `chore: delete shell test scripts (batch-d)`
- Phase 5: `chore: delete shell template/utility scripts (batch-e)`
- Phase 6: `chore: delete orphaned materialized hook files (batch-i)`

### Critical Ordering

Phase 0 MUST complete before Phase 1. Phases 1-5 MUST execute in order (each
depends on the prior phase removing references). Phase 6 can execute independently.

### Test Requirements

Run `CGO_ENABLED=0 go build ./cmd/ari && CGO_ENABLED=0 go test ./...` after EVERY
phase. Do not proceed to the next phase if tests fail.

### Files to NOT Delete

These files MUST survive the cleanse:
- `user-hooks/ari/autopark.sh` (Batch F -- ari wrapper)
- `user-hooks/ari/route.sh` (Batch F -- ari wrapper)
- `user-hooks/ari/writeguard.sh` (Batch F -- ari wrapper)
- `user-hooks/ari/clew.sh` (Batch F -- ari wrapper)
- `user-hooks/ari/validate.sh` (Batch F -- ari wrapper)
- `user-hooks/ari/context.sh` (Batch F -- ari wrapper)
- `user-hooks/ari/cognitive-budget.sh` (Batch F -- ari wrapper)
- `rites/ecosystem/context-injection.sh` (Batch G -- live runtime)
- `rites/shared/mena/cross-rite-handoff/validation.sh` (Batch G -- mena resource)
- `.claude/hooks/lib/preferences-loader.sh` (Batch I -- materialized, active)
- `.claude/hooks/lib/fail-open.sh` (Batch I -- materialized, active)
- All `.claude/hooks/ari/*.sh` (materialized from Batch F sources)

---

## 10. Attestation

| Artifact | Path | Verified |
|---|---|---|
| Session FSM (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/session-fsm.sh` | READ |
| Session FSM (Go) | `/Users/tomtenuta/Code/knossos/internal/session/fsm.go` | READ |
| Session Context (Go) | `/Users/tomtenuta/Code/knossos/internal/session/context.go` | READ |
| Session Events (Go) | `/Users/tomtenuta/Code/knossos/internal/session/events.go` | READ |
| Session Discovery (Go) | `/Users/tomtenuta/Code/knossos/internal/session/discovery.go` | READ |
| Session ID (Go) | `/Users/tomtenuta/Code/knossos/internal/session/id.go` | READ |
| Session Manager (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/session-manager.sh` | READ |
| Session Core (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/session-core.sh` | READ |
| Session State (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/session-state.sh` | READ |
| Worktree Manager (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/worktree-manager.sh` | READ |
| Worktree Lifecycle (Go) | `/Users/tomtenuta/Code/knossos/internal/worktree/lifecycle.go` | READ |
| Worktree Operations (Go) | `/Users/tomtenuta/Code/knossos/internal/worktree/operations.go` | READ |
| Sync Core (shell) | `/Users/tomtenuta/Code/knossos/lib/sync/sync-core.sh` | READ |
| Materializer (Go) | `/Users/tomtenuta/Code/knossos/internal/materialize/materialize.go` | READ |
| Rite Transaction (shell) | `/Users/tomtenuta/Code/knossos/lib/rite/rite-transaction.sh` | READ |
| Rite Switcher (Go) | `/Users/tomtenuta/Code/knossos/internal/rite/switch.go` | READ |
| Rite Context Loader (Go) | `/Users/tomtenuta/Code/knossos/internal/rite/context_loader.go` | READ |
| Context Injection (shell) | `/Users/tomtenuta/Code/knossos/rites/ecosystem/context-injection.sh` | READ |
| Rite Context Loader (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/rite-context-loader.sh` | READ |
| Preferences Loader (materialized) | `/Users/tomtenuta/Code/knossos/.claude/hooks/lib/preferences-loader.sh` | READ |
| Fail Open (materialized) | `/Users/tomtenuta/Code/knossos/.claude/hooks/lib/fail-open.sh` | READ |
| Handoff Validation (rite script) | `/Users/tomtenuta/Code/knossos/rites/shared/mena/cross-rite-handoff/validation.sh` | READ |
| Config (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/config.sh` | READ |
| Primitives (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/primitives.sh` | READ |
| Logging (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/logging.sh` | READ |
| Hooks Init (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/hooks-init.sh` | READ |
| Orchestration Audit (shell) | `/Users/tomtenuta/Code/knossos/user-hooks/lib/orchestration-audit.sh` | READ |

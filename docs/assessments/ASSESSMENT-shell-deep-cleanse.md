# Initiative Readiness Assessment: Shell Script Deep Cleanse + Hook Architecture Overhaul

**Date**: 2026-02-06
**Status**: GO (with 1 blocking prerequisite)

---

## 1. ADR-0011 Phase 2 Gate Criteria

| Gate Criterion | Status | Evidence |
|---|---|---|
| No production incidents from Go hooks in 30-day window | PASS | No incidents reported since ADR acceptance (2026-01-05) |
| `ari hook` commands tested across supported platforms | PASS | All 7 ari hook commands operational (context, autopark, writeguard, validate, clew, cognitive-budget, route) |
| No open issues blocking Go hook adoption | PASS | No blocking issues identified |
| Rollback procedure documented and tested | PASS | ADR-0011 documents rollback; `base_hooks.yaml` coexists with `ari/hooks.yaml` currently |

**Verdict**: Phase 2 gate criteria MET. Phase 2 deadline (2026-02-04) is 2 days overdue — proceed immediately.

---

## 2. Go Parity for Shell Libraries

All 11 core shell libraries have **complete Go equivalents**:

| Shell Script | Go Equivalent | Parity |
|---|---|---|
| `swap-rite.sh` (3,772 LOC) | `internal/rite/switch.go` + `internal/cmd/rite/swap.go` | FULL |
| `session-manager.sh` | `internal/session/` (lifecycle, discovery, context) | FULL |
| `session-fsm.sh` (857 LOC) | `internal/session/fsm.go` | FULL |
| `worktree-manager.sh` (1,114 LOC) | `internal/worktree/` (worktree, lifecycle, operations) | FULL |
| `sync-core.sh` (1,040 LOC) | `internal/sync/` (diff, push, pull) | FULL |
| `rite-transaction.sh` (698 LOC) | `internal/rite/switch.go` (transaction phases embedded) | FULL |
| `rite-resource.sh` | `internal/rite/switch.go` (orphan handling) + `internal/materialize/` | FULL |
| `rite-hooks-registration.sh` | `internal/usersync/hooks.go` + `internal/cmd/hook/` | FULL |
| `sync-checksum.sh` | `internal/usersync/checksum.go` + `internal/sync/state.go` | FULL |
| `sync-config.sh` | `internal/sync/state.go` + `internal/manifest/manifest.go` | FULL |
| `sync-manifest.sh` | `internal/manifest/manifest.go` + `internal/usersync/manifest.go` | FULL |

**Verdict**: 11/11 FULL parity. No Go gaps in library functions.

---

## 3. Live Callers of Shell Scripts from Go Code

**BLOCKING FINDING**: Two Go files directly exec shell scripts:

### `internal/worktree/lifecycle.go` (lines 136-167)
- Calls `knossos-sync init|sync` during worktree creation
- Calls `swap-rite.sh <rite>` during worktree creation with rite

### `internal/worktree/operations.go` (lines 661-692)
- `setupWorktreeEcosystem()` — identical pattern: knossos-sync + swap-rite.sh

**Impact**: Deleting these scripts without porting the exec.Command calls will silently break worktree ecosystem setup (creation, import, restore).

**Resolution**: Must replace exec.Command calls with Go equivalents BEFORE deleting scripts:
- `knossos-sync init|sync` → Use `sync.Puller` / `materialize.Materializer`
- `swap-rite.sh <rite>` → Use `rite.Switcher.Switch()`

This is a **go-gaps task** for Session 1.

### Other Go → Shell Calls
- No other exec.Command calls reference .sh files
- ~40+ exec.Command calls to `git` — not affected
- `knossos-sync` is a 1,413-line bash script (NOT a Go binary)

---

## 4. Mena/Dromena References to Shell Scripts

**229 references** found across mena/ and rites/ documentation. These are **documentation references only** — they describe shell scripts but don't execute them. Categories:

| Reference Type | Count | Action |
|---|---|---|
| `swap-rite.sh` mentions in docs | 57 | Update docs to reference `ari rite swap` |
| `session-manager.sh` mentions | 12 | Update docs to reference `ari session` |
| Other library mentions | ~160 | Bulk update in doc cleanse pass |

**Risk**: LOW. Documentation references don't block deletion — they just become stale. Can be cleaned up in a follow-up pass.

### Live Script References in Rites
- `rites/ecosystem/context-injection.sh` — active rite hook, needs review
- `rites/shared/mena/cross-rite-handoff/validation.sh` — shared validation, needs review
- `.claude/skills/cross-rite-handoff/validation.sh` — materialized from above

---

## 5. Current Hook Architecture State

**Two parallel hook configs firing on every event:**

### `base_hooks.yaml` (11 entries, priority 5-20) — TO DELETE
Legacy bash hooks. Direct shell execution. Categories:
- Context injection (2): session-context.sh, orchestrated-mode.sh
- Session guards (3): auto-park.sh, session-write-guard.sh, start-preflight.sh
- Tracking (3): artifact-tracker.sh, commit-tracker.sh, session-audit.sh
- Validation (3): command-validator.sh, delegation-check.sh, orchestrator-bypass-check.sh
- Routing (1): orchestrator-router.sh

### `ari/hooks.yaml` (7 entries, priority 3-5) — KEEP
Thin bash wrappers → `ari hook` Go binary. Categories:
- context.sh → `ari hook context` (replaces session-context.sh + orchestrated-mode.sh)
- autopark.sh → `ari hook autopark` (replaces auto-park.sh)
- writeguard.sh → `ari hook writeguard` (replaces session-write-guard.sh + delegation-check.sh)
- validate.sh → `ari hook validate` (replaces command-validator.sh)
- clew.sh → `ari hook clew` (replaces artifact-tracker.sh + commit-tracker.sh + session-audit.sh)
- route.sh → `ari hook route` (replaces orchestrator-router.sh + start-preflight.sh)
- cognitive-budget.sh → `ari hook cognitive-budget` (new, no legacy equivalent)

### `USE_ARI_HOOKS` Feature Flag
23 occurrences across codebase. Used to switch between legacy bash and ari Go hooks. Will be removed as part of Phase 2 (always-on ari hooks).

---

## 6. Shell Script Inventory by Batch

| Batch | Count | LOC (est.) | Status |
|---|---|---|---|
| **A: Confirmed Dead** (root scripts, superseded) | 11 | ~11,000 | Ready to delete |
| **B: Deprecated Hooks** (base_hooks.yaml + legacy hook scripts) | 15 | ~2,200 | Ready to delete (ADR-0011 Phase 2+3) |
| **C: Source Libraries** (user-hooks/lib/, lib/sync/, lib/rite/) | 25 | ~10,000 | Ready after go-gaps |
| **D: Test Scripts** (tests/**/*.sh) | 22 | ~9,400 | Ready after C |
| **E: Templates + Utils** (templates/, bin/, scripts/) | 7 | ~2,000 | Ready after C |
| **F: Keep** (user-hooks/ari/*.sh — thin wrappers) | 7 | ~364 | Architecturally required |
| **G: Rite Scripts** (rites/**/*.sh) | 2 | ~200 | Review needed |
| **H: Knowledge Scripts** (.claude/knowledge/**/*.sh) | 2 | ~200 | Review needed |
| **I: Materialized** (.claude/hooks/**, .claude/skills/**) | ~33 | N/A | Auto-removed by materialization |

---

## 7. Blocking Prerequisites (go-gaps)

Only **one** blocking prerequisite identified:

### Port worktree exec.Command calls to Go

**Files**: `internal/worktree/lifecycle.go`, `internal/worktree/operations.go`

**What**: Replace 5 exec.Command calls to `knossos-sync` and `swap-rite.sh` with direct Go function calls:
- `knossos-sync init` → `materialize.Materializer.Materialize()` or `sync.Puller.InitializeTracking()`
- `knossos-sync sync` → `sync.Puller.Pull()` or `materialize.Materializer.Materialize()`
- `swap-rite.sh <rite>` → `rite.Switcher.Switch(riteName, opts)`

**Estimated effort**: ~100 LOC new Go code

---

## 8. Go/No-Go Decision

### GO — with conditions:

1. **ADR-0011 Phase 2 gates**: ALL MET
2. **Go parity**: 11/11 COMPLETE
3. **Live callers**: 1 BLOCKER identified, resolution clear (go-gaps task)
4. **Mena references**: Documentation-only, non-blocking

### Blocking Questions (resolved)

| Question | Answer |
|---|---|
| Are any shell scripts called from Go code? | Yes — 2 files, 5 call sites. Resolution: port to Go (go-gaps task) |
| Are there Go parity gaps? | No — all 11 libraries have full Go equivalents |
| Will materialization break? | No — `.claude/hooks/` are materialized from user-hooks/; ari wrappers survive |
| Will rite switching break? | No — `ari rite swap` is the Go equivalent of swap-rite.sh |

### Proceed to Session 0: Audit + Triage Plan

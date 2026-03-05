# ADR-0029: Worktree Environment Contract

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-03-02 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

Git worktrees allow multiple checkouts of the same repository to coexist simultaneously. Claude Code's `EnterWorktree` lifecycle event, `ari worktree create`, and manual `git worktree add` each produce a linked worktree — a working directory that shares git history with the main worktree but has its own checkout of a different branch.

Knossos's runtime state lives in five gitignored directories (`.claude/`, `.sos/`, `.knossos/`, `.know/`, `.ledge/`). Because these directories are gitignored, `git worktree add` does not copy them into the linked worktree. The linked worktree starts with none of the framework directories that `ari sync`, hook scripts, and session management require.

**Spike Reference**: `docs/spikes/SPIKE-worktree-sync-breakage.md` — full call chain analysis, blast radius, and remediation specification.

### Root Causes Identified

Three distinct failure modes were identified in the spike:

| Root Cause | Effect | Severity |
|------------|--------|----------|
| RC-1: `.claude/` absent in linked worktree | `ari sync` falls to minimal mode; `MaterializeMinimal()` actively removes manually placed ACTIVE_RITE | CRITICAL |
| RC-2: `ari sync` does not inherit ACTIVE_RITE from main worktree | Rite materialisation skipped, CLAUDE.md degraded, agents unavailable | CRITICAL |
| RC-3: Collision checker reads empty provenance | User-scope writes bypass collision detection; cross-project contamination via USER_PROVENANCE_MANIFEST.yaml | HIGH |
| RC-4: `listAvailableRites()` read only `.knossos/rites/` | Context hook reported zero available rites even in main worktree | MEDIUM |

---

## Decision

### Worktree Lifecycle Contract

A **knossos worktree environment** is fully functional when the following framework directories exist and are seeded:

| Directory | Criticality | Created by |
|-----------|-------------|-----------|
| `.claude/` | CRITICAL | `ari sync` (via `ensureProjectDirs`) or worktree creation |
| `.sos/sessions/` | HIGH | `ari sync` (via `ensureProjectDirs`) or worktree creation |
| `.knossos/` | MEDIUM | `ari sync` (via `ensureProjectDirs`) or worktree creation |
| `.know/` | LOW | `/know` command only (optional, not seeded automatically) |
| `.ledge/` | LOW | `ari init` only (optional, not seeded automatically) |

**Invariant**: `ari sync` run from any directory (main or linked worktree) must create the CRITICAL and HIGH directories before any sync logic executes. This is implemented via `Materializer.ensureProjectDirs()` called at the top of `Sync()` in `internal/materialize/materialize.go`.

### Three Worktree Creation Paths

Three paths exist for creating a linked worktree. Each path results in a different seeding state:

#### Path 1: `ari worktree create`

`ari worktree create` calls `internal/worktree/lifecycle.go:Create()` which calls `setupWorktreeEcosystem()`. This seeds `.claude/`, `.sos/`, `.knossos/` and writes `ACTIVE_RITE` from the main worktree. This is the **fully seeded** path — all required directories exist immediately after creation.

#### Path 2: CC `EnterWorktree` (WorktreeCreate hook)

Claude Code fires a `WorktreeCreate` hook event when a worktree is created within a CC conversation. The hook handler (`ari hook worktree-seed`) reads the main worktree's ACTIVE_RITE and runs `ari sync --scope=rite` in the worktree, seeding `.claude/` with the inherited rite. The `ensureProjectDirs()` pre-flight ensures `.sos/sessions/` and `.knossos/` are also created.

This path requires the `WorktreeCreate` hook to be configured in `settings.local.json`. If not configured, CC creates the worktree without knossos seeding, and the user falls to Path 3 on their first explicit `ari sync`.

#### Path 3: Manual `git worktree add`

Manual worktree creation produces an unseeded state. The worktree has no `.claude/`, `.sos/`, or `.knossos/`. The worktree becomes functional when the user runs `ari sync` (with or without a `--rite` flag):

1. `ensureProjectDirs()` creates `.claude/`, `.sos/sessions/`, `.knossos/`
2. `syncRiteScope()` detects the linked worktree via `isGitWorktree()` and inherits ACTIVE_RITE from the main worktree
3. Full rite materialisation proceeds with the inherited rite
4. Session lifecycle becomes available after sync completes

### Rite Inheritance

Linked worktrees **inherit their active rite from the main worktree** when no explicit `--rite` flag is passed and no local ACTIVE_RITE exists.

**Implementation**: `syncRiteScope()` in `internal/materialize/materialize.go` calls `isGitWorktree()` and `getMainWorktreeDir()` before falling through to minimal mode. If the main worktree's `.knossos/ACTIVE_RITE` contains a rite name, it is used for the current sync and written into the worktree's `.knossos/ACTIVE_RITE`.

**Rationale**: Worktrees are created for parallel work on the same project, typically under the same rite context. Inheriting the rite is the lowest-friction default. Users who want a different rite can pass `--rite` explicitly or edit ACTIVE_RITE after sync.

**Worktree detection**: `isGitWorktree(dir)` uses `git rev-parse --git-common-dir`. A linked worktree returns an absolute path (the shared `.git/` directory). The main worktree returns the relative path `.git`. This distinction is O(1) via subprocess call (~5ms).

### Collision Checker Fallback

The user-scope sync uses a `CollisionChecker` that reads from the project's `PROVENANCE_MANIFEST.yaml` to detect which resources are rite-owned and must not be overwritten by user-scope writes.

In a fresh linked worktree, the project `.knossos/PROVENANCE_MANIFEST.yaml` does not exist (`.claude/` is absent). The checker falls back in two stages:

1. **Main worktree fallback**: If the checker is ineffective (manifest absent), and `worktreeMainDir()` succeeds, a new checker is constructed from the main worktree's `.claude/`. This provides accurate collision detection using the main worktree's rite provenance.

2. **Fail-closed guard**: If both the local and main worktree checkers are ineffective AND an ACTIVE_RITE is present, all user-scope writes are skipped. This prevents cross-project contamination of `USER_PROVENANCE_MANIFEST.yaml`.

3. **No-rite path**: If both checkers are ineffective AND no ACTIVE_RITE is set, there is no rite to protect. User-scope sync proceeds without collision checking — this is the correct behaviour for minimal/cross-cutting mode.

### Session Isolation

Each worktree maintains its own session lifecycle. Sessions are stored in `.sos/sessions/` which is gitignored. A linked worktree created via any path starts with an empty `.sos/sessions/`. The first `/start` or `ari session create` in the worktree creates a new session scoped to that worktree.

**Sessions are not shared between worktrees.** This is the intended design. Parallel work in parallel worktrees produces parallel, independent session histories. The CC session map (`.sos/sessions/.cc-map/`) maintains a per-CC-conversation pointer to the worktree's active session.

### Rite Discovery in Context Hook

The `SessionStart` context hook reports available rites in `ContextOutput.AvailableRites`. The hook uses `source.SourceResolver.ListAvailableRites()` which checks all four resolution tiers:

1. Project-local rites (`.knossos/rites/`)
2. User-level rites (`~/.claude/rites/`)
3. Knossos platform rites (`$KNOSSOS_HOME/rites/`)
4. Embedded rites (compiled into the `ari` binary)

This is correct for both main and linked worktrees. The previous `listAvailableRites(resolver.RitesDir())` only read `.knossos/rites/` and returned empty for all projects without satellite rites — a bug that this ADR's implementation corrects.

---

## Consequences

### Positive

1. **`ari sync` works in any worktree created by any mechanism.** The three-path contract covers `ari worktree create`, CC `EnterWorktree`, and manual `git worktree add` without requiring a specific creation path.

2. **Rite inheritance is zero-friction.** Developers working in parallel worktrees on the same project get the correct rite context without remembering to pass `--rite`.

3. **User-scope contamination is prevented.** The fail-closed collision checker guard stops `USER_PROVENANCE_MANIFEST.yaml` from being corrupted by worktree syncs with empty provenance.

4. **Context hook is always populated.** `AvailableRites` is non-empty for any project using knossos, regardless of whether `.knossos/rites/` has satellite rites.

5. **Session isolation is correct.** Each worktree has its own session history. No session state leaks between worktrees.

### Negative

1. **`isGitWorktree()` adds ~5ms to `ari sync` on the no-ACTIVE_RITE fallback path.** This only triggers when no `--rite` flag is passed and no local ACTIVE_RITE exists — the minimal-mode path. It does not affect the hot path (ACTIVE_RITE present).

2. **Manual worktrees require one explicit `ari sync` before hooks function.** The autopark, clew, and writeguard hooks all depend on `.sos/sessions/` and `.claude/`. They silently fail or skip until `ari sync` has run. This is unavoidable without a WorktreeCreate hook configured.

3. **The WorktreeCreate hook payload schema is not officially documented.** The `ari hook worktree-seed` implementation uses `CLAUDE_PROJECT_DIR` to determine the worktree path, which is set by CC before executing hooks. If CC changes this behaviour, the hook will require updating.

### Neutral

1. **`.know/` and `.ledge/` are intentionally NOT seeded by `ensureProjectDirs()`.** These directories have their own lifecycle: `/know` for codebase knowledge, `ari init` for work product structure. Auto-creating them in every worktree would be presumptuous.

2. **The `internal/worktree/` package and `internal/materialize/` package have overlapping git detection utilities.** `worktree.go` in each package contains `isGitWorktree`-equivalent logic to avoid a circular import (`userscope → materialize → worktree → materialize`). This is a pragmatic duplication consistent with Go's package design conventions.

---

## Implementation Files

| File | Role |
|------|------|
| `internal/materialize/worktree.go` | `isGitWorktree()`, `getMainWorktreeDir()`, `inheritRiteFromMainWorktree()` |
| `internal/materialize/materialize.go` | `ensureProjectDirs()`, `syncRiteScope()` worktree fallback |
| `internal/materialize/userscope/worktree.go` | `worktreeMainDir()` (local copy, avoids circular import) |
| `internal/materialize/userscope/sync.go` | Collision checker fallback + fail-closed guard |
| `internal/cmd/hook/context.go` | `SourceResolver.ListAvailableRites()` replacing `listAvailableRites()` |
| `test/worktree/testutil/worktree.go` | `SetupWorktreeTestFixture()` shared test helper |

## Test Coverage

| Package | Test | Priority |
|---------|------|----------|
| `internal/materialize` | `TestIsGitWorktree_*` (3 cases) | P2 |
| `internal/materialize` | `TestGetMainWorktreeDir_*` (2 cases) | P2 |
| `internal/materialize` | `TestSyncRiteScope_InWorktree_InheritsRite` | P0 |
| `internal/materialize` | `TestSyncRiteScope_InWorktree_NoMainRite_FallsToMinimal` | P0 |
| `internal/materialize/userscope` | `TestCollisionChecker_InWorktree_FallsBackToMainProvenance` | P1 |
| `internal/materialize/userscope` | `TestCollisionChecker_InWorktree_NoMainProvenance_FailsClosed` | P1 |
| `internal/materialize/userscope` | `TestWorktreeMainDir_*` (2 cases) | P2 |
| `internal/cmd/hook` | `TestContextHook_InWorktree_ReportsRites` | P1 |

---

## Alternatives Considered

### Alternative 1: Symlink `.claude/` from Main to Linked Worktree

Create a symlink at `linkedWorktree/.claude` pointing to `mainWorktree/.claude`.

**Rejected**: Shared `.claude/` means shared ACTIVE_RITE, shared agents, and shared settings. Worktrees are used for branch-parallel work that may require different rites. Symlinks also break if the main worktree is moved or renamed. Directory-per-worktree is the correct isolation model.

### Alternative 2: Require `ari worktree create` for All Worktrees

Document that `git worktree add` is unsupported; require all worktrees to be created via `ari worktree create`.

**Rejected**: CC's `EnterWorktree` creates worktrees outside ari's lifecycle. Requiring `ari worktree create` for all cases breaks the CC workflow and contradicts the goal of dogfooding knossos with CC. The three-path contract explicitly supports all creation mechanisms.

### Alternative 3: Store ACTIVE_RITE in Git Attributes

Use git attributes or worktree-config to store ACTIVE_RITE per-worktree in a tracked file.

**Rejected**: ACTIVE_RITE is runtime state, not repository state. Different developers on different machines may run different rites on the same branch. Storing ACTIVE_RITE in git would impose one rite choice on all developers. The gitignored `.knossos/ACTIVE_RITE` pattern is intentional.

---

## References

| Reference | Location |
|-----------|----------|
| Spike | `docs/spikes/SPIKE-worktree-sync-breakage.md` |
| Worktree lifecycle (existing) | `internal/worktree/lifecycle.go` — `Create()`, `setupWorktreeEcosystem()` |
| Git detection utilities | `internal/materialize/worktree.go` |
| Directory pre-flight | `internal/materialize/materialize.go` — `ensureProjectDirs()` |
| Collision checker | `internal/materialize/userscope/collision.go` — `IsEffective()` |
| Source resolver rite discovery | `internal/materialize/source/resolver.go` — `ListAvailableRites()` |
| ADR-0010 | `docs/decisions/ADR-0010-worktree-session-seeding.md` — earlier worktree session context |

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-03-02 | Claude Sonnet 4.6 (Integration Engineer) | Initial draft — worktree environment contract from spike + Sprint 1-3 implementation |

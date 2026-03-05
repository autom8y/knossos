# SPIKE: Worktree Sync Breakage -- Deep Investigation

**Date**: 2026-03-02
**Status**: Deep Investigation Complete
**Severity**: P1 (core framework function broken)
**Session**: session-20260302-012234-211412ef

## Question

Why does `ari sync` on a worktree (a) fail to materialize the rite and (b) potentially break other repos sharing the same KNOSSOS_HOME?

---

## Phase 1: Surface Area Analysis

### RC-1: `.claude/` is gitignored and not seeded in non-ari worktrees

#### Full Call Chain

| Step | File | Function | Line | Behavior |
|------|------|----------|------|----------|
| 1 | `.gitignore` | N/A | 7 | `.claude/` excluded from git |
| 2 | (external) | `git worktree add` / CC `EnterWorktree` | N/A | Creates worktree, `.claude/` absent |
| 3 | `internal/paths/paths.go` | `NewResolver(projectDir)` | 20 | Resolver points at worktree path |
| 4 | `internal/paths/paths.go` | `ClaudeDir()` | 59 | Returns `worktree/.claude/` (does not exist) |
| 5 | `internal/paths/paths.go` | `ReadActiveRite()` | 120-126 | Returns `""` (file not found, no error) |
| 6 | `internal/paths/paths.go` | `ActiveRiteFile()` | 114 | Returns `worktree/.knossos/ACTIVE_RITE` (absent) |

**All files involved:**
- `.gitignore` -- the root cause
- `internal/paths/paths.go` -- `Resolver`, `ClaudeDir()`, `ReadActiveRite()`
- `internal/worktree/operations.go` -- `setupWorktreeEcosystem()` (lines 656-680) is the *correct* seeding path
- `internal/worktree/lifecycle.go` -- `Create()` (line 135) correctly calls `setupWorktreeEcosystem()`
- `internal/worktree/metadata.go` -- `SavePerWorktreeMeta()` (line 234) writes `.worktree-meta.json`
- `.claude/settings.local.json` -- no `WorktreeCreate`/`WorktreeRemove` hooks defined

**Blast radius:**
1. **`ari sync`** -- falls through to minimal mode (RC-2 cascade)
2. **`ari hook context`** -- SessionStart returns degraded context, no rite, zero available rites
3. **`ari hook autopark`** -- Stop hook reads from `.sos/sessions/` which also does not exist in fresh worktrees; autopark silently skips
4. **`ari hook writeguard`** -- PreToolUse hooks reference `.claude/` paths; writeguard may fail or produce incorrect results
5. **`ari status`** -- `collectKnossos()` reads `.knossos/` which is also gitignored; reports non-existent
6. **`ari hook clew`** -- PostToolUse writes to `.sos/sessions/`; session dir does not exist, event drops silently

**Edge cases not in initial spike:**
- `.sos/` is also gitignored and absent in worktrees. Sessions cannot be created, parked, or continued in a fresh worktree. The entire session lifecycle is broken, not just sync.
- `.knossos/` is also gitignored. No satellite rites can be discovered.
- `.know/` is gitignored. The `knowStatus()` function in the context hook returns empty, losing codebase knowledge awareness.
- `.ledge/` is gitignored. `ari init`-scaffolded subdirectories (decisions/, specs/, reviews/, spikes/) are absent.
- **Hooks still fire** from `settings.local.json` because CC reads `settings.local.json` from the main worktree's `.claude/` (CC's `CLAUDE_PROJECT_DIR` points at the worktree, but settings merge uses git common dir). However, `ari hook context` fails to find session state, producing degraded output that is injected into every conversation.

#### Gitignored Directories -- Full Worktree Impact

| Directory | Gitignored | Worktree Effect | Criticality |
|-----------|------------|-----------------|-------------|
| `.claude/` | Yes | No rite, no CLAUDE.md, no agents, no hooks config, no provenance | CRITICAL |
| `.sos/` | Yes | No sessions, no locks, no CC map, no archive | HIGH |
| `.knossos/` | Yes | No satellite rites | MEDIUM |
| `.know/` | Yes | No codebase knowledge context | LOW |
| `.ledge/` | Yes | No work product artifact dirs | LOW |
| `.wip/` | Yes | No ephemeral scratch (by design) | NONE |

### RC-2: `ari sync` falls through to minimal mode without ACTIVE_RITE

#### Full Call Chain

| Step | File | Function | Line | Behavior |
|------|------|----------|------|----------|
| 1 | `internal/cmd/sync/sync.go` | `runSync()` | 130 | `projectDir, _ := os.Getwd()` |
| 2 | `internal/cmd/sync/sync.go` | `runSync()` | 137 | `resolver := paths.NewResolver(projectDir)` |
| 3 | `internal/cmd/sync/sync.go` | `runSync()` | 159 | `m.Sync(opts)` |
| 4 | `internal/materialize/materialize.go` | `Sync()` | 486-487 | `m.syncRiteScope(opts)` |
| 5 | `internal/materialize/materialize.go` | `syncRiteScope()` | 525-528 | Reads `ACTIVE_RITE` from `worktree/.knossos/ACTIVE_RITE` -- empty |
| 6 | `internal/materialize/materialize.go` | `syncRiteScope()` | 530-531 | `riteName == ""` and `previousRite == ""` |
| 7 | `internal/materialize/materialize.go` | `syncRiteScope()` | 536 | Falls to `syncRiteScopeMinimal(opts)` |
| 8 | `internal/materialize/materialize.go` | `syncRiteScopeMinimal()` | 578-580 | `legacyOpts := Options{Minimal: true}` |
| 9 | `internal/materialize/materialize.go` | `MaterializeMinimal()` | 227 | Creates minimal `.claude/` scaffold |
| 10 | `internal/materialize/materialize.go` | `MaterializeMinimal()` | 280-282 | **Actively removes** `ACTIVE_RITE`, `ACTIVE_WORKFLOW.yaml`, `INVOCATION_STATE.yaml` |

**Critical detail the spike identified but underemphasized**: Line 280-282 is destructive. `MaterializeMinimal()` calls `os.Remove()` on ACTIVE_RITE. This means if a user manually places an ACTIVE_RITE file and then runs `ari sync` with `scope=all` (default), the minimal path *deletes* it. The catch-22 is even worse than described: manual intervention is actively undone.

**Blast radius:**
1. **`ari sync` itself** -- produces incorrect output (status "minimal" instead of rite name)
2. **CLAUDE.md generation** -- `materializeMinimalCLAUDEmd()` produces a stripped-down CLAUDE.md without agent configs, rite-specific sections, or agent routing
3. **Orphan detection** -- skipped entirely in minimal mode (no rite manifest to compare against)
4. **Rite-switch cleanup** -- skipped (no previous rite to detect switch from)
5. **Provenance manifest** -- written with minimal entries, creating a divergent provenance state
6. **Settings generation** -- `materializeSettingsWithManifest()` receives `nil` manifest, producing bare settings

**Edge case not in initial spike:**
- `scope=rite` mode returns a hard error (`"no ACTIVE_RITE found, specify --rite"`) which is correct but unhelpful -- it doesn't mention the worktree context or suggest the workaround.

### RC-3: User-scope collision checker reads empty provenance from worktree

#### Full Call Chain

| Step | File | Function | Line | Behavior |
|------|------|----------|------|----------|
| 1 | `internal/materialize/materialize.go` | `Sync()` | 500 | `m.syncUserScope(opts)` |
| 2 | `internal/materialize/user_scope.go` | `syncUserScope()` | 9-22 | Delegates to `userscope.SyncUserScope()` with `Resolver` |
| 3 | `internal/materialize/userscope/sync.go` | `syncUserScope()` | 73 | `projectClaudeDir := s.resolver.ClaudeDir()` = `worktree/.claude/` |
| 4 | `internal/materialize/userscope/sync.go` | `syncUserScope()` | 74 | `collisionChecker := NewCollisionChecker(projectClaudeDir)` |
| 5 | `internal/materialize/userscope/collision.go` | `NewCollisionChecker()` | 18-24 | Calls `c.loadRiteManifest(claudeDir)` |
| 6 | `internal/materialize/userscope/collision.go` | `loadRiteManifest()` | 26-38 | `provenance.Load()` fails (no file), `riteEntries` stays empty |
| 7 | `internal/materialize/userscope/collision.go` | `CheckCollision()` | 46 | `len(c.riteEntries) == 0` -> returns `false, ""` for ALL checks |
| 8 | `internal/materialize/userscope/sync.go` | `syncUserResource()` | 126+ | All resources pass collision check, potentially overwriting |

**Critical nuance**: The collision checker *does* set `manifestLoaded = true` on line 27, even when the subsequent `provenance.Load()` fails. The check on line 46 (`!c.manifestLoaded || len(c.riteEntries) == 0`) catches this via the `len == 0` clause, so collision checking is effectively disabled but the checker *believes* it loaded successfully. This is a silent degradation -- no log, no warning, no error.

**Blast radius:**
1. **`~/.claude/agents/`** -- User-scope agents synced from KNOSSOS_HOME could overwrite files that should be shadowed by rite-scope agents in a proper sync
2. **`~/.claude/commands/` and `~/.claude/skills/`** -- Mena entries synced without collision awareness
3. **`USER_PROVENANCE_MANIFEST.yaml`** -- Entries written for resources that should have been collision-skipped; these persist across all projects
4. **Cross-project contamination** -- A worktree sync with empty collision checker writes entries to `~/.claude/` that then affect subsequent syncs from the main repo or other projects

**Edge case not in initial spike:**
- The user-scope sync's Phase 1 snapshot (line 213-220 in sync.go) reads existing `USER_PROVENANCE_MANIFEST.yaml` entries. If a previous correct sync had marked certain resources as knossos-owned, the orphan cleanup phase could incorrectly remove them because the collision checker lets everything through.
- When `scope=all` and rite scope produces "minimal", user scope still runs. There is no gate that says "if rite scope degraded, also degrade user scope." The two phases are independent.

### RC-4: `RitesDir()` now points to `.knossos/rites/` which does not exist in worktrees

#### Full Call Chain

| Step | File | Function | Line | Behavior |
|------|------|----------|------|----------|
| 1 | `internal/paths/paths.go` | `KnossosDir()` | 154-156 | Returns `projectRoot/.knossos/` |
| 2 | `internal/paths/paths.go` | `RitesDir()` | 159-161 | Returns `projectRoot/.knossos/rites/` |
| 3 | `internal/cmd/hook/context.go` | `runContextCore()` | 174 | `listAvailableRites(resolver.RitesDir())` |
| 4 | `internal/cmd/hook/context.go` | `listAvailableRites()` | 277-293 | `os.ReadDir()` fails, returns `nil` |
| 5 | `internal/cmd/hook/context.go` | `runContextCore()` | 178 | `AvailableRites: nil` in output |

**Additional consumers of `RitesDir()`:**
- `internal/paths/paths.go` `RiteDir(riteName)` (line 198-206): Uses `RitesDir()` to check project satellite rites first. Falls through to `UserRitesDir()` -- this works but is misleading.
- `internal/cmd/status/status.go` `collectKnossos()` (line 272+): Reads `.knossos/` health, reports rites directory state. Shows non-existent in worktrees.

**Blast radius:**
1. **SessionStart context injection** -- zero available rites in JSON output, degraded agent experience
2. **`ari status`** -- `.knossos/` section shows as non-existent
3. **`Resolver.RiteDir(riteName)`** -- falls through to user rites, works but incorrect provenance tracking

**Edge case not in initial spike:**
- This bug exists in the main repo too, not just worktrees. `.knossos/rites/` is empty in the main repo; rites come from `KNOSSOS_HOME/rites/`. The context hook has *always* returned empty `AvailableRites` for projects without satellite rites. The worktree just makes it more visible.
- The `SourceResolver.ListAvailableRites()` (line 262-317 in resolver.go) correctly checks all 4 tiers. This function is not used by the context hook, which is the gap.

---

## Phase 2: Remediation Assessment

### Fix 1: WorktreeCreate/WorktreeRemove hooks in settings.local.json

#### Spec

**New file**: `internal/cmd/hook/worktreeseed.go`

**New function**: `runWorktreeSeed(ctx *cmdContext) error`

Logic:
1. Read stdin JSON for CC hook payload (contains worktree path)
2. Use `git rev-parse --git-common-dir` or `git worktree list --porcelain` to find main worktree
3. Read `mainWorktree/.knossos/ACTIVE_RITE`
4. Call `materialize.NewMaterializer(paths.NewResolver(worktreePath)).Sync(SyncOptions{RiteName: rite, Scope: ScopeRite})`
5. Scaffold `.sos/`, `.knossos/`, `.know/` symlinks or copies

**Files to change:**

| File | Change | LOC |
|------|--------|-----|
| `internal/cmd/hook/worktreeseed.go` | New file: `newWorktreeSeedCmd()`, `runWorktreeSeed()` | ~80 |
| `internal/cmd/hook/hook.go` | Register new subcommand: `cmd.AddCommand(newWorktreeSeedCmd(ctx))` | ~2 |
| `internal/materialize/materialize.go` | Add hook entry to `writeDefaultSettings()` for `WorktreeCreate` | ~10 |
| Total | | ~92 |

**Test coverage:**
- Existing: None for worktree hooks
- Needed: `TestWorktreeSeedHook_InheritsRite`, `TestWorktreeSeedHook_NoMainRite`, `TestWorktreeSeedHook_AlreadySeeded`

**Dependencies**: None. Can be implemented independently.

**Risk of regression**: LOW. New hook, new code path. No existing behavior changed. Risk: CC's `WorktreeCreate` hook payload format is undocumented -- need to verify stdin JSON schema matches expectations.

**Limitation**: Only covers CC's `EnterWorktree` path. Does not cover manual `git worktree add`.

#### CC Hook Payload Uncertainty

The initial spike assumes CC sends worktree path in the hook payload. This needs verification. CC's hook stdin JSON schema for `WorktreeCreate` may differ from `SessionStart`/`PreToolUse`. The `internal/hook/env.go` `StdinPayload` struct needs to be checked against CC's actual behavior. If the worktree path is not in stdin, we may need to use `CLAUDE_PROJECT_DIR` env var (which CC sets to the worktree path after creation).

### Fix 2: Make `ari sync` worktree-aware

#### Spec

**Core change**: Add worktree detection to `syncRiteScope()` in `materialize.go`, not `runSync()` in `sync.go`. The materializer is the right layer because the hook seeding (Fix 1) also uses the materializer directly.

**New file**: `internal/materialize/worktree.go`

Functions:
```go
// isGitWorktree checks if projectDir is a linked git worktree (not main).
func isGitWorktree(projectDir string) bool

// getMainWorktreeDir returns the main worktree path for a linked worktree.
// Uses `git rev-parse --git-common-dir` which is faster than `git worktree list`.
func getMainWorktreeDir(projectDir string) (string, error)

// inheritRiteFromMainWorktree reads ACTIVE_RITE from the main worktree.
func inheritRiteFromMainWorktree(mainDir string) string
```

**Files to change:**

| File | Change | LOC |
|------|--------|-----|
| `internal/materialize/worktree.go` | New file: worktree detection utilities | ~50 |
| `internal/materialize/materialize.go` | Modify `syncRiteScope()`: add worktree fallback before minimal path (after line 531) | ~15 |
| `internal/materialize/userscope/sync.go` | Modify `syncUserScope()`: pass main worktree claudeDir to collision checker when in worktree | ~10 |
| Total | | ~75 |

**Key implementation detail -- `syncRiteScope()` change:**
```go
// After line 531 (if riteName == "" && previousRite == ""):
// Before falling through to minimal, check if we're in a worktree
if isGitWorktree(m.resolver.ProjectRoot()) {
    mainDir, err := getMainWorktreeDir(m.resolver.ProjectRoot())
    if err == nil {
        mainRite := inheritRiteFromMainWorktree(mainDir)
        if mainRite != "" {
            riteName = mainRite
            // Continue to full MaterializeWithOptions path
        }
    }
}
// Only fall through to minimal if riteName still empty
if riteName == "" {
    if opts.Scope == ScopeRite {
        return nil, fmt.Errorf("no ACTIVE_RITE found, specify --rite")
    }
    return m.syncRiteScopeMinimal(opts)
}
```

**Why `git rev-parse --git-common-dir` instead of the existing `GitOperations.GetMainWorktree()`:**
- `git rev-parse --git-common-dir` is a single subprocess call (~5ms)
- `git worktree list --porcelain` (used by `GetMainWorktree()`) is slower (~30ms) and outputs all worktrees
- `git-common-dir` returns the shared `.git/` directory; the main worktree is its parent's parent
- The materialize package should NOT import `internal/worktree` (circular dependency risk)

**Test coverage:**
- Existing: Zero tests for worktree scenarios in materialize
- Needed: `TestSyncRiteScope_WorktreeFallback`, `TestSyncRiteScope_WorktreeNoMainRite`, `TestIsGitWorktree`, `TestGetMainWorktreeDir`
- Integration: `TestSyncInWorktree_EndToEnd` (creates git repo, adds worktree, runs sync, verifies ACTIVE_RITE inherited)

**Dependencies**: None, but implementing Fix 3 and Fix 4 alongside is natural since the worktree detection code is shared.

**Risk of regression**: LOW-MEDIUM.
- The worktree detection adds a `git` subprocess call to every `syncRiteScope()` invocation where riteName is empty. This is only the fallback path (no `--rite` flag, no ACTIVE_RITE file), so it does not affect the hot path.
- If `git rev-parse --git-common-dir` fails (e.g., not a git repo), the code falls through to existing minimal behavior -- graceful degradation.
- Risk: `git rev-parse --git-common-dir` in the main worktree returns `.git` (relative), not the worktree path. Need to handle both relative and absolute paths, and detect main vs. linked worktrees.

### Fix 3: Fix `listAvailableRites()` to use SourceResolver

#### Spec

**Files to change:**

| File | Change | LOC |
|------|--------|-----|
| `internal/cmd/hook/context.go` | Replace `listAvailableRites(resolver.RitesDir())` with `SourceResolver.ListAvailableRites()` | ~15 |
| `internal/cmd/hook/context.go` | Add `source.NewSourceResolver()` setup, wire embedded FS | ~10 |
| `internal/cmd/hook/context_test.go` | Update `TestRunContext_WithActiveSession_IncludesRitesAndAgents` to use KNOSSOS_HOME rites | ~15 |
| Total | | ~40 |

**Implementation:**
```go
// Replace line 174:
// availableRites := listAvailableRites(resolver.RitesDir())
// With:
srcResolver := source.NewSourceResolver(resolver.ProjectRoot())
if embRites := common.EmbeddedRites(); embRites != nil {
    srcResolver.WithEmbeddedFS(embRites)
}
resolvedRites, _ := srcResolver.ListAvailableRites()
availableRites := make([]string, len(resolvedRites))
for i, r := range resolvedRites {
    availableRites[i] = r.Name
}
```

**Dependencies**: None.

**Risk of regression**: LOW but has performance implications.
- `SourceResolver.ListAvailableRites()` checks 4 tiers (project, user, knossos, embedded), reading multiple directories
- Current `listAvailableRites()` reads one directory
- In the context hook's 100ms time budget, this adds ~5-10ms of directory reads
- The embedded FS check (`fs.ReadDir(embFS, "rites")`) is in-memory and fast

**Test gap**: The existing test `TestRunContext_WithActiveSession_IncludesRitesAndAgents` creates rites in `.knossos/rites/`. This test passes because the test sets up that directory explicitly. After the fix, the test would also discover rites from KNOSSOS_HOME, so the test needs to set `KNOSSOS_HOME` to a temp dir to avoid environmental contamination.

### Fix 4: Guard user-scope collision checker

#### Spec

**Files to change:**

| File | Change | LOC |
|------|--------|-----|
| `internal/materialize/userscope/collision.go` | Add `IsEffective()` method and modify `loadRiteManifest()` to distinguish "no manifest" from "empty manifest" | ~10 |
| `internal/materialize/userscope/sync.go` | Add guard: if `!collisionChecker.IsEffective()`, log warning and set all collision checks to `true` (safe default = skip) | ~8 |
| `internal/materialize/userscope/collision_test.go` | Add `TestCollisionChecker_MissingManifest_ReportsIneffective`, `TestCollisionChecker_EmptyManifest_IsEffective` | ~25 |
| Total | | ~43 |

**Alternative approach (recommended): Fall back to main worktree provenance**

If Fix 2 is implemented (worktree detection), the collision checker can use the main worktree's provenance:

```go
projectClaudeDir := s.resolver.ClaudeDir()
collisionChecker := NewCollisionChecker(projectClaudeDir)
if !collisionChecker.IsEffective() && isGitWorktree(s.resolver.ProjectRoot()) {
    mainDir, _ := getMainWorktreeDir(s.resolver.ProjectRoot())
    if mainDir != "" {
        mainClaudeDir := filepath.Join(mainDir, ".claude")
        collisionChecker = NewCollisionChecker(mainClaudeDir)
    }
}
```

This provides accurate collision detection from the main worktree's rite provenance.

**Dependencies**: If using the fallback approach, depends on Fix 2's worktree detection utilities. The simple guard (log + skip) has no dependencies.

**Risk of regression**: NONE for the simple guard. LOW for the fallback approach (same worktree detection concerns as Fix 2).

---

## Phase 3: Improvement Opportunities

### 1. Worktree Environment Abstraction

**Current state**: No concept of "worktree environment" in the codebase. The `internal/worktree/` package manages worktree *lifecycle* (create, list, remove, export/import) but has no awareness of worktree *environment* (what directories exist, what state is available).

**Proposed abstraction**: `WorktreeEnv` struct in a new file `internal/worktree/env.go`:

```go
type WorktreeEnv struct {
    IsWorktree     bool
    WorktreePath   string   // Current worktree path
    MainPath       string   // Main worktree path
    HasClaude      bool     // .claude/ exists
    HasSOS         bool     // .sos/ exists
    HasKnossos     bool     // .knossos/ exists
    HasActiveRite  bool     // ACTIVE_RITE exists
    ActiveRite     string   // ACTIVE_RITE value
    MainActiveRite string   // Main worktree's ACTIVE_RITE
}

func DetectWorktreeEnv(projectDir string) (*WorktreeEnv, error)
```

This would be consumed by:
- `internal/cmd/sync/sync.go` -- to decide whether to inherit rite
- `internal/materialize/materialize.go` -- to gate minimal vs. full
- `internal/materialize/userscope/sync.go` -- to select collision checker source
- `internal/cmd/hook/context.go` -- to report worktree state

**Benefit**: Single detection point instead of scattered `git` subprocess calls. Cache-friendly (detect once, use everywhere).

**Risk**: Over-abstraction if the worktree use case remains niche. Recommend implementing Fix 2 first with inline utilities, then refactoring to `WorktreeEnv` if more consumers emerge.

### 2. Gitignored Directory Scaffolding Contract

**Current state**: `ari init` (in `internal/cmd/initialize/init.go`) scaffolds `.knossos/`, `.sos/`, `.ledge/` directories with `scaffoldProjectDirs()` (line 302-335). But this only runs on explicit `ari init` invocation. Worktrees created by any mechanism bypass this.

**Problem**: 5 of 6 gitignored directories (`.claude/`, `.sos/`, `.knossos/`, `.know/`, `.ledge/`) need to exist for the framework to function. Only `.wip/` is truly optional.

**Proposed improvement**: Add `ensureProjectDirs()` to the sync pipeline as a pre-flight step:

```go
func (m *Materializer) ensureProjectDirs() {
    dirs := []string{
        m.resolver.ClaudeDir(),     // .claude/
        m.resolver.SOSDir(),        // .sos/
        m.resolver.SessionsDir(),   // .sos/sessions/
        m.resolver.KnossosDir(),    // .knossos/
    }
    for _, d := range dirs {
        os.MkdirAll(d, 0755)
    }
}
```

This is idempotent, zero-cost when dirs exist, and ensures sync always has the minimum directory structure.

**Counter-argument**: `.know/` and `.ledge/` should NOT be auto-created by sync. They are optional and have their own lifecycle (`/know` for `.know/`, `ari init` for `.ledge/`).

### 3. Test Infrastructure for Worktree Scenarios

**Current gaps:**
- Zero tests in `internal/materialize/` for worktree scenarios
- Zero tests in `internal/cmd/sync/` for worktree scenarios
- Zero tests in `internal/materialize/userscope/` for empty provenance + worktree
- Zero tests in `internal/cmd/hook/` for worktree context
- The `internal/worktree/` tests (lifecycle_test.go, operations_test.go) test *ari's worktree management* but not *ari running inside a worktree it didn't create*

**Proposed test infrastructure:**

```go
// test/worktree/testutil.go
func SetupWorktreeTestFixture(t *testing.T) (mainDir, worktreeDir string, cleanup func()) {
    // 1. Create temp dir with git init
    // 2. Create initial commit
    // 3. Set up .knossos/ACTIVE_RITE in main
    // 4. Run ari sync in main (creates provenance)
    // 5. git worktree add to create linked worktree
    // 6. Return both paths
}
```

**Needed tests:**

| Package | Test | Priority |
|---------|------|----------|
| `internal/materialize` | `TestSyncRiteScope_InWorktree_InheritsRite` | P0 |
| `internal/materialize` | `TestSyncRiteScope_InWorktree_NoMainRite_FallsToMinimal` | P0 |
| `internal/materialize/userscope` | `TestCollisionChecker_InWorktree_FallsBackToMainProvenance` | P1 |
| `internal/materialize/userscope` | `TestSyncUserScope_InWorktree_EmptyProvenance_SafeDefault` | P1 |
| `internal/cmd/hook` | `TestContextHook_InWorktree_ReportsRites` | P1 |
| `internal/cmd/sync` | `TestSyncCmd_InWorktree_EndToEnd` | P2 |
| `internal/cmd/hook` | `TestWorktreeSeedHook_Integration` | P2 |

### 4. Documentation Gaps

| Gap | Recommended Action |
|-----|--------------------|
| No ADR for worktree lifecycle contract | Write ADR-0029: "Worktree Environment Contract" |
| `ari worktree create` vs CC `EnterWorktree` not documented | Add to CLI help text |
| `.gitignore` implications for worktrees not documented | Add comment block in `.gitignore` |
| Collision checker degradation modes not documented | Add to `collision.go` package doc |

### 5. `.sos/` Session State in Worktrees

**Overlooked by initial spike**: Sessions are stored in `.sos/sessions/` which is gitignored. A worktree has no sessions, no CC map, no locks, no archive. This means:

- `ari hook context` (SessionStart) fails to find sessions -> reports no session
- `/go` (start command) creates sessions in worktree's `.sos/` -- this is correct behavior (session per worktree)
- `/park` and `/continue` work on the worktree's local session state
- But there is no mechanism to share session context between main repo and worktree

**Recommendation**: This is actually the correct design. Each worktree should have its own session lifecycle. The only issue is that the first `ari sync` in a worktree needs to create `.sos/sessions/` before hooks can function. This is covered by improvement #2 (`ensureProjectDirs()`).

---

## Risk Matrix

| Fix | Regression Risk | Performance Impact | Dependency | Complexity |
|-----|----------------|-------------------|------------|------------|
| Fix 1 (WorktreeCreate hook) | LOW | None (new code path) | None | LOW |
| Fix 2 (Worktree-aware sync) | LOW-MEDIUM | +5ms on minimal path | None | MEDIUM |
| Fix 3 (SourceResolver for rites) | LOW | +5-10ms on context hook | None | LOW |
| Fix 4 (Guard collision checker) | NONE | None | Fix 2 (for fallback) | LOW |
| Improvement 1 (WorktreeEnv) | LOW | Negative (caching) | Fixes 2+4 done first | MEDIUM |
| Improvement 2 (ensureProjectDirs) | NONE | None | None | TRIVIAL |
| Improvement 3 (Test infrastructure) | NONE | N/A | None | MEDIUM |

---

## Recommended Implementation Order

### Sprint 1: Safety First (Estimated: ~135 LOC, ~2 hours)

1. **Fix 4 (Simple guard)** -- ~43 LOC
   - Guard collision checker with `IsEffective()` method
   - When provenance is missing, default to "all collisions = true" (skip user-scope writes)
   - Prevents user-scope corruption immediately
   - No dependencies

2. **Improvement 2 (ensureProjectDirs)** -- ~15 LOC
   - Add `ensureProjectDirs()` pre-flight to `Sync()`
   - Creates `.claude/`, `.sos/sessions/`, `.knossos/` if missing
   - Prevents hook failures from missing directories

3. **Fix 3 (SourceResolver for rites)** -- ~40 LOC
   - Replace `listAvailableRites(resolver.RitesDir())` with `SourceResolver.ListAvailableRites()`
   - Fixes rite discovery in context hook for all scenarios (not just worktrees)

### Sprint 2: Worktree Awareness (Estimated: ~170 LOC, ~3 hours)

4. **Fix 2 (Worktree-aware sync)** -- ~75 LOC
   - Add `isGitWorktree()` and `getMainWorktreeDir()` to materialize package
   - Modify `syncRiteScope()` to inherit rite from main worktree
   - This is the core fix that makes `ari sync` work in worktrees

5. **Fix 4 (Fallback to main provenance)** -- upgrade from simple guard (~20 LOC delta)
   - Now that worktree detection exists, collision checker falls back to main worktree provenance
   - Replaces the "skip all" guard with accurate collision detection

6. **Fix 1 (WorktreeCreate hook)** -- ~92 LOC
   - New `ari hook worktree-seed` command
   - Register in `settings.local.json` default template
   - Covers CC `EnterWorktree` path specifically

### Sprint 3: Quality & Documentation (Estimated: ~200 LOC tests, ~1 hour docs)

7. **Test infrastructure** -- ~200 LOC
   - `SetupWorktreeTestFixture()` helper
   - P0 and P1 tests from the table above

8. **Documentation**
   - ADR-0029: Worktree Environment Contract
   - `.gitignore` comment block explaining worktree implications
   - CLI help text for `ari worktree` vs CC `EnterWorktree`

---

## Key Files Reference

| File | Role | Lines | Fixes |
|------|------|-------|-------|
| `.gitignore` | Root cause: `.claude/` excluded | 27 | Context |
| `internal/cmd/sync/sync.go` | `ari sync` entry point | 241 | Fix 2 |
| `internal/materialize/materialize.go` | Sync pipeline, `syncRiteScope()`, `MaterializeMinimal()` | ~700 | Fix 2 |
| `internal/materialize/source/resolver.go` | 4-tier resolution, `ListAvailableRites()` | 345 | Fix 3 |
| `internal/materialize/userscope/collision.go` | Collision checker | 60 | Fix 4 |
| `internal/materialize/userscope/sync.go` | User-scope sync entry | ~300 | Fix 4 |
| `internal/materialize/user_scope.go` | Bridge to userscope package | 23 | Fix 4 |
| `internal/worktree/operations.go` | `setupWorktreeEcosystem()` | 681 | Reference |
| `internal/worktree/lifecycle.go` | `Create()`, worktree lifecycle | 380 | Reference |
| `internal/worktree/git.go` | `IsWorktree()`, `GetMainWorktree()` | 319 | Reference |
| `internal/worktree/metadata.go` | `SavePerWorktreeMeta()` | 351 | Fix 1 |
| `internal/paths/paths.go` | `Resolver`, `RitesDir()`, `ReadActiveRite()` | 358 | Fix 3 |
| `internal/cmd/hook/context.go` | `runContextCore()`, `listAvailableRites()` | 387 | Fix 3 |
| `internal/cmd/hook/hook.go` | Hook command registration, `resolveSession()` | ~164 | Fix 1 |
| `internal/cmd/initialize/init.go` | `scaffoldProjectDirs()` | ~350 | Improvement 2 |
| `.claude/settings.local.json` | Hook definitions (no WorktreeCreate yet) | 154 | Fix 1 |

## Appendix: RC-1 Extended Analysis -- All Gitignored Paths

The `.gitignore` file (27 lines) excludes 6 directories. Here is the full impact of each in worktree context:

```
ari               # binary -- irrelevant
.claude/          # RC-1 primary: all framework state
.wip/             # ephemeral scratch -- by design, no impact
.sos/             # sessions, locks, cc-map -- HIGH impact on hooks
.know/            # codebase knowledge -- LOW impact (cosmetic)
.ledge/           # work products -- LOW impact (optional)
```

The `.claude/agent-memory/` and `.claude/agent-memory-local/` exclusions (lines 23-24) are subdirectories of `.claude/` and are already covered by the `.claude/` exclusion on line 7.

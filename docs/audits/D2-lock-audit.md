# D2: Lock Audit Report

**Date**: 2026-02-05
**Auditor**: ecosystem-analyst
**Scope**: `internal/lock/`, `internal/cmd/session/`, `internal/cmd/common/context.go`, `internal/paths/paths.go`
**Status**: Analysis complete. No code modified.

---

## 1. Current Behavior

### 1.1 Lock Mechanism

The lock package (`internal/lock/lock.go`) implements advisory file locking using `syscall.Flock` (BSD flock). Key characteristics:

- **Lock types**: Shared (`LOCK_SH`) and Exclusive (`LOCK_EX`), defined at `lock.go:20-25`.
- **Lock granularity**: Per-session. Each session gets its own lock file at `{sessionID}.lock` -- `lock.go:49-51`.
- **Lock file location**: `.sos/sessions/.locks/{sessionID}.lock` -- `paths.go:68-69`.
- **Default timeout**: 10 seconds (`DefaultTimeout`, `lock.go:28`), configurable per call.
- **Retry interval**: 100ms polling loop with non-blocking `Flock` attempts -- `lock.go:74-108`.
- **PID tracking**: On exclusive lock acquisition, the holder's PID is written to the lock file -- `lock.go:80-84`.
- **Lock release**: `Flock(LOCK_UN)` + `file.Close()`. The lock file is NOT deleted on release -- `lock.go:117-128`.

### 1.2 Stale Lock Detection

The `isStale` method (`lock.go:131-149`) checks whether the lock holder process is alive:

1. Reads PID from lock file content via `getHolderPID` -- `lock.go:152-165`.
2. Uses `os.FindProcess` + `Signal(0)` to probe process existence -- `lock.go:138-148`.
3. If the process is dead, the lock file is removed (`os.Remove`) and the file descriptor is reopened -- `lock.go:96-103`.
4. The retry loop then attempts to acquire the lock on the fresh file.

### 1.3 Lock Usage Across Session Commands

| Command | Lock Type | Lock ID | File:Line |
|---------|-----------|---------|-----------|
| `create` | Exclusive | `__create__` (sentinel) | `create.go:85` |
| `park` | Exclusive | `{sessionID}` | `park.go:57` |
| `resume` | Exclusive | `{sessionID}` | `resume.go:46` |
| `wrap` | Exclusive | `{sessionID}` | `wrap.go:60` |
| `transition` | Exclusive | `{sessionID}` | `transition.go:65` |
| `migrate` | Exclusive | `{sessionID}` | `migrate.go:176` |
| `status` | Shared | `{sessionID}` | `status.go:67` |
| `audit` | Shared | `{sessionID}` | `audit.go:57` |
| `lock` (manual) | Exclusive | `{sessionID}` | `lock.go:62` |
| `unlock` (manual) | N/A (ForceRelease) | `{sessionID}` | `unlock.go:94` |
| `list` | None | N/A | `list.go` |

All mutating commands acquire exclusive locks on the session ID and use `defer lock.Release()`. Read-only commands (`status`, `audit`) use shared locks and gracefully degrade if acquisition fails (`status.go:69-73`, `audit.go:58-62`). `list` uses no lock at all.

### 1.4 Single-Session-Per-Repo Enforcement

The single-active-session constraint is enforced by application logic, NOT by the lock system:

1. `create` acquires an exclusive lock on the synthetic ID `__create__` -- `create.go:85`.
2. It reads `.sos/sessions/.current-session` to get the current session ID -- `create.go:93` via `GetCurrentSessionID()`.
3. If a current session exists and its `SESSION_CONTEXT.md` is loadable, it rejects creation with `ErrSessionExists` -- `create.go:107`.

The `__create__` lock serializes concurrent create attempts, but does NOT protect `.current-session` writes from other commands. Specifically:

- `resume` writes `.current-session` at `resume.go:85` under the session-specific lock (not `__create__`).
- `wrap` clears `.current-session` at `wrap.go:170` under the session-specific lock (not `__create__`).
- `SetCurrentSessionID()` and `ClearCurrentSessionID()` perform bare `os.WriteFile`/`os.Remove` at `context.go:81-97`.

### 1.5 Seeded Session Path

The `--seed` flow (`create.go:170-304`) bypasses the lock entirely:

- Creates a session in an ephemeral worktree with its own `.sos/sessions/.locks/` -- `create.go:222-228`.
- Sets status directly to PARKED without FSM transition -- `create.go:243-244`.
- Copies the session directory to the main repo via filesystem operations -- `create.go:272`.
- Does NOT acquire the `__create__` lock.
- Does NOT update `.current-session`.
- This is correct by design: seeded sessions are PARKED and invisible to `.current-session`.

### 1.6 ForceRelease and Manual Unlock

- **`ForceRelease`** (`lock.go:200-206`): Removes the lock file via `os.Remove`. Does NOT call `flock(LOCK_UN)` on the held file descriptor. On Unix, removing a file while a process holds `flock` on it means: the holder retains its lock on the now-unlinked inode, but any new process opening the same path gets a new inode. This means the old holder's lock becomes irrelevant to new callers.

- **`unlock --force`** (`unlock.go:33-112`): Reads the holder PID from the lock file, checks if caller is owner (or `--force`), and calls `ForceRelease` to remove the lock file. Does NOT send any signal to the holding process.

### 1.7 Error UX on Lock Contention

When lock acquisition times out, the error is produced at `lock.go:113`:

```go
errors.ErrLockTimeout(lockPath, holderPID)
```

This resolves to (`errors.go:218-227`):

```
"Could not acquire lock within timeout"
```

With details: `lock_path` and `holder_pid` (if available). Exit code: `3` (`ExitLockTimeout`, `errors.go:15`). The error message does NOT include:
- Whether the holder process is alive or dead
- What command to run to resolve the situation
- How long to wait or retry

---

## 2. Risk Assessment

### P0: Critical

**(None identified.)**

The flock mechanism is fundamentally sound for the single-user, single-machine use case that Knossos targets. No data loss or corruption scenarios exist in normal single-user operation.

### P1: High

**2.1 TOCTOU Race in `.current-session` During Concurrent Operations**

The `.current-session` file is the canonical pointer to the active session, but reads and writes to it are not covered by a common lock:

1. `create` holds `__create__` lock when reading/writing `.current-session` -- `create.go:85-148`.
2. `resume` holds the session-specific lock when writing `.current-session` -- `resume.go:46,85`.
3. `wrap` holds the session-specific lock when clearing `.current-session` -- `wrap.go:60,170`.

These are three DIFFERENT locks. Concurrent execution of `resume` (on session A) and `create` (for new session B) could produce:

1. `create` reads `.current-session` = "" (session A was parked, pointer is empty).
2. `resume` writes `.current-session` = "session-A".
3. `create` writes `.current-session` = "session-B", overwriting session-A.
4. Result: Two ACTIVE sessions, `.current-session` points to only session-B. Session-A is active but invisible.

**Likelihood**: Low in typical single-user interactive use. Higher in CI pipelines or scripted automation that calls multiple `ari` commands in parallel.

**Impact**: Corrupted single-session invariant. Subsequent commands operate on the wrong session.

**2.2 `ForceRelease` Can Create Split-Brain When Holder Is Alive**

`ForceRelease` (`lock.go:200-206`) removes the lock file unconditionally. If the holder process is still alive:

1. Process A holds `flock` on inode X (the original lock file).
2. `ForceRelease` removes the file at the lock path. Inode X is unlinked but still referenced by process A's fd.
3. Process B creates a new file at the same path (new inode Y), acquires `flock` on it.
4. Both A and B believe they hold the exclusive lock.
5. Both proceed to mutate `SESSION_CONTEXT.md`.

The `unlock` command at `unlock.go:74` requires `--force` if the caller is not the owner, but does not verify the holder is actually dead before calling `ForceRelease`.

**Likelihood**: Low. Requires `unlock --force` while the holder is alive. A user would typically only do this after verifying the holder is dead.

**Impact**: Potential concurrent mutation of session state files. Recoverable by re-parking and resuming.

### P2: Medium

**2.3 PID Reuse Defeats Stale Detection**

Stale detection (`lock.go:131-149`) relies on PID liveness. If a process crashes and its PID is reused by an unrelated process before stale detection runs, the lock appears non-stale. On macOS (current target), PIDs wrap after ~99999.

**Likelihood**: Very low under normal conditions. Higher on systems with many short-lived processes (e.g., CI runners).

**Impact**: Lock permanently stuck until manual `ari session unlock --force`. User sees "Could not acquire lock within timeout" with no guidance.

**2.4 Lock File Accumulation**

Lock files are never deleted by `Release()` (`lock.go:117-128`). They persist in `.sos/sessions/.locks/` indefinitely. When `wrap` archives a session (`wrap.go:211`), the lock file is NOT cleaned up.

**Likelihood**: Certain. Every session creates a lock file that persists.

**Impact**: Minimal. Files are tiny (PID string). Cosmetic clutter only.

**2.5 `flock` Is Local-Only (No NFS/Network FS Support)**

`syscall.Flock` provides advisory locking for local filesystems only. On NFS, SMB, or other network filesystems, `flock` may be silently ignored or behave inconsistently.

**Likelihood**: Low for typical Knossos usage (local developer machines). Relevant for shared CI environments.

**Impact**: Complete loss of lock protection on network filesystems. Concurrent mutations would proceed without serialization.

**2.6 `ari session lock` Blocks Forever**

The manual `lock` command (`lock.go:97`) uses `select {}` to block indefinitely:

```go
select {} // Block forever - lock released when process dies
```

No timeout, no heartbeat, no auto-release. A forgotten background terminal holds the lock permanently.

**Likelihood**: Low (debugging tool only). Higher if users script with it.

**Impact**: All session operations on that session time out. Recoverable via `unlock --force` or killing the process.

### P3: Low

**2.7 Lock Timeout Error Not Actionable**

The error message "Could not acquire lock within timeout" (`errors.go:225`) provides `lock_path` and `holder_pid` but no recovery instructions.

**Likelihood**: Moderate (any lock contention triggers this).

**Impact**: User confusion and manual debugging. No data loss.

**2.8 Shared Lock Does Not Write PID**

When a shared lock is acquired (`lock.go:80-84`), no PID is written. Stale detection cannot identify shared lock holders. Since shared locks are only used for non-critical reads that gracefully degrade, this is acceptable.

**Impact**: Negligible.

**2.9 `list` Reads Without Lock**

The `list` command (`list.go:41-126`) reads `SESSION_CONTEXT.md` files without any lock. Could read partially-written files during concurrent transitions.

**Impact**: Garbled display in rare race conditions. Self-corrects on next invocation.

**2.10 Spin-Wait Polling Is Inefficient**

The 100ms polling interval (`lock.go:107`) means worst-case 100ms latency after a lock becomes available.

**Impact**: Negligible for a CLI tool.

---

## 3. Recommendations (Prioritized)

### R1: Protect `.current-session` With a Common Lock [P1, addresses Risk 2.1]

**What**: All operations that read or write `.current-session` must hold a common lock (either reuse `__create__` or introduce a `__session-pointer__` sentinel).

**Where**: `resume.go:85`, `wrap.go:170`, `create.go:93-148`.

**Why**: Eliminates the TOCTOU race that can create multiple active sessions.

**Complexity**: PATCH. Add lock acquisition to 2 additional commands.

### R2: Validate Holder Liveness Before `ForceRelease` [P1, addresses Risk 2.2]

**What**: `unlock --force` should check if the holder PID is alive. If alive, warn the user: "Process {pid} is still running. Kill it first or use --force --kill." If dead, proceed with removal.

**Where**: `unlock.go:86-91`.

**Why**: Prevents split-brain from premature force-unlock.

**Complexity**: PATCH. Add PID liveness check before `ForceRelease`.

### R3: Make Lock Timeout Error Actionable [P3, addresses Risk 2.7]

**What**: Include recovery guidance in `ErrLockTimeout`:
```
"Could not acquire lock within timeout. Holder PID: {pid}.
If the holder is dead, run: ari session unlock --force"
```

**Where**: `errors.go:224-226`.

**Complexity**: PATCH. String change only.

### R4: Clean Up Lock Files on Archive [P2, addresses Risk 2.4]

**What**: Add `lockMgr.ForceRelease(sessionID)` to `runWrap` after archiving the session directory.

**Where**: `wrap.go:218` (after successful `os.Rename`).

**Complexity**: PATCH. One-line addition.

### R5: Add Lock Age Metadata [P2, addresses Risk 2.3]

**What**: Write `{pid} {unix-timestamp}` to lock files instead of just `{pid}`. Stale detection can then use age-based heuristics: locks older than N minutes trigger a warning regardless of PID liveness.

**Where**: `lock.go:82-84` (write), `lock.go:131-149` (read).

**Complexity**: PATCH. Small format change to lock file content.

### R6: Add Duration Limit to Manual Lock [P2, addresses Risk 2.6]

**What**: Add `--duration` flag to `ari session lock` with a default maximum (e.g., 5 minutes). Auto-release with warning when duration expires.

**Where**: `lock.go:18-97`.

**Complexity**: PATCH. Add timer goroutine.

### R7: Document Worktree Lock Isolation [Informational]

**What**: Document in the Session Model ADR that locks are per-worktree when `.claude/` is per-worktree (default git behavior). Warn that symlinked `.claude/` directories share lock state.

**Complexity**: No code change. Documentation only.

---

## 4. D5 Input: Key Architectural Questions for Session Model ADR

### Q1: What is the correct lock scope for `.current-session` mutations?

**Current**: `.current-session` writes use different locks depending on the calling command (`__create__` for create, session-specific for resume/wrap). This allows races.

**Options**:
- (a) Require `__create__` lock for all `.current-session` mutations (simple, slightly higher contention).
- (b) Introduce a dedicated `__session-pointer__` lock (cleaner separation of concerns).
- (c) Replace `.current-session` file with a derived scan of session directories (eliminates pointer file race entirely, but slower).

### Q2: Should the single-session-per-repo constraint be relaxed?

**Current**: One ACTIVE session per repo, enforced by `.current-session` + `create` logic. Seeded sessions bypass this (they are PARKED).

**Considerations**:
- Multiple Claude Code terminals in the same repo would each want their own session.
- Worktrees already provide isolation (separate `.claude/` directories).
- CI environments may need multiple concurrent sessions in the same checkout.
- Relaxing to "one ACTIVE session per terminal/process" requires a different pointer mechanism (e.g., PID-based or environment variable).

### Q3: What is the correct lock scope for worktree scenarios?

**Current**: Lock files live at `.sos/sessions/.locks/` relative to the project root. In git worktrees, `.claude/` is typically per-worktree (copies, not symlinks). This means locks are naturally isolated per worktree.

**Decision needed**: Is per-worktree isolation correct? If so, document it. If worktrees should share session state, a different lock location is needed (e.g., inside `.git/` which IS shared across worktrees).

### Q4: Should `flock` be replaced or supplemented?

**Current**: `syscall.Flock` is Unix-only, local-only, per-file-descriptor.

**Alternatives**:
- `fcntl` (POSIX record locks): Per-process semantics, but released on ANY close of any fd to the file.
- `lockfile` libraries (e.g., `github.com/nightlyone/lockfile`): Cross-platform, PID-based, no kernel lock.
- Keep `flock` but add NFS warning on startup if network FS detected.

For the current macOS-centric, single-user use case, `flock` is adequate.

### Q5: Should the default timeout be configurable?

**Current**: 10 seconds hardcoded as `DefaultTimeout`, with per-command timeout flags only on `lock` command.

**Proposal**: Add `ARIADNE_LOCK_TIMEOUT` environment variable, checked in `Acquire` when `timeout == 0`. This allows CI to increase the timeout without code changes.

### Q6: Should hook operations acquire locks?

**Current**: `ari hook context` and `ari hook clew` read `.current-session` and session state without any lock. These run on EVERY Claude Code tool invocation and must be fast.

**Trade-offs**: Adding shared locks would add ~0-100ms latency to every tool call. Lock-free reads risk stale data during transitions, but hooks are advisory (they inject context, not enforce state).

**Recommendation**: Lock-free reads for hooks, with a "last known good" caching strategy.

---

## 5. Summary Matrix

| Finding | Severity | Likelihood | Recommendation | Complexity |
|---------|----------|------------|----------------|------------|
| `.current-session` TOCTOU race | P1 | Low | R1: Common lock | PATCH |
| `ForceRelease` split-brain | P1 | Low | R2: Liveness check | PATCH |
| PID reuse defeats stale detection | P2 | Very Low | R5: Add timestamp | PATCH |
| Lock file accumulation | P2 | Certain | R4: Clean on archive | PATCH |
| `flock` local-only (no NFS) | P2 | Low | R7: Document | DOC |
| Manual `lock` blocks forever | P2 | Low | R6: Duration limit | PATCH |
| Lock timeout error not actionable | P3 | Moderate | R3: Better message | PATCH |
| Shared lock no PID | P3 | N/A | Acceptable | N/A |
| `list` reads without lock | P3 | Very Low | Acceptable | N/A |
| Spin-wait polling | P3 | N/A | Acceptable | N/A |

---

## 6. Files Examined

| File | Purpose |
|------|---------|
| `internal/lock/lock.go` | Lock implementation (flock + stale detection) |
| `internal/lock/lock_test.go` | Lock tests (concurrency, race, force release) |
| `internal/cmd/session/session.go` | Session command group registration |
| `internal/cmd/session/create.go` | Session creation with `__create__` sentinel lock |
| `internal/cmd/session/park.go` | Park with session-specific exclusive lock |
| `internal/cmd/session/resume.go` | Resume with session-specific exclusive lock |
| `internal/cmd/session/wrap.go` | Wrap/archive with session-specific exclusive lock |
| `internal/cmd/session/transition.go` | Phase transition with session-specific exclusive lock |
| `internal/cmd/session/migrate.go` | Schema migration with session-specific exclusive lock |
| `internal/cmd/session/status.go` | Status with shared lock (graceful degradation) |
| `internal/cmd/session/audit.go` | Audit with shared lock (graceful degradation) |
| `internal/cmd/session/lock.go` | Manual lock command (blocks forever) |
| `internal/cmd/session/unlock.go` | Manual unlock with `--force` |
| `internal/cmd/session/list.go` | List sessions (no lock) |
| `internal/cmd/common/context.go` | Lock manager factory, `.current-session` read/write |
| `internal/paths/paths.go` | Path resolution (locks dir, session files) |
| `internal/session/fsm.go` | State machine (transition context) |
| `internal/errors/errors.go` | Error types and exit codes for lock timeout |

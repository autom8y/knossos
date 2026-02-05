# ADR-0022: Session Model Architecture

| Field | Value |
|-------|-------|
| **Status** | Proposed |
| **Date** | 2026-02-05 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A (refines ADR-0001, ADR-0005/0013, ADR-0010) |
| **Superseded by** | N/A |

## Context

The Knossos session model has evolved through four ADRs: ADR-0001 (FSM redesign from bash to Go), ADR-0005 (Moirai centralized state authority), ADR-0010 (worktree session seeding), and ADR-0013 (Moirai consolidation). A systematic audit of the session infrastructure (D1-D3) was conducted to assess production readiness for framework productization. This ADR synthesizes those findings and resolves five architectural questions that D2 identified as requiring first-principles reasoning.

### Evidence Base

**D1: Session Lifecycle Tests** (56 tests, 2 bugs found)

D1 tested the FSM, context serialization, and full lifecycle flows at the `internal/session/` package level. Two low-severity bugs were found:

1. Extra newline in body round-trip (cosmetic, does not affect production behavior)
2. CreatedAt loses sub-second precision on round-trip (RFC3339 serialization truncates to seconds)

Both are test-observation-only issues. The session model's core data integrity is sound: all 16 FSM transition pairs pass, terminal state enforcement works, concurrent reads (20 goroutines) succeed, and corrupt frontmatter is handled gracefully.

**D2: Lock Audit** (10 risks found, 0 critical)

D2 performed a complete audit of `internal/lock/`, `internal/cmd/session/`, and `internal/cmd/common/context.go`. Key findings:

| Finding | Severity | Summary |
|---------|----------|---------|
| `.current-session` TOCTOU race | P1 | create/resume/wrap use different lock scopes when mutating the pointer file |
| `ForceRelease` split-brain | P1 | Removing a lock file while the holder is alive creates two concurrent "holders" |
| PID reuse defeats stale detection | P2 | Unlikely but possible on CI systems |
| Lock file accumulation | P2 | Lock files never cleaned up on archive |
| `flock` is local-only | P2 | Advisory locks do not work on NFS/SMB |
| Manual `lock` blocks forever | P2 | No timeout on `ari session lock` |
| Lock timeout error not actionable | P3 | No recovery guidance in error message |

No P0 (critical/data-loss) risks were found. The `flock`-based per-session locking mechanism is fundamentally correct for the single-user, single-machine use case.

**D3: Moirai Integration Tests** (28 tests, 0 bugs found)

D3 tested the full CLI command layer (`runCreate`, `runPark`, `runResume`, `runWrap`) that Moirai invokes. All 28 tests pass. Key validations:

- Golden path (create -> park -> resume -> wrap) succeeds with correct state at each step
- Context file integrity: fields survive round-trips across all transitions
- Audit trail: every transition emits correctly-typed JSONL events
- Error paths: all invalid transitions produce clear, specific error messages
- Concurrent safety: per-session `flock` serializes mutations correctly; concurrent park/resume/create races produce exactly one winner
- `.current-session` is correctly set on create/resume and cleared on wrap

D3 confirmed D2's architectural observation: per-session locking works well, but `.current-session` is a cross-session concern that falls outside the per-session lock scope.

### Forces

- **Productization trajectory**: The framework is moving toward multiple users, satellite repositories, and potentially CI-driven sessions. Decisions must account for this trajectory without over-engineering the current single-user case.
- **Proven foundation**: D1-D3 demonstrate that the current FSM, locking, and serialization are correct for their scope. Changes should be surgical, not rewrite-level.
- **Worktree reality**: Git worktrees create isolated `.claude/` directories by default. This is a feature, not a bug, but it must be explicitly designed for rather than accidentally relied upon.
- **Moirai's role**: ADR-0013 established Moirai as CLI wrapper, not CLI replacement. The relationship between Moirai (agent), CLI (authoritative), and hooks (advisory) must be clarified for productization.
- **Backward compatibility**: Existing sessions, satellites, and user workflows must continue to work unless a breaking change is explicitly justified.

## Decisions

### Decision 1: Maintain Single-Session-Per-Repo Constraint

**Keep the current constraint that only one session can be ACTIVE in a given `.claude/` directory at a time.**

The single-session constraint, enforced by the `.current-session` pointer file, has been validated by D3 as functionally correct. The question is whether to relax it for multi-terminal or CI use cases.

**Options considered:**

| Option | Description | Verdict |
|--------|-------------|---------|
| (a) Keep single-session-per-repo | Current behavior. One ACTIVE session per `.claude/` directory | **Selected** |
| (b) Allow multiple ACTIVE sessions per repo | Remove `.current-session`, let sessions be identified by PID or env var | Rejected |
| (c) Make configurable | Feature flag to toggle single vs. multi-session | Rejected |

**Why (b) is rejected:** Multiple ACTIVE sessions in the same `.claude/` directory would require every hook, agent, and CLI command to accept an explicit session ID for every operation. The `.current-session` pointer file exists precisely to avoid this -- it provides ambient context ("what session am I in?") without requiring the user or agent to track and pass session IDs. Removing ambient context would degrade the developer experience for the dominant single-user case to support an edge case that worktrees already solve.

**Why (c) is rejected:** Configuration creates two code paths that must both be tested and maintained. The worktree-based isolation model (Decision 4) handles multi-terminal use cases without configuration.

**The multi-terminal case is real but already solved.** When a developer opens a second terminal, they have three options today:
1. Park the current session and start a new one (sequential work)
2. Use `ari session create --seed` to prepare parallel sessions (ADR-0010)
3. Use a git worktree, which creates a separate `.claude/` with its own session space

These options cover all observed use cases without relaxing the constraint. The constraint serves as a guardrail against accidental state corruption, and the D2 audit confirms that relaxing it would exponentially increase the lock scope complexity.

**Consequence:** CI environments that need multiple concurrent sessions in the same checkout must use worktrees. This is documented in Decision 4.

### Decision 2: Introduce a Session Pointer Lock to Fix the TOCTOU Race

**Add a dedicated lock for `.current-session` mutations, separate from both the `__create__` sentinel and per-session locks.**

D2 identified a P1 TOCTOU race: `create` holds the `__create__` lock, `resume` holds the session-specific lock, and `wrap` holds the session-specific lock -- three different locks protecting the same resource (`.current-session`). Concurrent `resume` + `create` can corrupt the pointer.

**Options considered:**

| Option | Description | Verdict |
|--------|-------------|---------|
| (a) Reuse `__create__` lock for all `.current-session` mutations | Simple; all pointer operations hold the same lock | Rejected |
| (b) Introduce `__session-pointer__` sentinel lock | Dedicated lock for pointer; per-session locks remain for session data | **Selected** |
| (c) Eliminate `.current-session`; derive active session from directory scan | No pointer file, no race. Scan all session dirs, find the one with `status: ACTIVE` | Rejected |

**Why (a) is rejected:** The `__create__` lock protects session creation logic (ID generation, directory creation, uniqueness checks), not just the pointer. Overloading it for resume/wrap would force those commands to serialize against create, adding unnecessary contention. A park operation in terminal A would block a create attempt in terminal B even though they operate on different sessions.

**Why (c) is rejected:** Directory scanning is O(n) in the number of sessions. For a developer with dozens of sessions accumulated over weeks, this adds measurable latency to every hook invocation that checks the current session. The `.current-session` file is a deliberate O(1) optimization.

**Why (b) is selected:** A dedicated `__session-pointer__` sentinel lock clearly expresses intent: "I am mutating the pointer to the current session." It does not contend with per-session data mutations (park/resume/wrap still hold session-specific locks for data integrity) or with session creation logic (which still holds `__create__` for ID uniqueness). The lock is lightweight (one additional `flock` acquisition) and only required by three operations: `create` (after creating the session, before writing the pointer), `resume` (before writing the pointer), and `wrap` (before clearing the pointer).

**Acquisition order to prevent deadlocks:**

1. `create`: acquire `__create__` -> create session -> acquire `__session-pointer__` -> write pointer -> release both
2. `resume`: acquire `{sessionID}` -> validate FSM -> acquire `__session-pointer__` -> write pointer -> release both
3. `wrap`: acquire `{sessionID}` -> validate FSM -> acquire `__session-pointer__` -> clear pointer -> release both

The rule is: always acquire the session-specific or sentinel lock first, then acquire `__session-pointer__` second. Since `__session-pointer__` is always acquired last, no circular dependency can form.

**Backward compatibility:** COMPATIBLE. Existing sessions and satellites are unaffected. The new lock file (`__session-pointer__.lock`) is created automatically on first use.

### Decision 3: Keep `flock`-Based Locking, Harden ForceRelease

**Retain `syscall.Flock` as the locking mechanism. Harden `ForceRelease` with liveness checks and improve error messages.**

D2 validated that `flock` is correct for the current use case (single user, local filesystem, macOS/Linux). Three hardening changes are warranted:

**3a. ForceRelease liveness check**

`ForceRelease` currently removes the lock file unconditionally. If the holder process is alive, this creates a split-brain where two processes believe they hold the exclusive lock (D2 Risk 2.2). The fix: before removing the lock file, check if the holder PID is alive. If alive, require an explicit `--kill` flag or refuse the operation.

**3b. Lock timeout error improvement**

The current error message ("Could not acquire lock within timeout") provides no recovery guidance (D2 Risk 2.7). The improved message should include: holder PID, whether the holder is alive or dead, and the command to run for recovery (`ari session unlock --force` if dead, `kill {pid}` if alive but unresponsive).

**3c. Lock file cleanup on archive**

Lock files are never deleted by `Release()` and accumulate indefinitely (D2 Risk 2.4). Add `lockMgr.ForceRelease(sessionID)` to `runWrap` after successful archive. This is cosmetic but prevents confusion when inspecting the locks directory.

**3d. Lock age metadata**

Write `{pid} {unix-timestamp}` to lock files instead of just `{pid}`. This enables age-based stale detection as a fallback when PID-based detection fails (D2 Risk 2.3). A lock older than 30 minutes with an unreachable PID is almost certainly stale. The format change is backward-compatible: existing `getHolderPID` parses the first whitespace-delimited token, so adding a second token does not break readers that expect only a PID.

**Alternatives not pursued:**

| Alternative | Reason not pursued |
|-------------|-------------------|
| Replace `flock` with `fcntl` record locks | `fcntl` has per-process (not per-fd) semantics; closing any fd to the file releases the lock. This is worse for the CLI use case where multiple goroutines may open the same lock file. |
| Replace `flock` with cross-platform library | Adds external dependency. `flock` is available on all target platforms (macOS, Linux). Windows support is not a current requirement. |
| Add NFS detection and warning | Low priority. If productization targets network filesystems, revisit. Document the limitation (Decision 5). |

**Backward compatibility:** COMPATIBLE. Lock file format change is additive. ForceRelease behavior change only affects `ari session unlock --force`, which is a debugging command.

### Decision 4: Worktrees Provide Session Isolation by Design

**Affirm that each git worktree has its own `.claude/` directory and therefore its own session space. This is the correct and intended behavior. Worktrees do not share session state.**

Git's default behavior places `.claude/` in each worktree's working directory, not in the shared `.git/` directory. This means:

- Each worktree has its own `.claude/sessions/`, `.claude/sessions/.locks/`, and `.current-session`
- Sessions created in one worktree are invisible to another worktree
- Locks in one worktree do not contend with locks in another worktree

This is correct because:

1. **Isolation prevents corruption.** Shared session state across worktrees would require cross-worktree locking, which reintroduces the NFS problem (`flock` is per-filesystem-inode, and worktrees may be on different filesystems).
2. **Worktrees represent independent work contexts.** A developer using worktree A for feature work and worktree B for a hotfix has two independent work contexts. Independent contexts should have independent sessions.
3. **`--seed` bridges the gap.** ADR-0010's `ari session create --seed` creates a PARKED session in a worktree and copies it to the main repo. This is the mechanism for preparing sessions that will later be resumed in different worktrees.

**What about shared state?** Some data legitimately needs to be shared across worktrees:

| Data | Shared? | Location | Rationale |
|------|---------|----------|-----------|
| Session state | No | `.claude/sessions/` (per-worktree) | Independent work contexts |
| Session locks | No | `.claude/sessions/.locks/` (per-worktree) | Locks must be co-located with the data they protect |
| Rite definitions | Yes | `rites/` (tracked in git) | Source content shared via git |
| CLAUDE.md | Yes | `.claude/CLAUDE.md` (tracked in git, regenerated per-worktree by materialization) | Configuration shared via git, projection per-worktree |
| Archive | No | `.claude/.archive/` (per-worktree) | Historical data stays with the worktree that created it |

**Consequence for CI:** CI pipelines that need multiple concurrent sessions should use separate worktrees (or separate checkouts, which are functionally equivalent). This is documented but not enforced by tooling.

**Backward compatibility:** COMPATIBLE. This documents existing behavior; no code change required.

### Decision 5: Moirai as Convention-Enforced State Authority, CLI as Enforcement Mechanism

**Keep Moirai's role as the recommended path for agent-driven state mutations, enforced by PreToolUse hook convention. Do not introduce token-based or capability-based hardening. The CLI (`ari`) remains the authoritative enforcement layer.**

D3 validated that the CLI commands (`ari session create/park/resume/wrap`) correctly enforce FSM transitions, emit audit events, and serialize concurrent access. Moirai is a wrapper around these commands (ADR-0013). The question is whether Moirai's "sole authority" status (ADR-0005) should be hardened beyond the current hook-based convention.

**Options considered:**

| Option | Description | Verdict |
|--------|-------------|---------|
| (a) Keep as-is: convention + PreToolUse hook | Current approach. Hook blocks direct Write/Edit to `*_CONTEXT.md` and suggests Moirai. CLI commands bypass the hook by using native Go file I/O. | **Selected** |
| (b) Harden with Moirai token | Require a Moirai-issued token for CLI mutations. CLI refuses operations without a valid token. | Rejected |
| (c) Relax: allow direct CLI in non-orchestrated mode | When no session team is active, permit `ari session park` without Moirai. In orchestrated mode, require Moirai. | **Selected (additive)** |

**Why (b) is rejected:** Token-based enforcement adds infrastructure complexity (token generation, validation, expiry, revocation) for a threat model that does not exist. The "threat" is a Claude Code agent accidentally editing `SESSION_CONTEXT.md` directly. The PreToolUse hook already prevents this. The CLI is a trusted component -- it validates FSM transitions, writes audit events, and holds locks. Adding a token requirement to the CLI would break interactive use (a developer running `ari session park` from their terminal should not need a Moirai token).

**Why (a) + (c) together:** The current system has a philosophical tension. ADR-0005 says "Moirai is the sole authority for mutations." ADR-0013 says "CLI remains authoritative." In practice, both are true at different layers:

```
                 Agent Layer                    CLI Layer
                 (convention)                   (enforcement)
                      |                              |
User/Agent -----> Moirai -----> ari session park -----> FSM validation
                    |               |                      |
              PreToolUse hook    flock + PID           SaveContext()
              blocks direct      concurrency            audit events
              Write/Edit         safety
```

**The refined model:**

- **In orchestrated mode** (active session with a team): Moirai is the recommended path. The PreToolUse hook blocks direct edits. Agents should invoke Moirai, which invokes the CLI.
- **In non-orchestrated mode** (native mode, or no active team): Users and scripts may invoke the CLI directly. The CLI's own validation (FSM, locking, audit) is sufficient.
- **In all modes**: The CLI is the enforcement layer. It validates transitions, holds locks, emits events. Moirai adds orchestration-layer value (natural language parsing, operation routing, Fate-domain skill loading) but is not required for correctness.

This resolves the tension: Moirai is the sole authority at the agent convention layer. The CLI is the sole authority at the enforcement layer. Neither needs to be "hardened" because they operate at different layers.

**Consequence for the PreToolUse hook:** The hook's decision logic should check whether an orchestrated session is active. If yes, block direct writes and suggest Moirai. If no, allow the write (or suggest the CLI command as a better alternative). This is a documentation and hook logic change, not an architectural change.

**Backward compatibility:** COMPATIBLE. No behavioral change for orchestrated sessions. Non-orchestrated sessions gain explicit permission to use the CLI directly, which many users already do.

## Implementation

### Phase 1: Session Pointer Lock (addresses D2 R1)

| File | Change |
|------|--------|
| `internal/cmd/session/create.go` | After creating session, acquire `__session-pointer__` lock before calling `SetCurrentSessionID`. Release after write. |
| `internal/cmd/session/resume.go` | Acquire `__session-pointer__` lock before calling `SetCurrentSessionID`. Release after write. |
| `internal/cmd/session/wrap.go` | Acquire `__session-pointer__` lock before calling `ClearCurrentSessionID`. Release after clear. |
| `internal/lock/lock.go` | No changes. The existing `Acquire` method works with any sentinel ID. |

### Phase 2: ForceRelease Hardening (addresses D2 R2, R3)

| File | Change |
|------|--------|
| `internal/cmd/session/unlock.go` | Before calling `ForceRelease`, check if holder PID is alive. If alive and `--kill` not passed, print warning and refuse. |
| `internal/lock/lock.go` | `Acquire` method: write `{pid} {timestamp}` instead of `{pid}` to lock files. `getHolderPID` method: parse first token only (backward compatible). Add `getHolderTimestamp` method. |
| `internal/errors/errors.go` | `ErrLockTimeout`: include recovery guidance in error message. |

### Phase 3: Lock Cleanup on Archive (addresses D2 R4)

| File | Change |
|------|--------|
| `internal/cmd/session/wrap.go` | After successful archive (line ~221), call `lockMgr.ForceRelease(sessionID)` to remove the lock file. |

### Phase 4: Documentation

| File | Change |
|------|--------|
| This ADR | Serves as the canonical reference for session model architecture. |
| `docs/guides/parallel-sessions.md` | Update with worktree isolation model and CI guidance. |
| `.claude/CLAUDE.md` state management section | Clarify Moirai vs CLI authority layers. |

### Not Implemented (Deferred)

| Item | Reason |
|------|--------|
| Manual `lock` duration limit (D2 R6) | Low priority. The `ari session lock` command is a debugging tool rarely used in production. |
| Configurable lock timeout via env var (D2 Q5) | Low priority. The 10-second default has not been reported as problematic. Can be added later without architectural impact. |
| NFS/network FS detection (D2 Risk 2.5) | Not a current requirement. If productization targets shared filesystems, revisit with a different locking strategy entirely. |

## Consequences

### Positive

1. **TOCTOU race eliminated.** The `__session-pointer__` lock ensures that `.current-session` mutations are serialized regardless of which command initiates them. The race described in D2 Risk 2.1 becomes impossible.
2. **ForceRelease is safer.** Users cannot accidentally create split-brain by force-unlocking a live process. The new liveness check and `--kill` flag make the danger explicit.
3. **Worktree model is documented.** The per-worktree isolation behavior was implicit; it is now an explicit architectural decision with rationale, enabling satellite developers to reason about multi-worktree deployments.
4. **Moirai/CLI layering is clarified.** The philosophical tension between "Moirai is sole authority" and "CLI is authoritative" is resolved by distinguishing the convention layer from the enforcement layer.
5. **Lock error messages guide recovery.** Users encountering lock contention receive actionable instructions rather than opaque error messages.

### Negative

1. **Slightly more lock contention.** Operations that mutate `.current-session` now acquire two locks (session-specific + `__session-pointer__`). In practice this adds microseconds; `flock` on a local file is fast.
2. **ForceRelease is more restrictive.** Users who previously relied on `ari session unlock --force` to bypass stuck locks must now also pass `--kill` if the holder is alive. This is intentional friction.
3. **CI requires worktrees for parallelism.** CI pipelines cannot run multiple sessions in a single checkout. This is an explicit trade-off: protecting the single-session invariant for the common case at the cost of CI complexity for the uncommon case.

### Neutral

1. **No schema changes.** `SESSION_CONTEXT.md` format is unchanged.
2. **No FSM changes.** The state machine (NONE -> ACTIVE -> PARKED -> ARCHIVED) is unchanged.
3. **No behavioral change for single-user workflows.** A developer using one terminal with one session will not notice any difference.

## Alternatives Considered

### Alternative: Replace File-Based Sessions with SQLite

Store session state in a SQLite database instead of `SESSION_CONTEXT.md` files. This would provide transactional atomicity, eliminate file-level locking, and support complex queries.

**Rejected because:**

1. Requires `CGO_ENABLED=1` (SQLite binding), which contradicts the static binary requirement (ADR-0010, Section 7).
2. Session state is intentionally human-readable. `SESSION_CONTEXT.md` can be inspected with `cat`, debugged with a text editor, and version-controlled with git. SQLite eliminates all three.
3. The file-based model works. D1-D3 found no data integrity issues. The only concurrency issue (`.current-session` TOCTOU) is solved by a file lock, not a database.
4. The file-per-session model naturally partitions data, making per-session operations (archive, delete, copy to worktree) trivial filesystem operations.

### Alternative: Event-Sourced Session State

Replace `SESSION_CONTEXT.md` with an event log. Current state is derived by replaying events.

**Rejected because:**

1. The current model already emits events to `events.jsonl` for audit purposes. Adding event sourcing would create two sources of truth (or require migrating all reads to event replay).
2. Event replay adds latency to every status check. Hooks call `ari hook context` on every tool invocation; this must be fast.
3. The FSM is simple (3 states, 4 transitions). Event sourcing adds value for complex domain models with many state dimensions. The session model is not complex enough to justify the overhead.
4. D1 validated that the current serialize/deserialize round-trip is correct. There is no data-loss problem that event sourcing would solve.

### Alternative: Process-Scoped Sessions (One Session Per PID)

Replace `.current-session` with a PID-indexed pointer (e.g., `.current-session.{pid}`). Each terminal gets its own session automatically.

**Rejected because:**

1. Claude Code's process model is not stable. A single Claude Code session may spawn multiple processes (main thread, Task agents, hook subprocesses) with different PIDs. Determining "which PID is the session owner" is non-trivial.
2. PID files accumulate and require reaping. A developer who opens and closes 20 terminals leaves 20 stale PID files.
3. Worktrees already provide process-level isolation without PID tracking.
4. This approach solves multi-terminal at the cost of making single-terminal harder to reason about (which of my PIDs owns the session?).

## Related Decisions

- **ADR-0001**: Session State Machine Redesign (defines FSM; this ADR refines lock scope but does not change the FSM)
- **ADR-0005**: Moirai Centralized State Authority (this ADR clarifies the convention vs. enforcement layering)
- **ADR-0010**: Worktree Session Seeding (this ADR affirms and extends the worktree isolation model)
- **ADR-0013**: Moirai Consolidation (this ADR is consistent with CLI-as-authority; no changes needed)

## References

| Reference | Location |
|-----------|----------|
| D1 Bug Report | `docs/bugs/D1-session-bugs.md` |
| D2 Lock Audit | `docs/audits/D2-lock-audit.md` |
| D3 Integration Tests | `internal/cmd/session/moirai_integration_test.go` |
| FSM Implementation | `internal/session/fsm.go` |
| Lock Implementation | `internal/lock/lock.go` |
| Session Commands | `internal/cmd/session/{create,park,resume,wrap}.go` |
| Current-Session Pointer | `internal/cmd/common/context.go` (lines 67-97) |
| Paths | `internal/paths/paths.go` |

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-02-05 | Claude Code (Context Architect) | Initial proposal synthesizing D1-D3 findings |

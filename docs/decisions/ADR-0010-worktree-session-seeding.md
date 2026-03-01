# ADR-0010: Worktree Session Seeding

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-01-05 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

### The Single-Session-Per-Terminal Constraint

The Ariadne CLI enforces a single active session per terminal to prevent state corruption and race conditions. This is implemented via:

1. **Session existence check** (`ariadne/internal/cmd/session/create.go:76-95`):
   ```go
   if currentID != "" {
       if _, err := os.Stat(resolver.SessionDir(currentID)); err == nil {
           existingCtx, err := session.LoadContext(ctxPath)
           if err == nil {
               err := errors.ErrSessionExists(currentID, string(existingCtx.Status))
               printer.PrintError(err)
               return err
           }
       }
   }
   ```

2. **Current session tracking** (`.sos/sessions/.current-session`):
   - Stores the active session ID for the terminal
   - Checked before any session creation

This constraint exists for good reason: concurrent session modifications within a single context would create state machine conflicts and audit trail corruption.

### The Problem: Preparing Parallel Work

ADR-0006 established the value of parallel session execution for independent workstreams. However, preparing sessions for parallel execution requires creating multiple PARKED sessions before execution begins. The current architecture blocks this:

```bash
# Terminal 1: Create session for work item A
$ ari session create "Work Item A" --complexity=MODULE
# Creates session-20260105-100000-aaaa1111, sets as current

# Still Terminal 1: Try to create session for work item B
$ ari session create "Work Item B" --complexity=MODULE
# ERROR: Session session-20260105-100000-aaaa1111 already exists with status ACTIVE
```

The user must either:
1. Wrap/archive the first session before creating the second (loses parallelism)
2. Open a new terminal for each session (works but loses coordination context)
3. Use worktrees to create isolated git contexts (current manual approach)

### Forces

- **Session Isolation**: Sessions must not interfere with each other's state
- **Parallel Preparation**: Multiple PARKED sessions needed before parallel execution
- **Single Constraint Respected**: The single-session-per-terminal rule is correct and must remain
- **Git Cleanliness**: Worktree-based sessions should not leave artifacts in the main branch
- **Fail-Safe Hooks**: Hooks must operate correctly even when `ari` binary is unavailable
- **State-Mate Authority**: State mutations must flow through state-mate (ADR-0005)

## Decision

### 1. Single Session Per Terminal Constraint: Maintained

The existing single-session constraint remains unchanged. The `--seed` flag works **around** this constraint by creating sessions in ephemeral worktrees, not by relaxing the constraint itself.

**Rationale**: The constraint prevents legitimate race conditions. Relaxing it would require complex locking that adds more problems than it solves. The worktree approach provides filesystem isolation that inherently prevents conflicts.

### 2. Worktree Creates Ephemeral Isolation

Sessions created with `--seed` follow this lifecycle:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SEEDING LIFECYCLE                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. Create Worktree         2. Create Session        3. Seed & Cleanup      │
│  ─────────────────          ──────────────────       ────────────────       │
│                                                                             │
│  git worktree add           ari session create       cp -r worktree/        │
│    /tmp/roster-seed-xxx       "Initiative"             .sos/sessions/    │
│    --detach                                            session-xxx/         │
│                             ari session park         → main/.claude/        │
│  cd /tmp/roster-seed-xxx      "Ready for parallel"     sessions/            │
│                                                                             │
│                                                      git worktree remove    │
│                                                        /tmp/roster-seed-xxx │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Step-by-Step Flow**:

1. **Create ephemeral worktree**:
   ```bash
   WORKTREE_PATH="/tmp/roster-seed-$(date +%s)-$$"
   git worktree add "$WORKTREE_PATH" --detach HEAD
   ```

2. **Create session in worktree** (worktree has its own `.sos/sessions/`):
   ```bash
   cd "$WORKTREE_PATH"
   ari session create "$INITIATIVE" --complexity="$COMPLEXITY"
   SESSION_ID=$(ari session status --format=json | jq -r '.session_id')
   ```

3. **Park session immediately** (seeded sessions start PARKED):
   ```bash
   ari session park "Seeded for parallel execution"
   ```

4. **Copy session to main branch's sessions directory**:
   ```bash
   cp -r "$WORKTREE_PATH/.sos/sessions/$SESSION_ID" \
         "$MAIN_REPO/.sos/sessions/"
   ```

5. **Cleanup worktree**:
   ```bash
   cd "$MAIN_REPO"
   git worktree remove "$WORKTREE_PATH" --force
   ```

**Result**: Session exists in main branch's `.sos/sessions/` with status `PARKED`, ready for `/resume` in any terminal.

### 3. CLI Contract: `--seed` Flag

The `ari session create` command gains a `--seed` flag:

```bash
ari session create "Initiative Name" \
    --complexity=MODULE \
    --seed
```

**Behavior**:

| Flag | Creates in | Initial Status | Sets Current | Worktree Lifecycle |
|------|------------|----------------|--------------|-------------------|
| (none) | Current directory | ACTIVE | Yes | N/A |
| `--seed` | Ephemeral worktree | PARKED | No | Create → Copy → Delete |

**Output**:
```json
{
  "session_id": "session-20260105-100000-aaaa1111",
  "status": "PARKED",
  "seeded": true,
  "seeded_to": "/Users/tom/Code/roster/.sos/sessions/session-20260105-100000-aaaa1111",
  "park_reason": "Seeded for parallel execution"
}
```

**Flags**:
- `--seed`: Enable worktree seeding mode
- `--seed-prefix=PATH`: Custom prefix for worktree (default: `/tmp/roster-seed-`)
- `--seed-keep`: Keep worktree after seeding (for debugging)

### 4. Hooks Fail-Open with JSON Audit Trail

When hooks cannot invoke the `ari` binary (e.g., binary not built, PATH issues), they must fail-open to avoid blocking Claude Code's workflow while maintaining auditability.

**Fail-Open Policy**:
```bash
# Hook attempts ari invocation
if ! command -v ari &>/dev/null; then
    # Fail-open: allow operation but log to audit trail
    log_fail_open "ari binary not found"
    exit 0  # Allow operation to proceed
fi
```

**Audit Schema** (`.claude/audit/fail-open.jsonl`):

```json
{"timestamp": "2026-01-05T10:00:00Z", "hook": "session-write-guard.sh", "operation": "Edit", "error": "ari binary not found", "context": {"file_path": ".sos/sessions/session-xxx/SESSION_CONTEXT.md", "tool_name": "Edit"}}
{"timestamp": "2026-01-05T10:00:01Z", "hook": "orchestrator-router.sh", "operation": "TaskComplete", "error": "ari session status failed: exit 1", "context": {"session_id": "session-xxx", "exit_code": 1}}
```

**Schema Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | ISO 8601 | When the fail-open occurred |
| `hook` | string | Hook script name (e.g., `session-write-guard.sh`) |
| `operation` | string | Tool operation that triggered hook (e.g., `Edit`, `Write`) |
| `error` | string | Error message explaining why fail-open occurred |
| `context` | object | Additional context (file path, session ID, exit code, etc.) |

**Log Rotation**: Fail-open logs are appended. Rotation handled by external tooling or manual cleanup.

### 5. State-Mate Bypass Detection

The `session-write-guard.sh` hook blocks direct writes to `*_CONTEXT.md` files per ADR-0005. However, the state-mate agent must be allowed to perform these writes.

**Current Mechanism** (environment variable):
```bash
# session-write-guard.sh (lines 33-35)
if [[ "${STATE_MATE_BYPASS:-}" == "true" ]]; then
    exit 0  # Allow write
fi
```

**Enhanced Mechanism** (agent name detection):

When Claude Code invokes a Task agent, the hook can detect the agent name via environment variable:

```bash
# Check if invoked by state-mate agent
AGENT_NAME="${CLAUDE_TASK_AGENT_NAME:-}"
if [[ "$AGENT_NAME" == "state-mate" ]]; then
    exit 0  # Allow write - state-mate is authorized
fi

# Fallback to environment marker
if [[ "${STATE_MATE_BYPASS:-}" == "true" ]]; then
    exit 0
fi
```

**Detection Priority**:
1. `CLAUDE_TASK_AGENT_NAME` environment variable (set by Claude Code Task tool)
2. `STATE_MATE_BYPASS` environment marker (legacy/fallback)
3. Neither set: block operation with error message

**Note**: The `CLAUDE_TASK_AGENT_NAME` variable is a proposed enhancement. Until Claude Code implements this, the `STATE_MATE_BYPASS` environment marker remains the primary mechanism.

### 6. Sessions Created as PARKED by Default (Seeding Mode)

Seeded sessions start in `PARKED` status with a standard park reason:

```yaml
---
session_id: "session-20260105-100000-aaaa1111"
status: "PARKED"
parked_at: "2026-01-05T10:00:00Z"
parked_reason: "Seeded for parallel execution"
initiative: "Work Item A"
complexity: "MODULE"
schema_version: "2.0"
---
```

**Rationale**:
- PARKED status prevents the session from being treated as "current" in any terminal
- Explicit park reason documents the session's seeded origin
- `/resume` activates the session when a terminal is ready to work on it

### 7. CGO_ENABLED=0 Build Constraint

The `ari` binary is built with `CGO_ENABLED=0` (see `ariadne/justfile:8-9`):

```makefile
build:
    CGO_ENABLED=0 go build -o ari ./cmd/ari/main.go
```

**Rationale**:

On macOS arm64, CGO-enabled binaries can encounter dyld issues related to LC_UUID mismatches when the binary is built in one environment and executed in another (e.g., built in Rosetta, run natively). Specifically:

1. **LC_UUID Verification**: macOS verifies that shared libraries have matching UUIDs to prevent accidental mixing of library versions
2. **CGO Dependencies**: When CGO is enabled, Go links against system libraries (libc, libSystem) which have platform-specific UUIDs
3. **Cross-Compilation Issues**: Building on one architecture variant can embed UUIDs that don't match the runtime environment

**Consequence**: Pure Go builds (`CGO_ENABLED=0`) avoid system library linking entirely, producing fully static binaries that work across macOS environments without dyld verification failures.

**Tradeoff**: Some Go packages require CGO (e.g., `sqlite3`, `libgit2`). The Ariadne CLI intentionally avoids such dependencies to maintain portability.

## Consequences

### Positive

1. **Parallel Preparation Enabled**: Users can create multiple PARKED sessions for parallel execution without terminal juggling
2. **Constraint Preserved**: Single-session-per-terminal rule remains intact, maintaining state machine safety
3. **Audit Trail Maintained**: Fail-open behavior is logged, enabling post-hoc debugging
4. **Clean Main Branch**: Ephemeral worktrees leave no artifacts after seeding completes
5. **State-Mate Authority Preserved**: Bypass mechanism is explicit and detectable

### Negative

1. **Worktree Overhead**: Creating/deleting worktrees adds ~5-10 seconds per seeded session
2. **Disk Space During Seeding**: Ephemeral worktrees consume disk space (mitigated by immediate cleanup)
3. **Git Worktree Complexity**: Users unfamiliar with worktrees may find debugging harder if seeding fails mid-process

### Neutral

1. **Fail-Open Philosophy**: Hooks allowing operations when `ari` unavailable is a design choice balancing availability vs. enforcement
2. **Environment Variable Detection**: Reliance on `STATE_MATE_BYPASS` is a convention, not a hard security boundary

## Implementation Guidance

### P2: Implement `--seed` Flag in Session Create

**File**: `ariadne/internal/cmd/session/create.go`

**Changes**:
1. Add `--seed` flag to command definition
2. When `--seed` is set:
   - Create ephemeral worktree
   - Execute session create in worktree context
   - Park session immediately
   - Copy session directory to main repo
   - Remove worktree
3. Output should include `seeded: true` field

**Test Cases**:
- `ari session create "Test" --seed` creates PARKED session in main repo
- Multiple `--seed` invocations don't conflict
- Cleanup removes worktree even on partial failure

### P3: Implement Fail-Open Audit Logging

**File**: `.claude/hooks/lib/hooks-init.sh` (or new `fail-open.sh`)

**Changes**:
1. Create `log_fail_open()` function
2. Ensure `.claude/audit/` directory exists
3. Append JSON line to `fail-open.jsonl`
4. All hooks using `ari` should call `log_fail_open()` on binary unavailability

**Schema Validation**: Consider JSON Schema for fail-open entries to ensure consistency.

### P4: Enhance State-Mate Bypass Detection

**File**: `.claude/hooks/session-guards/session-write-guard.sh`

**Changes**:
1. Check for `CLAUDE_TASK_AGENT_NAME` first
2. Fall back to `STATE_MATE_BYPASS` if agent name not set
3. Log which bypass mechanism was used (for debugging)

**Note**: `CLAUDE_TASK_AGENT_NAME` requires Claude Code enhancement. Implement fallback-first, upgrade when available.

### P5: Document Seeding Workflow in User Guide

**File**: `docs/guides/parallel-sessions.md` (new)

**Content**:
1. When to use session seeding
2. Step-by-step: Creating multiple parallel sessions
3. Resuming seeded sessions in separate terminals
4. Troubleshooting worktree failures
5. Relationship to ADR-0006 parallel execution pattern

## Related Decisions

- **ADR-0001**: Session State Machine Redesign (defines PARKED status)
- **ADR-0005**: State-Mate Centralized State Authority (defines write guard mechanism)
- **ADR-0006**: Parallel Session Orchestration Pattern (motivates need for session seeding)

## References

- Session creation logic: `ariadne/internal/cmd/session/create.go`
- Write guard hook: `.claude/hooks/session-guards/session-write-guard.sh`
- Build configuration: `ariadne/justfile`
- Current session tracking: `.sos/sessions/.current-session`
- Worktree skill: `.claude/skills/session-lifecycle/worktree-ref.md`

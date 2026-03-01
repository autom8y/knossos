# Parallel Session Execution

> Create multiple PARKED sessions for parallel execution in separate terminals

## Overview

Session seeding enables creating multiple PARKED sessions without violating the single-session-per-terminal constraint. Sessions are created in ephemeral worktrees and "seeded" back to the main branch.

**Use Cases**:
- Preparing multiple sessions for parallel execution
- Breaking down a large initiative into independent workstreams
- Running multiple Claude Code terminals simultaneously on different tasks

## Quick Start

### Create Seeded Sessions

```bash
# Create multiple sessions for parallel work
ari session create "Feature A" --complexity=MODULE --seed
ari session create "Feature B" --complexity=MODULE --seed
ari session create "Feature C" --complexity=PATCH --seed
```

Each command creates a PARKED session without activating it, allowing you to create as many as needed from a single terminal.

### Resume in Separate Terminals

```bash
# Terminal 1
ari session resume session-xxx-feature-a

# Terminal 2
ari session resume session-xxx-feature-b

# Terminal 3
ari session resume session-xxx-feature-c
```

Each terminal now has an independent active session.

## How It Works

The `--seed` flag creates sessions through an ephemeral worktree lifecycle:

```
+-----------------------------------------------------------------------------+
|                           SEEDING LIFECYCLE                                  |
+-----------------------------------------------------------------------------+
|                                                                             |
|  1. Create Worktree         2. Create Session        3. Seed & Cleanup      |
|  -----------------          ------------------       ----------------       |
|                                                                             |
|  git worktree add           ari session create       cp -r worktree/        |
|    /tmp/roster-seed-xxx       "Initiative"             .sos/sessions/    |
|    --detach                                            session-xxx/         |
|                             ari session park         -> main/.claude/       |
|  cd /tmp/roster-seed-xxx      "Ready for parallel"     sessions/            |
|                                                                             |
|                                                      git worktree remove    |
|                                                        /tmp/roster-seed-xxx |
|                                                                             |
+-----------------------------------------------------------------------------+
```

### Step-by-Step Flow

1. **Create ephemeral worktree**: A detached worktree is created at a temporary location
2. **Create session in worktree**: Normal session creation runs in the isolated worktree (which has its own `.sos/sessions/`)
3. **Park session immediately**: Seeded sessions are automatically parked with reason "Seeded for parallel execution"
4. **Copy session to main branch**: The session directory is copied back to the main repository's `.sos/sessions/`
5. **Cleanup worktree**: The ephemeral worktree is removed, leaving no artifacts

**Result**: Session exists in main branch's `.sos/sessions/` with status `PARKED`, ready for `/resume` in any terminal.

## Command Reference

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--seed` | Enable worktree seeding mode | false |
| `--seed-prefix=PATH` | Custom worktree location | `/tmp/roster-seed-` |
| `--seed-keep` | Keep worktree after seeding (for debugging) | false |

### Behavior Comparison

| Flag | Creates in | Initial Status | Sets Current | Worktree Lifecycle |
|------|------------|----------------|--------------|-------------------|
| (none) | Current directory | ACTIVE | Yes | N/A |
| `--seed` | Ephemeral worktree | PARKED | No | Create -> Copy -> Delete |

### Output Example

```json
{
  "session_id": "session-20260105-100000-aaaa1111",
  "status": "PARKED",
  "seeded": true,
  "seeded_to": "/Users/tom/Code/roster/.sos/sessions/session-20260105-100000-aaaa1111",
  "park_reason": "Seeded for parallel execution"
}
```

## When to Use Session Seeding

### Good Use Cases

- **Multi-feature sprints**: Prepare 3-5 independent features to work on in parallel
- **Team coordination**: Create sessions for different team members to pick up
- **Independent workstreams**: Tasks with no shared file dependencies

### Not Recommended For

- **Sequential dependencies**: If Task B depends on Task A's output, don't parallelize
- **Shared file modifications**: If sessions modify the same files, conflicts will occur
- **Hook infrastructure changes**: Sessions modifying hooks should run serially

## Parallel Execution Pattern

For complex initiatives, consider the hybrid parallelization pattern from ADR-0006:

### Layer 1: Serial Execution (Dependencies)

```
T=0h                T=2.5h
|------ Session A -------|
                  |--- Session B ---|
```

Sessions with dependencies (e.g., shared hook modifications) run sequentially.

### Layer 2: Parallel Execution (Independent)

```
T=2.5h   T=3.5h  T=4.5h  T=5.5h  T=6h
         |--- Session C (skills) ---|
         |--- Session D (docs) -----|
         |--- Session E (tests) ----|
```

Independent sessions run simultaneously in separate terminals.

### Identifying Parallelizable Sessions

Sessions can run in parallel when they have:
- **No shared file modifications** (use Glob/Grep to verify disjoint file regions)
- **No hook interference** (Layer 1 hook work complete)
- **No state machine conflicts** (each session has independent locks via state-mate)
- **Read-only shared resources** (e.g., orchestrator.yaml, preferences.json)

## Troubleshooting

### Worktree Creation Fails

**Symptom**: `fatal: '/tmp/roster-seed-xxx' already exists`

**Solution**:
```bash
# List existing worktrees
git worktree list

# Remove stale worktree
git worktree remove /tmp/roster-seed-xxx --force

# Retry seeding
ari session create "Feature" --complexity=MODULE --seed
```

### Session Not Found After Seeding

**Symptom**: `ari session resume <id>` fails with "session not found"

**Possible Causes**:
1. Seeding process was interrupted before copy completed
2. Running from a different repository than where seeding occurred

**Solution**:
```bash
# Verify session directory exists
ls -la .sos/sessions/

# Check if worktree still exists (use --seed-keep for debugging)
git worktree list

# If worktree still exists, manually copy
cp -r /tmp/roster-seed-xxx/.sos/sessions/session-* .sos/sessions/
git worktree remove /tmp/roster-seed-xxx --force
```

### Worktree Not Cleaned Up

**Symptom**: Stale worktrees accumulating in `/tmp/`

**Solution**:
```bash
# List all worktrees
git worktree list

# Remove orphaned worktrees
git worktree prune

# Force remove specific worktree
git worktree remove /tmp/roster-seed-xxx --force
```

### Permission Errors

**Symptom**: `permission denied` during copy or worktree operations

**Solution**:
```bash
# Check session directory permissions
ls -la .sos/sessions/

# Fix permissions if needed
chmod 755 .sos/sessions/

# Ensure .claude exists
mkdir -p .claude/sessions
```

## Integration with state-mate

Seeded sessions integrate with state-mate for state management:

- **Session creation**: state-mate validates session schema
- **Park operation**: state-mate records park reason in audit trail
- **Resume**: state-mate transitions status from PARKED to ACTIVE
- **Per-session locking**: Each parallel session acquires independent lock

No special configuration is needed - state-mate automatically handles parallel session state mutations via per-session locking.

## Disk Space Considerations

- **During seeding**: Each worktree consumes ~equivalent space to main repo
- **After seeding**: Only session directory (~1-10KB) persists
- **Cleanup**: Worktrees are automatically removed unless `--seed-keep` is set

For large repositories, consider using `--seed-prefix` to direct worktrees to a location with adequate space.

## Related Documentation

- [ADR-0010: Worktree Session Seeding](../decisions/ADR-0010-worktree-session-seeding.md) - Architecture decision record
- [ADR-0006: Parallel Session Orchestration](../decisions/ADR-0006-parallel-session-orchestration.md) - Parallel execution pattern
- [Knossos Integration Guide](knossos-integration.md) - Ariadne CLI overview
- [User Preferences](user-preferences.md) - Configure session behavior

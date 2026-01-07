---
name: rite-ref
description: "Switch agent rite packs or list available rites via roster system. Use when: changing active rite, listing available packs, checking current rite status. Triggers: /rite, switch rite, change rite pack, rite management, roster."
---

# /rite - Agent Rite Pack Switcher

> **Category**: Rite Management | **Phase**: Rite Switching

## Purpose

Switch between agent rite packs to access specialized workflows. Each rite pack provides a curated set of agents optimized for specific types of work (development, documentation, code hygiene, technical debt).

This command integrates with the roster system at `$KNOSSOS_HOME/` and updates the active rite context for the current project.

---

## Usage

```bash
/rite                 # Show current active rite
/rite <pack-name>     # Switch to specified rite pack
/rite --list          # List all available rite packs
```

### Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `pack-name` | No | Name of rite pack to switch to |
| `--list` | No | List all available rite packs |

---

## Behavior

When `/rite` is invoked, the following sequence occurs:

### 1. Parse Arguments

Determine operation mode:

- **No arguments**: Query and display current active rite
- **`--list` or `-l`**: List all available rites
- **`<pack-name>`**: Switch to specified rite

### 2. Invoke Roster Script

Execute the swap-rite.sh script via Bash tool:

```bash
$KNOSSOS_HOME/swap-rite.sh [args]
```

The script handles:
- Validation of rite existence
- Backup of current agents (if any)
- Atomic swap of agent files
- Update of `.claude/ACTIVE_RITE` state file

### 3. Display Results

Show script output to user, which includes:

- **For query**: Current active rite name
- **For list**: All available rites
- **For swap**: Confirmation with agent count loaded

### 4. Update SESSION_CONTEXT (if active)

If a session is active (`.claude/sessions/{session_id}/SESSION_CONTEXT.md` exists):

- Update `active_rite` field to new rite name
- Append to handoff notes:
  ```
  Rite switched: {old-rite} → {new-rite} ({agent-count} agents)
  Reason: [User-provided or "Manual rite switch"]
  ```

---

## Available Rite Packs

Rite packs are discovered dynamically from `$KNOSSOS_HOME/rites/`. Reference the `rite-discovery` skill for structured metadata access.

### Current Inventory

To list all rites at runtime:
```bash
ls -d $KNOSSOS_HOME/rites/*-pack 2>/dev/null | xargs -n1 basename
```

As of this writing, the roster contains 11 rites:
- 10x-dev-pack (software development)
- debt-triage-pack (technical debt)
- doc-rite-pack (documentation)
- ecosystem-pack (roster infrastructure)
- forge-pack (rite creation)
- hygiene-pack (code quality)
- intelligence-pack (analytics/research)
- rnd-pack (exploration/prototyping)
- security-pack (security assessment)
- sre-pack (operations/reliability)
- strategy-pack (business analysis)

**Important**: This list is informational. For current, accurate rite data, use `rite-discovery` skill or read directly from `$KNOSSOS_HOME/rites/*/orchestrator.yaml`.

### Rite Details

For detailed rite profiles including agents, routing conditions, and use cases:
- Run `/consult --rite` for formatted display
- Reference `rite-discovery` skill for structured data
- Read `rites/{name}/README.md` for extended documentation

---

## State Changes

### Files Modified

| File | Change | Description |
|------|--------|-------------|
| `.claude/ACTIVE_RITE` | Overwritten | Contains single line with active rite name |
| `.claude/agents/` | Replaced | All agent files swapped atomically |
| `.claude/agents.backup/` | Created | Backup of previous agents (safety net) |
| `.claude/sessions/{session_id}/SESSION_CONTEXT.md` | Updated | If session active, rite field updated |

### Exit Codes

The swap-rite.sh script returns:

- `0` - Success
- `1` - Invalid arguments
- `2` - Validation failure (pack doesn't exist or invalid structure)
- `3` - Backup failure (disk full, permissions)
- `4` - Swap failure (incomplete copy, file count mismatch)

---

## Examples

### Example 1: Query Current Rite

```bash
/rite
```

Output:
```
[Roster] Active rite: 10x-dev-pack
```

### Example 2: List Available Rites

```bash
/rite --list
```

Output:
```
[Roster] Available rites:
  - 10x-dev-pack
  - debt-triage-pack
  - doc-rite-pack
  - hygiene-pack
```

### Example 3: Switch to Doc Rite

```bash
/rite doc-rite-pack
```

Output:
```
[Roster] Backed up current agents to .claude/agents.backup/
[Roster] Switched to doc-rite-pack (4 agents loaded)
```

After switch, `.claude/agents/` contains:
- `doc-auditor.md`
- `information-architect.md`
- `tech-writer.md`
- `doc-reviewer.md`

### Example 4: Switch During Active Session

```bash
/rite hygiene-pack
```

Output:
```
[Roster] Backed up current agents to .claude/agents.backup/
[Roster] Switched to hygiene-pack (4 agents loaded)

Session context updated:
  Active rite: hygiene-pack
  Handoff note added: "Rite switched: 10x-dev-pack → hygiene-pack (4 agents)"
```

### Example 5: Idempotent Switch

```bash
/rite doc-rite-pack   # Already active
```

Output:
```
[Roster] Already using doc-rite-pack (no changes needed)
```

**Idempotency**: The script automatically detects when already on the target rite and exits early. Use `--refresh` to pull latest agent definitions from roster.

---

## Refresh Mode

Use `--refresh` when you need to pull the latest agent definitions from the roster, even if already on that team.

### When to Use

- After updating agents in the roster repository
- When agents seem stale or behaving unexpectedly
- After running `git pull` in the roster repo
- To reset local agent modifications to upstream state

### Examples

```bash
# Refresh current rite (most common)
/rite --refresh

# Refresh specific rite
/rite 10x-dev-pack --refresh

# Preview what would change before refreshing
/rite --refresh --dry-run
```

### Dry-Run Output

```
[Roster] Dry-run: Would refresh ecosystem-pack

Agent changes:
  ~ orchestrator.md (modified in roster)
  = ecosystem-analyst.md (unchanged)
  = integration-engineer.md (unchanged)
  + new-agent.md (new)

No changes made (--dry-run mode)
```

---

## Error Handling

### Rite Pack Not Found

```bash
/rite nonexistent-pack
```

Output:
```
[Roster] Error: Rite pack 'nonexistent-pack' not found in $KNOSSOS_HOME/rites/
[Roster] Use './swap-rite.sh --list' to see available packs
```

**Resolution**: Use `/rite --list` to see valid rite names

### Invalid Pack Structure

If a rite exists but missing `agents/` directory:

```
[Roster] Error: Rite pack 'broken-pack' missing agents/ directory
```

**Resolution**: Fix rite structure in roster repository

### Backup Failure

If `.claude/` directory not writable:

```
[Roster] Error: Backup failed (disk full? permissions?)
```

**Resolution**: Check disk space, verify `.claude/` permissions

### Swap Failure

If file copy fails or count mismatch:

```
[Roster] Error: File count mismatch (expected 5, got 3)
[Roster] Restore from backup: cp -r .claude/agents.backup/* .claude/agents/
```

**Resolution**: Run restore command, investigate disk/permission issues

---

## Prerequisites

- Roster system installed at `$KNOSSOS_HOME/`
- At least one rite exists in `$KNOSSOS_HOME/rites/`
- `.claude/` directory exists and is writable (created automatically if missing)

---

## Integration with Other Commands

### /start - Session Initialization

The `/start` command supports `--rite=PACK` parameter:

```bash
/start "Add dark mode" --rite=10x-dev-pack
```

This internally calls `/rite 10x-dev-pack` before creating SESSION_CONTEXT.

### /handoff - Agent Coordination

When handing off between agents, if target agent not in current rite:

```
Agent 'debt-collector' not found in current rite.
Switch to 'debt-triage-pack' with /rite debt-triage-pack
```

### /wrap - Session Finalization

On session wrap, current rite recorded in session summary:

```
Session completed:
  Rite used: hygiene-pack
  Agents invoked: code-smeller, janitor
```

---

## Quick Switch Commands

Quick-switch commands are derived from rite names:

| Rite | Quick Switch | Derivation |
|------|--------------|------------|
| 10x-dev-pack | `/10x` | First token before hyphen |
| debt-triage-pack | `/debt` | First token before hyphen |
| doc-rite-pack | `/docs` | First token before hyphen |
| ecosystem-pack | `/ecosystem` | First token before hyphen |
| forge-pack | `/forge` | First token before hyphen |
| hygiene-pack | `/hygiene` | First token before hyphen |
| intelligence-pack | `/intelligence` | First token before hyphen |
| rnd-pack | `/rnd` | First token before hyphen |
| security-pack | `/security` | First token before hyphen |
| sre-pack | `/sre` | First token before hyphen |
| strategy-pack | `/strategy` | First token before hyphen |

These commands invoke `/rite {pack-name}` internally and display rite roster after switch.

---

## Success Criteria

- Correct rite loaded in `.claude/agents/`
- `.claude/ACTIVE_RITE` file updated with rite name
- Previous agents backed up to `.claude/agents.backup/`
- All expected agent files present (validated by count)
- If session active, SESSION_CONTEXT updated

---

## Related Commands

- `/10x` - Quick switch to 10x-dev-pack
- `/docs` - Quick switch to doc-rite-pack
- `/hygiene` - Quick switch to hygiene-pack
- `/debt` - Quick switch to debt-triage-pack
- `/start` - Initialize session with rite selection

---

## Related Documentation

- [swap-rite.sh]($KNOSSOS_HOME/swap-rite.sh) - Roster swap script
- [TDD-roster-system.md]($KNOSSOS_HOME/docs/design/TDD-0003-rite-swap.md) - Roster design
- [COMMAND_REGISTRY.md](../../COMMAND_REGISTRY.md) - All registered commands

---

## Notes

### Why Atomic Swaps?

The roster system uses backup-then-swap to prevent corruption:

1. Backup current agents to `.claude/agents.backup/`
2. Clear `.claude/agents/`
3. Copy new agents from roster
4. Validate file count matches
5. Update `.claude/ACTIVE_RITE`

If any step fails, previous agents can be restored from backup.

### Rite Pack Structure

Each rite in `$KNOSSOS_HOME/rites/` has:

```
rite-name/
  agents/
    agent1.md
    agent2.md
    ...
  README.md  (optional)
```

The `agents/` directory must contain at least one `.md` file.

### Session Context Awareness

This command is session-aware:
- Works with or without active session
- Updates SESSION_CONTEXT.active_rite if session exists
- Logs rite switch in handoff notes for audit trail

### Environment Variable Override

The roster location can be customized:

```bash
export KNOSSOS_HOME=/custom/path/to/roster
/rite --list
```

Default: `~/Code/roster`

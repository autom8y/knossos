---
name: start
description: Initialize a new work session
argument-hint: <initiative> [--complexity=LEVEL] [--rite=PACK]
allowed-tools: Bash, Read, Task
model: opus
---

## Pre-computed Context

The SessionStart hook has already injected all session state above. Check:
- **Session Status** table → whether a session exists
- **Has Session** → true/false
- **Session State** → IDLE, ACTIVE, or PARKED
- **Pre-computed Values** → suggested session ID, entry agent

## Your Task

$ARGUMENTS

## Behavior

### 1. Check Pre-conditions (Read from context above)

| If Session Status Shows | Action |
|------------------------|--------|
| `Has Session = false` | Proceed with session creation |
| `Has Session = true, Parked = true` | Offer options (see below) |
| `Has Session = true, Parked = false` | Offer options (see below) |

**When session already exists, offer these options:**

```
A session already exists in this terminal.

Options:
1. /continue - Resume the parked session
2. /park + /start - Park current, then start new
3. /wrap - Complete current session first
4. /worktree create "<name>" - Start in ISOLATED worktree (parallel work)

Tip: Use worktrees when you want to work on something different
without affecting the current session/team.
```

The `/worktree` option is especially useful when:
- Different team needed for the new work
- Want to keep current sprint context intact
- Need true parallel sessions on same project

### 2. Gather Parameters

If not provided in arguments, ask the user:

| Parameter | Description | Default |
|-----------|-------------|---------|
| **Initiative** | What are we building? | Required |
| **Complexity** | PATCH \| MODULE \| SYSTEM \| INITIATIVE \| MIGRATION | MODULE |
| **Rite** | Rite to use | Current rite from context |

### 3. Create Session (ONE command)

```bash
hooks/lib/session-manager.sh create "<initiative>" "<complexity>" "<rite>"
```

This atomically:
- Creates session directory
- Maps TTY to session
- Creates SESSION_CONTEXT.md
- Returns JSON with session_id, entry_agent

### 4. Rite Switch (only if --rite differs)

If user specified `--rite=NAME` and it differs from Active Rite:
```bash
ari sync materialize --rite <rite-name>
```

**Note**: If `ari` is not in PATH, use `~/bin/ari` or fall back to legacy:
`${ROSTER_HOME:-~/Code/roster}/swap-rite.sh <rite-name>`

### 5. Invoke Entry Point Agent

Read **Entry Agent** from context (or from session-manager response).

Use Task tool to invoke the entry agent:
- Default: `requirements-analyst`
- Task: "Create PRD for: <initiative>"
- Include complexity level in task description

### 6. Confirm Success

Output to user:
- Session ID created
- Entry agent invoked
- What happens next

## Complexity Levels

| Level | Phases | Use For |
|-------|--------|---------|
| PATCH | requirements → implementation | Single-file changes, quick fixes |
| MODULE | requirements → design → implementation | Multi-file features |
| SYSTEM | + validation | New services, APIs |
| INITIATIVE | + Session -1/0 assessment | Major initiatives |
| MIGRATION | Full lifecycle + migration planning | Cross-cutting migrations, large-scale refactors |

## Example Usage

```
/start "Add dark mode toggle"
/start "Refactor auth module" --complexity=MODULE
/start "New billing service" --complexity=SYSTEM --rite=10x-dev
```

## Reference

Full documentation: `.claude/skills/start-ref/skill.md`

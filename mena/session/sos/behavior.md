# /sos Behavior Specification

## Natural Language Routing

When $ARGUMENTS does not match a subcommand keyword, parse as natural language:

| User Says | Detected Intent | Maps To |
|-----------|----------------|---------|
| "park because waiting on feedback" | park + reason | `park "waiting on feedback"` |
| "hand this off to the architect" | handoff + agent | `handoff architect` |
| "I need to pause" | park (no reason) | `park "user requested pause"` |
| "resume where I left off" | resume | `resume` |
| "fork this for a spike" | fray | `fray` |
| "what's the status" | status | (no args) |
| "quick fix for the auth bug" | start PATCH | `start "quick fix for the auth bug"` |
| "bind to session-xxx" | claim | `claim session-xxx` |

### Keyword Detection Order

1. Park keywords: park, pause, stop, hold, save -> **park**
2. Resume keywords: resume, continue, unpause, pick up -> **resume**
3. Delegation keywords (+ agent name): hand off, delegate, give to -> **handoff**
4. Fork keywords: fork, fray, branch, spike -> **fray**
5. Claim keywords: bind, claim, attach -> **claim**
6. Start keywords: start, begin, fix, quick -> **start** (PATCH default)
7. Wrap keywords: wrap, finish, done, archive, complete -> **wrap**
8. No match -> **ambiguity fallback**

### Ambiguity Fallback

When the intent is unclear, do NOT guess. Present the available operations:

```
I could not determine the intended operation. Available subcommands:
  /sos park "reason"    - Pause current session
  /sos resume           - Resume parked session
  /sos handoff <agent>  - Transfer to agent
  /sos fray             - Fork to parallel strand
  /sos claim <id>       - Bind CC to session
  /sos start "desc"     - Start lightweight session
  /sos wrap             - Archive session
```

## Multi-Operation Protocol

Some operations compose naturally. Execute as sequential Task(moirai) calls:

| Compound Operation | Sequence | Use Case |
|-------------------|----------|----------|
| park + start | park_session -> create_session | Switch context |
| resume + handoff | resume_session -> handoff | Resume and delegate |

**Anti-pattern**: Do NOT combine wrap + start. Use `/land` for wrap-with-synthesis, then `/sos start` separately. Wrap has knowledge extraction concerns that /sos does not handle.

## Fray + Claim Integration

When `/sos fray` creates a child session, include claim instructions in the output:

```
Frayed: {parent_id} -> {child_id}
Worktree: {worktree_path}

Next steps:
  cd {worktree_path} && claude
  # In the new CC instance, run:
  /sos claim {child_id}
```

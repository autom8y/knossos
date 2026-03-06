# /sos Behavior Specification

## Natural Language Routing

When $ARGUMENTS does not match a subcommand keyword, parse as natural language:

| User Says | Detected Intent | Maps To |
|-----------|----------------|---------|
| "park because waiting on feedback" | park + reason | `park "waiting on feedback"` |
| "I need to pause" | park (no reason) | `park "user requested pause"` |
| "resume where I left off" | resume | `resume` |
| "what's the status" | status | (no args) |
| "quick fix for the auth bug" | start PATCH | `start "quick fix for the auth bug"` |

### Keyword Detection Order

1. Park keywords: park, pause, stop, hold, save -> **park**
2. Resume keywords: resume, continue, unpause, pick up -> **resume**
3. Start keywords: start, begin, fix, quick -> **start** (PATCH default)
4. Wrap keywords: wrap, finish, done, archive, complete -> **wrap**
5. Delegation keywords: hand off, delegate, give to -> redirect to `/handoff`
6. Fork keywords: fork, fray, branch, spike -> redirect to `/fray`
7. No match -> **ambiguity fallback**

### Ambiguity Fallback

When the intent is unclear, do NOT guess. Present the available operations:

```
I could not determine the intended operation. Available subcommands:
  /sos park "reason"    - Pause current session
  /sos resume           - Resume parked session
  /sos start "desc"     - Start lightweight session
  /sos wrap             - Archive session

Related commands:
  /handoff <agent>      - Transfer to agent
  /fray                 - Fork to parallel strand
```

## Multi-Operation Protocol

Some operations compose naturally. Execute as sequential Task(moirai) calls:

| Compound Operation | Sequence | Use Case |
|-------------------|----------|----------|
| park + start | park_session -> create_session | Switch context |
| resume + handoff | resume_session -> /handoff | Resume, then redirect to /handoff |

**Anti-pattern**: Do NOT combine wrap + start. Use `/land` for wrap-with-synthesis, then `/sos start` separately. Wrap has knowledge extraction concerns that /sos does not handle.


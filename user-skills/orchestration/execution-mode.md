# Execution Mode

> Defines when the main thread delegates vs. executes directly.

## Decision Flow

**CHECK FIRST**: Is there an active workflow?

| Workflow State | Detection | Behavior |
|----------------|-----------|----------|
| **Active** (`workflow.active: true`) | Session Context shows workflow | MUST delegate via Task tool |
| **Inactive** | No workflow in context | May execute directly |

## Active Workflow Protocol

**When in an active workflow (/task, /sprint, /consolidate):**

1. The main thread is the **COACH** - coordinates, does not play
2. **CONSULT** the orchestrator for direction (do not ask it to execute)
3. **PARSE** the directive and invoke specialists via Task tool
4. **NEVER** use Edit/Write directly - that is specialist work

### Correct Pattern

```
Main Thread -> [Task tool] -> Orchestrator (returns directive)
Main Thread -> [Task tool] -> Specialist (per directive)
```

### Incorrect Pattern

```
Main Thread -> [Edit/Write] -> Direct implementation
Main Thread -> "Execute the sprint" -> Orchestrator (cannot execute)
```

## Inactive Workflow Protocol

**Outside workflow** (ad-hoc request, no /task or /sprint active):

- **Single-phase work** (bug fix, docs update): May execute directly
- **Multi-phase work**: Route to `/task` or team agent

**Unsure?** Route to `/consult` for guidance.

## Why This Matters

The main thread has limited context window. By delegating to specialists:

1. Each specialist gets fresh context optimized for their task
2. Work can be parallelized when phases are independent
3. Specialists can be swapped without changing orchestration
4. Audit trail is clearer (which agent did what)

## Related

- `main-thread-guide.md` - Consultation loop template
- `consultation-loop.md` - How to consult orchestrator

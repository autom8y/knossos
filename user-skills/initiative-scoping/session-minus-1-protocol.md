# Session -1 Protocol: Initiative Assessment

> Orchestrator assesses initiative readiness. Main agent is a prompter only.

## Purpose

Session -1 is **context ingestion only**. The main agent:
1. Receives an initiative description from the user
2. Invokes the Orchestrator to assess it
3. Returns the Orchestrator's guidance verbatim

**No implementation, no planning** - just Orchestrator assessment.

## Protocol

### 1. Acknowledge

```
I've received your initiative. Let me invoke the Orchestrator to assess readiness.
```

### 2. Invoke Orchestrator

Use the Task tool with `@orchestrator`:

```markdown
## Session -1: Initiative Assessment

**Task**: Assess whether this initiative is ready for the 4-agent workflow.

### Initiative
{USER'S INITIATIVE DESCRIPTION}

### Required Assessment
Using the `10x-workflow` skill:

1. **North Star**: What this achieves and what success looks like (1-3 sentences)

2. **Go/No-Go**:
   - Problem validated?
   - Scope bounded?
   - Blocking dependencies?
   - Complexity level? (Script/Module/Service/Platform)
   - Recommendation: GO / CONDITIONAL GO / NO-GO

3. **If GO/CONDITIONAL GO**:
   - Which agents, in what order?
   - Right-sized workflow for complexity?
   - Context needed before Session 0?

4. **Blocking Questions**: Must-answer before proceeding

5. **Risks/Assumptions**: Key failure modes
```

### 3. Return Guidance

Present the Orchestrator's assessment verbatim. Ask user to:
- Proceed to Session 0 (if GO)
- Address conditions (if CONDITIONAL GO)
- Resolve blockers (if NO-GO)

## Session Output

| Output | Purpose |
|--------|---------|
| North Star | Objective and success criteria |
| Go/No-Go | Whether to proceed |
| Workflow Sizing | Which agents, in what order |
| Blocking Questions | What must be answered |
| Risks/Assumptions | Failure modes to watch |

This output becomes **input for Session 0**.

## Related

- [shared-principles.md](shared-principles.md) - Main agent behavior rules
- [session-0-protocol.md](session-0-protocol.md) - Next session if GO

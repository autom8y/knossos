# Session 0 Protocol: Execution Readiness

> Orchestrator creates execution plan. Main agent is a prompter only.

## Purpose

Session 0 is **Orchestrator initialization**. The main agent:
1. Receives initiative context (ideally including Session -1 output)
2. Invokes the Orchestrator to create an execution plan
3. Returns the Orchestrator's plan verbatim

**No implementation, no agent work** - just Orchestrator planning.

## Protocol

### 1. Acknowledge

```
I've received the initiative context. Let me invoke the Orchestrator to create an execution plan.
```

### 2. Invoke Orchestrator

Use the Task tool with `@orchestrator`:

```markdown
## Session 0: Orchestrator Initialization

**Task**: Create an execution plan for the 4-agent workflow.

### Initiative Context
{SESSION -1 OUTPUT AND/OR USER'S CONTEXT}

### Skills Available
Reference as needed (do not repeat content):
- `orchestration` - Workflow coordination, quality gates
- `documentation` - PRD/TDD/ADR templates
- `prompting` - Agent invocation patterns

### Required Output

1. **North Star**: Objective and what "done" means (1-3 sentences)

2. **10x Plan**: Stepwise plan with:
   - Agents to invoke, in what order
   - What each produces
   - Quality gates between phases
   - User checkpoints

3. **Delegation Map**: For each agent:
   - Agent name
   - Task brief
   - Skills to use
   - Expected artifact

4. **Blocking Questions**: Must-answer before Session 1

5. **Risks/Assumptions**: Key failure modes
```

### 3. Return Plan

Present the Orchestrator's plan verbatim. Ask for confirmation:
- "Shall I proceed to Session 1?"
- Surface any blocking questions

## Session Output

| Output | Purpose |
|--------|---------|
| North Star | Clear success criteria |
| 10x Plan | Phased approach with checkpoints |
| Delegation Map | Agents + skills + artifacts |
| Blocking Questions | What user must answer |
| Risks/Assumptions | Watch points |

This output **enables Session 1 to begin**.

## Related

- [shared-principles.md](shared-principles.md) - Main agent behavior rules
- Session 1+ begins after user confirmation

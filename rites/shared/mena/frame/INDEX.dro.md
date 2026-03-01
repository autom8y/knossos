---
name: frame
description: "Invoke context-engineer to decompose initiatives into workstreams or frame rite handoffs. Produces .claude/wip/frames/{slug}.md."
argument-hint: "<brief>"
allowed-tools: Read, Glob, Grep, Write, Task, Skill
model: opus
---

# /frame -- Initiative Framing

Dispatches context-engineer to analyze the current conversation and codebase state, producing a structured framing document that can feed directly into `/start`, `/sprint`, or `/rite` invocations.

## Context

This command runs in the main thread. Main-thread execution is required because context-engineer needs full visibility into the conversation history -- a forked context would lose the conversation that makes framing meaningful. Single Task dispatch, no Argus pattern.

## Pre-flight

1. **Parse `$ARGUMENTS`**:
   - Treat the full argument string as the user's brief verbatim.
   - If empty: ERROR "Usage: /frame <brief> -- Provide a short description of the initiative or handoff to frame."

2. **Normalize to slug**:
   - Convert the brief to kebab-case: lowercase, spaces to hyphens, strip non-alphanumeric except hyphens.
   - Truncate to 60 characters if longer.
   - The output path will be: `.claude/wip/frames/{slug}.md`

3. **Ensure output directory exists**:
   - Write a placeholder or use Bash: `mkdir -p .claude/wip/frames`
   - This is a side effect -- execute it before the Task dispatch.

4. **Read session context** (if available):
   - Check for `.claude/SESSION_CONTEXT.md` (injected by hooks when a session is active).
   - If it exists, read it and extract: active rite, sprint name/number, current phase, recent decisions.
   - If it does not exist: note "no active session" -- framing proceeds without session context.
   - Do NOT call `ari` commands or run shell introspection to discover session state.

## CE Dispatch

Construct the Task prompt and dispatch context-engineer. Include everything the CE needs to produce a useful framing document without asking clarifying questions.

```
Task(subagent_type="context-engineer", prompt="
## Framing Request

### User's Brief

{brief verbatim from $ARGUMENTS}

### Session Context

{If SESSION_CONTEXT.md was found:}
- Active rite: {rite}
- Sprint: {sprint}
- Phase: {phase}
- Recent decisions: {summary from SESSION_CONTEXT.md}

{If no session context:}
- No active session. Frame as a standalone initiative.

### Your Directive

Analyze the conversation history above and the user's brief to produce a framing document.

The framing document should decompose the initiative into actionable workstreams, identify the right starting point, and end with concrete suggested next commands (e.g., `/start`, `/sprint`, `/rite`, `/go` invocations the user can run immediately).

You have full discretion over the framing document's structure -- there are no prescribed sections. Design the schema that best serves this specific initiative.

### Output

Write the framing document to: .claude/wip/frames/{slug}.md

The document must end with a '## Next Commands' section listing the exact commands the user should run next (with arguments), in priority order.
")
```

## Report

After context-engineer returns:

1. Confirm the artifact was written:
   ```
   Read(".claude/wip/frames/{slug}.md", limit=20)
   ```

2. Display to the user:
   ```
   ## Frame: .claude/wip/frames/{slug}.md

   {CE's suggested next commands, extracted from the ## Next Commands section}

   Read the full framing document: .claude/wip/frames/{slug}.md
   ```

If the file was not written (CE did not produce output at the expected path), WARN: "context-engineer did not write the expected file. Check the CE output above for the framing document."

## Error Handling

| Scenario | Action |
|----------|--------|
| No `$ARGUMENTS` provided | ERROR with usage message |
| SESSION_CONTEXT.md unreadable | Proceed without session context, note omission |
| CE Task dispatch fails | ERROR "Framing failed: {reason}" |
| Output file not found after CE returns | WARN with path; display CE output directly |

## Anti-Patterns

- **Reading source files yourself**: You are the dispatcher. Let CE observe the codebase and conversation. Do not pre-load architecture files or run codebase scans.
- **Prescribing the document schema**: CE has full discretion over artifact structure. The only required section is `## Next Commands`.
- **Running ari commands for session state**: Only read SESSION_CONTEXT.md if it exists. Do not shell out to discover session information.
- **Forking context**: This dromenon intentionally runs in the main thread. Do not add `context: fork`.

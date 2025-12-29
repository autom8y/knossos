# Consultation Loop

> The execution pattern all orchestrated commands implement

## The Loop

```
CONSULTATION LOOP
=================

┌─────────────────────────────────────────────────────────────────┐
│  START: User invokes /task, /sprint, or /consolidate           │
└───────────────────────────┬─────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. BUILD REQUEST                                               │
│     Main agent constructs CONSULTATION_REQUEST                  │
│     - Gather current state from session context                 │
│     - Summarize artifacts (NOT full content)                    │
│     - Set request type (initial/checkpoint/decision/failure)    │
└───────────────────────────┬─────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. CONSULT ORCHESTRATOR                                        │
│     Main agent invokes orchestrator via Task tool               │
│     - Pass CONSULTATION_REQUEST as prompt context               │
│     - Orchestrator returns CONSULTATION_RESPONSE                │
└───────────────────────────┬─────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. PARSE DIRECTIVE                                             │
│     Main agent reads directive.action                           │
└───────────────────────────┬─────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┬─────────────────┐
        ▼                   ▼                   ▼                 ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌─────────┐
│ invoke_       │   │ request_      │   │ await_        │   │complete │
│ specialist    │   │ info          │   │ user          │   │         │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └────┬────┘
        │                   │                   │                │
        ▼                   ▼                   ▼                │
┌───────────────┐   ┌───────────────┐   ┌───────────────┐        │
│ Execute Task  │   │ Gather info   │   │ Ask user      │        │
│ with prompt   │   │ from files,   │   │ question      │        │
│ from response │   │ codebase, or  │   │ from response │        │
│               │   │ user          │   │               │        │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘        │
        │                   │                   │                │
        ▼                   ▼                   ▼                │
┌───────────────┐   ┌───────────────┐   ┌───────────────┐        │
│ Build         │   │ Build updated │   │ Build         │        │
│ checkpoint    │   │ request with  │   │ decision      │        │
│ request       │   │ info          │   │ request       │        │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘        │
        │                   │                   │                │
        └───────────────────┴───────────────────┘                │
                            │                                    │
                            ▼                                    │
                    [Return to step 2]                           │
                                                                 │
                            ┌────────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  END: Finalize session, report completion to user              │
└─────────────────────────────────────────────────────────────────┘
```

## Request Types

| Type | When Used | Key Fields |
|------|-----------|------------|
| `initial` | First consultation for new initiative | initiative, context_summary |
| `checkpoint` | After specialist completes work | results.phase_completed, results.artifact_summary |
| `decision` | After user answers a question | results with user decision |
| `failure` | When specialist fails or blocks | results.failure_reason |

## Loop Invariants

1. **Main agent controls execution** - Only main agent uses Task tool
2. **Orchestrator is stateless** - All state passed in request
3. **Summaries not files** - Main agent summarizes, orchestrator trusts
4. **Structured only** - No prose conversations in the loop

## Cycle Count

Typical iteration counts by complexity:

| Complexity | Phases | Expected Cycles |
|------------|--------|-----------------|
| SCRIPT | 1-2 | 2-4 |
| MODULE | 3-4 | 6-10 |
| SERVICE | 4-6 | 10-15 |
| PLATFORM | 6+ | 15-25 |

Each checkpoint adds one cycle. Each `request_info` or `await_user` adds one cycle.

## Error Recovery

When a specialist fails:

```
1. Main agent detects failure (Task tool error, incomplete artifact)
2. Build failure request:
   - type: failure
   - results.failure_reason: "Specialist could not complete: [reason]"
3. Consult orchestrator
4. Orchestrator returns recovery directive:
   - New prompt addressing the issue, OR
   - Request for more info, OR
   - Rollback to previous phase
```

## Token Economics

| Component | Target | Notes |
|-----------|--------|-------|
| CONSULTATION_REQUEST | 200-400 tokens | Context summarization |
| CONSULTATION_RESPONSE | 400-500 tokens | Full directive + prompt |
| Specialist prompt (embedded) | 200-300 tokens | Within response |

**Key constraint**: Main agent provides summaries (1-2 sentences per artifact), not full file contents. Orchestrator trusts summaries without verification.

## Invariants (Canonical)

These invariants are non-negotiable across all orchestrated workflows:

1. **Main agent owns Task tool** - Only main agent invokes specialists or orchestrator
2. **Orchestrator is stateless** - All state comes from CONSULTATION_REQUEST
3. **Summaries not files** - Main agent summarizes artifacts; orchestrator never reads files
4. **Structured formats only** - No prose conversation; YAML request/response only
5. **Throughline tracking** - Every response includes decision + rationale for audit trail

## See Also

- [request-format.md](request-format.md) - CONSULTATION_REQUEST schema
- [response-format.md](response-format.md) - CONSULTATION_RESPONSE schema
- [command-integration.md](command-integration.md) - How commands use this loop
- [specialist-returns.md](specialist-returns.md) - What specialists return

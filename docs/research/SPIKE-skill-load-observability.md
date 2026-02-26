---
spike_id: "spike-20260226-172907"
created_at: "2026-02-26T17:29:07Z"
question: "Does CC's PostToolUse payload contain enough data to correlate Skill invocations with agents, enabling preload waste/gap reporting?"
timebox: "30m"
deliverable_type: "report"
status: "complete"
---

# SPIKE: Skill Load Observability

## Question

Can a PostToolUse hook for Skill invocations log enough data to the clew to enable:
1. **Waste detection**: preloaded skills (tokens spent) that agents never reference
2. **Gap detection**: skills loaded on-demand that should be preloaded

## Findings

### Finding 1: Project-level PostToolUse hooks do NOT fire for subagent tool calls

CC fires project-level hooks (defined in `settings.json` / `settings.local.json`) only for **main-thread** tool calls. When a subagent calls `Skill("foo")`, the project-level PostToolUse hook does not fire. Subagent hooks must be defined in the subagent's own frontmatter to observe its tool calls.

**Impact**: Fatal for the proposed approach. Most skill loads happen inside subagents (agents call `Skill("name")` on-demand), which are invisible to project-level hooks.

### Finding 2: PostToolUse payload contains no agent identity

The `StdinPayload` schema (`internal/hook/env.go:49-63`) has no `agent_name`, `agent_id`, or `subagent_id` field:

```
session_id, transcript_path, cwd, permission_mode, hook_event_name,
tool_name, tool_input, tool_response, tool_use_id, prompt, source,
stop_hook_active, trigger
```

Even if PostToolUse fired for subagent calls, we couldn't attribute the Skill load to a specific agent without temporal correlation with SubagentStart/SubagentStop events (unreliable with concurrent agents).

### Finding 3: Preloaded skills bypass the Skill tool entirely

Skills listed in agent frontmatter (`skills: ["foo", "bar"]`) are injected directly into the system prompt at agent startup. CC reads the skill files and includes their content before the first turn. **The Skill tool is never invoked.** No PostToolUse event is emitted.

**Impact**: Preload waste detection is impossible via any hook mechanism. Preloaded content is invisible to the event system — it's baked into the initial prompt, not a tool call.

### Finding 4: Existing infrastructure is closer than expected

The codebase already has:
- `CommandInvokedData` type (`internal/hook/clewcontract/typed_data.go:100-106`) with `command` and `type` fields
- `NewTypedCommandInvokedEvent` constructor (`typed_constructors.go:127-132`) — "emitted by PostToolUse hook when Skill tool call detected"
- **But no producer**: The clew hook (`internal/cmd/hook/clew.go`) only matches `Edit|Write|Bash`, not `Skill`

This infra would work for main-thread Skill calls but doesn't solve the subagent blindspot.

### Finding 5: SubagentStop provides transcript path

The `SubagentStop` payload includes `agent_transcript_path`, which is the full transcript JSONL for the completed agent. This transcript contains every tool call the agent made, including Skill invocations.

## Assessment

| Detection Goal | PostToolUse Hook | Feasible? |
|----------------|-----------------|-----------|
| Main-thread Skill loads | Would work (add `Skill` to clew matcher) | Yes, but low value — main thread rarely calls Skill |
| Subagent on-demand Skill loads | Project hooks don't fire for subagents | **No** |
| Preloaded skill usage | Preloads bypass Skill tool entirely | **No** |
| Preloaded skill waste | No event emitted at all | **No** |

**Verdict: PostToolUse hook approach is not feasible for the stated goals.**

## Alternative Approaches

### A. Transcript parsing at SubagentStop (Recommended)

Parse `agent_transcript_path` at SubagentStop to extract Skill tool calls. Cross-reference with agent's `skills:` frontmatter to compute:
- **Waste**: skills in frontmatter that never appeared as Skill tool calls (never needed on-demand, and preload content may have been sufficient or unused)
- **Gaps**: Skill tool calls for skills NOT in the frontmatter (agent loaded them on-demand)

**Pros**: Accurate, complete data; leverages existing SubagentStop hook infrastructure.
**Cons**: Heavyweight (parsing JSONL transcripts on every agent completion); can't distinguish "preload was sufficient" from "preload was wasted" (agent may have used the preloaded knowledge without calling Skill again).

**Fundamental limitation**: Even transcript parsing can't tell us whether an agent _used_ the preloaded skill content. It can only detect whether the agent made _additional_ Skill tool calls. A preloaded skill that the agent reads from context (no tool call needed) looks identical to a preloaded skill that was never referenced.

### B. Agent frontmatter hooks

Define PostToolUse hooks in each agent's markdown frontmatter. CC fires these for the specific agent's tool calls.

**Pros**: Per-agent tool call visibility.
**Cons**: Requires modifying all agent prompts; adds hook overhead to every agent invocation; still can't detect preload usage (only on-demand calls).

### C. Post-hoc session analysis (Simplest)

After session wrap, batch-analyze all agent transcripts. No real-time overhead.

**Pros**: No hook changes needed; can aggregate across sessions for statistical patterns.
**Cons**: Delayed feedback (end-of-session only); requires transcript storage.

### D. Instrumented skill wrapper

Create a thin wrapper skill that logs its own load to a known file path, then includes the real skill content. Each preloaded skill becomes `skills: ["observed/foo"]` where `observed/foo.md` starts with a side-effect marker.

**Pros**: Could detect preload loads.
**Cons**: Doubles skill count; fragile; doesn't detect _usage_, only _load_.

## Recommendation

**Do not build a PostToolUse hook for Skill observability.** The approach has three independent blockers (no subagent visibility, no agent identity, preloads invisible).

**Next step**: If skill observability remains a priority, pursue **Option C** (post-hoc transcript analysis) as a low-cost `ari report skill-usage` command. This avoids real-time hook overhead and provides statistical patterns across sessions. The key insight is that _observing_ preload usage is fundamentally impossible at the tool-call level — only usage _patterns_ (Skill calls that suggest preloading, or agents that never call skills suggesting waste) can be inferred statistically.

## Files Examined

| File | Relevance |
|------|-----------|
| `internal/hook/env.go` | StdinPayload schema — confirmed no agent identity field |
| `config/hooks.yaml` | Current hook registrations — clew only matches Edit\|Write\|Bash |
| `internal/cmd/hook/clew.go` | Clew hook implementation — no Skill handling |
| `internal/cmd/hook/subagent.go` | SubagentStart/Stop — agent identity available here |
| `internal/hook/clewcontract/typed_data.go` | CommandInvokedData type exists but has no producer |
| `internal/hook/clewcontract/typed_constructors.go` | NewTypedCommandInvokedEvent constructor exists |

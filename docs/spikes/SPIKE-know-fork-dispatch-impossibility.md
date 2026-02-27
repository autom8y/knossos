# SPIKE: /know Fork-Dispatch Impossibility

**Date**: 2026-02-27
**Status**: Complete
**Question**: Why does `/know --all` consistently fail to dispatch theoros subagents on satellites?
**Decision Informs**: Whether to remove `context: fork` from the `/know` dromenon

---

## TL;DR

The `/know` dromenon has `context: fork`, which runs it as a subagent. CC's fundamental constraint: **agents cannot spawn agents — only the main thread has Task tool access.** The forked agent literally does not have the Task tool. All prompt engineering ("YOU MUST USE TASK TOOL") is irrelevant — the tool isn't in its toolbox. The fix is architectural: remove `context: fork`.

---

## Evidence

### Debug Log Analysis

Two satellite dogfood runs examined:
- `4b9cb079` (autom8y-ads, 3418 lines)
- `2e3681d4` (autom8y, 3657 lines)

Both show identical pattern:

```
Executing forked slash command /know with agent general-purpose
```

**Tool sequence in both logs (no Task tool calls anywhere):**
1. `Skill(pinakes)` — loads domain registry ✓
2. `Bash` ×4-6 — git rev-parse, mkdir, etc. ✓
3. `Read` ×5-6 — loads criteria files ✓
4. `TaskCreate` ×1-5 — creates **todo items** (not subagent dispatch)
5. `Read` ×30-50 — **does observation itself** (source files, test files, configs)
6. `Bash` ×3-6 — additional exploration
7. `Write` ×4-5 — writes .know/ files **directly**
8. `TaskUpdate` ×1-5 — marks todos complete

**Zero `Task` tool calls** (subagent dispatch) in either run. The model uses `TaskCreate`/`TaskUpdate` (todo tracking) but never `Task` (subagent dispatch) because it's not available.

### The Architectural Constraint

Three independent sources confirm the same rule:

**1. The Doctrine** (`docs/doctrine/philosophy/knossos-doctrine.md:218`):
> "The pattern's architectural constraint mirrors Argus's nature -- agents cannot spawn agents, only the main thread dispatches. One giant, many eyes. One Theseus, many theoroi."

**2. The CLAUDE.md inscription** (`.claude/CLAUDE.md`):
> "Agents cannot spawn other agents — only the main thread has Task tool access."

**3. CC Platform Behavior** (empirical, from debug logs):
The forked agent has: Bash, Read, Write, Glob, Grep, Skill, TaskCreate, TaskUpdate, TaskList, TaskGet.
The forked agent does NOT have: **Task** (subagent dispatch).

### Why Prompt Engineering Cannot Fix This

The dromenon currently contains:
```markdown
**YOU MUST USE THE TASK TOOL TO DISPATCH THEOROS SUBAGENTS.**
Do NOT attempt to observe the codebase yourself.
```

And an anti-pattern:
```markdown
**Performing observation yourself instead of dispatching theoros**:
You are the ORCHESTRATOR, not the observer.
```

These instructions are correct in intent but impossible to follow. The model reads them, has no Task tool available, and rationally falls back to the only option: doing the work itself. The model's behavior is not defiant — it's adaptive.

### The Dromenon Frontmatter

```yaml
allowed-tools: Bash, Read, Write, Glob, Grep, Task, Skill
model: opus
context: fork
```

`allowed-tools: Task` declares intent but CC does not honor this for forked contexts. The platform strips Task from subagent tool sets regardless of what the dromenon frontmatter requests.

---

## Root Cause

**The `/know` dromenon design violates a fundamental CC platform invariant.** It uses `context: fork` (creating a subagent) while requiring `Task` tool (subagent dispatch), which is only available to the main thread. This is an architectural impossibility, not a prompt compliance issue.

The myth itself tells this story: the **polis** (main thread) sends the theoria (delegation). A theoros (observer) cannot send other theoroi. The authority to delegate is inherent to the dispatcher, not the dispatched.

---

## Recommended Fix

**Remove `context: fork` from the `/know` dromenon.**

When `/know` runs in the main thread:
1. Main thread has Task tool ✓
2. Main thread can dispatch theoros subagents (Argus Pattern) ✓
3. Each theoros gets its own 150-turn context window ✓
4. Heavy codebase observation happens in theoros windows, not main ✓
5. Main thread only does lightweight orchestration (criteria load + assembly) ✓

### Concern: Main Thread Context Pollution

The original `context: fork` existed to protect the main context window. But the `/know` orchestrator is lightweight:
- Phase 1: Read 5 criteria files (~2k tokens each = ~10k)
- Phase 2: Dispatch 5 Task calls (minimal tokens)
- Phase 3: Receive 5 theoros outputs, write 5 files
- Phase 4: Print summary

Total main-thread token cost: ~30-40k tokens for orchestration. The heavy lifting (40+ Read calls per domain) happens in theoros subagent windows.

**Trade-off**: ~30-40k main context tokens vs. a working Argus Pattern. Acceptable.

### Alternative: Keep fork, Redesign to Single-Agent

If `context: fork` must be preserved, redesign the dromenon to be a competent single-agent observer (current behavior, but optimized). Remove all Task dispatch instructions. Accept reduced quality from single-context observation.

**Not recommended**: contradicts the doctrine's theoria model and the Argus Pattern. Single-agent observation demonstrably produces 4/5 domains (context exhaustion) vs. 5/5 with dispatch.

---

## Implementation

Single-line change in `rites/shared/mena/know/INDEX.dro.md`:

```yaml
# Before
context: fork

# After (remove the line entirely, or:)
context: main
```

Then `ari sync` to project the change to satellites.

No other changes needed — the dromenon's Phase 2 instructions already correctly describe the Argus Pattern dispatch. Once the main thread executes them (with Task tool available), the design works as intended.

---

## Follow-Up Actions

1. Remove `context: fork` from `/know` dromenon
2. Run `ari sync` on knossos
3. Re-test `/know --all` on a satellite to verify theoros dispatch
4. If main-thread context cost is problematic, explore CC's `context` options for a middle ground
5. Update the dromenon's anti-patterns section to document why fork + Task is impossible

---

## Lessons

1. **Architectural constraints beat prompt engineering.** No instruction can make a tool appear that the platform doesn't provide.
2. **The doctrine contained the answer.** "Agents cannot spawn agents" was documented in philosophy, CLAUDE.md, AND CC's own system prompt — three consistent sources that the dromenon design contradicted.
3. **Debug logs reveal tool availability.** The absence of Task tool calls + presence of TaskCreate calls was the diagnostic signal. The model uses what it has.
4. **The myth is the architecture.** The theoros metaphor (polis sends delegation) correctly models the constraint. When the design violates the myth, the implementation breaks.

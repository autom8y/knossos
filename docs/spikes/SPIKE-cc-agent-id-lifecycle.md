# SPIKE: CC Agent ID Lifecycle

**Date**: 2026-02-19
**Status**: Findings Complete
**Question**: How do Claude Code agent IDs work, what is their lifecycle, and is the knossos resume protocol viable for cross-rite rollout?
**Method**: Codebase exploration + CC documentation analysis + GitHub issue triage
**Timebox**: Single session
**Informs**: Resume cross-rite rollout decision (deferred from CC-OPP), ADR-0028

---

## Executive Summary

CC agent IDs are **session-scoped unique identifiers** assigned to each subagent when spawned via the Task tool. They persist as transcript files (`agent-{agentId}.jsonl`) within the session directory. The `resume` parameter on the Task tool allows re-invoking a subagent with its full conversation history restored. The knossos throughline resume protocol (OPP-004) correctly models these semantics, but the CC-side implementation had a critical bug (Issue #11712: user prompts not stored in transcripts) that was reported fixed in December 2025. The protocol is architecturally sound for cross-rite rollout, but empirical validation is needed before committing.

---

## The Question

Three sub-questions:

1. **Mechanics**: How are CC agent IDs generated, what is their scope, and when do they expire?
2. **Resume reliability**: Is the `resume` parameter on the Task tool reliably functional?
3. **Cross-rite viability**: Should the throughline resume protocol (currently ecosystem-only) be rolled out to all rites?

---

## Findings

### Finding 1: Agent ID Lifecycle

**Generation**: CC assigns a unique agent ID when the Task tool spawns a subagent. The ID format appears to be an opaque string (not documented as UUID or any specific format). The ID is returned to the main thread as part of the Task tool result when the subagent completes.

**Storage**: Each subagent's conversation is persisted as a JSONL transcript file at:
```
~/.claude/projects/{project-hash}/{sessionId}/subagents/agent-{agentId}.jsonl
```

Transcript entries include:
- `timestamp`: ISO 8601
- `type`: `user_message`, `assistant_message`, `tool_use`, `tool_result`
- `content`: Message/tool payload
- `metadata`: Model, token counts, compact boundaries

**Scope**: Agent IDs are scoped to the **CC session** (the conversation instance). They are:
- Valid within the current CC session
- Persistent within that session even across Claude Code restarts (via `--resume`)
- Invalid across different sessions
- Cleaned up based on `cleanupPeriodDays` setting (default: 30 days)

**Compaction**: Subagent transcripts support auto-compaction independently of the main conversation. Compaction triggers at ~95% capacity (configurable via `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`). Compaction events are logged in the transcript.

### Finding 2: The Resume Parameter

The Task tool accepts a `resume` parameter with an agent ID from a previous invocation. When provided:

1. CC loads the transcript file for that agent ID
2. The subagent receives its full conversation history (all prior tool calls, results, and reasoning)
3. The subagent picks up where it left off rather than starting fresh
4. New instructions are appended as the next user message

**Key capability**: Resumed subagents retain their full context window. This means a Pythia orchestrator consulted 5 times in sequence would have access to all 5 prior consultation requests, its own prior responses, and the reasoning thread it built. This is the exact behavior the knossos throughline protocol requires.

**Foreground vs Background**: Subagents can run in foreground (blocking) or background (concurrent). Resumed subagents follow the same model. Background subagents auto-deny permissions not pre-approved. If a background subagent fails due to missing permissions, it can be resumed in the foreground.

### Finding 3: The December 2025 Bug and Fix

**Issue #11712** (filed Nov 2025, closed Dec 8 2025): Agent transcript files were missing all user prompts. Only assistant responses and tool results were stored. This caused resumed agents to:
- Lose the original dispatch instruction
- Lose all resume instructions
- Reconstruct context from their own prior responses only
- Hallucinate after 2-3 resumes due to accumulated context drift

**Root cause**: The transcript writing logic for main sessions was not being called for subagent sessions. Main sessions correctly stored user prompts; subagents simply skipped this step.

**Fix status**: Anthropic engineer confirmed fix, issue closed Dec 8 2025 as COMPLETED.

**Issue #11892** (filed Dec 2025, closed as duplicate): Reported that the Task tool's system prompt says "each agent invocation is stateless" contradicting the `resume` parameter. Closed as duplicate of #10864 (Task tool not returning agent IDs in results). Both reported fixed.

**Current status (Feb 2026)**: The resume functionality should be operational. However, no empirical validation has been performed within the knossos ecosystem since the fix was deployed.

### Finding 4: Knossos Throughline Protocol Analysis

The throughline resume protocol (OPP-004, ecosystem-only) is documented in `knossos/templates/sections/agent-routing.md.tpl`:

```
The main thread MAY track subagent IDs for throughline agents (Pythia, Moirai)
and pass `resume: {agentId}` on subsequent Task calls. This gives the agent
full history of its prior consultations within the workflow.

- Agent IDs are valid only within the current CC session
- Clear stored IDs on rite switch or session wrap
- If resume fails (invalid ID, session changed), fall back to fresh invocation
- Resume is opportunistic -- orchestrated workflows function correctly without it
```

**Assessment**: This protocol is well-designed:
- Correctly identifies session-scoped validity
- Correctly specifies fallback behavior (opportunistic, not required)
- Correctly targets throughline agents only (Pythia, Moirai)
- Does not create hard dependencies on resume working

The ecosystem Pythia agent (`rites/ecosystem/agents/pythia.md`) implements the receiving side:
- Documents "When resumed" vs "When starting fresh" behavior
- Includes "Context Checkpoint" pattern (persisting rationale in `throughline.rationale`)
- Explicitly states "Resume is opportunistic. The system works correctly without it."

### Finding 5: What the Knossos Platform Already Tracks

The `ari` binary already handles subagent lifecycle events via hooks:

- **SubagentStart hook** (`internal/cmd/hook/subagent.go`): Logs `agent.task_start` events to `events.jsonl` with agent_name, agent_type, and task_id
- **SubagentStop hook**: Logs `agent.task_end` events
- **Clew events**: The clew contract includes `EventTypeTaskStart` and `EventTypeTaskEnd` types
- **Throughline extraction** (`internal/hook/clewcontract/orchestrator.go`): Parses `throughline.decision` and `throughline.rationale` from orchestrator responses

**Gap**: The hooks capture agent_name and task_id but do NOT capture the CC-assigned agent_id (the opaque ID needed for resume). The `subagentPayload` struct has `AgentName`, `AgentType`, and `TaskID` but no `AgentID` field. This means knossos cannot currently recover agent IDs from its own event log -- the main thread must track them in-memory.

### Finding 6: ADR-0006 Empirical Evidence

ADR-0006 (Parallel Session Orchestration) documents real-world use of throughline IDs:

> "9 background agents required careful coordination... Per-session mapping of state-mate/orchestrator/specialist IDs prevents confusion"

The ADR confirms that agent IDs were successfully tracked across parallel session executions in January 2026. However, this was done manually by the main thread, not via any automated knossos mechanism.

---

## Comparison Matrix

| Aspect | Current (Ecosystem-only) | Cross-Rite Rollout | Wait for More Evidence |
|--------|--------------------------|--------------------|-----------------------|
| Protocol correctness | Validated | Same protocol | N/A |
| CC resume bug status | Fixed (Dec 2025) | Fixed | Independently verifiable |
| Pythia agent patterns | 1 rite tested | 13 Pythias need review | 0 additional risk |
| Main thread tracking | In-memory only | In-memory only | Could add hook capture |
| Fallback on failure | Clean (fresh invoke) | Same mechanism | Same mechanism |
| Implementation effort | Done | Template already shared | N/A |
| Risk if resume broken | None (opportunistic) | None (opportunistic) | None |

---

## Recommendation

**Verdict**: PROCEED with cross-rite rollout, with one prerequisite.

**Rationale**:

1. **The protocol is already rolled out.** The `agent-routing.md.tpl` template that contains the throughline resume protocol is synced to ALL rites via materialization. Every rite's CLAUDE.md already contains the resume protocol documentation. The question is not "should we roll it out" but "is the receiving side (Pythia agents) ready."

2. **Risk is zero.** The protocol is explicitly opportunistic. If resume fails, the system falls back to fresh invocation. Workflows work correctly without it. There is no downside to having it available.

3. **The bug is fixed.** Issue #11712 was closed as COMPLETED in December 2025. Issue #11892 (contradictory system prompt) was also resolved. The CC infrastructure should support resume correctly.

4. **Ecosystem Pythia already demonstrates the pattern.** The `rites/ecosystem/agents/pythia.md` shows exactly how to document "When resumed" vs "When starting fresh" behavior with a context checkpoint pattern. This can be templated for other rites.

### Prerequisite: Empirical Validation

Before declaring cross-rite rollout complete, run a manual validation:

1. Start an orchestrated workflow in the ecosystem rite
2. Consult Pythia 3+ times with explicit resume
3. On the 3rd consultation, verify Pythia references decisions from consultation 1
4. Verify context checkpoint in `throughline.rationale` carries forward
5. Test the fallback path: pass an invalid agent ID and verify clean fresh invocation

This validation should take 30-60 minutes and will provide the empirical evidence needed for ADR-0028.

### Optional Enhancement: Capture Agent ID in Hooks

The SubagentStart/SubagentStop hooks currently do not capture the CC-assigned agent ID. If CC sends this in the hook payload (check `tool_input` JSON), adding an `agent_id` field to the `subagentPayload` struct and logging it to events.jsonl would enable:
- Post-hoc throughline analysis
- Automated agent genealogy tracking
- Resume recovery from event logs

This is a low-priority enhancement -- the main thread's in-memory tracking is sufficient for the resume protocol.

---

## Follow-Up Actions

| Priority | Action | Effort | Blocked By |
|----------|--------|--------|------------|
| P1 | Run empirical resume validation (ecosystem rite) | 30-60 min | Nothing |
| P2 | Review all 13 Pythia agents for "When resumed" pattern | 2-3 hours | P1 |
| P2 | Write ADR-0028 with pilot + rollout + validation evidence | 2-3 hours | P1 |
| P3 | Investigate if CC sends agent_id in SubagentStart hook payload | 30 min | Nothing |
| P3 | Add agent_id capture to subagentPayload if available | 30 min | P3 investigation |

---

## Sources

- [CC Subagents Documentation](https://code.claude.com/docs/en/sub-agents)
- [CC Task Tool System Prompt](https://github.com/Piebald-AI/claude-code-system-prompts/blob/main/system-prompts/tool-description-task.md)
- [Issue #11712: Subagent Resume Missing All User Prompts](https://github.com/anthropics/claude-code/issues/11712)
- [Issue #11892: Task Tool Guidance Contradicts Resumable Functionality](https://github.com/anthropics/claude-code/issues/11892)
- [DeepWiki: Agent System & Subagents](https://deepwiki.com/anthropics/claude-code/3.1-agent-system-and-subagents)
- Internal: `knossos/templates/sections/agent-routing.md.tpl` (throughline protocol)
- Internal: `rites/ecosystem/agents/pythia.md` (reference implementation)
- Internal: `internal/cmd/hook/subagent.go` (hook capture)
- Internal: `internal/hook/clewcontract/orchestrator.go` (throughline extraction)
- Internal: `docs/decisions/ADR-0006-parallel-session-orchestration.md` (empirical ID tracking)
- Internal: `docs/spikes/SPIKE-session-lifecycle-functional-audit.md` (prior audit)
- Internal: `docs/assessments/SCOUT-claude-code-agent-discovery.md` (CC agent cache behavior)

---

*Generated by single-session spike, 2026-02-19.*
*Spike by Claude Opus 4.6.*

---
domain: feat/clew-conversation-management
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/slack/conversation/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.92
format_version: "1.0"
---

# Clew Thread Context and Conversation History

## Purpose and Design Rationale

Provides multi-turn coherence for Clew. Two problems: within-session continuity ("follow up on that") and cross-restart continuity (container deploy wipes memory). Explicit tradeoff: pure in-memory state for speed, Slack API resurrection as deploy fallback. Hybrid windowing: last N messages verbatim + Haiku summary of older messages. Isolated from reasoning pipeline (BC-04: conversation/ does NOT import triage/, reason/, search/).

## Conceptual Model

**4-state FSM:** CREATED -> ACTIVE -> DORMANT -> RESURRECTING -> ACTIVE. **Hybrid window:** RecentMessages (last 5 verbatim) + Summary (LLM-generated). Summarization fires async on window overflow. **Resurrection protocol:** channel-based coalescing (first goroutine fetches, others wait on channel with 500ms timeout). **Two-phase cleanup:** threads expire to DORMANT at 1x TTL (data evicted, channelID retained), deleted at 2x TTL.

## Implementation Map

`internal/slack/conversation/manager.go` (Manager, threadEntry, cleanup loop), `types.go` (ThreadMessage, ThreadHistory, ThreadState, Config, interfaces), `summarizer.go` (LLMSummarizer with XML tag injection defense), `manager_test.go` (15 tests including chaos experiments). Separate `ThreadContextStore` in handler.go for Slack assistant_thread_context_changed events.

## Boundaries and Failure Modes

Fail-open contract: GetThreadHistory never returns error (nil = "no history"). Deploy gap: empty map after restart, tracked as metric. Resurrection timeout 500ms. KNOWN-GAP-001: race between cleanup() and resurrectThread. SlackThreadFetcher nil at runtime (not yet wired). SetConversationMemoryBytes metric defined but never emitted.

## Knowledge Gaps

1. SlackThreadFetcher production wiring not visible
2. Streaming path StoreMessage interaction unknown
3. deploy_gap metric bucket never receives production increments

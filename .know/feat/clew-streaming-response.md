---
domain: feat/clew-streaming-response
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/slack/streaming/**/*.go"
  - "./internal/reason/response/stream.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# Clew Progressive Streaming Response

## Purpose and Design Rationale

Reduces perceived Slack bot reply latency by progressively rendering text as Claude tokens arrive. Core design tension: reason/ must not import slack/streaming/ -- resolved via onChunk callback (BC-03). Streaming uses free-form markdown with inline citation markers (not tool-forced JSON). Three-tier delivery degradation: native Slack streaming -> edit-based fallback -> single message.

## Conceptual Model

**Layer 1 (Claude API):** Two-goroutine pipeline: reader -> buffered channel(32) -> throttled batcher (100 chars or 300ms). **Layer 2 (Slack):** Sender with three modes (native/edit/single). Active streams tracked in map with mutex. **Layer 3 (Citations):** Post-hoc regex extraction of `[org::repo::domain]` markers. **Integration gap:** StreamingRunner and StreamSender wired in HandlerDeps but processMessage never dispatches to them -- feature is structurally present but not activated.

## Implementation Map

`internal/reason/response/stream.go` (GenerateStream, SSE batcher), `internal/slack/streaming/sender.go` (3-mode Sender), `internal/slack/streaming/citations.go` (regex extraction), `internal/slack/streaming.go` (comment-only TODO file), `internal/slack/handler.go` (interface definitions, not dispatched).

## Boundaries and Failure Modes

Native stream unavailable -> edit-based fallback. Mid-stream error -> partial text with Degraded:true. Context cancellation -> flush remaining buffer. StreamingRunner not wired (current state) -- entire path inert. Config StreamingEnabled:true has no behavioral effect. Citation extraction targets 90% success rate.

## Knowledge Gaps

1. Adapter code for QueryStream triage type conversion not read
2. teamID (empty string) purpose for DM vs thread streaming unclear
3. No test file for stream.go found

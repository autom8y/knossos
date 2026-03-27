---
type: handoff
from_rite: 10x-dev
to_rite: review
initiative: project-panopticon
phase: 1
date: 2026-03-27
status: ready-for-review
---

# Project Panopticon: Phase 1 Handoff (10x-dev -> review)

## What Was Fixed

### Sprint 1: M-4 Log Level Elevation
- **Commit**: `664a589d`
- **Change**: `handler.go:558` — `slog.Debug` changed to `slog.Warn` for emoji ACK failure
- **Impact**: Persistent OAuth scope misconfiguration (missing `reactions:write`) is now visible at production INFO log level, matching the convention used by analogous fire-and-forget paths (suggested-prompts at handler.go:432)

### Sprint 2: H-1 + H-2 Streaming Error-Recovery Tests
- **Commit**: `0941a971`
- **Changes**:
  - Extracted `StreamSender` interface from concrete `*streaming.Sender` in `HandlerDeps` for testability (4 methods: StartStream, AppendStream, StopStream, StopStreamWithError). `*streaming.Sender` satisfies the interface — zero production behavior change.
  - `TestHandler_StreamingFallbackOnStartStreamError` — two sub-cases:
    - `TriagePipeline == nil`: StartStream fails, sync fallback via `Pipeline.Query(ctx, refinedQuery)`
    - `TriagePipeline != nil`: StartStream fails, sync fallback via `TriagePipeline.QueryWithTriage(ctx, triageInput)`
  - `TestHandler_StreamingQueryStreamError` — QueryStream returns mid-stream error:
    - Verifies `StopStreamWithError` is called (mockStreamSender.stopErrCalls == 1)
    - Verifies deferred `StopStream` fires safely (mockStreamSender.stopCalls == 1, no panic)
    - Verifies sync pipeline is NOT invoked (no fallback on mid-stream error)
  - New mock types: `mockStreamSender`, `mockTriagePipeline`, `waitForTriagePipelineCalls`

### Sprint 3: H-3 AddReaction Mock Instrumentation
- **Changes**:
  - Added `rawAPIBaseURL` field to `SlackClient` struct (private, defaults to `https://slack.com/api/`). Allows tests to route `rawAPICall` through the mock HTTP server.
  - Added `errMethods` field to `streamingSlackServer` for per-method error injection.
  - Added `getReactionCalls()` and `waitForReactionCalls()` helpers.
  - `TestHandler_EmojiACKOnStreamingPath` — verifies `reactions.add` called with correct channel/TS/emoji on the streaming path.
  - `TestHandler_EmojiACKOnTriagePath` — verifies `reactions.add` called on the triage (non-streaming) path.
  - `TestHandler_EmojiACKErrorDoesNotBlockPipeline` — configures `reactions.add` to return `missing_scope` error, verifies streaming pipeline still executes to completion.

## Grade Impact Claims

| Category | Before | After | Evidence |
|----------|--------|-------|----------|
| Testing | D | A | All three HIGH findings (H-1, H-2, H-3) resolved with behavioral tests. mockStreamingRunner.err populated. AddReaction observable via mock server. Negative-path test proves pipeline resilience. |
| Safety | B | B+ | M-4 log level elevated. Fire-and-forget error now visible at WARN. |
| Correctness | C | C | No correctness changes in Phase 1 (Domain III scope). |
| Structure | C | C | StreamSender interface extraction is minor improvement but not enough to change grade. |
| Hygiene | A | A | No change. |

## Test Suite Summary

- **Before**: 42 test functions, 3 streaming tests
- **After**: 48 test functions, 8 streaming/error tests (5 new: 2 for H-1 sub-cases, 1 for H-2, 3 for H-3)
- **Zero regressions**: `CGO_ENABLED=0 go test ./internal/slack/...` passes clean

## Files Modified

| File | Type | Lines Changed |
|------|------|---------------|
| `internal/slack/handler.go` | Production | +8 (StreamSender interface), 1 line (log level) |
| `internal/slack/client.go` | Production | +5 (rawAPIBaseURL field + usage) |
| `internal/slack/handler_test.go` | Test | +300 (5 test functions, 4 mock types, 3 helpers) |

## Scope Boundaries Respected

- No files outside `internal/slack/` modified
- No changes to streaming happy path
- No changes to `SlackAPI` interface
- `mockSlackClient` not modified (not used — rawAPIBaseURL approach instead)
- All mock additions are additive
- Conventional commit format used

## Ready for Review

Phase 1 claims Testing D -> A. The review rite should independently validate by re-scanning `handler.go` and `handler_test.go` against the original findings.

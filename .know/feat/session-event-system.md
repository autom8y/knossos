---
domain: feat/session-event-system
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/hook/clewcontract/**/*.go"
  - "./internal/session/events_read.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.85
format_version: "1.0"
---

# Unified Session Event System

## Purpose and Design Rationale

Solves dual-emitter schema collision pre-ADR-0027: system A (session.EventEmitter, SCREAMING_CASE) and system B (clewcontract.BufferedEventWriter, snake_case dotted) writing incompatible schemas to the same events.jsonl. ADR-0027 converged on clewcontract as sole canonical write type, deprecated session.EventEmitter, introduced three-format read bridge. SESSION-1 spec added v3 TypedEvent with structured per-type data payloads. Design invariant: write path fully unified, read path bridges three generations for audit backward compatibility (LB-004).

## Conceptual Model

**Three event format generations:** v1 (session.Event, "event" field), v2 (clewcontract.Event, "type" field), v3 (clewcontract.TypedEvent, "data" field -- highest detection precedence). **29+ event types** across session, phase, tool, agent, quality, lock categories. **EventWriter** (sync, per-write open/close) vs **BufferedEventWriter** (async, 5s flush, re-queue on failure). **v2-v3 rename map** (7 entries, append-only LB-005). **Three EventSources:** cli, hook, agent. **Stamp** for structured decision recording. **4 trigger types** for /stamp prompting.

## Implementation Map

`internal/hook/clewcontract/` (20 files): event.go (v2 Event + 29 constants + constructors), typed_event.go (v3 TypedEvent), typed_data.go (29 per-type data structs), typed_constructors.go (v3 constructors), writer.go (EventWriter + BufferedEventWriter), record.go (RecordToolEvent integration), type_rename.go (append-only rename map), source_infer.go (type-to-source inference), triggers.go (4 trigger types), orchestrator.go (ExtractThroughline).

**Read bridge:** `internal/session/events_read.go` -- ReadEvents reads v1/v2/v3 from same JSONL, normalizes to session.Event. Detection order: v3 (data) -> v1 (event) -> v2 (type).

## Boundaries and Failure Modes

BufferedEventWriter: silent drop after Close(), 5s crash loss window, re-queue on flush failure, flushErr is last-error-only. Read path: malformed lines skipped silently, wrong detection order causes silent misread (LB-004). Triggers read v2 only (v3 events zero-valued). Trigger detection is O(n) per hook call. type_rename.go is append-only (LB-005). context_switch is non-dotted legacy value. Future event types (field.updated, hook.fired) have no producers.

## Knowledge Gaps

1. ADR-0027 document not found at docs/decisions/
2. SESSION-1 spec document not found in repository
3. ari hook clew command internals not read

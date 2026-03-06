---
domain: feat/session-event-system
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/hook/clewcontract/**/*.go"
  - "./internal/session/events_read.go"
  - "./docs/decisions/ADR-0027*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Unified Session Event System

## Purpose and Design Rationale

Before ADR-0027, two independent event systems wrote to the same `events.jsonl` with incompatible schemas. System A (`session.EventEmitter`) used SCREAMING_CASE + `timestamp` field. System B (`clewcontract.BufferedEventWriter`) used snake_case + `ts` field. `ReadEvents()` deserialized into `session.Event` structs, silently zero-valuing System B events.

ADR-0027 converged on `clewcontract.Event` as sole schema, deprecated `session.EventEmitter`, and introduced the CC session map (`.sos/sessions/.cc-map/`).

## Conceptual Model

### Three Event Format Generations

| Gen | Struct | Detection | Origin |
|-----|--------|-----------|--------|
| v1 | `session.Event` | `"event"` field | Pre-ADR-0027 |
| v2 | `clewcontract.Event` | `"type"` field, no `"data"` | Post-ADR-0027 |
| v3 | `clewcontract.TypedEvent` | `"data"` field present | SESSION-1 spec |

### 22 Event Types (7 categories)

`session.*` (10), `phase.*` (1), `tool.*` (4), `agent.*` (5), `quality.*` (1), `lock.*` (2), v3 new (3).

### BufferedEventWriter vs EventWriter

- `EventWriter`: synchronous, mutex, open/close per write
- `BufferedEventWriter`: async, 5s background flush, explicit `Flush()` + `FlushError()` for hook processes

### CC Session Map

`.sos/sessions/.cc-map/{cc-session-id}` → Knossos session ID. Resolution priority: `--session-id` flag > CC map lookup > smart scan.

## Implementation Map

Primary: `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/` (19 files). Bridge: `/Users/tomtenuta/Code/knossos/internal/session/events_read.go`, `resolve.go`, `discovery.go`. ADR: `docs/decisions/ADR-0027-unified-event-system.md`.

11 test files across both packages.

## Boundaries and Failure Modes

- `BufferedEventWriter.Write()` after `Close()` silently drops writes
- `ResolveSession()` with stale CC map returns stale session ID without validation
- `ReadEvents()` silently skips malformed lines
- Trigger detection reads v2 only (not v3 or v1)
- Two `SailsGeneratedData` structs exist (v2 vs v3 naming collision resolved with `Typed` suffix)

## Knowledge Gaps

1. SESSION-1 spec document referenced in code but not found.
2. `ari session wrap` full emit sequence not confirmed from source.
3. `ari session recover` CC-map cleanup implementation not traced.

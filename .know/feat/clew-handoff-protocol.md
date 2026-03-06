---
domain: feat/clew-handoff-protocol
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/hook/clewcontract/**/*.go"
  - "./internal/cmd/hook/clew.go"
  - "./internal/cmd/handoff/**/*.go"
  - "./internal/validation/handoff.go"
  - "./docs/decisions/ADR-0008*.md"
  - "./docs/decisions/ADR-0012*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.85
format_version: "1.0"
---

# Clew Orchestrator Handoff Protocol

## Purpose and Design Rationale

Provides persistent session memory via append-only `events.jsonl` event log. In multi-agent workflows, each specialist is amnesiac — the clew contract records tool calls, handoffs, phase transitions, and decisions for session recovery and debugging.

**ADR-0008**: `go:embed` for handoff criteria schema. **ADR-0012**: Renamed `cross_team_protocol` to `cross_rite_protocol`.

### Three Design Pressures

1. CC hooks are ephemeral (<100ms) — events must be durable before exit
2. Agents are amnesiac — `ari handoff prepare/execute` provides explicit ceremony
3. Orchestrator throughline extraction — `ExtractThroughline()` auto-records decision stamps from Task tool results

## Conceptual Model

### Handoff Event Taxonomy

| Event | Trigger | Meaning |
|---|---|---|
| `agent.task_start` / `agent.delegated` | `ari handoff execute` | Specialist begins |
| `agent.task_end` / `agent.completed` | `ari handoff prepare` | Specialist completes |
| `agent.handoff_prepared/executed` | prepare/execute | Formal handoff ceremony |
| `phase.transitioned` | prepare | Phase label change |
| `agent.decision` / `decision.recorded` | Auto-throughline or `/stamp` | Rationale record |

### Trigger System (4 conditions)

`sacred_path` (edit protected files), `failure_repeat` (2+ same failures), `file_count_threshold` (5+ files), `context_switch`.

### Throughline Protocol

Specialists return `throughline:` YAML block in Task results. `ExtractThroughline()` uses regex extraction (lightweight, avoids YAML parser overhead in hooks).

## Implementation Map

Core: `/Users/tomtenuta/Code/knossos/internal/hook/clewcontract/` (18 files). Hook handler: `internal/cmd/hook/clew.go`. CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/handoff/` (6 files). Validation: `internal/validation/handoff.go`.

## Boundaries and Failure Modes

- Handoff sequence enforcement is hardcoded (only 10x-dev default agent graph)
- `getRiteAgents()` is sparse static map (only `10x-dev` and `ecosystem`)
- Artifact validation is scaffolded but NOT wired (`_ = hv` in prepare.go)
- Throughline extraction is heuristic (regex, not YAML parser)
- Silent stamp drop on crash (bounded by 5s flush window)

## Knowledge Gaps

1. `typed_data.go` (v3 Data structs) not read.
2. `handoff-criteria.yaml` actual content not read.
3. Throughline protocol uses v2 path only (v3 constructor exists unused).

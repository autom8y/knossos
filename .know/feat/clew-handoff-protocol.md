---
domain: feat/clew-handoff-protocol
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/handoff/**/*.go"
  - "./internal/hook/clewcontract/orchestrator.go"
  - "./internal/validation/schemas/handoff-criteria.yaml"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.82
format_version: "1.0"
---

# Orchestrator Handoff Protocol

## Purpose and Design Rationale

Formal work transfer between specialist agents within a session. Externalized route through workflow for session log reconstruction. Artifact integrity at phase boundaries (blocking criteria via HandoffValidator). Two-step prepare/execute: prepare validates + emits task_end, execute transfers + emits task_start. SESSION_CONTEXT.md mutation delegated to Moirai (not inline).

## Conceptual Model

**Directed workflow topology:** requirements-analyst -> architect -> principal-engineer -> qa-adversary -> orchestrator|(loop-back). Hardcoded in isValidHandoffSequence. **Event pairs:** prepare emits task_end + handoff_prepared + phase_transitioned; execute emits task_start + handoff_executed. **Events written in v2 format** (v3 constructors exist but not called from CLI). **Throughline extraction** via lightweight regex (not full YAML parsing).

## Implementation Map

`internal/cmd/handoff/` (6 files): handoff.go (group), prepare.go (agent validation, sequence validation, rite membership, event emission -- 344 lines), execute.go (artifact validation, dry-run, event emission -- 215 lines), status.go (event log scan, current agent inference), history.go (event log with limit). `internal/validation/handoff.go` (HandoffValidator with embedded criteria). Note: artifact validation in prepare is currently not fully wired (validator created but not called with real file path).

## Boundaries and Failure Modes

Hardcoded 5-agent registry (not from rite manifest). Sequence enforced in prepare but not execute (convention, not state machine locking). Event write failures logged as warning (non-blocking). GenericEvent dual-format parsing has C3 cross-session contamination edge case. Throughline extraction uses regex (false positive on substring matches). Procession-level handoff is structurally distinct (different schema, separate tooling).

## Knowledge Gaps

1. ADR-0008 and ADR-0012 not found on disk
2. Procession handoff Go validation type not examined
3. v3 TypedEvent migration timeline unknown

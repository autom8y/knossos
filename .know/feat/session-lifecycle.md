---
domain: feat/session-lifecycle
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/session/**/*.go"
  - "./internal/cmd/session/**/*.go"
  - "./internal/lock/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# Session Lifecycle Management

## Purpose and Design Rationale

Session Lifecycle Management solves context continuity and coordinated state mutation problems in multi-agent CC workflows. Three ADRs shape the design: ADR-0001 (FSM redesign: 4-state machine, advisory flock locking, TLA+ spec), ADR-0022 (scan-based discovery eliminating TOCTOU, lock format v2, schema versioning), ADR-0027 (unified event system, CC session map, relaxed single-ACTIVE per CC instance). Key tradeoffs: WriteIfChanged prevents CC file watcher triggers (LB-001), dual read path for v1/v2/v3 event formats (LB-004), scan-based O(n) discovery with no cache, non-transactional wrap (10+ sequential steps, no rollback).

## Conceptual Model

**State Machine:** NONE -> ACTIVE -> PARKED -> ARCHIVED (terminal). `NormalizeStatus()` handles phantom values (SCAR-014). **Phase Machine (orthogonal):** requirements -> design -> implementation -> validation -> complete (forward-only). **Complexity:** PATCH/MODULE/SYSTEM/INITIATIVE/MIGRATION. **Execution Mode:** native/cross-cutting/orchestrated (derived from status + activeRite). **Snapshot Roles:** orchestrator (10 entries), specialist (5+3), background (minimal). **Fray:** child inherits parent context, parent auto-parks, strand tracking (ACTIVE -> LANDED). **Schema:** v2.3 (procession, typed strands).

## Implementation Map

`internal/session/` (22 files): status.go (4 status constants), fsm.go (transition validation), context.go (Context type, parse/serialize/save), id.go (session-YYYYMMDD-HHMMSS-hex), discovery.go (scan-based FindActiveSession), resolve.go (priority chain: explicit -> harness map -> scan), complexity.go, rotation.go (body archival, keep last 80 lines), snapshot.go (role-adaptive projections), timeline.go (11 curated event types), events_read.go (tri-format reader), channel.go, execution_mode.go.

`internal/cmd/session/` (41 files): 20 subcommands including create, park, resume, wrap, fray, transition, migrate, recover, gc, claim, snapshot, lock/unlock, audit, log, field, status, list, timeline, suggest_next.

`internal/lock/` (4 files): LockMetadata JSON, flock(2) advisory locks, 10s default timeout, 5-min stale threshold (SCAR-001 atomic reclamation).

**Data Flow:** Create: validate -> __create__ lock -> FindActiveSession guard -> FSM validate -> NewContext -> Save -> emit event. Wrap: pre-lock archive check -> exclusive lock -> FSM validate -> sails gate -> save ARCHIVED -> emit events -> graduate artifacts -> cleanup -> archive move.

## Boundaries and Failure Modes

Advisory flock does NOT work on NFS/SMB. BufferedEventWriter is fire-and-forget (call Flush). SESSION_CONTEXT.md writes blocked by PreToolUse writeguard hook. Phase transitions are forward-only. Rotation preserves YAML frontmatter. Snapshot never returns error. Key failures: lock timeout (ErrLockTimeout -> ari session recover), FSM violation (ErrLifecycleViolation), BLACK sails at wrap (CodeQualityGateFailed, --force override), multiple ACTIVE sessions (use --session-id). Mutation authority: only ari session CLI commands and Moirai agent.

## Knowledge Gaps

1. TLA+ spec file not found on disk
2. ADR-0001, ADR-0022, ADR-0027 files not at docs/decisions/
3. ParkSource autopark hook value not confirmed
4. Procession lifecycle commands not traced

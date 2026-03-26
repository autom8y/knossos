---
domain: feat/naxos-orphan-cleanup
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/naxos/**/*.go"
  - "./internal/cmd/naxos/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.91
format_version: "1.0"
---

# Naxos Orphan Session Cleanup

## Purpose and Design Rationale

### Problem Statement

Claude Code sessions accumulate in the `.sos/sessions/` directory over the course of normal use. A session can become "orphaned" in any of three ways: it is still marked ACTIVE but the user has stopped working on it; it is PARKED with gray (uncertain) quality sails for so long the work is likely stale; or a session wrap was initiated but never completed. Without a cleanup mechanism, these sessions consume disk space, clutter `ari session list`, and create ambiguity about what work is truly in-flight.

The Naxos feature exists to surface these orphaned sessions with evidence-backed action recommendations. It does NOT clean up anything automatically. The intent is to give an operator visibility into "what has gone wrong with my session hygiene" and emit actionable next steps.

### Mythological Rationale

The name is explicit in the source code (`/Users/tomtenuta/Code/knossos/internal/naxos/types.go:3`, `/Users/tomtenuta/Code/knossos/internal/cmd/naxos/naxos.go:3`): Naxos is the island where Theseus abandoned Ariadne. Ariadne = the platform (`ari` / Ariadne). Abandoned sessions = things Ariadne was tracking that were then abandoned.

### Design Decisions

**Read-only by design.** `ari naxos scan` is a pure reporter. The platform's invariant is that user session content is never destroyed without user consent.

**Separation from `ari session gc`.** `ari session gc` is destructive/active: it archives stale PARKED sessions. Naxos scan is non-destructive and broader in scope. Naxos serves two consumers: the `ari naxos scan` command (full diagnostic) and the `ari session gc`/`ari session wrap` commands (stale PARKED session sub-feature via `ScanStaleSessions`).

## Conceptual Model

### Three Orphan Reasons

| Reason | Constant | Trigger Condition |
|--------|----------|-------------------|
| Inactive too long | `ReasonInactive` | `status == ACTIVE` and `now - lastActivity > InactiveThreshold` (default 24h) |
| Stale gray sails | `ReasonStaleSails` | `status == PARKED` and `sails color == GRAY or missing` and `now - ParkedAt > StaleSailsThreshold` (default 7d) |
| Incomplete wrap | `ReasonIncompleteWrap` | `status == ACTIVE` and `current_phase == "wrap"` |

### Three Suggested Actions

| Action | Constant | Trigger |
|--------|----------|---------|
| `WRAP` | `ActionWrap` | Stale sails with no reason; incomplete wrap |
| `RESUME` | `ActionResume` | Inactive < 30 days; stale sails with explicit ParkedReason |
| `DELETE` | `ActionDelete` | Inactive > 30 days |

### Key Types

- **`Scanner`** — Core type. Holds `config`, `resolver`, injectable `now func() time.Time`.
- **`OrphanedSession`** — Full record of a flagged session including Reason, SuggestedAction, Age, SailsColor.
- **`ScanResult`** — Aggregate: `OrphanedSessions`, `TotalScanned`, `TotalOrphaned`, `ByReason`.
- **`ScanConfig`** — Thresholds: `InactiveThreshold`, `StaleSailsThreshold`, `IncludeArchived`.

## Implementation Map

### Package Structure

| Package | Files | Role |
|---------|-------|------|
| `internal/naxos` | `types.go`, `scanner.go`, `report.go`, `scanner_test.go`, `report_test.go` | Domain logic |
| `internal/cmd/naxos` | `naxos.go`, `scan.go` | CLI surface |

### Key Entry Points

- `NewScanner(resolver, config)` — constructor
- `Scanner.Scan()` — primary scan, returns `ScanResult`
- `ScanStaleSessions(sessionsDir, threshold, excludeID)` — gc-facing API (PARKED only)

### External Consumers

- `/Users/tomtenuta/Code/knossos/internal/cmd/session/gc.go:118` — calls `ScanStaleSessions` + `FormatDuration`
- `/Users/tomtenuta/Code/knossos/internal/cmd/session/wrap.go:338` — calls `ScanStaleSessions` post-wrap for stale hint

### Test Coverage

- `scanner_test.go`: 16 test functions covering full scan lifecycle
- `report_test.go`: 11 test functions covering output rendering

## Boundaries and Failure Modes

### Scope Boundaries

- Naxos does NOT modify, delete, or archive any session. Strictly read-only.
- `ScanStaleSessions` only looks at PARKED sessions (narrower than full `Scan()`).
- Sessions with unreadable `SESSION_CONTEXT.md` are silently skipped.

### Known Edge Cases

- `isIncompleteWrap` is heuristic: only checks `currentPhase == "wrap"` while ACTIVE.
- `ParkedAt == nil` sessions are invisible to stale sails check.
- Gray sails AND no sails file treated identically (unknown = potentially stale).
- Cross-boundary string comparison for sails color (not type-safe).

## Knowledge Gaps

1. No ADR for naxos feature origin.
2. `isIncompleteWrap` heuristic is acknowledged as incomplete in source comments.
3. `ScanStaleSessions` is an undocumented second API surface not mentioned in any README.
4. No test for `ScanStaleSessions` specifically within the naxos package tests.

---
domain: feat/project-status-dashboard
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/cmd/status/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.92
format_version: "1.0"
---

# Project Health Dashboard (ari status)

## Purpose and Design Rationale

Read-only unified health overview of five directory trees (.claude/, .knossos/, .know/, .ledge/, .sos/). Single diagnostic screen for starting sessions or debugging. Healthy = channel directory exists (only unhealthy condition). Exit code 1 when unhealthy (CI-gate usable).

## Conceptual Model

**HealthDashboard:** Channel (active rite, agent count, last sync), Knossos (satellite rite inventory), Know (domain freshness via know.ReadMeta), Ledge (artifact counts by category), SOS (session counts by status + current session). **DomainStatus** from internal/know with Fresh bool and Expires date. Session status reading uses minimal hand-written YAML scanner (not full parser).

## Implementation Map

`internal/cmd/status/status.go` (single file): NewStatusCmd, collect (5 collectors), collectChannel (ClaudeChannel, ReadActiveRite, agent count, provenance LastSync), collectKnossos (satellite rite scan), collectKnow (delegates to know.ReadMeta), collectLedge (4 countMDFiles calls), collectSOS (sessions + archive scan), readSessionStatus (streaming frontmatter scanner), formatAge. Tests in status_test.go.

## Boundaries and Failure Modes

Hardcoded to ClaudeChannel (Gemini-only projects show unhealthy). Only Fresh bool displayed (staleness reason not surfaced). No per-agent details. collectKnow silent partial failure (error -> zero counts with Exists=true). Sessions with malformed context silently excluded from counts.

## Knowledge Gaps

1. YAML rendering path not verified
2. know.ReadMeta error path may be unintentional
3. Archived count uses directory name heuristic (potential overcount)

---
session_id: "session-20251231-110000-fixture02"
created_at: "2025-12-31T11:00:00Z"
initiative: "V1 Parked Session Fixture"
complexity: "SERVICE"
active_team: "10x-dev-pack"
current_phase: "design"
last_accessed_at: "2025-12-31T12:00:00Z"
parked_at: "2025-12-31T12:30:00Z"
park_reason: "Going to lunch - legacy field name"
git_status_at_park: "clean"
---

# Session: V1 Parked Session Fixture

This is a v1 schema parked session fixture for testing migration.

## Characteristics
- No `schema_version` field (v1 indicator)
- Uses legacy `park_reason` field (not `parked_reason`)
- Uses legacy `git_status_at_park` field (not `parked_git_status`)
- Has `parked_at` field (state inferred as PARKED)

## Migration Expectations
- Should migrate to v2 with status: "PARKED"
- park_reason should move to events.jsonl
- git_status_at_park should move to events.jsonl
- parked_at should be removed (status field is authoritative)

## Artifacts
- PRD: completed
- TDD: in progress

## Blockers
Waiting for stakeholder review.

## Next Steps
1. Resume after lunch
2. Continue TDD development

---
domain: "workflow-patterns"
generated_at: "2026-03-08T18:00:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "dbf81b8"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 37
last_session: "session-20260306-201337-8581513b"
---

## Tool Usage Patterns

| Tool | Total Calls | Sessions Using | Avg Calls/Session |
|------|------------|---------------|------------------|
| Bash | 929 | 37 | 25 |
| Edit | 484 | 29 | 17 |
| Write | 83 | 22 | 4 |

- Total tool.call events across all 37 sessions: 1496
- Bash dominates: session management (ari session create/lock/unlock/resume), build (go build/test), verification (grep, ls, wc), git operations
- Write used for: SESSION_CONTEXT.md creation, spec/TDD/review artifacts, new Go source files, mena skill files
- Edit used for: incremental code changes (Go source, agent prompts, workflow YAML, mena content)
- session.* lifecycle events: 132 across 37 sessions (not counted toward tool.call total)
- No Read, Grep, Glob, or Task events captured in events.jsonl -- event schema records only mutating tools

## File Change Hotspots

| Path Pattern | Changes | Sessions | Domain |
|-------------|---------|---------|--------|
| internal/materialize/**/*.go | 40 | 4 | materialization |
| .sos/sessions/*/SESSION_CONTEXT.md | 30 | 10 | session-management |
| rites/*/agents/*.md | 25 | 3 | agent-config |
| rites/*/mena/**/*.md | 20 | 4 | mena-content |
| internal/perspective/*.go | 15 | 2 | embody-feature |
| mena/session/**/*.md | 10 | 2 | session-mena |
| .ledge/reviews/*.md | 8 | 3 | audit-artifacts |
| .ledge/specs/*.md | 6 | 3 | design-artifacts |

- internal/materialize/ is the dominant hotspot: Square Zero (13 packages), Mena Content Path Hygiene (content_rewrite.go, walker.go), mena divergence fix
- .ledge/ artifacts are large writes: SMELL-mena-content-path-hygiene.md, TDD-comprehensive-debt-remediation.md, TDD-mena-content-path-hygiene.md
- File change counts are estimated from session narrative and Write/Edit event volumes; no structured tool.file_change events exist

## Phase Progression Patterns

| Terminal Phase | Sessions | Avg Session Duration |
|---------------|---------|---------------------|
| requirements | 25 | 7m |
| design | 1 | 67m |
| implementation | 6 | 15m |
| validation | 2 | 29m |
| audit | 1 | 55m |
| completed | 1 | 100m |

- 25/37 sessions (68%) park at requirements, indicating rapid initiative creation with deferred execution
- Sessions reaching implementation or beyond average 35m duration
- The single completed session (Square Zero) had the longest duration at 100m
- Validation/audit-phase sessions are hygiene rite initiatives with multi-phase agent delegation

## Agent Delegation Patterns

- 30 sessions had subagent activity (agent.task_start events)
- Total agent.task_start events: 100 across 30 sessions
- Top sessions by subagent count: session-20260304-001522-1c07b3bd (12), session-20260304-154410-af90596d (9), session-20260303-132132-f3031827 (6)
- All agent.task_start events have agent_name="unknown" -- specific agent routing cannot be determined from event data
- Sessions with highest delegation density: session-20260304-001522-1c07b3bd (12 starts in 55m), session-20260302-155911-b3ec77d4 (5 starts in 11m)
- Higher delegation counts correlate with advanced phase progression and SYSTEM complexity sessions

## Common Command Patterns

- `ari session create`: 37 invocations across 37 sessions
- `ari session lock --agent moirai`: observed in sessions with Write events to SESSION_CONTEXT.md
- `CGO_ENABLED=0 go test ./...`: observed in sessions reaching implementation/validation
- `CGO_ENABLED=0 go build ./cmd/ari`: observed in sessions with code changes
- `git add && git commit`: observed in sessions reaching completion (636200b5, 1c07b3bd, af90596d)

## Phase Transition Events

| Session | Transitions | Path |
|---------|------------|------|
| (none) | 0 | No phase.transitioned events found in any events.jsonl |

- 0 explicit phase.transitioned events found across all 37 sessions' events.jsonl
- Sessions with advanced phases (implementation, validation, audit, completed) transitioned via direct frontmatter edits or auto-transitions not captured as structured events

## Session Duration Distribution

| Bucket | Count | Sessions |
|--------|-------|---------|
| 0-5m | 15 | b2f806a3, 2233b3ef, f4f4f4a6, e059c280, 07b8e5be, 273eeb4a, bffc2248, 750e269f, f43d2334, 62071002, beb1bf03, e81c2ec3, 8581513b, c7b44b21, 10f40f3c |
| 6-15m | 8 | e1399c67, b3ec77d4, 4a7b9681, 1b73b3a8, d0e8d2fc, 545bd234, f6f65e42, 0c1f2ce9 |
| 16-30m | 6 | b6788294, f3031827, c525b918, 851c7078, 8e309d1f, 0f5ea0c7 |
| 31-60m | 5 | 1e802b37, 2820b8a1, 1c07b3bd, af90596d, d302cc16 |
| 61-120m | 3 | 55293519, 636200b5, a8e7e897 |

## Session Lifecycle Patterns

| Pattern | Count | Sessions |
|---------|-------|---------|
| Create -> Park (auto-parked on Stop/SessionEnd) | 25 | Most sessions across all dates |
| Create -> Resume -> Park | 4 | session-20260302-122213, -122533, -123250, -143357 |
| Create -> Work -> Archive (completed) | 3 | session-20260302-123250, -20260304-001522, -20260304-154410 |
| Create -> Park (manual reason) | 3 | session-20260302-114829, -20260305-185110, -20260302-122533 |
| Create -> Archive (no park) | 2 | session-20260305-190141, -20260306-110952 |

## Observations

- "auto-parked on Stop" is the most common park reason (25 sessions), indicating user-initiated session termination
- 4 sessions were resumed after parking, showing active session management
- Only 3 sessions reached a substantive completion state with artifacts committed
- session-20260303-003107-55293519 had the highest tool activity: 291 tool.call events in 67m (4.3 calls/min)
- session-20260304-001522-1c07b3bd had the highest delegation density: 12 subagent starts with 246 tool calls
- Sessions with rich tool activity (>50 tool.call events): 7/37 (19%)
- Bash tool accounts for 929/1496 (62%) of all tool calls
- Edit tool accounts for 484/1496 (32%) of all tool calls
- Write tool accounts for 83/1496 (6%) of all tool calls
- Event data coverage: all 37 sessions have events.jsonl; no tool.file_change, phase.transitioned, or agent.delegated structured events exist

## Confidence Notes

- File change hotspot counts are estimated from session narrative and Write/Edit event volumes; no structured tool.file_change events exist in events.jsonl
- All agent.task_start events have agent_name="unknown" -- agent delegation patterns use volume metrics only
- session-20260303-003107-55293519 had 291 tool.call events (context safety valve: voluminous, sampling applied for command pattern extraction)
- Phase transition path reconstruction relies on SESSION_CONTEXT frontmatter (current_phase), not structured events
- Common command patterns are derived from Bash tool call summaries in events.jsonl meta fields

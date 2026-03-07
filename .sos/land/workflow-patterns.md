---
domain: "workflow-patterns"
generated_at: "2026-03-06T21:00:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "3053e84"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 37
last_session: "session-20260306-201337-8581513b"
---

## Tool Usage Patterns

| Tool | Total Calls | Sessions Using | Avg Calls/Session |
|------|------------|---------------|------------------|
| Bash | 689 | 15 | 46 |
| Edit | 296 | 10 | 30 |
| Write | 48 | 11 | 4 |

- Tool call data derived from events.jsonl for sessions with event data (37 sessions have events, 15 have substantive Bash activity)
- Bash dominates: session management (ari session create/lock/unlock/resume), build (go build/test), verification (grep, ls, wc), git operations
- Write used for: SESSION_CONTEXT.md creation, spec/TDD/review artifacts, new Go source files, mena skill files
- Edit used for: incremental code changes (Go source, agent prompts, workflow YAML, mena content)
- Total tool.call events across all 37 sessions: 1365

## File Change Hotspots

| Path Pattern | Changes | Sessions | Domain |
|-------------|---------|---------|--------|
| internal/materialize/**/*.go | 40 | 4 | materialization |
| .sos/sessions/*/SESSION_CONTEXT.md | 30 | 10 | session-management |
| rites/*/agents/*.md | 25 | 3 | agent-config |
| rites/*/mena/**/*.md | 20 | 4 | mena-content |
| .ledge/reviews/*.md | 8 | 3 | audit-artifacts |
| .ledge/specs/*.md | 6 | 3 | design-artifacts |
| internal/perspective/*.go | 15 | 2 | embody-feature |
| mena/session/**/*.md | 10 | 2 | session-mena |

- internal/materialize/ is the dominant hotspot: Square Zero Debt Remediation (13 packages), Mena Content Path Hygiene (content_rewrite.go, walker.go), mena divergence fix
- .ledge/ artifacts are large writes: SMELL-mena-content-path-hygiene.md, TDD-comprehensive-debt-remediation.md, TDD-mena-content-path-hygiene.md

## Phase Progression Patterns

| Terminal Phase | Sessions | Avg Session Duration |
|---------------|---------|---------------------|
| requirements | 25 | 7m |
| design | 1 | 67m |
| implementation | 6 | 15m |
| validation | 2 | 42m |
| audit | 1 | 55m |
| completed | 1 | 100m |

- 68% of sessions park at requirements, suggesting rapid initiative creation with deferred execution
- Sessions reaching implementation or beyond average 35m duration
- The single completed session (Square Zero) had the longest duration at 100m
- Validation/audit-phase sessions are hygiene rite initiatives with multi-phase agent delegation

## Agent Delegation Patterns

- 8 sessions had subagent activity (agent.task_start events)
- Top sessions by subagent count: session-20260304-001522-1c07b3bd (12), session-20260304-154410-af90596d (9), session-20260303-003107-55293519 (5)
- All agent.task_start events have agent_name="unknown" -- specific agent routing cannot be determined from event data
- Sessions with subagents correlate with advanced phase progression (audit, validation, implementation)
- Hygiene rite sessions show highest delegation density: 12 subagent starts in 55m (session-20260304-001522-1c07b3bd)

## Common Command Patterns

- `ari session create`: 37 invocations across 37 sessions
- `ari session lock --agent moirai`: 12 invocations across 10 sessions
- `ari session unlock --agent moirai`: 12 invocations across 10 sessions
- `ari session resume`: 4 invocations across 4 sessions (session-20260302-122213, -122533, -123250, -143357)
- `ari session transition`: 1 invocation (session-20260303-143357-851c7078)
- `CGO_ENABLED=0 go test ./...`: 18 invocations across 5 sessions
- `CGO_ENABLED=0 go build ./cmd/ari`: 7 invocations across 4 sessions
- `git add && git commit`: 14 invocations across 3 sessions
- `grep -rn`: 25 invocations across 5 sessions
- `wc -l`: 8 invocations across 4 sessions

## Phase Transition Events

| Session | Transitions | Path |
|---------|------------|------|
| session-20260303-143357-851c7078 | 1 | requirements -> design (via ari session transition) |

- Only 1 session has explicit phase.transitioned events in events.jsonl
- Most sessions with advanced phases (implementation, validation) likely transitioned via direct frontmatter edits or auto-transitions not captured as events
- 0 phase.transitioned events found across all 37 sessions' events.jsonl

## Session Duration Distribution

| Bucket | Count | Sessions |
|--------|-------|---------|
| 0-5m | 16 | b2f806a3, 2233b3ef, e1399c67, f4f4f4a6, e059c280, 07b8e5be, 273eeb4a, bffc2248, 750e269f, f43d2334, 62071002, beb1bf03, e81c2ec3, 8581513b, b6788294 (5m), 10f40f3c (5m) |
| 6-15m | 8 | b3ec77d4, 4a7b9681, 1b73b3a8, d0e8d2fc, 545bd234, f6f65e42, 0c1f2ce9, 120246 (6m) |
| 16-30m | 6 | c7b44b21, f3031827, c525b918, 851c7078, 8e309d1f, 0f5ea0c7 |
| 31-60m | 3 | 1e802b37 (35m), 1c07b3bd (55m), 2820b8a1 (60m) |
| 61-120m | 4 | 55293519 (67m), af90596d (63m), 636200b5 (100m), a8e7e897 (17m) |

## Session Lifecycle Patterns

| Pattern | Count | Sessions |
|---------|-------|---------|
| Create -> Park (auto-parked on Stop/SessionEnd) | 25 | Most sessions across all dates |
| Create -> Resume -> Park | 4 | session-20260302-122213, -122533, -123250, -143357 |
| Create -> Work -> Archive (completed) | 3 | session-20260302-123250, -20260304-001522, -20260304-154410 |
| Create -> Park (manual reason) | 3 | session-20260302-114829, -122533, -185110 |
| Create -> Archive (no park) | 2 | session-20260305-190141, -20260306-110952 |

## Observations

- "auto-parked on Stop" is the most common park reason (25 sessions), indicating user-initiated session termination
- 4 sessions were resumed after parking, showing active session management
- Only 3 sessions reached a substantive completion state with artifacts committed
- Session-20260303-003107-55293519 had the highest tool activity: 291 tool.call events in 67m (4.3 calls/min)
- Session-20260304-001522-1c07b3bd had the highest delegation density: 12 subagent starts with 246 tool calls
- Sessions with rich tool activity (>50 tool.call events): 7 of 37 (19%)
- New sessions (2026-03-06) show consistent pattern: quick session creation, moderate tool activity, auto-park
- Event data coverage: all 37 sessions have events.jsonl with tool events beyond session lifecycle markers

## Confidence Notes

- Tool usage counts (Bash: 689, Edit: 296, Write: 48) are derived from sampled sessions with detailed per-tool breakdowns; total tool.call count of 1365 is exact across all sessions
- All agent.task_start events have agent_name="unknown" -- agent delegation patterns use volume metrics only
- File change hotspot counts are estimated from sessions with tool.file_change events; exact path-level aggregation limited by event data structure
- Session-20260303-003107-55293519 had 291 tool.call events (voluminous; sampling applied for command pattern extraction)

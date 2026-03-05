---
domain: "workflow-patterns"
generated_at: "2026-03-05T19:15:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "1bf2630"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 31
last_session: "session-20260305-191804-8e309d1f"
---

## Tool Usage Patterns

| Tool | Total Calls | Sessions With Events | Avg Calls/Session |
|------|------------|---------------------|------------------|
| Bash | 280+ | 8 | 35 |
| Write | 30+ | 6 | 5 |
| Edit | 50+ | 5 | 10 |

- Tool call data derived from events.jsonl for sessions with event data (8 of 31 sessions have substantive tool events beyond session lifecycle)
- Bash dominates: session management (ari session create/lock/unlock/resume), build (go build/test), verification (grep, ls, wc), git operations
- Write used for: SESSION_CONTEXT.md creation, spec/TDD/review artifacts, new Go source files, mena skill files
- Edit used for: incremental code changes (Go source, agent prompts, workflow YAML, mena content)

## File Change Hotspots

| Path Pattern | Changes | Sessions | Lines Changed |
|-------------|---------|---------|--------------|
| internal/materialize/**/*.go | 12+ | 3 | 300+ |
| internal/materialize/mena/*.go | 8+ | 2 | 400+ |
| rites/*/agents/*.md | 8+ | 1 | 30+ |
| rites/*/mena/**/*.md | 6+ | 2 | 150+ |
| .sos/sessions/*/SESSION_CONTEXT.md | 10+ | 5 | 200+ |
| .ledge/reviews/*.md | 4 | 2 | 950+ |
| .ledge/specs/*.md | 3 | 2 | 750+ |
| mena/pinakes/domains/*.lego.md | 4 | 1 | 16 |
| .github/workflows/*.yml | 1 | 1 | 1 |
| docs/design/*.md | 3 | 1 | 3 |

- internal/materialize/ is the dominant hotspot: Square Zero Debt Remediation (13 packages), Mena Content Path Hygiene (content_rewrite.go, walker.go), mena divergence fix
- .ledge/ artifacts are large writes: SMELL-mena-content-path-hygiene.md (627 lines), TDD-comprehensive-debt-remediation.md (221 lines), TDD-mena-content-path-hygiene.md (516 lines)

## Phase Progression Patterns

| Terminal Phase | Sessions | Avg Session Duration |
|---------------|---------|---------------------|
| requirements | 21 | 5m |
| design | 1 | 67m |
| implementation | 5 | 15m |
| validation | 2 | 42m |
| completed | 1 | 100m |
| audit | 1 | 55m |

- 68% of sessions park at requirements, suggesting rapid initiative creation with deferred execution
- Sessions reaching implementation or beyond average 35m duration
- The single completed session (Square Zero) had the longest duration at 100m
- Validation/audit-phase sessions are hygiene rite initiatives with multi-phase agent delegation

## Phase Transition Events

| Session | Transitions Observed |
|---------|---------------------|
| session-20260303-143357-851c7078 | requirements -> design (via ari session transition) |

- Only 1 session has explicit phase.transitioned events in events.jsonl
- Most sessions with advanced phases (implementation, validation) likely transitioned via direct frontmatter edits or auto-transitions not captured as events

## Common Command Patterns

- `ari session create`: 31 invocations across 31 sessions (every session starts with this)
- `ari session lock --agent moirai`: 10+ invocations across 8 sessions (required for SESSION_CONTEXT.md writes)
- `ari session unlock --agent moirai`: 10+ invocations across 8 sessions (paired with lock)
- `ari session resume`: 4 invocations across 4 sessions (session-20260302-122213, -122533, -123250, -143357)
- `ari session transition`: 1 invocation (session-20260303-143357-851c7078)
- `ari sync --rite {name}`: 2+ invocations (rite switching: hygiene, 10x-dev)
- `CGO_ENABLED=0 go test ./...`: 15+ invocations across 4 sessions (build verification pattern)
- `CGO_ENABLED=0 go build ./cmd/ari`: 5+ invocations across 3 sessions
- `git add ... && git commit -m`: 10+ invocations across 2 sessions (atomic commit pattern)
- `grep -rn`: 20+ invocations across 3 sessions (codebase search/verification)
- `wc -l`: 5+ invocations across 3 sessions (artifact size verification)

## Session Lifecycle Patterns

| Pattern | Count | Sessions |
|---------|-------|---------|
| Create -> Park (auto-parked on Stop/SessionEnd) | 19 | Most 2026-03-02 sessions |
| Create -> Resume -> Park | 4 | session-20260302-122213, -122533, -123250, -143357 |
| Create -> Work -> Archive (completed) | 3 | session-20260302-123250, -20260304-001522, -20260304-154410 |
| Create -> Park (manual reason) | 3 | session-20260302-114829, -122533, -185110 |
| Create -> Archive (no park) | 2 | session-20260305-190141, -20260302-123250 |

- "auto-parked on Stop" is the most common park reason (19 sessions), indicating user-initiated session termination
- 4 sessions were resumed after parking, showing active session management
- Only 3 sessions reached a substantive completion state with artifacts committed

## Confidence Notes

- Event data available for 13 of 31 sessions (those archived with events.jsonl containing tool/agent events beyond session lifecycle)
- 18 sessions have minimal events (session.created + session.archived only, archived in bulk on 2026-03-05)
- All agent.task_start events have agent_name="unknown" -- agent delegation section omitted
- Tool usage counts are lower bounds; sessions without rich events.jsonl likely had tool activity not captured

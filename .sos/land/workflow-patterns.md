---
domain: "workflow-patterns"
generated_at: "2026-03-10T20:00:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "5ec4a18"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 43
last_session: "session-20260310-192540-28b102b0"
---

## Tool Usage Patterns

| Tool | Total Calls | Sessions Using | Avg Calls/Session |
|------|------------|---------------|------------------|
| tool.call (all) | 1806 | 43 | 42 |
| tool.file_change | 745 | 35 | 21 |

## File Change Hotspots

| Path Pattern | Changes | Sessions | Domain |
|-------------|---------|---------|--------|
| .sos/ (session artifacts) | 165 | 1 | mythology-architecture naming |
| mena/ | 131 | 3 | mena content, skills, commands |
| internal/materialize/ | 87 | 5 | sync pipeline, mena materialization |
| .sos/wip/frames/ | 42 | 8 | session framing documents |
| rites/ | 40 | 4 | rite configuration |
| internal/cmd/ | 30 | 4 | CLI commands |
| .ledge/reviews/ | 25 | 3 | hygiene audit artifacts |
| docs/ | 20 | 3 | documentation, doctrine |

## Phase Progression Patterns

| Terminal Phase | Sessions | Avg Session Duration |
|---------------|---------|---------------------|
| requirements | 30 | 8m |
| implementation | 7 | 17m |
| validation | 2 | 29m |
| design | 1 | 67m |
| audit | 1 | 55m |
| completed | 2 | 55m |

## Agent Delegation Patterns

- 41/43 sessions had agent delegation events (agent.task_start/agent.task_end)
- All agent.task_start events have agent_name="unknown", preventing per-agent-type analysis
- Top sessions by delegation count: session-20260304-001522-1c07b3bd (28), session-20260309-000328-ead5f874 (28), session-20260304-154410-af90596d (18), session-20260303-132132-f3031827 (16), session-20260303-003107-55293519 (12)
- Sessions with 0 delegation events: session-20260306-110952-e81c2ec3, session-20260302-150407-07b8e5be
- Multi-handoff chains (3+ handoffs) documented in SESSION_CONTEXT: session-20260304-001522-1c07b3bd (4 handoffs: code-smeller -> architect-enforcer -> janitor -> audit-lead), session-20260304-154410-af90596d (Phase 1 hygiene + Phase 2 SMELLS remediation)
- agent_name data limitation: all events report agent_name="unknown" due to CC hook payload not containing agent identity

## Common Command Patterns

- Session lifecycle (create, park, wrap): present in all 43 sessions
- ari sync: referenced in 4 sessions (sync noise remediation, framework agnosticism, mena materialization)
- CGO_ENABLED=0 go build/test: referenced in 4 sessions as validation steps
- git rev-parse: referenced in 2 sessions for source hash resolution
- ari session lock/unlock --agent moirai: present in sessions with SESSION_CONTEXT mutations

## Phase Transition Events

| Session | Transitions | Path |
|---------|------------|------|
| session-20260303-132132-f3031827 | 1 | requirements -> implementation |
| session-20260303-135856-c525b918 | 1 | requirements -> implementation |
| session-20260303-143357-851c7078 | 3 | requirements -> design -> implementation -> validation |
| session-20260304-135616-f6f65e42 | 2 | requirements -> design -> implementation |
| session-20260304-001522-1c07b3bd | 0 | assessment -> planning -> execution -> audit (custom hygiene phases, not in events) |
| session-20260304-154410-af90596d | 0 | assessment -> planning -> execution -> validation (custom hygiene phases, not in events) |
| session-20260305-191804-8e309d1f | 1 | requirements -> implementation |
| session-20260310-184202-048a25e4 | 1 | requirements -> implementation |

## Session Duration Distribution

| Bucket | Count | Sessions |
|--------|-------|---------|
| 0-5m | 17 | session-20260302-115457, session-20260302-121512, session-20260302-121832, session-20260302-150407, session-20260302-151610, session-20260302-151730, session-20260302-223308, session-20260302-224914, session-20260303-114830, session-20260303-163517, session-20260305-190141, session-20260306-110952, session-20260306-201337, session-20260308-223033, session-20260310-122716, session-20260310-174131, session-20260310-192540 |
| 5-15m | 7 | session-20260302-114829, session-20260302-120246, session-20260302-210626, session-20260302-232344, session-20260305-172543, session-20260305-185110, session-20260306-092108 |
| 15-30m | 10 | session-20260302-223852, session-20260303-132132, session-20260303-135856, session-20260303-143357, session-20260304-135616, session-20260305-191804, session-20260306-084705, session-20260306-152208, session-20260309-000328, session-20260310-184202 |
| 30-60m | 6 | session-20260302-122213, session-20260302-131246, session-20260302-155911, session-20260304-001522, session-20260304-154410, session-20260302-122533 |
| 60m+ | 3 | session-20260302-123250, session-20260303-003107, session-20260306-164058 |

## Observations

- 40% of sessions (17/43) lasted less than 5 minutes, indicating high frequency of session creation followed by rapid parking
- Sessions that reached implementation or later phases averaged 17+ minutes and produced concrete artifacts (commits, specs, reviews)
- The heaviest sessions by tool calls were session-20260303-003107 (291 tool calls, comprehensive debt remediation), session-20260304-001522 (246 tool calls, framework agnosticism audit), and session-20260309-000328 (235 tool calls, mythology-architecture naming alignment)
- The heaviest sessions by file changes were session-20260309-000328 (165 file changes, mythology-architecture naming alignment), session-20260303-003107 (131 file changes, comprehensive debt remediation), and session-20260304-001522 (87 file changes, framework agnosticism audit)
- No phase.transitioned events exist in any events.jsonl; phase transitions are tracked exclusively via SESSION_CONTEXT.md frontmatter mutations
- Agent delegation was present in 41/43 sessions (95%), but all events report agent_name="unknown" due to hook payload limitations
- hygiene rite sessions consistently produced the most structured workflows with formal handoff chains between specialist agents
- The GRAY sails across all 43 sessions indicate no CI proof infrastructure was captured, limiting quality friction analysis
- 4 new sessions on 2026-03-10 added 69 tool calls and 12 file changes, focused on DI and test infrastructure
- session-20260310-184202-048a25e4 (Materializer Constructor DI) had 43 tool calls and 11 file changes in 21 minutes, the highest tool density among the new sessions

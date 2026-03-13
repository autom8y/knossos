---
domain: "workflow-patterns"
generated_at: "2026-03-13T12:00:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "59a0de2"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 48
last_session: "session-20260312-183235-bbbf5231"
---

## Tool Usage Patterns

| Tool | Total Calls | Sessions Using | Avg Calls/Session |
|------|------------|---------------|------------------|
| tool.call (all) | 2021 | 48 | 42 |
| tool.file_change | 835 | 38 | 22 |

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
| requirements | 32 | 8m |
| implementation | 9 | 22m |
| design | 2 | 59m |
| validation | 2 | 29m |
| audit | 1 | 55m |
| completed | 2 | 55m |

## Agent Delegation Patterns

- 0/48 sessions had agent.delegated events in events.jsonl (event type not present in archive format)
- 0/48 sessions had agent.task_start events in events.jsonl (event type not present in archive format)
- Agent delegation is tracked exclusively via SESSION_CONTEXT.md narrative (Handoffs sections), not via events.jsonl
- Multi-handoff chains documented in SESSION_CONTEXT: session-20260304-001522-1c07b3bd (4 handoffs: code-smeller -> architect-enforcer -> janitor -> audit-lead), session-20260304-154410-af90596d (Phase 1 hygiene + Phase 2 SMELLS remediation)
- agent delegation event data limitation: archive event schema does not include agent.delegated or agent.task_start event types

## Common Command Patterns

- Session lifecycle (create, park, wrap): present in all 48 sessions
- ari sync: referenced in 4 sessions (sync noise remediation, framework agnosticism, mena materialization)
- CGO_ENABLED=0 go build/test: referenced in 5 sessions as validation steps
- git rev-parse: referenced in 2 sessions for source hash resolution
- ari session lock/unlock --agent moirai: present in sessions with SESSION_CONTEXT mutations

## Phase Transition Events

| Session | Transitions | Path |
|---------|------------|------|
| session-20260303-132132-f3031827 | 1 | requirements -> implementation |
| session-20260303-135856-c525b918 | 1 | requirements -> implementation |
| session-20260303-143357-851c7078 | 3 | requirements -> design -> implementation -> validation |
| session-20260304-001522-1c07b3bd | 0 | assessment -> planning -> execution -> audit (custom hygiene phases, not in events) |
| session-20260304-135616-f6f65e42 | 2 | requirements -> design -> implementation |
| session-20260304-154410-af90596d | 0 | assessment -> planning -> execution -> validation (custom hygiene phases, not in events) |
| session-20260305-191804-8e309d1f | 1 | requirements -> implementation |
| session-20260310-184202-048a25e4 | 1 | requirements -> implementation |
| session-20260311-231845-47df6625 | 1 | requirements -> design |
| session-20260312-183235-bbbf5231 | 1 | requirements -> implementation (with 2 rite transitions: hygiene -> 10x-dev -> docs) |

## Session Duration Distribution

| Bucket | Count | Sessions |
|--------|-------|---------|
| 0-5m | 17 | session-20260302-115457, session-20260302-121512, session-20260302-121832, session-20260302-150407, session-20260302-151610, session-20260302-151730, session-20260302-223308, session-20260302-224914, session-20260303-114830, session-20260303-163517, session-20260305-190141, session-20260306-110952, session-20260306-201337, session-20260310-122716, session-20260310-174131, session-20260310-192540, session-20260311-012734 |
| 5-15m | 8 | session-20260302-114829, session-20260302-120246, session-20260302-210626, session-20260302-232344, session-20260305-172543, session-20260305-185110, session-20260306-092108, session-20260312-113227 |
| 15-30m | 10 | session-20260302-223852, session-20260303-132132, session-20260303-135856, session-20260303-143357, session-20260304-135616, session-20260305-191804, session-20260306-084705, session-20260306-152208, session-20260309-000328, session-20260310-184202 |
| 30-60m | 9 | session-20260302-122213, session-20260302-131246, session-20260302-155911, session-20260302-122533, session-20260304-001522, session-20260304-154410, session-20260308-223033, session-20260311-231845, session-20260312-123128 |
| 60m+ | 4 | session-20260302-123250, session-20260303-003107, session-20260306-164058, session-20260312-183235 |

## Observations

- 35% of sessions (17/48) lasted less than 5 minutes, indicating high frequency of session creation followed by rapid parking
- Sessions that reached implementation or later phases averaged 22+ minutes and produced concrete artifacts (commits, specs, reviews)
- The heaviest sessions by tool calls were session-20260303-003107 (291 tool calls, comprehensive debt remediation), session-20260304-001522 (246 tool calls, framework agnosticism audit), session-20260309-000328 (235 tool calls, mythology-architecture naming alignment), and session-20260312-183235 (153 tool calls, preferential-language-scrub)
- The heaviest sessions by file changes were session-20260309-000328 (165 file changes), session-20260303-003107 (131 file changes), session-20260304-001522 (87 file changes), and session-20260312-183235 (76 file changes)
- No phase.transitioned events exist in any events.jsonl; phase transitions are tracked exclusively via SESSION_CONTEXT.md frontmatter mutations
- No agent.delegated or agent.task_start events exist in any events.jsonl; agent delegation is tracked via SESSION_CONTEXT narrative
- hygiene rite sessions consistently produced the most structured workflows with formal handoff chains between specialist agents
- All 48 sessions have GRAY sails (all proofs UNKNOWN), limiting quality friction analysis
- 5 new sessions (2026-03-11 through 2026-03-12) contributed 215 tool calls and 90 file changes
- session-20260312-183235-bbbf5231 (preferential-language-scrub) had 153 tool calls and 76 file changes across 59 minutes, the most active new session
- session-20260311-231845-47df6625 (multi-channel-integration) reached design phase in 50 minutes with 26 tool calls, indicating substantial planning work

---
domain: "workflow-patterns"
generated_at: "2026-03-09T10:15:06Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "2849e0b"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 39
last_session: "session-20260309-000328-ead5f874"
---

## Tool Usage Patterns

| Tool | Total Calls | Sessions Using | Avg Calls/Session |
|------|------------|---------------|------------------|
| tool.call (all) | 1549 | 39 | 40 |
| file_change | 836 | 33 | 25 |

## File Change Hotspots

| Path Pattern | Changes | Sessions | Domain |
|-------------|---------|---------|--------|
| internal/materialize/ | 87+ | 5 | sync pipeline, mena materialization |
| .sos/wip/frames/ | 42+ | 8 | session framing documents |
| internal/cmd/ | 30+ | 4 | CLI commands |
| .ledge/reviews/ | 25+ | 3 | hygiene audit artifacts |
| mena/ | 165+ | 3 | mena content, skills, commands |
| docs/ | 20+ | 3 | documentation, doctrine |
| rites/ | 40+ | 4 | rite configuration |

## Phase Progression Patterns

| Terminal Phase | Sessions | Avg Session Duration |
|---------------|---------|---------------------|
| requirements | 28 | 5m |
| implementation | 5 | 15m |
| validation | 2 | 47m |
| design | 1 | 67m |
| audit | 1 | 55m |
| completed | 2 | 62m |

## Agent Delegation Patterns

| Pattern | Frequency | Notes |
|---------|-----------|-------|
| Pythia-coordinated multi-agent workflow | 5 sessions | Sessions with phase transitions used Pythia to coordinate specialist agents |
| Single-agent execution | 28 sessions | Most requirements-phase sessions involved minimal agent routing |
| Multi-handoff chain (3+ handoffs) | 3 sessions | Framework agnosticism audit (4 handoffs), mena content path hygiene (2 phases), comprehensive debt remediation |

- 27/39 sessions had agent delegation events
- Top sessions by delegation count: session-20260304-001522-1c07b3bd (28), session-20260309-000328-ead5f874 (28), session-20260304-154410-af90596d (18)
- agent_name data limitation: event type names vary between agent.delegated and agent.task_start formats

## Common Command Patterns

- Session lifecycle (create, park, wrap): present in all 39 sessions
- ari sync: referenced in 4 sessions (sync noise remediation, framework agnosticism, mena materialization)
- CGO_ENABLED=0 go build/test: referenced in 3 sessions as validation steps
- git rev-parse: referenced in 2 sessions for source hash resolution

## Phase Transition Events

| Session | Transitions | Path |
|---------|------------|------|
| session-20260303-132132-f3031827 | 1 | requirements -> implementation |
| session-20260303-135856-c525b918 | 1 | requirements -> implementation |
| session-20260303-143357-851c7078 | 3 | requirements -> design -> implementation -> validation |
| session-20260304-135616-f6f65e42 | 2 | requirements -> design -> implementation |
| session-20260304-001522-1c07b3bd | 0 | assessment -> planning -> execution -> audit (custom hygiene phases, not in events) |
| session-20260304-154410-af90596d | 0 | assessment -> planning -> execution -> validation (custom hygiene phases, not in events) |

## Session Duration Distribution

| Bucket | Count | Sessions |
|--------|-------|---------|
| 0-5m | 18 | session-20260302-115457, session-20260302-120246, session-20260302-121512, session-20260302-121832, session-20260302-150407, session-20260302-151610, session-20260302-151730, session-20260302-223308, session-20260302-224914, session-20260303-114830, session-20260303-163517, session-20260305-185110, session-20260305-190141, session-20260306-110952, session-20260306-201337, session-20260302-114829, session-20260302-232344, session-20260302-120246 |
| 5-15m | 5 | session-20260302-210626, session-20260305-185110, session-20260306-092108, session-20260305-172543, session-20260308-223033 |
| 15-30m | 6 | session-20260302-223852, session-20260303-132132, session-20260303-135856, session-20260304-135616, session-20260305-191804, session-20260306-152208 |
| 30-60m | 6 | session-20260302-122213, session-20260302-122533, session-20260302-131246, session-20260302-155911, session-20260304-001522, session-20260306-084705 |
| 60m+ | 4 | session-20260302-123250, session-20260303-003107, session-20260304-154410, session-20260306-164058 |

## Observations

- 46% of sessions (18/39) lasted less than 5 minutes, indicating high frequency of session creation followed by rapid parking
- Sessions that reached implementation or later phases averaged 15+ minutes and produced concrete artifacts (commits, specs, reviews)
- The heaviest sessions by tool calls were session-20260303-003107 (291 tool calls, comprehensive debt remediation) and session-20260304-001522 (246 tool calls, framework agnosticism audit)
- The heaviest sessions by file changes were session-20260309-000328 (165 file changes, mythology-architecture naming alignment) and session-20260303-003107 (131 file changes, comprehensive debt remediation)
- Only 7 sessions (18%) had recorded phase transitions in events.jsonl; the remaining used custom phase flows or parked before transitioning
- Agent delegation was present in 27/39 sessions (69%), but most delegation happened within the initial session setup (Pythia routing)
- hygiene rite sessions consistently produced the most structured workflows with formal handoff chains between specialist agents
- The GRAY sails across all 39 sessions indicate no CI proof infrastructure was captured, limiting quality friction analysis

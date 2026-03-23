---
domain: "workflow-patterns"
generated_at: "2026-03-23T15:50:00Z"
expires_after: "14d"
source_scope: [".sos/archive/**"]
generator: "dionysus"
source_hash: "0ffd8061"
confidence: 0.85
format_version: "1.0"
sessions_synthesized: 55
last_session: "session-20260323-153258-fd8f7c41"
---

## Tool Usage Patterns

| Tool | Total Calls | Sessions Using | Avg Calls/Session |
|------|------------|---------------|------------------|
| tool.call (all tools) | 2439 | 55 | 44 |
| tool.file_change | 982 | 45 | 22 |

### Top Sessions by Tool Calls

| Session | tool.call | tool.file_change | Initiative |
|---------|----------|-----------------|-----------|
| session-20260303-003107-55293519 | 291 | 131 | Comprehensive Debt Remediation |
| session-20260304-001522-1c07b3bd | 246 | 87 | Framework Agnosticism Audit |
| session-20260309-000328-ead5f874 | 235 | 165 | Mythology-Architecture Naming Alignment |
| session-20260305-021309-e79a141b | 228 | 74 | claude-boundary-enforcement |
| session-20260312-183235-bbbf5231 | 153 | 76 | preferential-language-scrub |
| session-20260323-123506-4c58fdce | 97 | 44 | Cassandra Comprehensive Remediation |
| session-20260304-154410-af90596d | 84 | 21 | Mena Content Path Hygiene |
| session-20260303-132132-f3031827 | 69 | 24 | .know/ Incremental Refresh |
| session-20260306-164058-a8e7e897 | 63 | 5 | Release hardening |
| session-20260306-084705-0f5ea0c7 | 53 | 24 | embody-phase-2-workflow-awareness |

## File Change Hotspots

| Path Pattern | Changes | Sessions | Domain |
|-------------|---------|---------|--------|
| .sos/sessions/*/SESSION_CONTEXT.md | 982 | 45 | session-management |
| internal/materialize/**/* | 131 | 3 | materialization |
| internal/cmd/**/* | 44 | 2 | cli-infrastructure |
| rites/**/* | 76 | 3 | rite-content |
| .ledge/**/* | 21 | 3 | work-artifacts |
| docs/**/* | 165 | 2 | documentation |
| agents/**/* | 15 | 2 | agent-prompts |

## Phase Progression Patterns

| Terminal Phase | Sessions | Avg Session Duration |
|---------------|---------|---------------------|
| requirements | 42 | 25h3m |
| implementation | 7 | 14h0m |
| design | 2 | 37h16m |
| validation | 2 | 46m |
| completed | 1 | 1h40m |
| audit | 1 | 55m |
| execution | 1 | 4h12m |
| PLAN | 1 | 6h39m |

### Phase Transition Events (from events.jsonl)

| Session | Transitions | Path |
|---------|------------|------|
| session-20260303-132132-f3031827 | 1 | requirements -> implementation |
| session-20260303-135856-c525b918 | 1 | requirements -> implementation |
| session-20260303-143357-851c7078 | 3 | requirements -> design -> implementation -> validation |
| session-20260304-135616-f6f65e42 | 2 | requirements -> design -> implementation |
| session-20260310-184202-048a25e4 | 1 | requirements -> implementation |
| session-20260311-231845-47df6625 | 1 | requirements -> design |

## Agent Delegation Patterns

- 42/55 sessions (76%) had subagent activity (agent.task_start events)
- 171 total agent.task_start events across 42 sessions (avg 4 per session with subagents)
- 217 total agent.task_end events across 52 sessions
- 1 agent.delegated event (session-20260304-001522-1c07b3bd: Framework Agnosticism Audit)
- All agent_name fields are "unknown" in agent.task_end events
- agent_name data limitation: agent routing details not captured in event stream

### Top Sessions by Subagent Volume

| Session | task_start | task_end | Initiative |
|---------|-----------|---------|-----------|
| session-20260309-000328-ead5f874 | 13 | 15 | Mythology-Architecture Naming Alignment |
| session-20260304-001522-1c07b3bd | 12 | 15 | Framework Agnosticism Audit |
| session-20260312-183235-bbbf5231 | 11 | 12 | preferential-language-scrub |
| session-20260304-154410-af90596d | 9 | 9 | Mena Content Path Hygiene |
| session-20260317-232920-e16ac076 | 9 | 4 | CSS architecture first principles |
| session-20260312-123128-56a5e48f | 8 | 9 | harness-agnosticism |
| session-20260303-132132-f3031827 | 6 | 10 | .know/ Incremental Refresh |
| session-20260323-114009-c6891170 | 6 | 7 | Cassandra P0+P1 Health Review |

## Session Duration Distribution

| Bucket | Count | Sessions |
|--------|-------|---------|
| 0-10m | 7 | session-20260305-185110-545bd234, session-20260305-190141-62071002, session-20260306-110952-e81c2ec3, session-20260302-121512-f4f4f4a6, session-20260302-121832-e059c280, session-20260323-153258-fd8f7c41, session-20260302-224914-750e269f |
| 10-30m | 8 | session-20260302-122213-1e802b37, session-20260302-223852-c7b44b21, session-20260303-143357-851c7078, session-20260304-135616-f6f65e42, session-20260305-172543-d0e8d2fc, session-20260305-191804-8e309d1f, session-20260306-092108-0c1f2ce9, session-20260323-151705-26cbb521 |
| 30m-1h | 6 | session-20260302-131246-b6788294, session-20260302-155911-b3ec77d4, session-20260304-001522-1c07b3bd, session-20260306-084705-0f5ea0c7, session-20260309-000328-ead5f874, session-20260317-232920-e16ac076 |
| 1h-2h | 7 | session-20260302-123250-636200b5, session-20260304-154410-af90596d, session-20260306-152208-d302cc16, session-20260306-164058-a8e7e897, session-20260306-201337-8581513b, session-20260310-174131-33955b41, session-20260310-192540-28b102b0 |
| 2h-6h | 5 | session-20260311-231845-47df6625, session-20260312-183235-bbbf5231, session-20260323-123506-4c58fdce, session-20260323-114009-c6891170, session-20260312-123128-56a5e48f |
| 6h-24h | 3 | session-20260319-042521-06e927c4, session-20260311-012734-9847ff6f, session-20260312-113227-1c6e9313 |
| 24h+ | 19 | session-20260302-114829-b2f806a3, session-20260302-115457-2233b3ef, session-20260302-120246-e1399c67, session-20260302-122533-2820b8a1, session-20260302-150407-07b8e5be, session-20260302-151610-273eeb4a, session-20260302-151730-bffc2248, session-20260302-210626-4a7b9681, session-20260302-223308-10f40f3c, session-20260302-232344-1b73b3a8, session-20260303-003107-55293519, session-20260303-114830-beb1bf03, session-20260303-132132-f3031827, session-20260303-135856-c525b918, session-20260303-163517-f43d2334, session-20260305-021309-e79a141b, session-20260310-184202-048a25e4, session-20260310-122716-65d5c100, session-20260308-223033-625c43ec |

## Observations

- 2439 tool.call events across 55 sessions; 982 tool.file_change events across 45 sessions
- 10 sessions account for 0 file changes (session-level management only, no code modifications)
- 6/55 sessions (11%) have recorded phase.transitioned events; session-20260303-143357-851c7078 traversed the most phases (requirements -> design -> implementation -> validation)
- session-20260303-003107-55293519 (Comprehensive Debt Remediation) had the highest tool.call count (291) and second-highest file_change count (131)
- session-20260309-000328-ead5f874 (Mythology-Architecture Naming Alignment) had the highest file_change count (165)
- 19/55 sessions (35%) have wall-clock duration >24h, inflated by park/archive lag rather than active work time
- The 0-30m bucket contains 15/55 sessions (27%), reflecting many short scoping/exploration sessions
- 76% of sessions with subagent activity aligns with orchestrated mode being the dominant execution model
- No CI proof data available in any session (all GRAY sails with UNKNOWN proofs)

# QA Test Plan: Knossos v0.1.0 Comprehensive Validation

**Initiative**: knossos-v0.1.0-qa
**Session**: session-20260107-200019-bbf01130
**Agent**: qa-adversary
**Created**: 2026-01-07
**Pass Criteria**: STRICT - Zero errors. Any defect is a finding.

---

## 1. Executive Summary

This test plan provides comprehensive validation coverage for Knossos v0.1.0, extending beyond the initial release QA (QA-REPORT-v0.1.0.md) to cover:

1. **Rite Materialization Matrix** - All 12 rites with all flag combinations
2. **ari CLI Command Groups** - Complete command coverage
3. **Session Lifecycle Workflows** - Full lifecycle scenarios
4. **Moirai Agent Flows** - Session state management via Fates
5. **White Sails** - Quality gate validation
6. **Hook Shell Wrapper Validation** - Infrastructure integrity
7. **Regression Tests** - Known issues and fixes
8. **Greenfield Satellite Project Setup** - Fresh project initialization
9. **Edge Cases** - Error handling and boundary conditions

---

## 2. Test Environment

| Component | Value |
|-----------|-------|
| Platform | darwin (macOS) |
| OS Version | Darwin 25.1.0 |
| ari Binary | /Users/tomtenuta/Code/roster/ari |
| ari Version | 0.1.0 |
| Project Root | /Users/tomtenuta/Code/roster |
| Active Rite | 10x-dev |
| Current Session | session-20260107-200019-bbf01130 |

---

## 3. Rite Materialization Matrix

### 3.1 Rite Inventory (12 Rites)

| # | Rite | Type | Description |
|---|------|------|-------------|
| 1 | 10x-dev | Sequential | Full development lifecycle (PRD -> TDD -> Code -> QA) |
| 2 | debt-triage | Sequential | Technical debt management |
| 3 | docs | Sequential | Documentation workflows |
| 4 | ecosystem | Sequential | Ecosystem integration |
| 5 | forge | Sequential | Infrastructure/tooling development |
| 6 | hygiene | Sequential | Code quality workflows |
| 7 | intelligence | Sequential | Analytics and insights |
| 8 | rnd | Sequential | Research and development |
| 9 | security | Sequential | Security assessment |
| 10 | shared | Shared | Cross-rite shared skills |
| 11 | sre | Sequential | Site reliability engineering |
| 12 | strategy | Sequential | Strategic planning |

### 3.2 Test Cases - Fresh Materialization

| TC-ID | Rite | Precondition | Command | Expected Result |
|-------|------|--------------|---------|-----------------|
| RITE-001 | 10x-dev | No prior rite | `ari sync materialize --rite 10x-dev` | Success, .claude/ populated |
| RITE-002 | debt-triage | No prior rite | `ari sync materialize --rite debt-triage` | Success, .claude/ populated |
| RITE-003 | docs | No prior rite | `ari sync materialize --rite docs` | Success, .claude/ populated |
| RITE-004 | ecosystem | No prior rite | `ari sync materialize --rite ecosystem` | Success, .claude/ populated |
| RITE-005 | forge | No prior rite | `ari sync materialize --rite forge` | Success, .claude/ populated |
| RITE-006 | hygiene | No prior rite | `ari sync materialize --rite hygiene` | Success, .claude/ populated |
| RITE-007 | intelligence | No prior rite | `ari sync materialize --rite intelligence` | Success, .claude/ populated |
| RITE-008 | rnd | No prior rite | `ari sync materialize --rite rnd` | Success, .claude/ populated |
| RITE-009 | security | No prior rite | `ari sync materialize --rite security` | Success, .claude/ populated |
| RITE-010 | sre | No prior rite | `ari sync materialize --rite sre` | Success, .claude/ populated |
| RITE-011 | strategy | No prior rite | `ari sync materialize --rite strategy` | Success, .claude/ populated |
| RITE-012 | shared | No prior rite | `ari sync materialize --rite shared` | Error or partial (shared is special) |

### 3.3 Test Cases - Rite Switching

| TC-ID | From | To | Command | Expected Result |
|-------|------|-----|---------|-----------------|
| RITE-SW-001 | 10x-dev | hygiene | `ari sync materialize --rite hygiene` | Agents/skills replaced |
| RITE-SW-002 | hygiene | security | `ari sync materialize --rite security` | Agents/skills replaced |
| RITE-SW-003 | security | 10x-dev | `ari sync materialize --rite 10x-dev` | Restored to original |

### 3.4 Test Cases - Flag Combinations

| TC-ID | Flags | Command | Expected Result |
|-------|-------|---------|-----------------|
| RITE-FLG-001 | --force | `ari sync materialize --force` | Regenerate, overwrite local changes |
| RITE-FLG-002 | --rite + --force | `ari sync materialize --rite hygiene --force` | Switch and force regenerate |

### 3.5 Verification Points

For each materialization:
- [ ] ACTIVE_RITE updated to rite name
- [ ] .claude/agents/ contains expected agents
- [ ] .claude/skills/ contains expected skills (rite + shared)
- [ ] .claude/CLAUDE.md updated with quick-start section
- [ ] .claude/sync/state.json updated with timestamp

---

## 4. ari CLI Command Groups

### 4.1 sync Commands

| TC-ID | Command | Expected Behavior |
|-------|---------|-------------------|
| CLI-SYNC-001 | `ari sync status` | Show sync status for tracked paths |
| CLI-SYNC-002 | `ari sync materialize` | Generate .claude/ from templates |
| CLI-SYNC-003 | `ari sync pull` | Pull remote changes with conflict detection |
| CLI-SYNC-004 | `ari sync push` | Push local changes to remote |
| CLI-SYNC-005 | `ari sync diff` | Show local vs remote differences |
| CLI-SYNC-006 | `ari sync resolve` | Resolve sync conflicts |
| CLI-SYNC-007 | `ari sync history` | Show sync history/audit log |
| CLI-SYNC-008 | `ari sync reset` | Reset sync state (dangerous) |

### 4.2 session Commands

| TC-ID | Command | Expected Behavior |
|-------|---------|-------------------|
| CLI-SESS-001 | `ari session create "test" MODULE` | Create new session |
| CLI-SESS-002 | `ari session status` | Show session state |
| CLI-SESS-003 | `ari session list` | List all sessions |
| CLI-SESS-004 | `ari session park --reason "test"` | Suspend current session |
| CLI-SESS-005 | `ari session resume` | Resume parked session |
| CLI-SESS-006 | `ari session wrap` | Complete and archive session |
| CLI-SESS-007 | `ari session audit` | Show session event history |
| CLI-SESS-008 | `ari session transition` | Transition between workflow phases |
| CLI-SESS-009 | `ari session lock` | Manually acquire session lock |
| CLI-SESS-010 | `ari session unlock` | Manually release session lock |
| CLI-SESS-011 | `ari session migrate` | Migrate sessions to v2 schema |

### 4.3 hook Commands

| TC-ID | Command | Expected Behavior |
|-------|---------|-------------------|
| CLI-HOOK-001 | `ari hook clew` | Record tool events (PostToolUse) |
| CLI-HOOK-002 | `ari hook context` | Inject session context (SessionStart) |
| CLI-HOOK-003 | `ari hook autopark` | Auto-park on Stop event |
| CLI-HOOK-004 | `ari hook route` | Route slash commands |
| CLI-HOOK-005 | `ari hook validate` | Validate bash commands |
| CLI-HOOK-006 | `ari hook writeguard` | Block direct writes to context files |

### 4.4 sails Commands

| TC-ID | Command | Expected Behavior |
|-------|---------|-------------------|
| CLI-SAILS-001 | `ari sails check` | Check quality gate for session |
| CLI-SAILS-002 | `ari sails check --session <id>` | Check specific session |

### 4.5 handoff Commands

| TC-ID | Command | Expected Behavior |
|-------|---------|-------------------|
| CLI-HAND-001 | `ari handoff prepare --from X --to Y` | Validate readiness |
| CLI-HAND-002 | `ari handoff execute --to Y` | Trigger transition |
| CLI-HAND-003 | `ari handoff status` | Query current handoff state |
| CLI-HAND-004 | `ari handoff history` | Query handoff events |

### 4.6 rite Commands

| TC-ID | Command | Expected Behavior |
|-------|---------|-------------------|
| CLI-RITE-001 | `ari rite list` | List available rites |
| CLI-RITE-002 | `ari rite current` | Show active rite |
| CLI-RITE-003 | `ari rite info <rite>` | Show detailed rite info |
| CLI-RITE-004 | `ari rite status` | Show rite status |
| CLI-RITE-005 | `ari rite context` | Show rite context for injection |
| CLI-RITE-006 | `ari rite invoke <rite>` | Borrow components |
| CLI-RITE-007 | `ari rite release` | Release borrowed components |
| CLI-RITE-008 | `ari rite swap <rite>` | Full context switch |
| CLI-RITE-009 | `ari rite validate` | Validate rite integrity |

### 4.7 Other Commands

| TC-ID | Command | Expected Behavior |
|-------|---------|-------------------|
| CLI-OTHER-001 | `ari version` | Show version information |
| CLI-OTHER-002 | `ari --help` | Show help for all commands |
| CLI-OTHER-003 | `ari artifact list` | List workflow artifacts |
| CLI-OTHER-004 | `ari artifact register` | Register new artifact |
| CLI-OTHER-005 | `ari worktree list` | List git worktrees |
| CLI-OTHER-006 | `ari worktree create` | Create new worktree |
| CLI-OTHER-007 | `ari naxos scan` | Scan for abandoned sessions |
| CLI-OTHER-008 | `ari manifest validate` | Validate manifest files |
| CLI-OTHER-009 | `ari inscription sync` | Sync CLAUDE.md inscriptions |
| CLI-OTHER-010 | `ari tribute generate` | Generate session summaries |

---

## 5. Session Lifecycle Workflows

### 5.1 Full Lifecycle: Create -> Work -> Park -> Resume -> Wrap

| Step | Command | Expected State |
|------|---------|----------------|
| 1 | `ari session create "test-lifecycle" MODULE` | ACTIVE, phase=requirements |
| 2 | Simulate work (write file) | Artifacts tracked |
| 3 | `ari session park --reason "break"` | PARKED |
| 4 | `ari session resume` | ACTIVE |
| 5 | `ari session wrap` | WRAPPED, archived |

### 5.2 Direct Completion: Create -> Work -> Wrap

| Step | Command | Expected State |
|------|---------|----------------|
| 1 | `ari session create "direct-complete" SCRIPT` | ACTIVE |
| 2 | Simulate work | Artifacts tracked |
| 3 | `ari session wrap` | WRAPPED |

### 5.3 Complexity Levels

| TC-ID | Complexity | Expected Phases |
|-------|------------|-----------------|
| LIFE-CMPLX-001 | SCRIPT | requirements, implementation, validation |
| LIFE-CMPLX-002 | MODULE | requirements, design, implementation, validation |
| LIFE-CMPLX-003 | SERVICE | requirements, design, implementation, validation |
| LIFE-CMPLX-004 | PLATFORM | requirements, design, implementation, validation |

### 5.4 Cross-Rite Switching Mid-Session

| Step | Command | Expected Behavior |
|------|---------|-------------------|
| 1 | Session active with 10x-dev | Normal operation |
| 2 | `ari sync materialize --rite hygiene` | Rite switched |
| 3 | Session continues | Existing session unaffected |

---

## 6. Moirai Agent Flows

### 6.1 Known Issue: QA-001 (HIGH)

**Description**: Moirai returns conceptual response instead of executing session creation.

**Test Case**:
- Invoke: `Task(moirai, "create_sprint name=test-sprint")`
- Expected: Sprint created
- Actual: Conceptual explanation returned

### 6.2 Operation Matrix

| TC-ID | Operation | Fate | Test Command |
|-------|-----------|------|--------------|
| MOIRAI-001 | create_sprint | Clotho | `Task(moirai, "create_sprint name=test")` |
| MOIRAI-002 | start_sprint | Clotho | `Task(moirai, "start_sprint sprint-id")` |
| MOIRAI-003 | mark_complete | Lachesis | `Task(moirai, "mark_complete task-001")` |
| MOIRAI-004 | transition_phase | Lachesis | `Task(moirai, "transition_phase design")` |
| MOIRAI-005 | update_field | Lachesis | `Task(moirai, "update_field status=done")` |
| MOIRAI-006 | park_session | Lachesis | `Task(moirai, "park_session reason=test")` |
| MOIRAI-007 | resume_session | Lachesis | `Task(moirai, "resume_session")` |
| MOIRAI-008 | handoff | Lachesis | `Task(moirai, "handoff to=principal-engineer")` |
| MOIRAI-009 | wrap_session | Atropos | `Task(moirai, "wrap_session")` |
| MOIRAI-010 | generate_sails | Atropos | `Task(moirai, "generate_sails")` |

### 6.3 Natural Language Parsing

| TC-ID | Input | Expected Operation |
|-------|-------|-------------------|
| MOIRAI-NL-001 | "Mark the PRD task complete" | mark_complete |
| MOIRAI-NL-002 | "Pause the session" | park_session |
| MOIRAI-NL-003 | "Wrap up the work" | wrap_session |
| MOIRAI-NL-004 | "Hand off to QA" | handoff |

---

## 7. White Sails

### 7.1 sails check Functionality

| TC-ID | Scenario | Expected Result |
|-------|----------|-----------------|
| SAILS-001 | No WHITE_SAILS.yaml | Exit non-zero, "not generated" |
| SAILS-002 | Valid WHITE_SAILS.yaml | Exit 0 for WHITE |
| SAILS-003 | GRAY confidence | Exit non-zero |
| SAILS-004 | BLACK confidence | Exit non-zero |

### 7.2 Confidence Signal Generation

| TC-ID | Trigger | Expected Signal |
|-------|---------|-----------------|
| SAILS-GEN-001 | `ari sails check` on clean session | WHITE or GRAY |
| SAILS-GEN-002 | Session with failures | GRAY or BLACK |

---

## 8. Hook Shell Wrapper Validation

### 8.1 hooks.json Structure Validation

| File | Expected Hooks |
|------|----------------|
| base_hooks.yaml | SessionStart, Stop, PostToolUse, PreToolUse, UserPromptSubmit |
| ari/hooks.yaml | SessionStart, Stop, PreToolUse, PostToolUse, UserPromptSubmit |

### 8.2 Hook Registration

| TC-ID | Event | Hook | Expected |
|-------|-------|------|----------|
| HOOK-001 | SessionStart | context-injection/session-context.sh | Injects session context |
| HOOK-002 | SessionStart | ari/context.sh | Injects via ari hook context |
| HOOK-003 | Stop | session-guards/auto-park.sh | Auto-parks session |
| HOOK-004 | Stop | ari/autopark.sh | Auto-parks via ari |
| HOOK-005 | PostToolUse (Write) | tracking/artifact-tracker.sh | Tracks artifacts |
| HOOK-006 | PostToolUse | ari/clew.sh | Tracks via ari hook clew |
| HOOK-007 | PreToolUse (Bash) | validation/command-validator.sh | Validates commands |
| HOOK-008 | PreToolUse (Edit/Write) | ari/writeguard.sh | Guards context files |
| HOOK-009 | UserPromptSubmit | ari/route.sh | Routes slash commands |

### 8.3 Hook Script Execution

| TC-ID | Script | Test | Expected |
|-------|--------|------|----------|
| HOOK-EXEC-001 | clew.sh | Source and run | Executes without error |
| HOOK-EXEC-002 | context.sh | Source and run | Outputs context JSON |
| HOOK-EXEC-003 | autopark.sh | Source and run | Returns park guidance |
| HOOK-EXEC-004 | route.sh | Source with /sync | Routes to ari sync |
| HOOK-EXEC-005 | writeguard.sh | Write to SESSION_CONTEXT.md | Blocks write |

---

## 9. Regression Tests

### 9.1 /sync Routes to ari (Not Legacy roster-sync)

| TC-ID | Input | Expected |
|-------|-------|----------|
| REG-001 | `/sync` | Routes to `ari sync` commands |
| REG-002 | `/sync status` | Executes `ari sync status` |
| REG-003 | `/sync materialize` | Executes `ari sync materialize` |

### 9.2 PATH Fallback for ari Binary

| TC-ID | Scenario | Expected |
|-------|----------|----------|
| REG-PATH-001 | ari in PATH | Found and executed |
| REG-PATH-002 | ari at ~/.bin/ari | Found via fallback |
| REG-PATH-003 | ari at project/ari | Found via project-relative |
| REG-PATH-004 | ari not found | Graceful error message |

### 9.3 Skill Definitions Update After Materialize

| TC-ID | Scenario | Expected |
|-------|----------|----------|
| REG-SKILL-001 | Materialize 10x-dev | Skills reflect 10x-dev |
| REG-SKILL-002 | Materialize hygiene | Skills reflect hygiene |
| REG-SKILL-003 | CLAUDE.md quick-start | Reflects current agents |

---

## 10. Greenfield Satellite Project Setup

### 10.1 New Project Initialization

| Step | Action | Expected |
|------|--------|----------|
| 1 | Create new directory | Empty directory |
| 2 | Initialize git | .git created |
| 3 | Run `ari sync materialize --rite 10x-dev` | .claude/ created |
| 4 | Verify ACTIVE_RITE | Contains 10x-dev |
| 5 | Create session | session-* directory created |

### 10.2 Satellite Validation

| TC-ID | Check | Expected |
|-------|-------|----------|
| GREEN-001 | .claude/agents/ | Populated |
| GREEN-002 | .claude/skills/ | Populated |
| GREEN-003 | .claude/CLAUDE.md | Present with sections |
| GREEN-004 | Session creation | Works |
| GREEN-005 | Rite switching | Works |

---

## 11. Edge Cases

### 11.1 Missing Dependencies

| TC-ID | Scenario | Expected |
|-------|----------|----------|
| EDGE-001 | ari binary missing | Graceful error |
| EDGE-002 | Git not installed | Graceful error |
| EDGE-003 | Rite manifest missing | Exit with clear message |
| EDGE-004 | Shared rite missing | Exit with dependency error |

### 11.2 Malformed SESSION_CONTEXT.md

| TC-ID | Scenario | Expected |
|-------|----------|----------|
| EDGE-005 | Empty SESSION_CONTEXT.md | Graceful handling |
| EDGE-006 | Invalid YAML in SESSION_CONTEXT.md | Parse error message |
| EDGE-007 | Missing required fields | Validation error |

### 11.3 Concurrent Session Operations

| TC-ID | Scenario | Expected |
|-------|----------|----------|
| EDGE-008 | Two terminals, same session | Lock prevents conflict |
| EDGE-009 | Create while active | "Session already active" |
| EDGE-010 | Park while parked | "Session already parked" |
| EDGE-011 | Resume while active | "Session not parked" |

### 11.4 Input Validation

| TC-ID | Input | Expected |
|-------|-------|----------|
| EDGE-012 | Very long session name (200 chars) | Handled gracefully |
| EDGE-013 | Session name with special chars | Sanitized or rejected |
| EDGE-014 | Empty complexity | Error message |
| EDGE-015 | Invalid complexity | Error message |

### 11.5 File System Edge Cases

| TC-ID | Scenario | Expected |
|-------|----------|----------|
| EDGE-016 | Read-only .claude/ | Permission error |
| EDGE-017 | Disk full | Clear error message |
| EDGE-018 | Symlink in sessions path | Handled correctly |

---

## 12. Test Execution Plan

### 12.1 Execution Order

1. **Pre-flight checks** - Verify environment
2. **CLI Command Groups** - Basic functionality
3. **Rite Materialization Matrix** - All 12 rites
4. **Session Lifecycle** - Full workflows
5. **Moirai Agent Flows** - State management
6. **White Sails** - Quality gates
7. **Hook Validation** - Infrastructure
8. **Regression Tests** - Known issues
9. **Greenfield Project** - Fresh setup
10. **Edge Cases** - Error handling

### 12.2 Pass/Fail Criteria

- **PASS**: Expected behavior observed
- **FAIL**: Unexpected behavior, error, or crash
- **SKIP**: Cannot test due to precondition
- **BLOCKED**: Dependency issue prevents test

### 12.3 Defect Tracking

All defects will be documented with:
- Unique ID (DEF-XXX)
- Severity (CRITICAL, HIGH, MEDIUM, LOW)
- Steps to reproduce
- Expected vs actual behavior
- Evidence (command output)

---

## 13. Artifact Verification

All test results will be verified via Read tool and documented in the final QA report at:
`/Users/tomtenuta/Code/roster/docs/audits/QA-REPORT-knossos-v0.1.0-full.md`

---

## 14. Sign-off

This test plan covers comprehensive validation of Knossos v0.1.0. Execution will proceed systematically with all findings documented.

**Test Plan Author**: qa-adversary
**Date**: 2026-01-07
**Session**: session-20260107-200019-bbf01130

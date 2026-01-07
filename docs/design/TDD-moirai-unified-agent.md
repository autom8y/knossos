---
artifact_id: TDD-moirai-unified-agent
title: "TDD: Unified Moirai Agent"
created_at: "2026-01-07T23:00:00Z"
author: architect
status: draft
complexity: MODULE
prd_reference: docs/requirements/PRD-moirai-consolidation.md
success_criteria_mapping:
  SC-001: "Sections 2-4: Single agent handles all operations"
  SC-002: "Section 5: Fates as internal skills"
  SC-003: "Out of scope (hook TDD separate)"
  SC-004: "Section 8: Slash command integration"
  SC-005: "Out of scope (hook TDD separate)"
  SC-006: "Section 7: Audit protocol"
  SC-007: "Section 6: CLI integration"
  SC-008: "Section 7: Audit protocol with reasoning"
---

# TDD: Unified Moirai Agent

## 1. Overview

This Technical Design Document specifies the unified Moirai agent that consolidates the current 4-agent architecture (Moirai router + Clotho/Lachesis/Atropos) into a single agent with on-demand skill loading. The design eliminates routing overhead while preserving the Fate taxonomy through progressive disclosure.

### 1.1 Context

| Reference | Location |
|-----------|----------|
| PRD | `docs/requirements/PRD-moirai-consolidation.md` |
| Current Router | `user-agents/moirai.md` |
| Shared Infrastructure | `user-agents/moirai-shared.md` |
| Session FSM | `docs/design/TDD-session-state-machine.md` |
| Session Schema | `schemas/artifacts/session-context.schema.json` |
| Sprint Schema | `schemas/artifacts/sprint-context.schema.json` |

### 1.2 Problem Statement

The current architecture has four deficiencies:

1. **Router Overhead**: Extra subprocess hop through Moirai router before reaching Fate adds 50-100ms latency
2. **Cognitive Burden**: Users must understand Fate taxonomy despite invoking Moirai generically
3. **Fragmented Logic**: 4 separate agent files with shared infrastructure creates maintenance complexity
4. **Routing Errors**: Potential for FATE_MISMATCH errors when operations are misrouted

### 1.3 Design Goals

1. Single unified agent at `.claude/agents/moirai.md`
2. Sub-500ms end-to-end operation latency
3. Progressive disclosure via Fate skills (internal, not user-invokable)
4. CLI-authoritative for session state changes
5. Complete audit trail with reasoning for all mutations

---

## 2. Architecture

### 2.1 Component Diagram

```
                    +-----------------------+
                    |   Caller (Main Thread |
                    |   or Slash Command)   |
                    +-----------+-----------+
                                |
                                | Task(moirai, "operation...")
                                v
                    +---------------------------+
                    |      UNIFIED MOIRAI       |
                    |                           |
                    |  +---------------------+  |
                    |  | Operation Parser    |  |
                    |  +---------------------+  |
                    |           |               |
                    |  +--------v----------+    |
                    |  | Fate Skill Loader |    |
                    |  +-------------------+    |
                    |   |       |       |       |
                    |   v       v       v       |
                    | +---+  +----+  +-----+    |
                    | |Clo|  |Lach|  |Atro|    |
                    | +---+  +----+  +-----+    |
                    |   (skills, not agents)   |
                    +---------------------------+
                           |           |
              +------------+           +------------+
              |                                     |
              v                                     v
    +------------------+               +------------------------+
    |     ari CLI      |               |   Direct File Mutation |
    | (authoritative)  |               | (sprint operations)    |
    +------------------+               +------------------------+
              |                                     |
              v                                     v
    +------------------+               +------------------------+
    | SESSION_CONTEXT  |               |   SPRINT_CONTEXT.md    |
    |     .md          |               +------------------------+
    +------------------+
```

### 2.2 Agent Definition

```yaml
# .claude/agents/moirai.md frontmatter
---
name: moirai
description: |
  The Moirai--unified voice of the three Fates. Handles all session and sprint
  state mutations through a single interface with on-demand Fate skill loading.
tools: Read, Write, Edit, Glob, Grep, Bash, Skill
model: sonnet
color: indigo
aliases:
  - fates
  - state-mate
---
```

### 2.3 Components

| Component | Responsibility | Implementation |
|-----------|---------------|----------------|
| **Operation Parser** | Parse structured/NL input to operation name | Regex + keyword matching |
| **Fate Skill Loader** | Load appropriate skill on-demand | `Skill(moirai/clotho)` etc. |
| **CLI Executor** | Execute ari commands for session operations | `Bash(ari session ...)` |
| **Direct Mutator** | File operations for sprint-only changes | `Edit`, `Write` tools |
| **Response Formatter** | Format JSON responses per moirai-shared.md | Structured output |
| **Audit Logger** | Append mutations to audit trail | File append |

---

## 3. Interface Contract

### 3.1 Input Format

Moirai accepts both structured commands and natural language:

**Structured Command Format**:
```
operation_name [arg1=value1] [arg2=value2] [--flag]

Session Context:
- Session ID: {session-id}
- Session Path: .claude/sessions/{session-id}/SESSION_CONTEXT.md
```

**Natural Language Format**:
```
<natural language description of desired operation>

Session Context:
- Session ID: {session-id}
- Session Path: .claude/sessions/{session-id}/SESSION_CONTEXT.md
```

### 3.2 Operation Syntax

| Operation | Syntax | Required Context |
|-----------|--------|------------------|
| `create_session` | `create_session initiative="..." complexity=... [rite=...]` | None |
| `create_sprint` | `create_sprint name="..." [depends_on=...]` | Session |
| `start_sprint` | `start_sprint sprint_id` | Session |
| `mark_complete` | `mark_complete task_id artifact=path` | Session |
| `transition_phase` | `transition_phase from=X to=Y` | Session |
| `update_field` | `update_field field=value [field2=value2]` | Session |
| `park_session` | `park_session reason="..."` | Session |
| `resume_session` | `resume_session` | Session |
| `handoff` | `handoff to=agent [note="..."]` | Session |
| `record_decision` | `record_decision "text"` | Session |
| `append_content` | `append_content "text"` | Session |
| `wrap_session` | `wrap_session [--emergency]` | Session |
| `generate_sails` | `generate_sails [--skip-proofs]` | Session |
| `delete_sprint` | `delete_sprint sprint_id [--archive]` | Session |

### 3.3 Output Format

**Success Response**:
```json
{
  "success": true,
  "operation": "string",
  "message": "string",
  "reasoning": "string",
  "fate": "clotho|lachesis|atropos",
  "state_before": {},
  "state_after": {},
  "changes": {}
}
```

**Error Response**:
```json
{
  "success": false,
  "operation": "string",
  "error_code": "string",
  "message": "string",
  "reasoning": "string",
  "hint": "string"
}
```

**Dry-Run Response**:
```json
{
  "success": true,
  "operation": "string",
  "dry_run": true,
  "message": "Preview: Would {action}",
  "diff": {}
}
```

### 3.4 Error Codes

| Code | Description | Recovery |
|------|-------------|----------|
| `INVALID_OPERATION` | Unrecognized operation | Return valid operation list |
| `AMBIGUOUS_INPUT` | Cannot parse natural language | Request explicit syntax |
| `FILE_NOT_FOUND` | Session/sprint not found | Suggest creation |
| `SCHEMA_VIOLATION` | Output would fail validation | Show validation errors |
| `LIFECYCLE_VIOLATION` | State transition not allowed | Show valid transitions |
| `DEPENDENCY_BLOCKED` | Dependency not satisfied | Show blocking dependency |
| `LOCK_TIMEOUT` | Could not acquire lock | Retry after delay |
| `CONCURRENT_MODIFICATION` | File changed during operation | Re-read and retry |
| `QUALITY_GATE_FAILED` | BLACK sails block wrap | Suggest --emergency or fix |

---

## 4. State Machine

### 4.1 Session State Transitions

```
                    +--------------+
                    |              |
                    v              |
   (new) ---> ACTIVE ---> PARKED --+---> ARCHIVED
                |                         ^
                +-------------------------+
                     (direct wrap)
```

| From | To | Operation | Fate Domain | CLI Command |
|------|-----|-----------|-------------|-------------|
| (new) | ACTIVE | create_session | - | `ari session create` |
| ACTIVE | PARKED | park_session | Lachesis | `ari session park` |
| PARKED | ACTIVE | resume_session | Lachesis | `ari session resume` |
| ACTIVE | ARCHIVED | wrap_session | Atropos | `ari session wrap` |
| PARKED | ARCHIVED | wrap_session (--override) | Atropos | `ari session wrap --force` |

### 4.2 Sprint Status Transitions

```
   pending ---> active ---> blocked ---> completed ---> archived
                  |            |              ^
                  |            +--------------+
                  |                (unblock)
                  +---------------------------+
                       (can complete directly)
```

| From | To | Operation | Fate Domain |
|------|-----|-----------|-------------|
| pending | active | start_sprint | Clotho |
| active | blocked | update_field blocker=... | Lachesis |
| blocked | active | update_field blocker=null | Lachesis |
| active | completed | (auto when all tasks done) | Lachesis |
| completed | archived | delete_sprint --archive | Atropos |

### 4.3 Transition Validation

Moirai enforces state machine rules before executing operations:

```python
# Pseudo-code for transition validation
def validate_transition(current_state, operation):
    allowed = {
        "ACTIVE": ["park_session", "wrap_session", "transition_phase", ...],
        "PARKED": ["resume_session", "wrap_session (--override)"],
        "ARCHIVED": []  # Terminal state
    }
    if operation not in allowed[current_state]:
        return LIFECYCLE_VIOLATION
```

---

## 5. Skill Loading Protocol

### 5.1 Fate-to-Skill Mapping

| Operation | Fate Domain | Skill |
|-----------|-------------|-------|
| `create_session` | - | None (direct CLI) |
| `create_sprint` | Creation | `moirai/clotho` |
| `start_sprint` | Creation | `moirai/clotho` |
| `mark_complete` | Measurement | `moirai/lachesis` |
| `transition_phase` | Measurement | `moirai/lachesis` |
| `update_field` | Measurement | `moirai/lachesis` |
| `park_session` | Measurement | `moirai/lachesis` |
| `resume_session` | Measurement | `moirai/lachesis` |
| `handoff` | Measurement | `moirai/lachesis` |
| `record_decision` | Measurement | `moirai/lachesis` |
| `append_content` | Measurement | `moirai/lachesis` |
| `wrap_session` | Termination | `moirai/atropos` |
| `generate_sails` | Termination | `moirai/atropos` |
| `delete_sprint` | Termination | `moirai/atropos` |

### 5.2 Skill Loading Flow

```
1. Parse operation from input
2. Look up Fate domain from mapping
3. If CLI-backed operation:
   a. Execute CLI directly
   b. Parse CLI output
   c. Return structured response
4. If file-mutation operation:
   a. Load appropriate skill: Skill(moirai/{fate})
   b. Apply skill guidance for validation/mutation
   c. Execute file mutation
   d. Return structured response
5. Log to audit trail (always)
```

### 5.3 Skill Content Structure

Each Fate skill provides domain-specific guidance:

**`.claude/skills/moirai/clotho.md`** (Creation):
- Sprint creation validation rules
- Sprint dependency checking
- Schema requirements for SPRINT_CONTEXT.md
- Example create/start patterns

**`.claude/skills/moirai/lachesis.md`** (Measurement):
- Task completion validation
- Phase transition rules
- Field update constraints
- Park/resume state machine
- Handoff recording format

**`.claude/skills/moirai/atropos.md`** (Termination):
- Wrap session quality gates
- White Sails generation algorithm
- Sprint archival procedure
- Emergency override handling

### 5.4 Lazy Loading

Skills are loaded ONLY when needed:

```
Input: "mark_complete task-001 artifact=docs/PRD.md"

1. Parse: operation = mark_complete
2. Lookup: mark_complete -> Lachesis domain
3. Load: Skill(moirai/lachesis)  <- ONLY loaded now
4. Execute: Apply lachesis guidance for mark_complete
5. Return: Structured JSON response
```

This keeps the base Moirai prompt compact (<50 lines) while allowing rich domain logic via skills.

---

## 6. CLI Integration

### 6.1 CLI Command Mapping

| Moirai Operation | CLI Command | Moirai's Role |
|------------------|-------------|---------------|
| `create_session` | `ari session create "{initiative}" {complexity} {rite}` | Parse input, invoke CLI, return JSON |
| `park_session` | `ari session park --reason="{reason}"` | Validate, invoke CLI, return JSON |
| `resume_session` | `ari session resume` | Validate state, invoke CLI, return JSON |
| `wrap_session` | `ari session wrap [--force]` | Quality gate, invoke CLI, return JSON |
| `transition_phase` | `ari session transition --to={phase}` | Validate, invoke CLI, return JSON |
| `handoff` | `ari handoff execute --to={agent} --artifact={artifact}` | Validate, invoke CLI, return JSON |
| `generate_sails` | (computed during `ari session wrap`) | - |

### 6.2 CLI Bypass for Moirai

When Moirai mutates `*_CONTEXT.md` files directly (for sprint operations), it must bypass the write guard:

```bash
# Set bypass flag for Moirai's own writes
export MOIRAI_BYPASS=true
```

The write guard hook checks this environment variable:

```bash
# In session-write-guard.sh
if [[ "${MOIRAI_BYPASS:-}" == "true" ]]; then
    # Allow write from Moirai
    echo '{"decision": "allow", "reason": "MOIRAI_BYPASS set"}'
    exit 0
fi
```

### 6.3 CLI Output Parsing

Moirai parses CLI JSON output to construct responses:

```bash
# Execute CLI command
result=$(/Users/tomtenuta/Code/roster/ari session park --reason="$reason" -o json)
exit_code=$?

# Parse and wrap in Moirai response format
if [[ $exit_code -eq 0 ]]; then
    # Success - extract fields from CLI output
    # Return Moirai-formatted JSON
else
    # Error - extract error from CLI output
    # Return Moirai error JSON
fi
```

### 6.4 CLI Error Handling

| CLI Exit Code | Meaning | Moirai Error Code |
|---------------|---------|-------------------|
| 0 | Success | (none) |
| 1 | General error | `CLI_ERROR` |
| 2 | Validation error | `SCHEMA_VIOLATION` or `LIFECYCLE_VIOLATION` |
| 3 | Lock timeout | `LOCK_TIMEOUT` |
| 4 | File not found | `FILE_NOT_FOUND` |

---

## 7. Audit Protocol

### 7.1 Audit Log Location

```
.claude/sessions/.audit/session-mutations.log
```

### 7.2 Audit Log Format

```
TIMESTAMP | SESSION_ID | OPERATION | SOURCE | DETAILS | STATUS | FATE | reasoning="..."
```

**Example entries**:
```
2026-01-07T23:00:00Z | session-abc123 | mark_complete | moirai | task=task-001, artifact=docs/PRD.md | SUCCESS | lachesis | reasoning="Task marked complete per request"
2026-01-07T23:01:00Z | session-abc123 | wrap_session | moirai | sails=WHITE | SUCCESS | atropos | reasoning="All work complete, quality gate passed"
2026-01-07T23:02:00Z | session-abc123 | park_session | moirai | reason="Taking break" | SUCCESS | lachesis | reasoning="User requested park"
```

### 7.3 Audit Requirements

1. **Every mutation logged**: No silent mutations, even with `--emergency`
2. **Reasoning required**: All entries include reasoning field for debugging
3. **Fate attribution**: Even though unified, logs preserve Fate domain for traceability
4. **Before/after state**: For state transitions, log both states

### 7.4 Audit Logging Implementation

```bash
log_mutation() {
    local session_id="$1"
    local operation="$2"
    local details="$3"
    local status="$4"
    local fate="$5"
    local reasoning="$6"

    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local log_file=".claude/sessions/.audit/session-mutations.log"
    mkdir -p "$(dirname "$log_file")"

    echo "$timestamp | $session_id | $operation | moirai | $details | $status | $fate | reasoning=\"$reasoning\"" >> "$log_file"
}
```

---

## 8. Slash Command Integration

### 8.1 Slash Command to Moirai Routing

Slash commands internally invoke Moirai via Task tool:

**`/park "reason"`** -> `Task(moirai, "park_session reason=\"{reason}\"\n\nSession Context:\n- Session ID: {id}\n- Session Path: {path}")`

**`/wrap`** -> `Task(moirai, "wrap_session\n\nSession Context:\n- Session ID: {id}\n- Session Path: {path}")`

**`/handoff agent`** -> `Task(moirai, "handoff to={agent}\n\nSession Context:\n- Session ID: {id}\n- Session Path: {path}")`

### 8.2 Slash Command Skill Updates

The following skills require updates to route through Moirai:

| Skill | Current Implementation | New Implementation |
|-------|------------------------|-------------------|
| `user-commands/session/park.md` | `session-manager.sh mutate park` | `Task(moirai, "park_session...")` |
| `user-commands/session/wrap.md` | `session-manager.sh mutate wrap` | `Task(moirai, "wrap_session...")` |
| `user-commands/session/handoff.md` | `session-manager.sh mutate handoff` | `Task(moirai, "handoff...")` |

### 8.3 Session Context Injection

Slash command skills obtain session context before invoking Moirai:

```bash
# Get current session
session_id=$(/Users/tomtenuta/Code/roster/ari session status -o json | jq -r '.session_id')
session_path=".claude/sessions/$session_id/SESSION_CONTEXT.md"

# Build Moirai invocation
cat <<EOF
Task(moirai, "park_session reason=\"$reason\"

Session Context:
- Session ID: $session_id
- Session Path: $session_path")
EOF
```

---

## 9. Lock Protocol

### 9.1 Lock Acquisition

Moirai follows the locking protocol from `moirai-shared.md`:

```bash
LOCK_DIR=".claude/sessions/${SESSION_ID}/.locks"
LOCK_FILE="${LOCK_DIR}/context.lock"
mkdir -p "$LOCK_DIR"

# Atomic lock attempt
if mkdir "$LOCK_FILE" 2>/dev/null; then
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "${LOCK_FILE}/timestamp"
    # Lock acquired
else
    # Check for stale lock or wait
fi
```

### 9.2 Lock Timeout

- Default timeout: 10 seconds
- On timeout: Return `LOCK_TIMEOUT` error
- Stale lock (>60s): Force-release with warning in audit log

### 9.3 Lock Release

```bash
rm -rf "$LOCK_FILE"
```

### 9.4 Lock Scope

| Operation Type | Lock Required |
|----------------|---------------|
| CLI-backed (park, resume, wrap, transition) | CLI handles locking |
| Direct file mutation (sprint ops) | Moirai must acquire lock |
| Read-only (status queries) | No lock needed |

---

## 10. Error Handling

### 10.1 Error Categories

| Category | Examples | Recovery Strategy |
|----------|----------|-------------------|
| Input Errors | INVALID_OPERATION, AMBIGUOUS_INPUT | Return helpful error with valid options |
| State Errors | LIFECYCLE_VIOLATION, DEPENDENCY_BLOCKED | Return current state and allowed transitions |
| Concurrency Errors | LOCK_TIMEOUT, CONCURRENT_MODIFICATION | Suggest retry after delay |
| Quality Gate Errors | QUALITY_GATE_FAILED | Suggest --emergency or fix blockers |
| System Errors | FILE_NOT_FOUND, PERMISSION_DENIED | Return system error with recovery hint |

### 10.2 Partial Write Recovery

For direct file mutations, Moirai implements backup/rollback:

```bash
# 1. Create backup
cp "$ctx_file" "$ctx_file.backup.$$"

# 2. Perform mutation
edit_file "$ctx_file" "$mutation"

# 3. Validate result
if ! validate_schema "$ctx_file"; then
    mv "$ctx_file.backup.$$" "$ctx_file"
    return SCHEMA_VIOLATION
fi

# 4. Success - remove backup
rm -f "$ctx_file.backup.$$"
```

### 10.3 Error Response Examples

**INVALID_OPERATION**:
```json
{
  "success": false,
  "operation": "unknown_op",
  "error_code": "INVALID_OPERATION",
  "message": "Unknown operation: 'unknown_op'",
  "reasoning": "Operation not recognized in Moirai operation table",
  "hint": "Valid operations: create_session, create_sprint, start_sprint, mark_complete, transition_phase, update_field, park_session, resume_session, handoff, record_decision, append_content, wrap_session, generate_sails, delete_sprint"
}
```

**LIFECYCLE_VIOLATION**:
```json
{
  "success": false,
  "operation": "park_session",
  "error_code": "LIFECYCLE_VIOLATION",
  "message": "Cannot park ARCHIVED session",
  "reasoning": "Session is in terminal ARCHIVED state, no transitions allowed",
  "hint": "ARCHIVED sessions are immutable. Create a new session to continue work."
}
```

---

## 11. Test Strategy

### 11.1 Unit Tests

Location: `tests/unit/moirai-agent.bats`

| Test ID | Description | Validates |
|---------|-------------|-----------|
| `moirai_001` | Parse structured command syntax | Operation Parser |
| `moirai_002` | Parse natural language input | Operation Parser |
| `moirai_003` | Route creation ops to Clotho skill | Fate Skill Loader |
| `moirai_004` | Route measurement ops to Lachesis skill | Fate Skill Loader |
| `moirai_005` | Route termination ops to Atropos skill | Fate Skill Loader |
| `moirai_006` | Execute CLI for session operations | CLI Executor |
| `moirai_007` | Return JSON error for invalid operation | Error Handling |
| `moirai_008` | Log all mutations to audit trail | Audit Logger |
| `moirai_009` | Include reasoning in all responses | Response Format |
| `moirai_010` | Handle dry-run flag | Control Flags |

### 11.2 Integration Tests

Location: `tests/integration/moirai-integration.bats`

| Test ID | Description | Validates |
|---------|-------------|-----------|
| `int_001` | Full session lifecycle (create -> work -> park -> resume -> wrap) | End-to-end flow |
| `int_002` | Sprint lifecycle within session | Sprint operations |
| `int_003` | Slash command /park invokes Moirai | Slash command routing |
| `int_004` | Slash command /wrap generates sails | White Sails integration |
| `int_005` | Concurrent operations are serialized | Lock protocol |
| `int_006` | Write guard blocks direct writes | Hook integration |
| `int_007` | MOIRAI_BYPASS allows Moirai writes | Bypass mechanism |

### 11.3 Example Test Cases

```bash
@test "moirai_001: Parse structured command syntax" {
    local input="mark_complete task-001 artifact=docs/PRD.md"

    run parse_operation "$input"

    [ "$status" -eq 0 ]
    [[ "$output" == *'"operation": "mark_complete"'* ]]
    [[ "$output" == *'"task_id": "task-001"'* ]]
    [[ "$output" == *'"artifact": "docs/PRD.md"'* ]]
}

@test "moirai_003: Route creation ops to Clotho skill" {
    local input="create_sprint name=\"Implementation\""

    run determine_fate "$input"

    [ "$status" -eq 0 ]
    [ "$output" = "clotho" ]
}

@test "int_001: Full session lifecycle" {
    # Create session
    run Task moirai "create_session initiative=\"Test\" complexity=MODULE"
    [ "$status" -eq 0 ]
    [[ "$output" == *'"success": true'* ]]
    local session_id=$(echo "$output" | jq -r '.session_id')

    # Park session
    run Task moirai "park_session reason=\"Break\"\n\nSession Context:\n- Session ID: $session_id"
    [ "$status" -eq 0 ]
    [[ "$output" == *'"success": true'* ]]

    # Resume session
    run Task moirai "resume_session\n\nSession Context:\n- Session ID: $session_id"
    [ "$status" -eq 0 ]
    [[ "$output" == *'"success": true'* ]]

    # Wrap session
    run Task moirai "wrap_session\n\nSession Context:\n- Session ID: $session_id"
    [ "$status" -eq 0 ]
    [[ "$output" == *'"success": true'* ]]
    [[ "$output" == *'"sails_color"'* ]]
}
```

---

## 12. Handoff Criteria

The TDD is ready for Principal Engineer when:

- [x] Architecture diagram shows unified agent with skill loading
- [x] All 14 operations documented with syntax
- [x] Input/output contract specified with examples
- [x] State machine defined for session and sprint
- [x] CLI command mapping complete
- [x] Skill loading protocol documented
- [x] Lock protocol from moirai-shared.md preserved
- [x] Audit protocol with reasoning field specified
- [x] Error handling with all error codes
- [x] Test strategy with unit and integration tests
- [x] Slash command integration documented

### Implementation Order

1. **Phase 1**: Create unified agent file `.claude/agents/moirai.md`
   - Operation parser (structured + natural language)
   - Fate domain routing table
   - Basic response formatting

2. **Phase 2**: Create Fate skills
   - `.claude/skills/moirai/clotho.md` (creation)
   - `.claude/skills/moirai/lachesis.md` (measurement)
   - `.claude/skills/moirai/atropos.md` (termination)

3. **Phase 3**: CLI integration
   - Execute `ari` commands for session operations
   - Parse CLI output into Moirai responses
   - Handle CLI errors

4. **Phase 4**: Direct file mutations
   - Lock acquisition for sprint operations
   - Backup/rollback for file mutations
   - Schema validation

5. **Phase 5**: Audit logging
   - Log format implementation
   - Reasoning capture
   - Fate attribution

6. **Phase 6**: Slash command updates
   - Update park.md, wrap.md, handoff.md skills
   - Test transparent routing

---

## 13. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing Task(clotho/lachesis/atropos) calls | High | High | Document breaking change; update all references |
| Skill loading adds latency | Medium | Low | Skills are compact (<100 lines); measure latency |
| CLI output format changes | Low | Medium | Pin ari version; add output parsing tests |
| Concurrent sprint mutations corrupt state | Medium | High | Lock protocol from moirai-shared.md |
| Write guard not bypassed correctly | Medium | High | Integration test for MOIRAI_BYPASS |

---

## 14. ADRs

| ADR | Topic | Status |
|-----|-------|--------|
| ADR-0005 | state-mate Centralized State Authority | Accepted (Moirai supersedes state-mate) |
| (new) | Moirai Consolidation Architecture | To be created |

**Proposed ADR**: Document the decision to consolidate 4 agents into 1 with skill-based progressive disclosure, including the tradeoff of breaking backward compatibility for simpler architecture.

---

## 15. File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-moirai-unified-agent.md` | Created |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-moirai-consolidation.md` | Read |
| Current Router | `/Users/tomtenuta/Code/roster/user-agents/moirai.md` | Read |
| Shared Infrastructure | `/Users/tomtenuta/Code/roster/user-agents/moirai-shared.md` | Read |
| Clotho Agent | `/Users/tomtenuta/Code/roster/user-agents/clotho.md` | Read |
| Lachesis Agent | `/Users/tomtenuta/Code/roster/user-agents/lachesis.md` | Read |
| Atropos Agent | `/Users/tomtenuta/Code/roster/user-agents/atropos.md` | Read |
| Session FSM TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-session-state-machine.md` | Read |

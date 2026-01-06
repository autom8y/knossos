# TDD: Orchestrator Entry Pattern Codification

## Overview

This Technical Design Document specifies the implementation of the Orchestrator Entry Pattern, which ensures that `/start`, `/sprint`, and `/task` commands route through the Orchestrator agent when present in the active team. The design formalizes hook-triggered state mutations (via state-mate) rather than direct Task calls from the main agent.

## Context

| Reference | Location |
|-----------|----------|
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-orchestrator-entry-pattern.md` |
| base_hooks.yaml | `/Users/tomtenuta/Code/roster/.claude/hooks/base_hooks.yaml` |
| session-manager.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` |
| session-fsm.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-fsm.sh` |
| execution-mode.md | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/execution-mode.md` |
| orchestrator.md | `/Users/tomtenuta/Code/roster/.claude/agents/orchestrator.md` |
| response-format.md | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/response-format.md` |

### Problem Statement

When users invoke `/start`, `/sprint`, or `/task`, the main agent frequently:

1. Manages SESSION_CONTEXT mutations directly instead of delegating through state-mate
2. Attempts to orchestrate work itself rather than consulting the Orchestrator first
3. Invokes state-mate via explicit `Task(state-mate, ...)` calls rather than having hooks trigger it

This blurs the separation between:
- **Orchestrator**: Routing and phase coordination (stateless advisor)
- **state-mate**: State mutations (triggered by hooks, not orchestration)

### Design Goals

1. Route `/start`, `/sprint`, `/task` commands through Orchestrator when present
2. Hook-triggered state-mate invocation (not direct Task calls)
3. Workflow-aware error messages from session-write-guard.sh
4. Schema extension for `state_update.trigger_hooks` in CONSULTATION_RESPONSE
5. Validation hook for orchestrator bypass detection (warn-only)
6. Entry pattern documentation in skill file

### Requirements Coverage

| Requirement | Description | Addressed In |
|-------------|-------------|--------------|
| FR-1 | UserPromptSubmit hook extension | Hook Event Flow Design |
| FR-2 | Hook-triggered state-mate for session creation | Session Creation Flow |
| FR-3 | session-write-guard.sh workflow-aware errors | Write Guard Extension |
| FR-4 | state_update.trigger_hooks field | Schema Extension |
| FR-5 | Entry pattern documentation | Skill File Specification |
| FR-6 | PreToolUse validation hook (warn-only) | Bypass Detection Hook |

---

## System Design

### Architecture Diagram

```
User invokes /start, /sprint, /task
         |
         v
+------------------------------------------+
| UserPromptSubmit: orchestrator-router.sh |
| (NEW - priority 5, before start-preflight)|
+------------------------------------------+
         |
         | Detects command + checks for orchestrator
         v
+------------------------------------------+
| Inject context:                          |
| - ORCHESTRATOR_ROUTING_REQUIRED: true    |
| - CONSULTATION_REQUEST template          |
+------------------------------------------+
         |
         v
+------------------------------------------+
| Main Agent receives augmented prompt     |
| with routing directive                   |
+------------------------------------------+
         |
         v (Main agent consults Orchestrator)
+------------------------------------------+
| Orchestrator returns CONSULTATION_RESPONSE|
| with state_update.trigger_hooks: true    |
+------------------------------------------+
         |
         v (Main agent executes directive)
+------------------------------------------+
| PostToolUse hooks detect state events:   |
| - Session creation needed                |
| - Phase transition needed                |
| - Artifact registration needed           |
+------------------------------------------+
         |
         v (Hooks invoke state-mate)
+------------------------------------------+
| state-mate applies mutations             |
| Audit log shows trigger_source: "hook"   |
+------------------------------------------+
```

### Components

| Component | Responsibility | Location | Status |
|-----------|---------------|----------|--------|
| **orchestrator-router.sh** | Detect workflow commands, inject orchestrator routing context | `.claude/hooks/bin/orchestrator-router.sh` | NEW |
| **orchestrator-bypass-check.sh** | Warn when main agent skips orchestrator consultation | `.claude/hooks/bin/orchestrator-bypass-check.sh` | NEW |
| **session-write-guard.sh** | Block direct writes with workflow-aware messages | `.claude/hooks/bin/session-write-guard.sh` | MODIFY |
| **base_hooks.yaml** | Register new hooks | `.claude/hooks/base_hooks.yaml` | MODIFY |
| **response-format.md** | Document state_update.trigger_hooks field | `.claude/skills/orchestration/response-format.md` | MODIFY |
| **entry-pattern.md** | Document entry pattern with anti-patterns | `.claude/skills/orchestration/entry-pattern.md` | NEW |

---

## Hook Event Flow Design

### Hook Execution Order

The hook priority system (lower = earlier) determines execution order:

```
UserPromptSubmit Event for "/start foo":

Priority 5:  orchestrator-router.sh    (NEW - routing context injection)
Priority 10: start-preflight.sh        (existing - session validation)

Result: Routing context injected BEFORE preflight validation
```

### orchestrator-router.sh Specification

**Purpose**: Detect `/start`, `/sprint`, `/task` commands and inject orchestrator routing context when an orchestrator agent is present in the active team.

**Event**: UserPromptSubmit
**Matcher**: `^/(start|sprint|task)`
**Priority**: 5 (before start-preflight.sh at priority 10)
**Timeout**: 5 seconds

**Input (stdin JSON)**:
```json
{
  "prompt": "/start Add dark mode toggle"
}
```

**Output (stdout)**:
```json
{
  "continue": true,
  "prefix": "## Orchestrator Routing Required\n\nThis command requires consultation with the Orchestrator before proceeding...\n\nCONSULTATION_REQUEST:\n```yaml\ntype: initial\ninitiative:\n  name: \"Add dark mode toggle\"\n  complexity: \"MODULE\"\nstate:\n  current_phase: null\n  completed_phases: []\n  artifacts_produced: []\ncontext_summary: |\n  User invoked /start for new initiative.\n```\n\n**IMPORTANT**: Invoke the orchestrator with this CONSULTATION_REQUEST before any specialist work."
}
```

**Algorithm**:

```bash
#!/bin/bash
# orchestrator-router.sh - Inject orchestrator routing context for workflow commands

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"

# Source utilities
source "$SCRIPT_DIR/../lib/session-utils.sh" 2>/dev/null || true

# Read input
input=$(cat)
prompt=$(echo "$input" | jq -r '.prompt // ""')

# Check if this is a workflow command
if ! echo "$prompt" | grep -qE '^/(start|sprint|task)'; then
    echo '{"continue": true}'
    exit 0
fi

# Check if orchestrator is present in active team
has_orchestrator() {
    local agents_dir="$PROJECT_DIR/.claude/agents"
    [[ -f "$agents_dir/orchestrator.md" ]]
}

if ! has_orchestrator; then
    # No orchestrator = direct execution is valid
    echo '{"continue": true}'
    exit 0
fi

# Extract command and initiative from prompt
command=$(echo "$prompt" | grep -oE '^/(start|sprint|task)' | tr -d '/')
initiative=$(echo "$prompt" | sed -E 's|^/(start|sprint|task)\s*||')

# Determine request type based on command
case "$command" in
    start)
        request_type="initial"
        ;;
    sprint|task)
        request_type="checkpoint"
        ;;
esac

# Build routing context
cat <<EOF
{
  "continue": true,
  "prefix": "## Orchestrator Routing Required\n\nThis $command command requires consultation with the Orchestrator before proceeding.\n\n### CONSULTATION_REQUEST\n\n\`\`\`yaml\ntype: $request_type\ninitiative:\n  name: \"$initiative\"\n  complexity: \"MODULE\"  # Assess and adjust\nstate:\n  current_phase: null\n  completed_phases: []\n  artifacts_produced: []\ncontext_summary: |\n  User invoked /$command. Assess complexity and determine phase sequence.\n\`\`\`\n\n**IMPORTANT**: Invoke the orchestrator via Task tool with this CONSULTATION_REQUEST before any specialist work. Let hooks handle state mutations - do not call state-mate directly.\n\n---\n\n"
}
EOF
```

### Hook Registration Update

**File**: `.claude/hooks/base_hooks.yaml`

**Change**: Add orchestrator-router.sh registration

```yaml
# ===========================================================================
# UserPromptSubmit Hooks
# ===========================================================================
# NEW: Orchestrator routing for workflow commands (priority 5, before preflight)
- event: UserPromptSubmit
  matcher: "^/(start|sprint|task)"
  path: orchestrator-router.sh
  timeout: 5
  priority: 5
  description: "Injects orchestrator routing context for workflow commands"

# Existing
- event: UserPromptSubmit
  matcher: "^/"
  path: start-preflight.sh
  timeout: 5
  priority: 10
  description: "Preflight check for slash commands"
```

---

## Session Creation Flow

### Design Decision: Hook-Triggered vs Direct Task

**Options Considered**:

1. **Option A: Dedicated SessionCreate hook event** - Add new hook event type
   - Rejected: Requires Claude Code changes we don't control

2. **Option B: PostToolUse detection of session creation intent**
   - Rejected: Too late in flow, creates race conditions

3. **Option C: UserPromptSubmit triggers session creation directly**
   - Selected: Aligns with existing patterns, minimal infrastructure change

**Selected Approach**: UserPromptSubmit hook (`start-preflight.sh`) already validates session state. Extend it to trigger session creation via `session-manager.sh` when:
- Command is `/start`
- No existing session
- Pre-flight validation passes

### start-preflight.sh Extension

**Current Behavior**: Validates session state, blocks if session exists

**New Behavior**: Additionally creates session via session-manager.sh when conditions met

**Algorithm Extension**:

```bash
# In start-preflight.sh, after validation passes for /start:

if [[ "$command" == "start" ]] && ! has_session; then
    # Parse initiative from prompt
    local initiative
    initiative=$(echo "$prompt" | sed 's|^/start\s*||')
    [[ -z "$initiative" ]] && initiative="Unnamed initiative"

    # Default complexity (main agent can adjust)
    local complexity="MODULE"

    # Get active team
    local team
    team=$(cat "$PROJECT_DIR/.claude/ACTIVE_RITE" 2>/dev/null || echo "10x-dev-pack")

    # Create session via session-manager (uses FSM for schema-validated write)
    local result
    result=$("$SCRIPT_DIR/../lib/session-manager.sh" create "$initiative" "$complexity" "$team")

    if echo "$result" | jq -e '.success == true' >/dev/null 2>&1; then
        local session_id
        session_id=$(echo "$result" | jq -r '.session_id')

        # Log as hook-triggered creation
        local audit_log="$PROJECT_DIR/.claude/sessions/.audit/session-mutations.log"
        mkdir -p "$(dirname "$audit_log")"
        echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) | $session_id | CREATE | HOOK | start-preflight.sh" >> "$audit_log"

        # Include session info in prefix
        prefix="## Session Created (Hook-Triggered)\n\nSession ID: $session_id\nInitiative: $initiative\nComplexity: $complexity\n\n$prefix"
    else
        local error
        error=$(echo "$result" | jq -r '.error // "Unknown error"')
        # Include error in output but continue
        prefix="## Session Creation Warning\n\nCould not create session: $error\n\n$prefix"
    fi
fi
```

### Audit Trail Differentiation

**Requirement**: Audit log must indicate whether state mutation was hook-triggered or direct

**Implementation**: All hook-triggered state mutations include `trigger_source` field:

```
# Audit log format
TIMESTAMP | SESSION_ID | OPERATION | TRIGGER_SOURCE | DETAIL

# Examples
2024-01-15T10:30:00Z | session-20240115-103000-abc123 | CREATE | hook | start-preflight.sh
2024-01-15T10:35:00Z | session-20240115-103000-abc123 | PHASE_TRANSITION | hook | artifact-tracker.sh
2024-01-15T10:40:00Z | session-20240115-103000-abc123 | PARK | direct | Task(state-mate)
```

---

## State Update Schema Extension

### Current CONSULTATION_RESPONSE state_update Schema

```yaml
state_update:
  current_phase: string
  next_phases: string[]
  routing_rationale: string
```

### Extended Schema with trigger_hooks

```yaml
state_update:
  current_phase: string
  next_phases: string[]
  routing_rationale: string
  # NEW: Signal to main agent that hooks should handle state mutations
  trigger_hooks: boolean  # default: true when orchestrator is present
  # NEW: Expected state transitions for hook coordination
  expected_transitions:
    - type: "session_state" | "phase" | "artifact"
      from: string | null
      to: string
      artifact_path: string | null  # only for type: artifact
```

### Field Definitions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `trigger_hooks` | boolean | No | true | When true, main agent should let hooks handle state mutations |
| `expected_transitions` | array | No | [] | State changes the orchestrator expects to occur |
| `expected_transitions[].type` | enum | Yes | - | Type of state change: session_state, phase, artifact |
| `expected_transitions[].from` | string | No | null | Current state (null for creation) |
| `expected_transitions[].to` | string | Yes | - | Target state |
| `expected_transitions[].artifact_path` | string | No | null | For artifact type, expected file path |

### Example CONSULTATION_RESPONSE with trigger_hooks

```yaml
directive:
  action: invoke_specialist
specialist:
  name: requirements-analyst
  prompt: |
    # Context
    New initiative: Add dark mode toggle. User wants theming support.

    # Task
    Create PRD with user stories and acceptance criteria.

    # Deliverable
    docs/requirements/PRD-dark-mode.md

    # Handoff Criteria
    - [ ] User stories cover all theme scenarios
    - [ ] Acceptance criteria are testable
state_update:
  current_phase: requirements
  next_phases: [design, implementation, validation]
  routing_rationale: "Initial phase - requirements gathering needed first"
  trigger_hooks: true
  expected_transitions:
    - type: phase
      from: null
      to: requirements
    - type: artifact
      to: registered
      artifact_path: docs/requirements/PRD-dark-mode.md
throughline:
  decision: "Route to requirements-analyst for PRD creation"
  rationale: "New initiative, complexity MODULE, requires formal requirements"
```

### Validation Rules

1. `trigger_hooks` must be boolean if present
2. `expected_transitions` array elements must have valid `type` enum
3. `expected_transitions[].to` is required
4. `expected_transitions[].artifact_path` required when type is "artifact"

---

## session-write-guard.sh Extension

### Current Behavior

Blocks direct writes to `*_CONTEXT.md` files with message:
```
State mutations are handled by state-mate. Use Task(state-mate, "...")
```

### New Behavior: Workflow-Aware Messages

Detect whether an active workflow is present and adjust message:

```bash
#!/bin/bash
# session-write-guard.sh - Guard SESSION_CONTEXT writes

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"

# Source utilities
source "$SCRIPT_DIR/../lib/session-utils.sh" 2>/dev/null || true

# Read input
input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // .tool_input.path // ""')

# Check if this is a session context file
if ! echo "$file_path" | grep -qE '(SESSION|SPRINT)_CONTEXT\.md$'; then
    echo '{"continue": true}'
    exit 0
fi

# Check for active workflow
has_active_workflow() {
    local session_dir
    session_dir=$(get_session_dir 2>/dev/null || echo "")
    [[ -z "$session_dir" ]] && return 1

    local ctx_file="$session_dir/SESSION_CONTEXT.md"
    [[ ! -f "$ctx_file" ]] && return 1

    # Check workflow.active field or infer from current_phase
    grep -qE "^(workflow_active:|current_phase:)" "$ctx_file" 2>/dev/null
}

# Check if orchestrator is present
has_orchestrator() {
    [[ -f "$PROJECT_DIR/.claude/agents/orchestrator.md" ]]
}

# Build appropriate error message
if has_active_workflow && has_orchestrator; then
    # Active workflow with orchestrator = hooks handle state
    message="## State Mutation Blocked

State mutations are handled **automatically by hooks** during active workflows.

**Why?** The orchestrator coordinates phase transitions, and hooks invoke state-mate to maintain the audit trail.

**If you need an explicit mutation**, use the appropriate command:
- \`/park\` - Pause current session
- \`/wrap\` - Complete and archive session
- \`/handoff\` - Transfer to another agent

**Do not** call \`Task(state-mate, ...)\` directly during orchestrated workflows."

else
    # No workflow or no orchestrator = suggest state-mate
    message="## State Mutation Blocked

Direct writes to \`*_CONTEXT.md\` files are not allowed.

**Use state-mate for all session/sprint mutations:**

\`\`\`
Task(state-mate, \"<your mutation request>\")
\`\`\`

**Examples:**
- \`Task(state-mate, \"mark task-001 complete\")\`
- \`Task(state-mate, \"transition to design phase\")\`
- \`Task(state-mate, \"register artifact docs/PRD-foo.md\")\`

See \`~/.claude/agents/state-mate.md\` for full documentation (synced from roster/user-agents/)."

fi

# Block the operation
cat <<EOF
{
  "continue": false,
  "decision": "block",
  "reason": "$message"
}
EOF
```

---

## Orchestrator Bypass Detection Hook

### Purpose

Detect when main agent attempts to invoke specialists without prior Orchestrator consultation. Warn-only to avoid breaking orchestrator-less workflows.

### orchestrator-bypass-check.sh Specification

**Event**: PreToolUse
**Matcher**: Task
**Priority**: 20 (after other PreToolUse guards)
**Timeout**: 3 seconds

**Input (stdin JSON)**:
```json
{
  "tool_name": "Task",
  "tool_input": {
    "task": "requirements-analyst",
    "prompt": "Create PRD for dark mode..."
  }
}
```

**Output**:
```json
{
  "continue": true,
  "message": "## Warning: Orchestrator Consultation Recommended\n\nYou are invoking a specialist (requirements-analyst) without prior orchestrator consultation..."
}
```

**Algorithm**:

```bash
#!/bin/bash
# orchestrator-bypass-check.sh - Warn on orchestrator bypass

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"

# Read input
input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name // ""')

# Only check Task tool invocations
if [[ "$tool_name" != "Task" ]]; then
    echo '{"continue": true}'
    exit 0
fi

# Get the agent being invoked
agent=$(echo "$input" | jq -r '.tool_input.task // .tool_input.agent // ""')

# Skip if invoking orchestrator itself
if [[ "$agent" == "orchestrator" ]]; then
    echo '{"continue": true}'
    exit 0
fi

# Skip if no orchestrator in team
if [[ ! -f "$PROJECT_DIR/.claude/agents/orchestrator.md" ]]; then
    echo '{"continue": true}'
    exit 0
fi

# Check for active workflow
has_active_workflow() {
    local session_dir
    session_dir=$(get_session_dir 2>/dev/null || echo "")
    [[ -z "$session_dir" ]] && return 1
    [[ -f "$session_dir/SESSION_CONTEXT.md" ]]
}

if ! has_active_workflow; then
    # No workflow = no orchestration required
    echo '{"continue": true}'
    exit 0
fi

# Check session for recent orchestrator consultation
# Look for orchestrator consultation marker in last 5 minutes
check_recent_consultation() {
    local session_dir
    session_dir=$(get_session_dir 2>/dev/null || echo "")
    [[ -z "$session_dir" ]] && return 1

    local events_file="$session_dir/events.jsonl"
    [[ ! -f "$events_file" ]] && return 1

    # Check for orchestrator consultation event in last 5 minutes
    local five_min_ago
    five_min_ago=$(date -u -v-5M +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -d '5 minutes ago' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "1970-01-01T00:00:00Z")

    tail -20 "$events_file" 2>/dev/null | grep -q '"event":"ORCHESTRATOR_CONSULTED"' && return 0
    return 1
}

if check_recent_consultation; then
    # Recent consultation found - allow
    echo '{"continue": true}'
    exit 0
fi

# Emit warning (do not block)
cat <<EOF
{
  "continue": true,
  "message": "## Warning: Orchestrator Consultation Recommended

You are invoking specialist **$agent** without recent orchestrator consultation.

**Best Practice**: During active workflows, consult the orchestrator first:

\`\`\`
Task(orchestrator, \"CONSULTATION_REQUEST with current state...\")
\`\`\`

Then invoke specialists based on the orchestrator's directive.

*This is a warning only - proceeding with specialist invocation.*"
}
EOF
```

### Hook Registration

**File**: `.claude/hooks/base_hooks.yaml`

```yaml
# NEW: Orchestrator bypass detection (warn-only)
- event: PreToolUse
  matcher: "Task"
  path: orchestrator-bypass-check.sh
  timeout: 3
  priority: 20
  description: "Warns when invoking specialists without orchestrator consultation"
```

---

## Entry Pattern Documentation

### New Skill File: entry-pattern.md

**Location**: `.claude/skills/orchestration/entry-pattern.md`

```markdown
# Entry Pattern

> How /start, /sprint, /task commands route through the system

## The Pattern

When an orchestrator is present in the active team:

```
User: /start Add dark mode toggle
         |
         v
+----------------------------------+
| Hook: orchestrator-router.sh     |
| Injects CONSULTATION_REQUEST     |
+----------------------------------+
         |
         v
+----------------------------------+
| Hook: start-preflight.sh         |
| Creates session (hook-triggered) |
+----------------------------------+
         |
         v
+----------------------------------+
| Main Agent sees routing context  |
| Knows to consult orchestrator    |
+----------------------------------+
         |
         v
+----------------------------------+
| Task(orchestrator, request)      |
| Orchestrator returns directive   |
+----------------------------------+
         |
         v
+----------------------------------+
| Main Agent executes directive    |
| Task(specialist, prompt)         |
+----------------------------------+
         |
         v
+----------------------------------+
| Hooks detect state events        |
| Auto-invoke state-mate           |
+----------------------------------+
```

## When Orchestrator is Absent

Teams without an orchestrator agent are valid and use direct execution:

```
User: /start Add dark mode toggle
         |
         v
+----------------------------------+
| Hook: start-preflight.sh         |
| Creates session directly         |
+----------------------------------+
         |
         v
+----------------------------------+
| Main Agent routes to specialist  |
| (no orchestrator consultation)   |
+----------------------------------+
         |
         v
+----------------------------------+
| Task(state-mate, ...)           |
| Direct state mutations allowed   |
+----------------------------------+
```

## Hook-Triggered vs Direct state-mate

| Scenario | state-mate Invocation | Audit Log |
|----------|----------------------|-----------|
| Orchestrator present, active workflow | Hook-triggered | trigger_source: hook |
| Orchestrator present, no workflow | Direct allowed | trigger_source: direct |
| Orchestrator absent | Direct allowed | trigger_source: direct |
| Emergency override | Direct with flag | trigger_source: emergency |

## Anti-Patterns

### DO NOT: Invoke state-mate directly during orchestrated workflows

```
# WRONG
Task(state-mate, "transition to design phase")

# RIGHT
Let the artifact-tracker.sh hook detect the PRD write
and trigger the phase transition automatically
```

### DO NOT: Skip orchestrator and go directly to specialist

```
# WRONG
Task(requirements-analyst, "Create PRD for dark mode")

# RIGHT
Task(orchestrator, "CONSULTATION_REQUEST for dark mode initiative")
# Then invoke specialist per orchestrator directive
```

### DO NOT: Manually write SESSION_CONTEXT.md

```
# WRONG
Edit(SESSION_CONTEXT.md, ...)

# RIGHT
Use /park, /wrap, or let hooks handle mutations
```

## state_update.trigger_hooks

When orchestrator returns `state_update.trigger_hooks: true`:

1. Main agent should NOT call state-mate directly
2. Hooks will detect relevant events and invoke state-mate
3. Audit trail shows hook-triggered mutations

When `trigger_hooks: false` (or absent):

1. Direct state-mate calls are acceptable
2. Typically for orchestrator-less teams

## Fallback Behavior

If hooks fail or are disabled, the system degrades gracefully:

| Failure | Behavior |
|---------|----------|
| orchestrator-router.sh fails | start-preflight.sh continues, no routing context |
| start-preflight.sh fails | Main agent can create session manually |
| orchestrator-bypass-check.sh fails | Bypass check skipped, specialist invoked |
| state-mate hook fails | Logged as error, session continues in degraded mode |

## See Also

- [execution-mode.md](execution-mode.md) - When to delegate vs execute
- [consultation-loop.md](consultation-loop.md) - The consultation pattern
- [command-integration.md](command-integration.md) - How commands use the loop
```

---

## Backward Compatibility Classification

### Component-by-Component Analysis

| Component | Change Type | Compatibility | Migration Required |
|-----------|-------------|---------------|-------------------|
| **orchestrator-router.sh** | NEW | COMPATIBLE | No - additive hook |
| **orchestrator-bypass-check.sh** | NEW | COMPATIBLE | No - warn-only, non-blocking |
| **session-write-guard.sh** | MODIFY | COMPATIBLE | No - improved messages only |
| **base_hooks.yaml** | MODIFY | COMPATIBLE | No - additive registrations |
| **response-format.md** | MODIFY | COMPATIBLE | No - optional field added |
| **entry-pattern.md** | NEW | COMPATIBLE | No - documentation only |
| **start-preflight.sh** | MODIFY | COMPATIBLE | No - additive behavior |
| **state_update schema** | MODIFY | COMPATIBLE | No - new fields are optional |

### Existing Session Compatibility

| Session State | Expected Behavior |
|---------------|-------------------|
| V2 session, no workflow | Works unchanged |
| V2 session, active workflow | Enhanced with hook-triggered mutations |
| V1 session (legacy) | Auto-migrated per existing session-migrate.sh |
| No session | Works unchanged |

### Direct Task(state-mate) Compatibility

Direct invocation of state-mate remains valid as an escape hatch:

1. `--emergency` flag bypasses non-critical validations
2. Audit log differentiates: `trigger_source: direct`
3. Documentation notes when direct invocation is acceptable

---

## Integration Test Matrix

### Test Categories

| Category | Test Count | Priority |
|----------|------------|----------|
| Hook Execution Order | 3 | HIGH |
| Orchestrator Routing | 4 | HIGH |
| Session Creation Flow | 3 | HIGH |
| state_update Schema | 3 | MEDIUM |
| Write Guard Messages | 3 | MEDIUM |
| Bypass Detection | 4 | MEDIUM |
| Backward Compatibility | 4 | HIGH |

### Test Specifications

#### Hook Execution Order Tests

| Test ID | Description | Setup | Expected |
|---------|-------------|-------|----------|
| `oep_heo_001` | orchestrator-router fires before start-preflight | `/start foo` with orchestrator | Router output (priority 5) precedes preflight (priority 10) |
| `oep_heo_002` | Hooks fire in priority order | Multiple UserPromptSubmit hooks | Lower priority numbers execute first |
| `oep_heo_003` | Hook failure does not block subsequent hooks | First hook errors | Subsequent hooks still execute |

#### Orchestrator Routing Tests

| Test ID | Description | Setup | Expected |
|---------|-------------|-------|----------|
| `oep_or_001` | /start injects routing context when orchestrator present | orchestrator.md exists, `/start foo` | Prefix contains CONSULTATION_REQUEST |
| `oep_or_002` | /start skips routing when no orchestrator | No orchestrator.md, `/start foo` | No routing prefix injected |
| `oep_or_003` | /sprint injects routing context | orchestrator.md exists, `/sprint` | Prefix contains CONSULTATION_REQUEST |
| `oep_or_004` | /task injects routing context | orchestrator.md exists, `/task` | Prefix contains CONSULTATION_REQUEST |

#### Session Creation Flow Tests

| Test ID | Description | Setup | Expected |
|---------|-------------|-------|----------|
| `oep_sc_001` | /start creates session via hook | No existing session, `/start foo` | Session created, audit shows trigger_source: hook |
| `oep_sc_002` | /start with existing session blocked | Active session, `/start foo` | Error message, suggests /resume or /wrap |
| `oep_sc_003` | Session creation failure logged | Session creation fails | Warning in prefix, operation continues |

#### state_update Schema Tests

| Test ID | Description | Setup | Expected |
|---------|-------------|-------|----------|
| `oep_su_001` | trigger_hooks: true parsed correctly | CONSULTATION_RESPONSE with trigger_hooks | Main agent recognizes hook delegation |
| `oep_su_002` | expected_transitions parsed | CONSULTATION_RESPONSE with transitions | Transitions extracted for hook coordination |
| `oep_su_003` | Backward compatible without trigger_hooks | Legacy response format | Default behavior (hooks active) |

#### Write Guard Message Tests

| Test ID | Description | Setup | Expected |
|---------|-------------|-------|----------|
| `oep_wg_001` | Workflow-aware message with orchestrator | Active workflow + orchestrator, attempt write | Message mentions "hooks handle mutations" |
| `oep_wg_002` | Standard message without orchestrator | No orchestrator, attempt write | Message mentions "use state-mate" |
| `oep_wg_003` | Message suggests correct commands | Active workflow, attempt write | /park, /wrap, /handoff suggested |

#### Bypass Detection Tests

| Test ID | Description | Setup | Expected |
|---------|-------------|-------|----------|
| `oep_bd_001` | Warning on direct specialist invocation | Active workflow, Task(architect) | Warning message in output |
| `oep_bd_002` | No warning when invoking orchestrator | Active workflow, Task(orchestrator) | No warning |
| `oep_bd_003` | No warning without orchestrator in team | No orchestrator.md, Task(architect) | No warning |
| `oep_bd_004` | No warning after recent consultation | Recent orchestrator event, Task(architect) | No warning |

#### Backward Compatibility Tests

| Test ID | Description | Setup | Expected |
|---------|-------------|-------|----------|
| `oep_bc_001` | V1 sessions auto-migrate | V1 session format | Migrated to V2 on first access |
| `oep_bc_002` | Direct state-mate still works | No workflow, Task(state-mate) | Operation succeeds |
| `oep_bc_003` | Emergency override bypasses | --emergency flag | Operation proceeds with logging |
| `oep_bc_004` | Teams without orchestrator work | No orchestrator.md | Full functionality, direct execution |

### Satellite Diversity Coverage

| Satellite Type | Configuration | Tests Required |
|----------------|---------------|----------------|
| **Minimal** | No orchestrator, no workflow | oep_or_002, oep_wg_002, oep_bd_003 |
| **Standard** | 10x-dev-pack with orchestrator | All oep_or_*, oep_wg_001, oep_bd_001 |
| **Complex** | Ecosystem-pack, nested workflows | oep_sc_*, oep_su_*, oep_bc_* |
| **Legacy** | V1 session format | oep_bc_001 |

---

## Implementation Guidance

### Recommended Implementation Order

1. **Phase 1: Core Hooks (Priority: HIGH)**
   - Create `orchestrator-router.sh`
   - Update `base_hooks.yaml` with new registration
   - Test hook execution order

2. **Phase 2: Session Creation Flow (Priority: HIGH)**
   - Extend `start-preflight.sh` with hook-triggered session creation
   - Add audit trail differentiation
   - Test with various session states

3. **Phase 3: Write Guard Enhancement (Priority: MEDIUM)**
   - Modify `session-write-guard.sh` for workflow-aware messages
   - Test both message variants

4. **Phase 4: Schema Extension (Priority: MEDIUM)**
   - Update `response-format.md` with trigger_hooks documentation
   - Update orchestrator agent prompt (if needed for awareness)

5. **Phase 5: Bypass Detection (Priority: MEDIUM)**
   - Create `orchestrator-bypass-check.sh`
   - Register in base_hooks.yaml
   - Test warn-only behavior

6. **Phase 6: Documentation (Priority: HIGH)**
   - Create `entry-pattern.md` skill file
   - Update `execution-mode.md` with references
   - Update SKILL.md manifest

7. **Phase 7: Integration Testing (Priority: HIGH)**
   - Execute test matrix
   - Verify satellite diversity coverage
   - Document any edge case findings

### File Change Summary

| File | Action | Lines Changed (est.) |
|------|--------|---------------------|
| `.claude/hooks/bin/orchestrator-router.sh` | CREATE | ~80 |
| `.claude/hooks/bin/orchestrator-bypass-check.sh` | CREATE | ~100 |
| `.claude/hooks/bin/session-write-guard.sh` | MODIFY | ~50 |
| `.claude/hooks/bin/start-preflight.sh` | MODIFY | ~40 |
| `.claude/hooks/base_hooks.yaml` | MODIFY | ~20 |
| `.claude/skills/orchestration/response-format.md` | MODIFY | ~40 |
| `.claude/skills/orchestration/entry-pattern.md` | CREATE | ~200 |
| `.claude/skills/orchestration/SKILL.md` | MODIFY | ~5 |

---

## Performance Requirements

| Metric | Requirement | Measurement |
|--------|-------------|-------------|
| Hook execution latency | < 50ms per hook | Time from hook start to JSON output |
| Total routing overhead | < 200ms | Time from /start to main agent receiving context |
| Session creation time | < 100ms | Time for session-manager.sh create |
| Bypass check time | < 30ms | Time for consultation check |

### Performance Optimization Notes

1. **orchestrator-router.sh**: Minimize file I/O, use cached ACTIVE_RITE
2. **orchestrator-bypass-check.sh**: Only check last 20 events, not full history
3. **Session creation**: FSM already optimized with atomic writes
4. **JSON processing**: Use jq streaming where possible

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Hook execution order race | Low | High | Priority system enforces order; document dependencies |
| Main agent ignores routing context | Medium | Medium | Bypass check warns; training in prompts |
| Session creation timing | Low | Medium | Idempotent creation; locks prevent race |
| Backward compatibility breaks | Low | High | All changes additive; comprehensive test matrix |
| Performance regression | Low | Medium | 50ms timeout enforcement; monitoring |

---

## Open Items

None. All design decisions resolved.

---

## Artifact Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| TDD | `/Users/tomtenuta/Code/roster/docs/design/TDD-orchestrator-entry-pattern.md` | Created |
| PRD | `/Users/tomtenuta/Code/roster/docs/requirements/PRD-orchestrator-entry-pattern.md` | Read |
| base_hooks.yaml | `/Users/tomtenuta/Code/roster/.claude/hooks/base_hooks.yaml` | Read |
| session-manager.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-manager.sh` | Read |
| session-fsm.sh | `/Users/tomtenuta/Code/roster/.claude/hooks/lib/session-fsm.sh` | Read |
| execution-mode.md | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/execution-mode.md` | Read |
| orchestrator.md | `/Users/tomtenuta/Code/roster/.claude/agents/orchestrator.md` | Read |
| response-format.md | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/response-format.md` | Read |
| request-format.md | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/request-format.md` | Read |
| consultation-loop.md | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/consultation-loop.md` | Read |
| command-integration.md | `/Users/tomtenuta/Code/roster/.claude/skills/orchestration/command-integration.md` | Read |

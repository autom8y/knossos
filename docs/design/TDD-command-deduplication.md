---
title: "Context Design: Session-Lifecycle Command Pattern Deduplication"
type: context-design
complexity: MODULE
created_at: "2026-01-01T00:00:00Z"
status: ready-for-implementation
gap_analysis: N/A
affected_systems:
  - roster
author: context-architect
backward_compatible: true
migration_required: false
work_packages:
  - id: WP1
    name: "Create shared-sections Directory Structure"
    description: "Establish shared-sections/ directory with partial templates"
    files:
      - path: "user-skills/session-lifecycle/shared-sections/"
        action: create
        description: "Create directory for shared partials"
      - path: "user-skills/session-lifecycle/shared-sections/session-resolution.md"
        action: create
        description: "Session existence and state validation partial"
      - path: "user-skills/session-lifecycle/shared-sections/workflow-resolution.md"
        action: create
        description: "Team and agent validation partial"
      - path: "user-skills/session-lifecycle/shared-sections/state-mate-invocation.md"
        action: create
        description: "state-mate Task invocation partial"
    estimated_effort: "2 hours"
  - id: WP2
    name: "Update behavior.md Files to Reference Partials"
    description: "Refactor 5 behavior.md files to reference shared sections"
    files:
      - path: "user-skills/session-lifecycle/start-ref/behavior.md"
        action: modify
        description: "Add partial references, remove duplicated content"
      - path: "user-skills/session-lifecycle/park-ref/behavior.md"
        action: modify
        description: "Add partial references, remove duplicated content"
      - path: "user-skills/session-lifecycle/resume/behavior.md"
        action: modify
        description: "Add partial references, remove duplicated content"
      - path: "user-skills/session-lifecycle/wrap-ref/behavior.md"
        action: modify
        description: "Add partial references, remove duplicated content"
      - path: "user-skills/session-lifecycle/handoff-ref/behavior.md"
        action: modify
        description: "Add partial references, remove duplicated content"
    dependencies: [WP1]
    estimated_effort: "1 hour"
  - id: WP3
    name: "Create Partial Index"
    description: "Document shared-sections for discoverability"
    files:
      - path: "user-skills/session-lifecycle/shared-sections/INDEX.md"
        action: create
        description: "Index of available partials with usage guide"
    dependencies: [WP1]
    estimated_effort: "30 minutes"
schema_version: "1.0"
---

## Executive Summary

Five session-lifecycle skills (`start-ref`, `park-ref`, `resume`, `wrap-ref`, `handoff-ref`) contain duplicated patterns for session validation, workflow resolution, and state-mate invocation. This design extracts these patterns into a `shared-sections/` directory with referenceable partials, following the proven architecture pattern from `orchestrator-templates/shared-sections/`.

## Design Decisions

### Decision 1: Shared Section Location

**Options Considered**:
1. `session-common/shared-sections/` - Rejected: session-common is a schema reference module, not behavior templates
2. `session-lifecycle/shared-sections/` - Selected: co-located with consuming skills, follows orchestrator-templates pattern
3. `user-skills/shared/` - Rejected: creates cross-category coupling, harder to discover

**Selected**: `user-skills/session-lifecycle/shared-sections/`

**Rationale**: The orchestrator-templates skill already establishes the pattern of `shared-sections/` within a skill category. Session-lifecycle skills are the only consumers of these partials, so co-location maximizes discoverability and maintains category cohesion. The flat destination sync (per TDD-categorical-resource-organization) is unaffected because partials are reference documentation, not synced artifacts.

### Decision 2: Include Mechanism

**Options Considered**:
1. **Markdown transclusion** (e.g., `{{include: ../shared-sections/session-resolution.md}}`) - Rejected: Claude Code does not support transclusion; would require preprocessing
2. **Reference links** (e.g., `See [session-resolution](../shared-sections/session-resolution.md)`) - Selected: works natively, provides navigation
3. **Inline duplication with comments** - Rejected: defeats purpose of deduplication

**Selected**: Reference links with canonical source pattern

**Rationale**: Claude Code's skill loading reads files as-is without preprocessing. Reference links allow behavior.md to state "Pre-flight validation follows the [Session Resolution Pattern](../shared-sections/session-resolution.md)" and Claude will Read the partial when needed. This matches how session-lifecycle already references session-common: `See [session-validation](../../session-common/session-validation.md)`.

### Decision 3: Partial Granularity

**Options Considered**:
1. **One monolithic partial** - Rejected: forces reading irrelevant content; different commands need different subsets
2. **Many micro-partials** (one per check) - Rejected: overhead exceeds benefit; too fragmented
3. **Three domain partials** - Selected: session resolution, workflow resolution, state-mate invocation

**Selected**: Three domain-aligned partials

**Rationale**: Analysis of the five behavior.md files reveals three distinct patterns with clear boundaries:

| Partial | Used By | Content |
|---------|---------|---------|
| session-resolution | All 5 commands | Session exists, parked status checks |
| workflow-resolution | start, resume, handoff | Team/agent validation |
| state-mate-invocation | park, resume, wrap | Task tool delegation pattern |

This granularity matches actual usage patterns and avoids both over-fragmentation and monolithic bloat.

### Decision 4: Partial Schema Structure

**Selected Structure**:

```markdown
# {Pattern Name}

> One-line purpose

## When to Apply

- Command contexts where this pattern applies

## Validation Checks

| Check | Pass | Fail |
|-------|------|------|
| ... | ... | ... |

## Implementation

[Pseudocode or step sequence]

## Error Messages

| Condition | Message |
|-----------|---------|
| ... | ... |

## Customization Points

Parameters that consuming commands may vary.
```

**Rationale**: Follows the structure established in `orchestrator-templates/shared-sections/handling-failures.md` and `session-common/session-validation.md`. Provides complete specification for each pattern while remaining readable as a standalone reference.

### Decision 5: Backward Compatibility

**Classification**: COMPATIBLE

**Rationale**: This is a pure refactoring that extracts duplicated content into references. Existing behavior.md files will link to shared-sections instead of containing inline duplicates. No external interface changes occur:
- Skill loading unchanged
- Command behavior unchanged
- Error messages unchanged (now defined once in partial)
- No satellite impact

## Work Package Details

### WP1: Create shared-sections Directory Structure

**Objective**: Establish the shared partials that behavior.md files will reference.

**Directory Structure**:

```
user-skills/session-lifecycle/
  shared-sections/
    INDEX.md                    # Discovery and usage guide
    session-resolution.md       # Session existence and state validation
    workflow-resolution.md      # Team and agent validation
    state-mate-invocation.md    # state-mate Task delegation pattern
  start-ref/
  park-ref/
  resume/
  wrap-ref/
  handoff-ref/
```

#### session-resolution.md Schema

```markdown
# Session Resolution Pattern

> Validate session existence and state before command execution.

## When to Apply

All session-lifecycle commands that require an existing session:
- /park - requires active session
- /resume - requires parked session
- /wrap - requires active session
- /handoff - requires active session

/start is the exception: it requires NO existing session.

## Validation Checks

| Check | Function | Pass | Fail |
|-------|----------|------|------|
| Session exists | `get_session_dir()` | Directory exists | Error: No active session |
| Session not parked | `parked_at` field absent | Field not set | Error: Session parked |
| Session is parked | `parked_at` field present | Field set | Error: Session not parked |

## Implementation

```
1. Call get_session_dir() from session-utils.sh
   - Returns: Session directory path or empty
   - If empty: Error "No active session to {verb}. Use /start to begin."

2. Read SESSION_CONTEXT.md frontmatter
   - Extract parked_at field

3. Validate state against command requirements:
   - /park requires: parked_at NOT set
   - /resume requires: parked_at IS set
   - /wrap requires: parked_at NOT set (or offer auto-resume)
   - /handoff requires: parked_at NOT set
```

## Error Messages

| Condition | Message Template |
|-----------|------------------|
| No session | "No active session to {verb}. Use `/start` to begin." |
| Already parked | "Session parked at {timestamp}. Use `/resume` first." |
| Not parked | "Session is already active (not parked). Continue working." |

## Customization Points

| Parameter | Description | Commands Using |
|-----------|-------------|----------------|
| `verb` | Action verb for error message | All |
| `require_parked` | Whether session must be parked | resume only |
| `auto_resume_offer` | Offer to auto-resume if parked | wrap only |

## Cross-Reference

- Schema: [session-context-schema](../session-common/session-context-schema.md)
- State machine: [session-phases](../session-common/session-phases.md)
```

#### workflow-resolution.md Schema

```markdown
# Workflow Resolution Pattern

> Validate rite context and agent availability.

## When to Apply

Commands that invoke agents or switch teams:
- /start - validates target team, may switch
- /resume - validates session team matches active rite
- /handoff - validates target agent exists in team

## Validation Checks

| Check | Method | Pass | Fail |
|-------|--------|------|------|
| Team exists | `$ROSTER_HOME/rites/{team}` exists | Directory exists | Error: Team not found |
| Team matches session | Compare ACTIVE_RITE to session.active_team | Match | Warning + prompt |
| Agent exists | `.claude/agents/{agent}.md` exists | File exists | Error: Agent not found |

## Implementation

```
1. Read ACTIVE_RITE file
   - Path: .claude/ACTIVE_RITE
   - Returns: Current rite name

2. If command specifies team change:
   a. Verify team exists in roster
   b. Invoke swap-rite.sh
   c. Confirm ACTIVE_RITE updated

3. For session operations, check consistency:
   a. Read session.active_team from SESSION_CONTEXT
   b. Compare to ACTIVE_RITE
   c. If mismatch: Surface warning, offer switch or override

4. For agent invocation:
   a. Verify .claude/agents/{agent}.md exists
   b. If missing: Error with available agents list
```

## Error Messages

| Condition | Message Template |
|-----------|------------------|
| Team not found | "Team '{name}' not found. Use `/roster` to list available teams." |
| Team mismatch | "Session team ({session_team}) differs from active rite ({active_team})." |
| Agent not found | "Agent '{agent}' not found in team '{team}'." |
| Roster unavailable | "Roster system unavailable. Set ROSTER_HOME or check installation." |

## Customization Points

| Parameter | Description | Commands Using |
|-----------|-------------|----------------|
| `target_team` | Team to validate/switch to | start |
| `target_agent` | Agent to validate | handoff |
| `allow_override` | Allow continuing despite mismatch | resume |
```

#### state-mate-invocation.md Schema

```markdown
# state-mate Invocation Pattern

> Delegate session state mutations to state-mate agent via Task tool.

## When to Apply

Commands that mutate SESSION_CONTEXT:
- /park - sets parked_at, parked_reason, etc.
- /resume - clears park fields, sets resumed_at
- /wrap - transitions to ARCHIVED state

## Task Invocation Template

```
Task(moirai, "{operation}

Session Context:
- Session ID: {session_id}
- Session Path: .claude/sessions/{session_id}/SESSION_CONTEXT.md")
```

### Operations

| Operation | Command | Mutations |
|-----------|---------|-----------|
| `park_session reason='{reason}'` | /park | Set parked_at, parked_reason |
| `resume_session` | /resume | Clear parked_*, set resumed_at |
| `wrap_session` | /wrap | Set completed_at, archive |

## Response Handling

### Success Response

```json
{
  "success": true,
  "operation": "{operation_name}",
  "message": "Session {operation} successfully",
  "state_before": { "session_state": "..." },
  "state_after": { "session_state": "...", ... }
}
```

**Action**: Parse response, display confirmation to user.

### Failure Response

```json
{
  "success": false,
  "error_type": "LIFECYCLE_VIOLATION",
  "message": "Cannot {operation}: {reason}",
  "hint": "Use /{suggested_command} first"
}
```

**Action**: Surface error message and hint to user.

## Error Types

| Error Type | Cause | Recovery |
|------------|-------|----------|
| `LIFECYCLE_VIOLATION` | Invalid state transition | Follow hint (e.g., resume before wrap) |
| `VALIDATION_ERROR` | Missing required field | Provide missing data |
| `UNAVAILABLE` | state-mate not responding | Retry or check agent configuration |

## Implementation

```
1. Get session context:
   session_id=$(session-manager.sh status | jq -r '.session_id')
   session_path=".claude/sessions/${session_id}/SESSION_CONTEXT.md"

2. Invoke state-mate via Task tool:
   Task(moirai, "{operation}

   Session Context:
   - Session ID: {session_id}
   - Session Path: {session_path}")

3. Parse JSON response:
   - If success: true → Extract state_after, continue
   - If success: false → Extract message and hint, surface to user

4. Post-operation (command-specific):
   - /park: Display parking summary
   - /resume: Invoke selected agent
   - /wrap: Archive session directory
```

## Customization Points

| Parameter | Description | Commands Using |
|-----------|-------------|----------------|
| `operation` | state-mate operation name | All |
| `reason` | User-provided reason | park |
| `post_action` | Action after successful mutation | All |

## Cross-Reference

- state-mate agent: `.claude/agents/state-mate.md`
- ADR: `docs/decisions/ADR-0005-state-mate-centralized-state-authority.md`
```

### WP2: Update behavior.md Files to Reference Partials

**Objective**: Replace inline duplicated patterns with references to shared-sections.

**Before/After Comparison**:

| File | Before (Lines) | After (Lines) | Delta |
|------|----------------|---------------|-------|
| start-ref/behavior.md | 145 | ~100 | -45 |
| park-ref/behavior.md | 126 | ~85 | -41 |
| resume/behavior.md | 139 | ~95 | -44 |
| wrap-ref/behavior.md | 154 | ~110 | -44 |
| handoff-ref/behavior.md | 124 | ~85 | -39 |

**Reference Pattern**:

```markdown
### 1. Pre-flight Validation

Apply [Session Resolution Pattern](../shared-sections/session-resolution.md):
- Requires: Active session (not parked)
- Verb: "wrap"
- Auto-resume offer: Yes

Apply [Workflow Resolution Pattern](../shared-sections/workflow-resolution.md):
- Team consistency check: Optional
```

**Implementation Per File**:

| File | Partials Referenced |
|------|---------------------|
| start-ref/behavior.md | session-resolution (inverse), workflow-resolution |
| park-ref/behavior.md | session-resolution, state-mate-invocation |
| resume/behavior.md | session-resolution, workflow-resolution, state-mate-invocation |
| wrap-ref/behavior.md | session-resolution, state-mate-invocation |
| handoff-ref/behavior.md | session-resolution, workflow-resolution |

### WP3: Create Partial Index

**Objective**: Provide discovery and usage documentation.

**INDEX.md Content**:

```markdown
# Session-Lifecycle Shared Sections

> Reusable behavior patterns for session lifecycle commands.

## Available Partials

| Partial | Purpose | Used By |
|---------|---------|---------|
| [session-resolution](session-resolution.md) | Session existence and state validation | All 5 commands |
| [workflow-resolution](workflow-resolution.md) | Team and agent validation | start, resume, handoff |
| [state-mate-invocation](state-mate-invocation.md) | state-mate delegation pattern | park, resume, wrap |

## Usage Pattern

Reference partials from behavior.md files:

```markdown
### Pre-flight Validation

Apply [Session Resolution Pattern](../shared-sections/session-resolution.md):
- Requires: {state requirement}
- Verb: "{command verb}"
```

## Design Rationale

Partials extract duplicated patterns to:
1. **Single source of truth**: Error messages, validation logic defined once
2. **Consistent behavior**: All commands follow identical patterns
3. **Easier maintenance**: Update pattern in one place
4. **Progressive disclosure**: Skill users can drill into details

## Relationship to session-common

`shared-sections/` contains **behavioral patterns** (how to validate).
`session-common/` contains **schemas** (what fields exist).

Both are reference modules; neither is invoked directly.

## Adding New Partials

1. Identify pattern duplicated across 2+ behavior.md files
2. Extract to new `shared-sections/{pattern-name}.md`
3. Follow schema: When to Apply, Checks, Implementation, Errors, Customization
4. Update INDEX.md with new entry
5. Refactor behavior.md files to reference partial
```

## Backward Compatibility

**Classification**: COMPATIBLE

This design is purely additive refactoring:
- New `shared-sections/` directory created
- Existing behavior.md files modified to add references
- No external interfaces change
- No runtime behavior changes
- No satellite impact

## Integration Test Matrix

| Scenario | Test | Expected Outcome |
|----------|------|------------------|
| Skill loading | Load start-ref via Skill tool | Skill loads, behavior.md readable |
| Partial navigation | Follow link from behavior.md to shared-section | Partial content accessible via Read |
| Pattern coverage | Grep for duplicated error messages | All duplicates replaced with partial refs |
| session-common coexistence | Reference both session-common and shared-sections | Both resolve correctly |

## Future Extensibility

### Team-Specific Customization (Phase 2)

If teams need to override patterns, extend partial schema:

```markdown
## Team Customization

### ecosystem-pack Override

[Ecosystem-pack may override Session Resolution to include...]
```

This maintains backward compatibility while allowing specialization.

### Additional Partials (On Demand)

Future candidates for extraction:
- Confirmation message formatting
- Artifact listing pattern
- Git status capture pattern

## Handoff Criteria

- [x] Solution architecture documented with rationale
- [x] Schema definitions complete with validation rules
- [x] Backward compatibility classified (COMPATIBLE)
- [x] No migration required (pure refactoring)
- [x] Integration test matrix with expected outcomes
- [x] File changes specified at file/function level
- [x] No unresolved design decisions

## Cross-Reference

| Document | Path | Relationship |
|----------|------|--------------|
| orchestrator-templates/shared-sections | `user-skills/orchestration/orchestrator-templates/shared-sections/` | Pattern precedent |
| session-common | `user-skills/session-common/` | Schema definitions |
| TDD-categorical-resource-organization | `docs/design/TDD-categorical-resource-organization.md` | Category structure |

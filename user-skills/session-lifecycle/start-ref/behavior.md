# /start Behavior Specification

> Full step-by-step sequence for initializing a work session.

## Behavior Sequence

### 1. Pre-flight Validation

Apply [Session Resolution Pattern](../shared-sections/session-resolution.md) (inverse):
- Requires: NO existing session
- Error if session exists: "Session already active. Use `/resume` or `/wrap` first"

Apply [Workflow Resolution Pattern](../shared-sections/workflow-resolution.md):
- Target team: User-specified or current ACTIVE_TEAM
- Validate team exists in roster

See [session-validation](../../session-common/session-validation.md) for validation patterns.

### 2. Gather Session Parameters

Prompt user for any missing parameters:

- **Initiative name**: Clear, concise description (e.g., "Add dark mode toggle")
- **Complexity level**:
  - `SCRIPT` - Single file, < 200 LOC, no external dependencies
  - `MODULE` - Multiple files, < 2000 LOC, clear interfaces
  - `SERVICE` - Multiple modules, APIs, data persistence
  - `PLATFORM` - Multiple services, infrastructure, complex integration
- **Target team**: Defaults to current ACTIVE_TEAM (read from `.claude/ACTIVE_TEAM`)

### 3. Team Context Setup

- If `--team` specified and differs from ACTIVE_TEAM:
  - Invoke `$ROSTER_HOME/swap-team.sh <team-name>` via Bash tool
  - Verify ACTIVE_TEAM file updated
  - Confirm: "Switched to {team} for this session"

### 4. Create SESSION_CONTEXT

Use the Write tool to create `.claude/sessions/{session_id}/SESSION_CONTEXT.md` with this exact YAML frontmatter format:

```yaml
---
session_id: "session-20251224-HHMMSS"
created_at: "2025-12-24THH:MM:SSZ"
initiative: "{user-provided-initiative}"
complexity: "{SCRIPT|MODULE|SERVICE|PLATFORM}"
active_team: "{team-pack-name}"
current_phase: "requirements"
last_agent: null
artifacts: []
blockers: []
next_steps:
  - "Review PRD when complete"
context_version: "1.0"
---

# Session: {initiative}

## Artifacts
- PRD: pending
- TDD: pending

## Blockers
None yet.

## Next Steps
1. Complete requirements gathering
```

CRITICAL: The file MUST start with `---` on line 1. Without proper YAML frontmatter, session tracking will fail.

See [session-context-schema](../../session-common/session-context-schema.md) for field definitions.

### 5. Invoke Requirements Analyst

Use Task tool to delegate to Requirements Analyst. See [integration.md](integration.md) for delegation template.

Wait for analyst to produce PRD artifact.

### 6. Conditional Architect Invocation

If complexity is MODULE, SERVICE, or PLATFORM:
- Invoke **Architect** via Task tool
- Architect produces TDD and ADRs

See [integration.md](integration.md) for delegation template.

### 7. Update SESSION_CONTEXT

After artifacts are produced, update SESSION_CONTEXT:

```yaml
artifacts:
  - type: "PRD"
    path: "/docs/requirements/PRD-{slug}.md"
    status: "approved"
  - type: "TDD"  # if complexity > SCRIPT
    path: "/docs/design/TDD-{slug}.md"
    status: "draft"
last_agent: "architect"  # or "analyst" if SCRIPT
current_phase: "design"  # or "requirements" if SCRIPT
```

### 8. Confirmation

Display confirmation message with:
- Session name, complexity, team
- Artifacts produced (PRD, TDD, ADRs)
- Next steps
- Available commands (/park, /handoff, etc.)

---

## State Changes

### Files Created

- `.claude/sessions/{session_id}/SESSION_CONTEXT.md` - Session metadata and state
- `/docs/requirements/PRD-{slug}.md` - Product requirements document
- `/docs/design/TDD-{slug}.md` - Technical design (if complexity > SCRIPT)
- `/docs/decisions/ADR-{NNNN}-{slug}.md` - Architecture decisions (if applicable)

### Fields Set in SESSION_CONTEXT

| Field | Initial Value | Description |
|-------|---------------|-------------|
| `session_id` | Generated timestamp-based ID | Unique session identifier |
| `created_at` | Current ISO timestamp | Session start time |
| `initiative` | User-provided | Initiative name |
| `complexity` | User-provided | Complexity level |
| `active_team` | Current or specified team | Team pack for this session |
| `current_phase` | "requirements" or "design" | Current workflow phase |
| `last_agent` | "analyst" or "architect" | Last agent to work on session |
| `artifacts` | List of produced artifacts | Tracks deliverables |

---

## Error Cases

| Error | Condition | Resolution |
|-------|-----------|------------|
| Session already active | Active session for current project | Use `/wrap` to complete current session or `/resume` to continue it |
| Invalid team name | Team not found in roster | Use `/roster` to list available teams |
| Roster system unavailable | `$ROSTER_HOME/` not found | Set ROSTER_HOME environment variable or check installation |
| PRD creation failed | Analyst unable to produce PRD | Review error, provide more context, retry |
| Missing parameters | User cancels prompts | Command aborted, no state changed |

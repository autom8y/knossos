# SPRINT_CONTEXT Schema

> Canonical schema for `.claude/sessions/{session_id}/SPRINT_CONTEXT.md`

## YAML Frontmatter

```yaml
# Core fields (set by /sprint)
sprint_id: string           # "sprint-YYYYMMDD-HHMMSS"
session_id: string          # Parent session ID (links to SESSION_CONTEXT)
created_at: string          # ISO 8601 timestamp
sprint_name: string         # User-provided sprint name
sprint_goal?: string        # Optional high-level objective

# Task tracking
tasks: array                # [{id, name, status, phase, complexity, artifacts}]
completed_tasks: number     # Count of completed tasks
total_tasks: number         # Total task count

# Configuration
duration?: string           # "1w" | "2w" | "1m" (default: "2w")
start_date?: string         # ISO 8601 date
end_date?: string           # Calculated from duration

# Inherited from session
active_team: string         # Team pack name (copied from SESSION_CONTEXT)

# State tracking
blockers?: array            # [{description, task_id, severity}]

# Schema versioning
context_version: "1.0"      # Schema version for compatibility checks
```

## Required vs Optional Fields

| Field | Required | Default | Set By |
|-------|----------|---------|--------|
| `sprint_id` | Yes | - | /sprint |
| `session_id` | Yes | - | /sprint |
| `created_at` | Yes | - | /sprint |
| `sprint_name` | Yes | - | User |
| `sprint_goal` | No | null | User |
| `tasks` | Yes | [] | /sprint |
| `completed_tasks` | Yes | 0 | /sprint, task completion |
| `total_tasks` | Yes | 0 | /sprint, task addition |
| `duration` | No | "2w" | User |
| `start_date` | No | created_at | /sprint |
| `end_date` | No | calculated | /sprint |
| `active_team` | Yes | - | /sprint (from SESSION_CONTEXT) |
| `blockers` | No | [] | Task execution |
| `context_version` | Yes | "1.0" | /sprint |

## Task Object Schema

```yaml
tasks:
  - id: string              # "task-001", "task-002", etc.
    name: string            # Task description
    status: enum            # pending | in_progress | completed | blocked | skipped
    phase?: string          # Current workflow phase if in_progress
    complexity?: enum       # SCRIPT | MODULE | SERVICE | PLATFORM (null until estimated)
    artifacts: array        # [{type, path, status}]
    started_at?: string     # ISO 8601 when status changed to in_progress
    completed_at?: string   # ISO 8601 when status changed to completed
    blocker?: string        # Blocker description if status is blocked
```

## Blocker Object Schema

```yaml
blockers:
  - description: string     # What is blocked
    task_id: string         # Reference to task.id
    severity: enum          # low | medium | high | critical
    created_at: string      # ISO 8601 timestamp
    resolved_at?: string    # ISO 8601 if resolved
```

## Field Ownership

| Field | Set By | Modified By | Removed By |
|-------|--------|-------------|------------|
| `sprint_id` | /sprint | - | /wrap-sprint |
| `tasks` | /sprint | task completion | - |
| `completed_tasks` | /sprint | task completion | - |
| `blockers` | task execution | blocker resolution | - |

## Valid State Transitions

### Sprint State
```
[none] --/sprint--> active
active --/wrap-sprint--> archived
active + all tasks complete --/wrap-sprint--> archived
```

### Task State
```
pending --start--> in_progress
in_progress --complete--> completed
in_progress --block--> blocked
blocked --unblock--> in_progress
pending --skip--> skipped
```

## Example: Valid SPRINT_CONTEXT.md

```yaml
---
sprint_id: "sprint-20251229-143000"
session_id: "session-20251229-140000-a1b2c3d4"
created_at: "2025-12-29T14:30:00Z"
sprint_name: "Authentication Sprint"
sprint_goal: "Implement complete user authentication flow"
tasks:
  - id: "task-001"
    name: "Login API endpoint"
    status: "completed"
    complexity: "MODULE"
    artifacts:
      - type: "PRD"
        path: "docs/requirements/PRD-login-api.md"
        status: "approved"
    started_at: "2025-12-29T14:31:00Z"
    completed_at: "2025-12-29T15:00:00Z"
  - id: "task-002"
    name: "Session management"
    status: "in_progress"
    phase: "implementation"
    complexity: "MODULE"
    artifacts: []
    started_at: "2025-12-29T15:01:00Z"
  - id: "task-003"
    name: "Password reset flow"
    status: "pending"
    complexity: null
    artifacts: []
completed_tasks: 1
total_tasks: 3
duration: "2w"
start_date: "2025-12-29"
end_date: "2026-01-12"
active_team: "10x-dev-pack"
blockers: []
context_version: "1.0"
---

## Sprint Goal

Implement complete user authentication flow including login, session management, and password reset.

## Sprint Progress

- [x] Login API endpoint (completed 2025-12-29T15:00:00Z)
- [ ] Session management (in progress - implementation phase)
- [ ] Password reset flow (pending)

## Sprint Retrospective

(To be filled at sprint completion)
```

## Validation Rules

### Structure Validation
1. File MUST start with `---` on line 1
2. File MUST have closing `---` within first 100 lines
3. Content between delimiters MUST be valid YAML

### Field Validation
1. `sprint_id` MUST match pattern `^sprint-[0-9]{8}-[0-9]{6}$`
2. `session_id` MUST match pattern `^session-[0-9]{8}-[0-9]{6}-[a-f0-9]{8}$`
3. `created_at` MUST be valid ISO 8601 timestamp
4. `tasks` MUST be an array (may be empty)
5. `completed_tasks` MUST be integer >= 0 and <= `total_tasks`
6. `total_tasks` MUST be integer >= 0
7. `context_version` MUST be "1.0" (for forward compatibility)

### Task Validation
1. Each task MUST have `id`, `name`, `status`, `artifacts`
2. `status` MUST be one of: pending, in_progress, completed, blocked, skipped
3. If `status` is "in_progress", `started_at` SHOULD be set
4. If `status` is "completed", `completed_at` SHOULD be set
5. If `status` is "blocked", `blocker` SHOULD be set

## Relationship to SESSION_CONTEXT

SPRINT_CONTEXT is a **child context** of SESSION_CONTEXT:

```
SESSION_CONTEXT.md
    |
    +-- session_id: "session-20251229-140000-a1b2c3d4"
    |
    +-- SPRINT_CONTEXT.md (optional, 0 or 1 per session)
            |
            +-- session_id: "session-20251229-140000-a1b2c3d4" (references parent)
            +-- sprint_id: "sprint-20251229-143000"
```

### Inheritance Rules
- `active_team` in SPRINT_CONTEXT MUST match SESSION_CONTEXT at creation time
- If SESSION_CONTEXT team changes, SPRINT_CONTEXT team is NOT automatically updated
- `/wrap` on a session with active sprint MUST wrap sprint first

## Migration from Unversioned

Files without `context_version` field are assumed to be version "0.9" (pre-schema):
- Validation is lenient (field presence only)
- On next mutation, upgrade to "1.0" format

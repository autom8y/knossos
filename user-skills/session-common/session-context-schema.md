# SESSION_CONTEXT Schema

> Canonical schema for `.claude/sessions/{session_id}/SESSION_CONTEXT.md`

## YAML Frontmatter

```yaml
# Core fields (set by /start)
session_id: string        # "session-YYYYMMDD-HHMMSS-{hash}"
created_at: string        # ISO 8601 timestamp
initiative: string        # User-provided name
complexity: enum          # SCRIPT | MODULE | SERVICE | PLATFORM
active_rite: string       # Rite name
current_phase: enum       # requirements | design | implementation | validation
last_agent: string|null   # Agent identifier or null if not started
artifacts: array          # [{type, path, status}]
blockers: array           # [{description, severity}]
next_steps: array         # [string]

# Park fields (set by /park, removed by /resume)
parked_at?: string        # ISO 8601 when parked
parked_reason?: string    # User-provided or "Manual park"
parked_phase?: string     # Phase at park time
parked_git_status?: enum  # clean | dirty
parked_uncommitted_files?: int

# Resume fields (set by /resume)
resumed_at?: string       # Most recent resume timestamp
resume_count?: int        # Park/resume cycle count

# Handoff fields (set by /handoff)
handoff_count?: int       # Total handoffs in session
last_handoff_at?: string  # Most recent handoff timestamp
```

## Field Ownership

| Field | Set By | Modified By | Removed By |
|-------|--------|-------------|------------|
| `session_id` | /start | - | /wrap |
| `parked_at` | /park | - | /resume |
| `resume_count` | /resume | /resume | - |
| `handoff_count` | /handoff | /handoff | - |
| `last_agent` | /start | /handoff, /resume | - |

## Valid State Transitions

```
[none] --/start--> active
active --/park--> parked
parked --/resume--> active
active --/handoff--> active (agent changes)
active --/wrap--> archived
```

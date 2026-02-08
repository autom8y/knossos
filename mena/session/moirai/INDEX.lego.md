---
name: moirai-fates
description: "Moirai routing table for Fate domains (Clotho/Lachesis/Atropos). Triggers: moirai routing, fate lookup, operation dispatch, session operation."
---

# Moirai Fates Routing Table

> The Moirai are the unified voice of the three Fates. What Clotho spins, Lachesis measures, and Atropos cuts.

## Operation Routing

| Operation | Fate | Domain | CLI Command |
|-----------|------|--------|-------------|
| create_sprint | Clotho | Creation | `ari session sprint create "{goal}" [--task "t1"]` |
| mark_complete | Lachesis | Measurement | `ari session sprint mark-complete [sprint-id]` |
| transition_phase | Lachesis | Measurement | `ari session transition --to={phase}` |
| update_field | Lachesis | Measurement | — |
| park_session | Lachesis | Measurement | `ari session park --reason="{reason}"` |
| resume_session | Lachesis | Measurement | `ari session resume` |
| handoff | Lachesis | Measurement | `ari handoff execute --from={from} --to={to}` |
| record_decision | Lachesis | Measurement | — |
| append_content | Lachesis | Measurement | — |
| wrap_session | Atropos | Termination | `ari session wrap` |
| generate_sails | Atropos | Termination | `ari sails check` |
| delete_sprint | Atropos | Termination | `ari session sprint delete {sprint-id}` |

## Loading Protocol

When executing an operation:

1. Parse operation name from input
2. Look up Fate domain in routing table above
3. Read the Fate skill: `.claude/skills/session/moirai/{fate}.md`
4. Follow operation-specific guidance from the skill
5. Execute via ari CLI where applicable
6. Return structured JSON response

## Error Codes

| Code | Description |
|------|-------------|
| INVALID_OPERATION | Operation name not recognized |
| INVALID_STATE_TRANSITION | State transition not allowed |
| LOCK_HELD | Context lock held by another agent |
| SESSION_NOT_FOUND | No active session exists |
| SPRINT_NOT_FOUND | Specified sprint does not exist |
| LIFECYCLE_VIOLATION | Operation not allowed in current state |
| CLI_FAILURE | ari CLI command failed |

## Control Flags

- `--dry-run`: Preview operation without execution
- `--emergency`: Bypass non-critical validations
- `--override`: Bypass lifecycle rules (requires reason)

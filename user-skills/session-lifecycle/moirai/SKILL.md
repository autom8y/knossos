---
name: moirai
description: |
  Moirai internal skills for session lifecycle management. These skills provide
  operation-specific guidance for the unified Moirai agent. They are NOT user-invokable.
internal: true
---

# Moirai Skills

> The three Fates govern session lifecycle. This routing table maps operations to domains.

## Routing Table

| Operation | Fate | Domain | CLI Command |
|-----------|------|--------|-------------|
| create_sprint | clotho | creation | - |
| start_sprint | clotho | creation | - |
| mark_complete | lachesis | measurement | - |
| transition_phase | lachesis | measurement | ari session transition |
| update_field | lachesis | measurement | - |
| park_session | lachesis | measurement | ari session park |
| resume_session | lachesis | measurement | ari session resume |
| handoff | lachesis | measurement | ari handoff execute |
| record_decision | lachesis | measurement | - |
| append_content | lachesis | measurement | - |
| wrap_session | atropos | termination | ari session wrap |
| generate_sails | atropos | termination | ari sails check |
| delete_sprint | atropos | termination | - |

## Domain Files

- **clotho.md**: Creation operations (spinning new entities)
- **lachesis.md**: Measurement operations (tracking state changes)
- **atropos.md**: Termination operations (ending and archiving)

## Loading Protocol

1. Parse operation from user input
2. Lookup operation in routing table above
3. Read the corresponding domain file
4. Follow operation specification in domain file
5. Delegate to CLI command if specified
6. Return structured JSON response

## Error Codes

| Code | Description |
|------|-------------|
| INVALID_OPERATION | Operation not in routing table |
| SCHEMA_VIOLATION | State change would violate schema |
| LIFECYCLE_VIOLATION | State transition not allowed |
| DEPENDENCY_BLOCKED | Blocked by unmet dependency |
| LOCK_TIMEOUT | Could not acquire file lock |
| FILE_NOT_FOUND | Target context file missing |
| VALIDATION_FAILED | Pre-mutation validation failed |
| QUALITY_GATE_FAILED | Sails check prevents wrap |

## Control Flags

| Flag | Effect |
|------|--------|
| --dry-run | Preview mutation without applying |
| --emergency | Bypass non-critical validations |
| --override=reason | Bypass lifecycle rules with reason |

## Schema Locations

| Schema | Path |
|--------|------|
| Session Context | `schemas/artifacts/session-context.schema.json` |
| Sprint Context | `schemas/artifacts/sprint-context.schema.json` |
| White Sails | `ariadne/internal/validation/schemas/white-sails.schema.json` |

## File Paths

| Context | Path |
|---------|------|
| Session | `.claude/sessions/{session-id}/SESSION_CONTEXT.md` |
| Default Sprint | `.claude/sessions/{session-id}/SPRINT_CONTEXT.md` |
| Named Sprint | `.claude/sessions/{session-id}/sprints/{sprint-id}/SPRINT_CONTEXT.md` |
| Audit Log | `.claude/sessions/.audit/session-mutations.log` |

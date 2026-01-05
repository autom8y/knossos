# Knossos Migration Guide

> Migration path from roster to Knossos platform completion.

## Current State: 70% Integration

The Knossos platform (roster/.claude/) is 70% integrated with Ariadne CLI:

### What Works
- Session lifecycle (create, park, wrap, resume)
- Thread Contract v2 events (tool_call, file_change, command, decision, sails_generated)
- White Sails confidence signaling
- Hook invocation via `ari hook <name>`
- state-mate session mutations

### What's In Progress
- Full Thread Contract event coverage (SESSION_*, TASK_*, ARTIFACT_*, ERROR)
- Handoff delegation to state-mate
- Comprehensive documentation

## Target: 90% Integration

### Phase 1: Thread Contract Completion

All 6 handoff events must be emitted:

| Event | Status | Emitter |
|-------|--------|---------|
| SESSION_START | Implementing | `ari session create` |
| SESSION_END | Implementing | `ari session wrap` |
| TASK_START | Implementing | Task tool hook |
| TASK_END | Implementing | Task tool completion |
| ARTIFACT_CREATED | Implementing | Post-artifact hook |
| ERROR | Implementing | Error boundary hook |

### Phase 2: Handoff Integration

Handoff commands delegate to state-mate:

```bash
# Handoff prepares context for agent transition
ari handoff prepare --from=architect --to=engineer --artifact=TDD.md

# Handoff executes the transition
ari handoff execute --validation=schema

# Check handoff status
ari handoff status
```

### Phase 3: Self-Hosting

Ariadne managing Knossos sessions (dog-fooding milestone):

1. Create session via `ari session create`
2. Track all events via Thread Contract
3. Generate WHITE_SAILS.yaml via `ari sails generate`
4. Wrap session via `ari session wrap`

## Bash to Ari Migration

### Before (Shell Scripts)

```bash
# Old: Direct file manipulation
SESSION_DIR=".claude/sessions/current"
echo "status: ACTIVE" >> "$SESSION_DIR/SESSION_CONTEXT.md"
```

### After (Ariadne CLI)

```bash
# New: CLI with validation
ari session resume
ari session transition ACTIVE
```

### Migration Checklist

- [ ] Replace direct SESSION_CONTEXT.md writes with `ari session` commands (park, resume, transition, wrap)
- [ ] Replace manual event recording with `ari hook thread`
- [ ] Replace shell status checks with `ari session status`
- [ ] Update hooks to use `ari hook <name>` pattern

## Future: Repository Rename

The roster repository will be renamed to knossos:

### Impact

| Before | After |
|--------|-------|
| `roster/` | `knossos/` |
| `github.com/user/roster` | `github.com/user/knossos` |

### Preparation

1. All imports use relative paths (no absolute roster references)
2. Documentation references "Knossos" not "roster" where appropriate
3. CI/CD uses repository-agnostic configuration

### Timeline

Rename occurs when:
- 90% integration achieved
- All handoff events implemented
- Self-hosting milestone complete
- Documentation covers all integration patterns

## Rollback Procedures

### Session State Rollback

```bash
# If session state needs correction, use state-mate with override
Task(state-mate, "--override=reason='Recovery from failed mutation' resume_session")

# View session audit history
ari session audit
```

### Event Rollback

Events in events.jsonl are append-only. To "rollback":
1. Archive corrupted events.jsonl
2. Regenerate from SESSION_CONTEXT.md state
3. Re-emit corrective events

## Related Resources

- [White Sails Guide](white-sails.md)
- [PRD-ariadne.md](../requirements/PRD-ariadne.md) - Full CLI specification
- [TDD-knossos-v2.md](../design/TDD-knossos-v2.md) - Technical design

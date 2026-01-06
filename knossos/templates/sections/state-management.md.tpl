{{/* state-management section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START state-management -->
## State Management

**Mutating session/sprint state?** Use the **Moirai** (the Fates) for all `SESSION_CONTEXT.md` and `SPRINT_CONTEXT.md` changes.

### Moirai Usage

The Moirai are the centralized authority for session lifecycle--spinning sessions into existence (Clotho), measuring their allotment (Lachesis), and cutting when complete (Atropos). They enforce schema validation, lifecycle transitions, and maintain audit trails.

**When to Use**:
- Updating session state (park, resume, wrap)
- Marking tasks complete
- Transitioning workflow phases
- Creating or managing sprints
- Any modification to `*_CONTEXT.md` files
- Generating White Sails confidence signals

**Invocation Pattern** (requires session context):
```
Task(moirai, "mark_complete task-001 artifact=docs/requirements/PRD-foo.md

Session Context:
- Session ID: {from session-manager.sh status}
- Session Path: .claude/sessions/{session-id}/SESSION_CONTEXT.md")
```

Get session context: `.claude/hooks/lib/session-manager.sh status | jq -r '.session_id'`

**Natural Language Supported**:
```
Task(moirai, "Mark the PRD task complete with artifact at docs/requirements/PRD-foo.md")
```

**Control Flags**:
- `--dry-run`: Preview changes without applying
- `--emergency`: Bypass non-critical validations (logged)
- `--override=reason`: Bypass lifecycle rules with explicit reason

**Direct writes blocked**: PreToolUse hook intercepts `Write`/`Edit` to `*_CONTEXT.md` and instructs use of Moirai.

**Full documentation**: See `user-agents/moirai.md` and `docs/philosophy/knossos-doctrine.md`
<!-- KNOSSOS:END state-management -->

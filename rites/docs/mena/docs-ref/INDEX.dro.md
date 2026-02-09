---
name: docs
description: "Switch to docs rite (documentation workflow). Use when: user says /docs, wants tech writing, API docs, documentation audit. Triggers: /docs, documentation rite, doc workflow, tech writing."
context: fork
---

# /docs - Switch to Documentation Rite

Switch to docs, the technical writing and documentation quality rite.

## Behavior

### 1. Invoke Rite Switch

Execute via Bash tool:

```bash
ari sync --rite docs
```

### 2. Display Knossos

After successful switch, show the agent table:

| Agent | Role |
|-------|------|
| orchestrator | Coordinates documentation workflows |
| doc-auditor | Audits existing documentation quality |
| information-architect | Designs information structure and hierarchy |
| tech-writer | Writes and edits technical documentation |
| review-coordinator | Coordinates documentation reviews |

### 3. Update Session

If a session is active, update `active_rite` to `docs`.

## When to Use

- Technical writing and API documentation
- Documentation quality audits
- Information architecture planning
- Documentation review coordination

**Don't use for**: Feature development → `/10x` | Code quality → `/hygiene` | Debt assessment → `/debt`

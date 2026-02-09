---
name: docs-ref
description: "Switch to docs (documentation workflow). Triggers: /docs, documentation rite, doc workflow, tech writing."
---

# /docs - Quick Switch to Documentation Rite

> **Category**: Rite Management | **Phase**: Rite Switching

## Purpose

Instantly switch to the docs, a specialized rite focused on technical writing, API documentation, information architecture, and documentation quality audits.

This is a convenience wrapper around `/rite docs` that also displays the knossos after switching.

---

## Usage

```bash
/docs
```

No parameters required. This command:
1. Switches to docs
2. Displays rite catalog with agent descriptions

---

## Behavior

### 1. Invoke Rite Switch

Execute via Bash tool:

```bash
$KNOSSOS_HOME/ari sync --rite docs
```

### 2. Display Knossos

After successful switch, show the active knossos:

```
Switched to docs (4 agents loaded)

Knossos:
┌─────────────────────────┬──────────────────────────────────────────────┐
│ Agent                   │ Role                                         │
├─────────────────────────┼──────────────────────────────────────────────┤
│ doc-auditor             │ Reviews docs for accuracy and completeness   │
│ information-architect   │ Structures documentation for discoverability │
│ tech-writer             │ Creates clear, concise technical content     │
│ doc-reviewer            │ Validates quality, consistency, standards    │
└─────────────────────────┴──────────────────────────────────────────────┘

Use /handoff <agent> to delegate work.
```

### 3. Update SESSION_CONTEXT (if active)

If a session is active:
- Update `active_rite` field to `docs`
- Add handoff note documenting rite switch

---

## Rite Details

**Rite Name**: docs
**Agent Count**: 4
**Workflow**: Audit → Architecture → Writing → Review

### Agents

#### doc-auditor.md
**Role**: Documentation audit and gap analysis
**Invocation**: `Act as **Doc Auditor**`
**Purpose**: Identifies missing, outdated, or incomplete documentation

**When to use**:
- Beginning documentation initiatives
- Assessing current doc coverage
- Finding documentation gaps in codebases
- Planning documentation roadmaps

#### information-architect.md
**Role**: Documentation structure and organization
**Invocation**: `Act as **Information Architect**`
**Purpose**: Designs navigation, taxonomy, and information hierarchy

**When to use**:
- Structuring large documentation sets
- Creating documentation site architecture
- Organizing API references
- Designing user guides and tutorials
- Planning progressive disclosure strategies

#### tech-writer.md
**Role**: Technical content creation
**Invocation**: `Act as **Tech Writer**`
**Purpose**: Writes clear, accurate technical documentation

**When to use**:
- Creating README files
- Writing API documentation
- Producing user guides
- Drafting tutorials and how-tos
- Documenting architecture and design decisions
- Creating runbooks and operational docs

#### doc-reviewer.md
**Role**: Documentation quality validation
**Invocation**: `Act as **Doc Reviewer**`
**Purpose**: Ensures accuracy, consistency, and adherence to style guides

**When to use**:
- Final review before publishing
- Consistency checks across docs
- Style guide compliance validation
- Technical accuracy verification
- Readability and clarity assessment

---

## Examples

### Example 1: Basic Switch

```bash
/docs
```

Output:
```
[Knossos] Switched to docs (4 agents loaded)

Knossos:
  - doc-auditor: Reviews docs for accuracy and completeness
  - information-architect: Structures documentation for discoverability
  - tech-writer: Creates clear, concise technical content
  - doc-reviewer: Validates quality, consistency, standards

Ready for documentation workflow.
```

### Example 2: Documentation Initiative

```bash
/docs
/start "Document REST API" --complexity=MODULE
```

Output:
```
[Knossos] Switched to docs (4 agents loaded)
Session started: Document REST API
Complexity: MODULE

Next: Doc Auditor will assess current API documentation state.
```

### Example 3: Mid-Session Switch

After implementing a feature with `/10x`, switch to document it:

```bash
/docs
/handoff writer
```

Output:
```
[Knossos] Switched to docs (4 agents loaded)
Handing off to: tech-writer

Tech Writer reviewing implementation artifacts...
```

---

## Typical Workflow with Docs Rite

### Phase 1: Audit
```bash
/docs
/start "Document authentication system" --complexity=MODULE
# Doc Auditor assesses current state, identifies gaps
```

### Phase 2: Architecture
```bash
/handoff information-architect
# Information Architect designs doc structure
# - Getting Started
# - API Reference
# - Integration Guide
# - Troubleshooting
```

### Phase 3: Writing
```bash
/handoff tech-writer
# Tech Writer creates documentation content
# Following structure from Information Architect
```

### Phase 4: Review
```bash
/handoff doc-reviewer
# Doc Reviewer validates quality and consistency
# Checks against style guides
# Ensures technical accuracy
```

### Phase 5: Completion
```bash
/wrap
```

---

## When to Use Docs Rite

Use this rite for:

- **API documentation**: REST/GraphQL/gRPC endpoint documentation
- **README creation**: Project onboarding, quickstart guides
- **Architecture documentation**: System design docs, diagrams
- **User guides**: Feature usage, integration instructions
- **Runbooks**: Operational procedures, troubleshooting
- **ADR review**: Polishing architecture decision records
- **Tutorial creation**: Step-by-step learning content

**Don't use for**:
- Feature implementation → Use `/10x` instead
- Code refactoring → Use `/hygiene` instead
- Debt assessment → Use `/debt` instead

---

## State Changes

### Files Modified

| File | Change | Description |
|------|--------|-------------|
| `.claude/ACTIVE_RITE` | Set to `docs` | Active rite state |
| `.claude/agents/` | Populated | 4 agent files loaded |
| `.claude/sessions/{session_id}/SESSION_CONTEXT.md` | `active_rite` updated | If session active |

---

## Success Criteria

- Switched to docs rite
- 4 agent files present in `.claude/agents/`
- Rite catalog displayed to user
- If session active, SESSION_CONTEXT updated

---

## Error Handling

If swap fails:

```
[Knossos] Error: Rite 'docs' not found
[Knossos] Use '/rite --list' to see available packs
```

**Resolution**: Verify knossos installation at `$KNOSSOS_HOME/`

---

## Integration with Documentation Skill

This rite complements the `documentation` skill:

```bash
/docs
# Use documentation skill to access templates
Act as **Tech Writer**.

Create API documentation for user authentication endpoints.
Use template at `.claude/skills/documentation/templates/api-doc.md`.
```

The documentation skill provides templates, this rite provides specialized agents.

---

## Related Commands

- `/rite` - General rite switching with options
- `/10x` - Quick switch to development rite
- `/hygiene` - Quick switch to code hygiene rite
- `/debt` - Quick switch to technical debt rite
- `/handoff` - Delegate to specific agent in current rite

---

## Related Skills

- [documentation](../../../../mena/templates/documentation/INDEX.lego.md) - Templates for PRD/TDD/ADR/Test Plans
- [10x-workflow](../../../10x-dev/mena/10x-workflow/INDEX.lego.md) - Agent coordination patterns

---

## Related Documentation

- [COMMAND_REGISTRY.md](../../COMMAND_REGISTRY.md) - All registered commands
- [ari sync --rite]($KNOSSOS_HOME/ari sync --rite) - Rite sync implementation

---

## Notes

### Documentation-First Projects

Some projects prioritize documentation:
- Open source libraries (README, API docs critical)
- Internal tools (runbooks, user guides essential)
- Platform migrations (documentation before code changes)

For these, start with `/docs` instead of `/10x`.

### Difference from /rite

| Command | Behavior |
|---------|----------|
| `/rite docs` | Switches rite, shows swap confirmation |
| `/docs` | Switches rite, shows rite catalog with agent descriptions |

Use `/docs` when you want to see available agents after switching.

### Cross-Rite Handoffs

Common pattern: Implement with `/10x`, document with `/docs`:

```bash
/10x
/start "Add OAuth2 support" --complexity=MODULE
# ... implementation ...
/docs
/handoff tech-writer
# Tech Writer documents new OAuth2 endpoints
```

This leverages specialized rites for each phase.

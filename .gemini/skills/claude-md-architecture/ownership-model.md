# Section Ownership Model

CLAUDE.md sections have explicit owners that determine sync behavior. This document defines the ownership model and how the sync pipeline applies it.

---

## Ownership Categories

### SYNC Sections (Knossos-Owned)

Content that comes from knossos and overwrites satellite content during sync.

**Characteristics**:
- Source of truth: Knossos templates
- Change frequency: When ecosystem patterns evolve
- Satellite modifications: Not allowed (use `## Project:*` to extend)
- Propagation: Knossos -> All satellites

**Examples**:

| Section | Content |
|---------|---------|
| `## Agent Routing` | How to route work to agents |
| `## Skills Architecture` | Skill activation table |
| `## Hooks` | Hook documentation |
| `## Dynamic Context Syntax` | How `!` commands work |
| `## Getting Help` | Navigation reference |

**Rule**: If content describes HOW THE ECOSYSTEM WORKS, it syncs from knossos.

---

### PRESERVE Sections (Satellite-Owned)

Content that satellites own and the sync pipeline never overwrites.

**Characteristics**:
- Source of truth: Satellite itself
- Change frequency: When project scope evolves
- Knossos modifications: Ignored for this section
- Propagation: Never (satellite-specific)

**Examples**:

| Section | Content |
|---------|---------|
| `## Quick Start` | Satellite's rite (regenerated from knossos if missing) |
| `## Agent Configurations` | Satellite's agents (regenerated from knossos if missing) |
| Custom sections not matching knossos | Project-specific content |
| Unknown sections | Default to preserve for safety |

**Rule**: If content describes WHAT THIS PROJECT IS, satellite owns it.

---

### PROJECT Sections (Satellite Extensions)

Content that extends knossos patterns without conflicting. Uses `## Project:*` namespace.

**Characteristics**:
- Source of truth: Satellite
- Naming convention: `## Project: {name}` or `## Project:{name}`
- Sync behavior: Never touched by sync pipeline
- Purpose: Add project-specific extensions

**Examples**:

```markdown
## Project: Custom Skills

| Skill | When to Activate |
|-------|------------------|
| **my-domain** | Domain-specific logic for X |

## Project: Deployment

This project deploys via GitHub Actions to AWS ECS.
```

**Rule**: Use `## Project:*` when you need to extend ecosystem patterns with project-specific content.

---

### REGENERATE Sections (Knossos-Derived)

Content derived from knossos state (ACTIVE_RITE file + agents/ directory).

**Characteristics**:
- Source of truth: ACTIVE_RITE + agents/
- Regeneration trigger: `ari sync --rite` or ari sync with missing content
- Generated locally: Each satellite has its own rite
- Represents: Which agents are available in THIS project

**Examples**:

| Section | Source |
|---------|--------|
| Quick Start agent table | ACTIVE_RITE + agents/*.md |
| Agent Configurations list | agents/*.md |
| Rite name reference | ACTIVE_RITE |

**Regeneration Logic**:

```
IF satellite has ACTIVE_RITE + agents:
  REGENERATE rite sections from satellite knossos
ELSE IF section exists in satellite:
  PRESERVE existing content
ELSE:
  Leave empty (satellite needs to configure rite)
```

**Rule**: Rite content ALWAYS comes from satellite's own ACTIVE_RITE + agents/.

---

## Section Ownership Map

Complete mapping of sections to owners and sync behavior:

| Section Header | Owner | Sync Behavior | Notes |
|----------------|-------|---------------|-------|
| `# CLAUDE.md` | Knossos | SYNC | Title and tagline |
| `## Quick Start` | Rite | PRESERVE/REGENERATE | From satellite's ACTIVE_RITE |
| `## Agent Routing` | Knossos | SYNC | Infrastructure |
| `## Skills Architecture` | Knossos | SYNC | Infrastructure |
| `## Agent Configurations` | Rite | PRESERVE/REGENERATE | From satellite's agents/ |
| `## Hooks` | Knossos | SYNC | Infrastructure |
| `## Dynamic Context Syntax` | Knossos | SYNC | Infrastructure |
| `## Getting Help` | Knossos | SYNC | Infrastructure |
| `## Project:*` | Satellite | PRESERVE | Unlimited extensions |
| `## (unknown)` | Satellite | PRESERVE | Default to preserve |

---

## Decision Tree: Which Owner?

```
What does this content describe?

1. HOW the ecosystem works?
   - Agent routing patterns
   - Skill activation rules
   - Hook behavior
   - Dynamic context syntax
   └─> Knossos-owned, SYNC

2. WHAT this project is?
   - Project-specific conventions
   - Custom skills
   - Domain terminology
   - Integration patterns
   └─> Satellite-owned, PRESERVE or ## Project:*

3. WHO is working (which agents)?
   - Agent rite catalog
   - Rite name
   - Agent configurations
   └─> Knossos-derived, REGENERATE from ACTIVE_RITE

4. WHAT is happening now?
   - Current task
   - Session phase
   - Handoff context
   └─> Session state, NOT in CLAUDE.md
```

---

## The Sync Contract

### Knossos PROVIDES

1. **Workflow Infrastructure**: Agent routing, handoff protocols, phase transitions
2. **Capability Documentation**: Skills architecture, hooks documentation, dynamic context
3. **Reference Patterns**: Getting help, entry point structure, section organization

### Satellites OWN

1. **Rite Identity**: Quick Start (regenerated), Agent Configurations, rite-specific variations
2. **Project Extensions**: `## Project:*` sections, custom sections, project conventions
3. **Project Context**: Tech stack references, domain terminology, integration patterns

### Should NEVER Sync

1. **Session State**: Current task, work in progress, parked session info
2. **Transient Context**: Git state, file modification status, worktree context
3. **Rite Content from Wrong Source**: Rite content comes from satellite's own ACTIVE_RITE

---

## Anti-Pattern: Wrong Rite Content Source

**Wrong**:
```markdown
# satellite CLAUDE.md (after sync)

## Quick Start

This project uses a 6-agent workflow (ecosystem):
| ecosystem-analyst | ... |
| context-architect | ... |
```

**Why wrong**: Satellite has its own rite (e.g., docs). Rite content should come from satellite's ACTIVE_RITE + agents/.

**Correct approach**:
- Rite sections are PRESERVE (keep satellite content) or REGENERATE (rebuild from satellite's own ACTIVE_RITE)
- Never sync rite content from knossos's source templates

---

## Layer Mapping

Content lives in different layers with different scopes:

| Layer | Location | Content | Modified By |
|-------|----------|---------|-------------|
| Global | `~/.claude/CLAUDE.md` | Personal preferences | User |
| Project | `.claude/CLAUDE.md` | Rite + Project + Infrastructure | sync pipeline, knossos |
| Session | Hook output | Transient context | Hooks (read-only) |

---

## Marker Syntax

Use HTML comments to mark section ownership. The sync pipeline uses these markers to determine sync behavior.

### Format

```markdown
<!-- SYNC: knossos-owned -->
## Section Name

<!-- PRESERVE: satellite-owned, regenerated from ACTIVE_RITE + agents/ -->
## Quick Start
```

### Placement Rules

1. **Markers MUST appear immediately before the section header** (no blank line between)
2. Markers are single-line HTML comments: `<!-- SYNC: description -->`
3. Description is optional but recommended for clarity

### Valid Markers

| Marker | Meaning |
|--------|---------|
| `<!-- SYNC: knossos-owned -->` | Section syncs from knossos, overwrites satellite |
| `<!-- PRESERVE: satellite-owned -->` | Section preserved from satellite, never overwritten |
| `<!-- PRESERVE: satellite-owned, regenerated from ACTIVE_RITE + agents/ -->` | Preserved, with note about regeneration source |

### Example CLAUDE.md Structure

```markdown
# CLAUDE.md

> Entry point for Claude Code.

<!-- PRESERVE: satellite-owned, regenerated from ACTIVE_RITE + agents/ -->
## Quick Start

This project uses a 5-agent workflow (docs):
...

<!-- SYNC: knossos-owned -->
## Agent Routing

Before implementing work, check:
...

<!-- SYNC: knossos-owned -->
## Skills Architecture

Skills provide domain knowledge on-demand.
...
```

### Sync Pipeline Behavior

- Sync pipeline reads markers during sync to determine ownership
- If marker is missing, sync pipeline uses section name matching as fallback
- Markers are preserved during sync (extracted with their section)

---

## Related Files

- [first-principles.md](first-principles.md) - Core architectural principles
- [boundary-test.md](boundary-test.md) - Validation checklist
- [anti-patterns.md](anti-patterns.md) - What NOT to put in CLAUDE.md

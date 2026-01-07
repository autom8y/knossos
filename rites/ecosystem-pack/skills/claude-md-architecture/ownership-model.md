# Section Ownership Model

CLAUDE.md sections have explicit owners that determine sync behavior. This document defines the ownership model and how CEM applies it.

---

## Ownership Categories

### SYNC Sections (Roster-Owned)

Content that comes from roster and overwrites satellite content during sync.

**Characteristics**:
- Source of truth: Roster's templates
- Change frequency: When ecosystem patterns evolve
- Satellite modifications: Not allowed (use `## Project:*` to extend)
- Propagation: Roster -> All satellites

**Examples**:

| Section | Content |
|---------|---------|
| `## Agent Routing` | How to route work to agents |
| `## Skills Architecture` | Skill activation table |
| `## Hooks` | Hook documentation |
| `## Dynamic Context Syntax` | How `!` commands work |
| `## Getting Help` | Navigation reference |

**Rule**: If content describes HOW THE ECOSYSTEM WORKS, it syncs from roster.

---

### PRESERVE Sections (Satellite-Owned)

Content that satellites own and CEM never overwrites.

**Characteristics**:
- Source of truth: Satellite itself
- Change frequency: When project scope evolves
- Roster modifications: Ignored for this section
- Propagation: Never (satellite-specific)

**Examples**:

| Section | Content |
|---------|---------|
| `## Quick Start` | Satellite's team (regenerated from roster if missing) |
| `## Agent Configurations` | Satellite's agents (regenerated from roster if missing) |
| Custom sections not matching roster | Project-specific content |
| Unknown sections | Default to preserve for safety |

**Rule**: If content describes WHAT THIS PROJECT IS, satellite owns it.

---

### PROJECT Sections (Satellite Extensions)

Content that extends roster patterns without conflicting. Uses `## Project:*` namespace.

**Characteristics**:
- Source of truth: Satellite
- Naming convention: `## Project: {name}` or `## Project:{name}`
- Sync behavior: Never touched by CEM
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

### REGENERATE Sections (Roster-Derived)

Content derived from roster state (ACTIVE_RITE file + agents/ directory).

**Characteristics**:
- Source of truth: ACTIVE_RITE + agents/
- Regeneration trigger: `swap-rite.sh` or CEM sync with missing content
- Generated locally: Each satellite has its own team
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
  REGENERATE team sections from satellite roster
ELSE IF section exists in satellite:
  PRESERVE existing content
ELSE:
  Leave empty (satellite needs to configure team)
```

**Rule**: Team content ALWAYS comes from satellite's own ACTIVE_RITE + agents/.

---

## Section Ownership Map

Complete mapping of sections to owners and sync behavior:

| Section Header | Owner | Sync Behavior | Notes |
|----------------|-------|---------------|-------|
| `# CLAUDE.md` | Roster | SYNC | Title and tagline |
| `## Quick Start` | Team | PRESERVE/REGENERATE | From satellite's ACTIVE_RITE |
| `## Agent Routing` | Roster | SYNC | Infrastructure |
| `## Skills Architecture` | Roster | SYNC | Infrastructure |
| `## Agent Configurations` | Team | PRESERVE/REGENERATE | From satellite's agents/ |
| `## Hooks` | Roster | SYNC | Infrastructure |
| `## Dynamic Context Syntax` | Roster | SYNC | Infrastructure |
| `## Getting Help` | Roster | SYNC | Infrastructure |
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
   └─> Roster-owned, SYNC

2. WHAT this project is?
   - Project-specific conventions
   - Custom skills
   - Domain terminology
   - Integration patterns
   └─> Satellite-owned, PRESERVE or ## Project:*

3. WHO is working (which agents)?
   - Agent roster
   - Rite name
   - Agent configurations
   └─> Roster-derived, REGENERATE from ACTIVE_RITE

4. WHAT is happening now?
   - Current task
   - Session phase
   - Handoff context
   └─> Session state, NOT in CLAUDE.md
```

---

## The Sync Contract

### Roster PROVIDES

1. **Workflow Infrastructure**: Agent routing, handoff protocols, phase transitions
2. **Capability Documentation**: Skills architecture, hooks documentation, dynamic context
3. **Reference Patterns**: Getting help, entry point structure, section organization

### Satellites OWN

1. **Team Identity**: Quick Start (regenerated), Agent Configurations, rite-specific variations
2. **Project Extensions**: `## Project:*` sections, custom sections, project conventions
3. **Project Context**: Tech stack references, domain terminology, integration patterns

### Should NEVER Sync

1. **Session State**: Current task, work in progress, parked session info
2. **Transient Context**: Git state, file modification status, worktree context
3. **Team Content from Wrong Source**: Team content comes from satellite's own ACTIVE_RITE

---

## Anti-Pattern: Wrong Team Content Source

**Wrong**:
```markdown
# satellite CLAUDE.md (after sync)

## Quick Start

This project uses a 6-agent workflow (ecosystem-pack):
| ecosystem-analyst | ... |
| context-architect | ... |
```

**Why wrong**: Satellite has its own team (e.g., doc-rite-pack). Team content should come from satellite's ACTIVE_RITE + agents/.

**Correct approach**:
- Team sections are PRESERVE (keep satellite content) or REGENERATE (rebuild from satellite's own ACTIVE_RITE)
- Never sync team content from roster's source templates

---

## Layer Mapping

Content lives in different layers with different scopes:

| Layer | Location | Content | Modified By |
|-------|----------|---------|-------------|
| Global | `~/.claude/CLAUDE.md` | Personal preferences | User |
| Project | `.claude/CLAUDE.md` | Team + Project + Infrastructure | CEM, roster |
| Session | Hook output | Transient context | Hooks (read-only) |

---

## Marker Syntax

Use HTML comments to mark section ownership. CEM uses these markers to determine sync behavior.

### Format

```markdown
<!-- SYNC: roster-owned -->
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
| `<!-- SYNC: roster-owned -->` | Section syncs from roster, overwrites satellite |
| `<!-- PRESERVE: satellite-owned -->` | Section preserved from satellite, never overwritten |
| `<!-- PRESERVE: satellite-owned, regenerated from ACTIVE_RITE + agents/ -->` | Preserved, with note about regeneration source |

### Example CLAUDE.md Structure

```markdown
# CLAUDE.md

> Entry point for Claude Code.

<!-- PRESERVE: satellite-owned, regenerated from ACTIVE_RITE + agents/ -->
## Quick Start

This project uses a 5-agent workflow (doc-rite-pack):
...

<!-- SYNC: roster-owned -->
## Agent Routing

Before implementing work, check:
...

<!-- SYNC: roster-owned -->
## Skills Architecture

Skills provide domain knowledge on-demand.
...
```

### CEM Behavior

- CEM reads markers during sync to determine ownership
- If marker is missing, CEM uses section name matching as fallback
- Markers are preserved during sync (extracted with their section)

---

## Related Files

- [first-principles.md](first-principles.md) - Core architectural principles
- [boundary-test.md](boundary-test.md) - Validation checklist
- [anti-patterns.md](anti-patterns.md) - What NOT to put in CLAUDE.md

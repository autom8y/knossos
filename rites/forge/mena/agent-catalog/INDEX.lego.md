---
name: agent-catalog
description: "Rite catalog templates, Consultant sync checklists, and versioning scheme for the Agent Curator. Use when: integrating a new rite, updating Consultant knowledge base, creating rite profiles, deprecating rites, recording versions. Triggers: catalog entry, rite profile, Consultant sync, versioning, deprecation, rite integration."
---

# Agent Catalog Artifacts

> Templates, checklists, and versioning patterns for rite catalog integration by the Agent Curator.

## Purpose

Provides the canonical rite profile template, step-by-step Consultant sync checklists for new/modified/deprecated rites, and the versioning scheme with changelog format.

## Contents

| Resource | Purpose |
|----------|---------|
| [Rite Profile Template](#rite-profile-template) | Standard documentation format for rite profiles |
| [Consultant Sync Checklist](#consultant-sync-checklist) | File-by-file update guide for new/modified/deprecated rites |
| [Versioning Scheme](#versioning-scheme) | Version numbering and changelog format |

## Rite Profile Template

```markdown
# {rite-name}

> {One-line description}

## Overview
{2-3 sentences about rite purpose and when to use it}

## Quick Start
```bash
/{rite}          # Switch to this rite
/task "{goal}"   # Start a task
```

## Agents

| Agent | Model | Phase | Produces |
|-------|-------|-------|----------|
| {name} | {model} | {phase} | {artifact} |

## Workflow

```
{phase-1} -> {phase-2} -> {phase-3} -> {phase-4}
```

## Complexity Levels

| Level | Scope | Phases |
|-------|-------|--------|
| {LEVEL} | {description} | [{phases}] |

## Commands

| Command | Purpose |
|---------|---------|
| `/{rite}` | Switch to this rite |
| `/task` | Full lifecycle task |

## Best For
- {Use case 1}
- {Use case 2}

## Not For
- {Anti-use case 1}
- {Anti-use case 2}

## Related Rites
- [{other-rite}]({link}) - {relationship}
```

## Consultant Sync Checklist

### New Rite

Files to update:

1. **ecosystem-map.md**
   - Add rite to Rites table
   - Update rite count
   - Update total agent count

2. **agent-reference.md**
   - Add new section: `## {rite} ({N} agents)`
   - List all agents with model, phase, produces
   - Add workflow summary

3. **rite-profiles/{rite}.md** (NEW)
   - Create from template above
   - Include all sections

4. **routing/intent-patterns.md**
   - Add domain keywords
   - Map to rite and commands

5. **command-reference.md**
   - Add `/{rite}` to Rite Management section

### Modified Rite

Files to check:
- agent-reference.md (if agents changed)
- rite-profiles/{rite}.md (update details)
- ecosystem-map.md (if counts changed)

### Deprecated Rite

Actions:
1. Mark as deprecated in rite profile
2. Remove from active routing (intent-patterns)
3. Add migration note pointing to replacement
4. Keep rite-profile for historical reference
5. Update ecosystem counts

## Versioning Scheme

```
v{major}.{minor}.{patch}

major: Breaking changes, restructured workflow
minor: New agents, new capabilities
patch: Bug fixes, prompt refinements
```

Example changelog entry:
```markdown
## [1.1.0] - 2025-12-24

### Added
- New compliance-auditor agent
- PATCH complexity level

### Changed
- threat-modeler now produces structured threat model

### Fixed
- Handoff criteria for security-reviewer
```

---
type: design
session: session-20260303-114830-beb1bf03
status: complete
sprint: .know/feat Design & Implementation Sprint
workstream: WS-5
---

# Context Design: Feature Knowledge Skill Wiring

## Problem Statement

Agents need a discovery mechanism for feature knowledge files in `.know/feat/`. Without a skill (legomenon) that Claude Code can auto-invoke, agents have no way to discover that `.know/feat/` exists, what schema the files follow, or how to load the right feature reference before modifying feature-related code.

## Solution: Single `feat-knowledge` Legomenon

### Decision

Create a single shared legomenon at `rites/shared/mena/feat-knowledge/INDEX.lego.md` that serves as a discovery index and usage guide for `.know/feat/` files.

### Rationale

**Alternative A (chosen): Single index skill.** One legomenon catalogs all features and teaches agents how to use `.know/feat/`. Feature knowledge itself lives in `.know/feat/{slug}.md` files -- the skill is a pointer, not a container.

**Alternative B (rejected): N skills for N features.** Creating a legomenon per feature would inflate the skill count, require skill creation/deletion as features are added/removed, and complicate the mena directory. The skill layer should not mirror the knowledge layer.

**Alternative C (rejected): Embed feature knowledge in the existing `know` dromenon.** The `/know` command is a dromenon (transient, user-invoked). Feature knowledge discovery is a persistent reference need that should be model-invoked. Mixing these violates the dro/lego lifecycle distinction.

### Why This Works

- CC auto-invokes skills based on description triggers. The `feat-knowledge` description covers feature-related queries (feature purpose, feature design, implementation map, boundaries).
- Progressive disclosure: skill loads into context (~95 lines), then agents Read() individual `.know/feat/` files on demand. Context cost is minimal.
- Graceful degradation: skill explicitly handles the case where `.know/feat/` does not exist yet, guiding agents to suggest `/know --scope=feature`.

## Components Affected

### New File

| File | Purpose |
|------|---------|
| `rites/shared/mena/feat-knowledge/INDEX.lego.md` | Shared legomenon -- feature knowledge discovery index |

### Existing Files: No Modifications

No existing files are modified. This is a purely additive change.

## Schema

### Legomenon Frontmatter

```yaml
name: feat-knowledge  # Required. Kebab-case, matches directory name.
description: "..."     # Required. Contains "Use when:" and "Triggers:" clauses.
```

Follows the established legomenon frontmatter pattern (see `cross-rite-handoff/INDEX.lego.md`, `pinakes/INDEX.lego.md`). No `invokable` field -- majority of shared legomena omit it.

### Description Trigger Design

The description includes these trigger terms for CC auto-invocation:
- `feature knowledge`, `feature purpose`, `feature design` -- direct feature queries
- `feature implementation map`, `feature boundaries` -- section-specific queries
- `.know/feat/` -- path-based queries
- `product feature`, `feature rationale` -- conceptual queries

These do not overlap with existing skill triggers. The closest skill is `pinakes`, which triggers on `theoria`, `audit`, `domain criteria` -- complementary, not competing.

### Body Structure

1. **Quick Reference** table: 3-row resource table (census, per-feature, generation command)
2. **When to Use** decision tree: agents decide whether to load feature knowledge
3. **How to Use** steps: find feature -> load reference -> apply knowledge
4. **Feature Knowledge Schema** table: 4-section schema with purpose per section
5. **If Feature Knowledge Does Not Exist**: graceful degradation guidance
6. **Census Expiry**: staleness expectations (30d census, 14d per-feature)
7. **Related**: links to complementary `.know/` files and pinakes criteria

## Backward Compatibility: COMPATIBLE

This is a purely additive change. No existing files are modified. The new legomenon directory follows established conventions. Satellites are unaffected -- this skill is knossos-internal and distributed via `ari sync --scope=user --resource=mena`.

## Migration Path

None required. The skill is new and optional. Agents that do not encounter feature-related topics will never auto-invoke it.

## Integration Test Matrix

| Satellite Type | Test | Expected Outcome |
|----------------|------|------------------|
| **baseline** | `ari sync --scope=user --resource=mena` | Skill appears in `~/.claude/skills/feat-knowledge/INDEX.lego.md`. No regression in other skills. |
| **minimal** | Agent conversation about a feature | CC auto-invokes `feat-knowledge` skill. Skill guides agent to Read `.know/feat/INDEX.md`. If absent, skill suggests `/know --scope=feature`. |
| **complex** | Agent modifying feature code in a project with `.know/feat/` files | CC auto-invokes skill. Agent reads INDEX.md, identifies correct slug, reads `.know/feat/{slug}.md`, applies knowledge to code change. |
| **no-feature-knowledge** | Agent in project with no `.know/feat/` directory | Skill loads, graceful degradation section activates, agent suggests generation. No errors. |

## Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| Legomenon | `/Users/tomtenuta/Code/knossos/rites/shared/mena/feat-knowledge/INDEX.lego.md` | Created, verified via Read |
| Context Design | `/Users/tomtenuta/Code/knossos/docs/ecosystem/DESIGN-feat-knowledge-skill-wiring.md` | This file |

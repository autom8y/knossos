---
name: orchestrator-templates
description: "Durable abstraction layer for agent orchestrator generation via YAML specs and templates. Use when: creating team orchestrators, updating canonical templates, troubleshooting generation. Triggers: orchestrator template, orchestrator.yaml, orchestrator generation, team orchestrator, orchestrator-base.md.tpl."
---

# Orchestrator Templates Skill

> Master the durable abstraction layer for agent orchestrator generation.

## What It Is

The orchestrator templating system separates **semantic specifications** (YAML) from **implementation** (generation scripts). Instead of maintaining 10+ hand-written orchestrator.md files, teams define configuration once in `orchestrator.yaml` and the generator produces production-ready agents.

**Key insight**: The template is durable. The generation logic is an implementation detail.

```
orchestrator.yaml (team semantics, durable)
         │
         ▼
orchestrator-generate.sh (bash mechanics, replaceable)
         │
         ▼
orchestrator.md (production output, disposable)
```

## When to Use This Skill

### Creating a New Team Orchestrator
You're building a new agent team and need to create its orchestrator.md. Follow the step-by-step creation guide to:
1. Define `orchestrator.yaml` with your team's configuration
2. Run the generator
3. Validate and commit

**Time**: 15 minutes for straightforward teams, 30-45 minutes for complex routing.

### Updating the Canonical Template
Your team discovered a pattern all orchestrators should follow. Update `rite-base.md.tpl` and regenerate all teams:
1. Modify the template
2. Run batch regeneration
3. Review diffs (should be cosmetic if adding optional section)
4. Commit

**Time**: 5 minutes to update template, 10 minutes to regenerate, 15-30 minutes to review.

### Troubleshooting Generation Issues
The generator produced unexpected output or failed validation. Use the troubleshooting guide to diagnose:
- Schema validation errors
- Missing specialist definitions
- Placeholder replacement failures
- Formatting inconsistencies

**Time**: 5-15 minutes depending on issue complexity.

### Understanding the Architecture
You're designing a new team structure and want to understand how orchestrator templating fits. Read the architecture overview to see:
- How generated agents evolve over time
- Why YAML specs are durable
- How templating integrates with CEM, swap-rite.sh, and workflow.yaml

**Time**: 10-15 minutes for solid understanding.

## Core Concept: Durable Abstractions

Hand-written orchestrators mix **unchanging protocol** with **rite-specific details**. When protocol evolves, you update 10+ files manually with high risk of inconsistency.

**The solution**: Separate canonical protocol (template) from team specifics (YAML):

```yaml
# orchestrator.yaml - team owns this
team:
  name: rnd
routing:
  integration-researcher: "Needs research on technologies"
  technology-scout: "Evaluates emerging tools"
  prototype-engineer: "Builds working demo"
```

Template contains protocol once. YAML contains variation. Generator merges them.

**Benefits**:
- Update protocol once, affects all teams
- Consistent contracts across orchestrators
- 80% reduction in maintenance burden
- Teams focus on routing logic, not protocol details

## Workflow

### 1. Design Configuration

Create `rites/my-team/orchestrator.yaml`:

```yaml
team:
  name: my-team
  domain: "What your team does"
  color: purple

frontmatter:
  role: "One-line role"
  description: "Multi-line description"

routing:
  specialist-one: "When to route here"
  specialist-two: "When to route here"

handoff_criteria:
  specialist-one:
    - "Criterion 1"
    - "Artifacts verified via Read tool"

skills:
  - "@skill-one brief description"
```

See [schema-reference.md](schema-reference.md) for complete field reference.

### 2. Generate & Validate

```bash
# Generate
/roster/templates/orchestrator-generate.sh my-team

# Validate
/roster/templates/validate-orchestrator.sh \
  rites/my-team/agents/orchestrator.md
```

### 3. Commit & Activate

```bash
git add rites/my-team/orchestrator.yaml rites/my-team/agents/orchestrator.md
git commit -m "feat: add my-team orchestrator"
./swap-rite.sh my-team
```

See [create-new-rite-orchestrator.md](create-new-rite-orchestrator.md) for detailed walkthrough.

## File Structure

| File | Location | Purpose |
|------|----------|---------|
| **orchestrator.yaml** | `rites/{team}/orchestrator.yaml` | Team config (durable) |
| **rite-base.md.tpl** | `/roster/templates/` | Canonical template |
| **workflow.yaml** | `rites/{team}/workflow.yaml` | Phase definitions (reference) |
| **orchestrator.md** | `rites/{team}/agents/orchestrator.md` | Generated agent (production) |

## Generator Behavior

**Does**:
- Reads orchestrator.yaml, workflow.yaml, template
- Substitutes placeholders with team data
- Validates output (no unreplaced placeholders)
- Writes orchestrator.md

**Doesn't**:
- Edit YAML files
- Auto-commit to git
- Update AGENT_MANIFEST
- Run validation (separate step)

## Validation Rules

After generation, run `validate-orchestrator.sh` to verify:

| Rule | Check | Catches |
|------|-------|---------|
| File exists | Can read generated file | Missing output file |
| No placeholders | All `{{}}` replaced | Incomplete generation |
| Valid frontmatter | YAML between --- delimiters | Corrupted YAML header |
| Required sections | All 10 required headers present | Missing sections |
| Specialist names | Consistent throughout | Typos in routing table |
| No duplicates | No duplicate section headers | Generator bug |
| Handoff criteria | Checkbox format in phase sections | Missing acceptance gates |
| Consultation protocol | CONSULTATION_REQUEST/RESPONSE present | Missing protocol sections |
| Skill references | @skill-name format | Broken skill links |
| Markdown syntax | Balanced code blocks | Syntax errors |

**Success**: Exit code 0. Output is production-ready.
**Failure**: Exit code 2-3. Review error message and fix.

## Common Patterns

**Linear Pipeline** (4-5 specialists): analyst → designer → implementer → reviewer
- Example teams: rnd, security, docs

**Hub Coordination** (5+ specialists): analyzer coordinates multiple parallel paths
- Example teams: ecosystem

**Domain-Specific Complexity**: Generator pulls complexity enum from workflow.yaml
- Security: PATCH | FEATURE | SYSTEM
- Documentation: PAGE | SECTION | SITE

See [architecture-overview.md](architecture-overview.md) for pattern details.

## Integration Points

**swap-rite.sh**: Reads frontmatter, extracts name/role/description, symlinks orchestrator.md

**workflow.yaml**: Generator validates specialists in routing exist in workflow.yaml phases

**CEM Sync**: Generated orchestrator.md treated as standard agent file (no special handling)

**AGENT_MANIFEST**: Automatically tracks orchestrator.md (no manual edits needed)

See [architecture-overview.md](architecture-overview.md) for integration details.

## Troubleshooting Quick Reference

**Specialist not found**: Update orchestrator.yaml routing to match workflow.yaml phase names exactly

**Unreplaced placeholders**: Verify orchestrator.yaml is valid YAML with all required fields

**Frontmatter parsing fails**: Run validation, check frontmatter format matches schema

**Large diff after template update**: Normal if canonical sections changed, regenerate all teams

**Team-specific content lost**: Use `extension_points` in YAML for custom content

See [troubleshooting.md](troubleshooting.md) for 25+ detailed scenarios with solutions.

## Future Enhancements

**Template versioning**: Enable v1 → v2 migrations with schema version field

**CI integration**: Validate YAML against schema, verify generated files match

See [architecture-overview.md](architecture-overview.md) for roadmap details.

## Related Skills

- `orchestration` - Phase coordination and handoff protocols
- `documentation` - Templates for PRDs, TDDs, ADRs
- `standards` - Naming conventions and code style

## Quick Reference

See [QUICK-REFERENCE.md](QUICK-REFERENCE.md) for commands and FAQ.

---

**Last Updated**: 2025-12-29
**Status**: Production


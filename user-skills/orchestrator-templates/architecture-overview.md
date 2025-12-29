# Architecture Overview: Orchestrator Templating System

> Understand why orchestrator templating exists, how it evolves, and how it integrates with the broader agent architecture.

## The Big Picture

Agent orchestrators encode two things:

1. **Canonical Protocol** (identical across all teams)
   - How orchestrators interact with main agents (CONSULTATION_REQUEST/RESPONSE)
   - How phases are structured and sequenced
   - How handoff criteria define phase gates
   - → Lives in template (shared, updated once for all teams)

2. **Team-Specific Details** (different per team)
   - Which specialists are available
   - When to route to each specialist
   - Team-specific handoff criteria
   - Team-specific skills and focus areas
   - → Lives in YAML config (team owns, updates as team evolves)

**The insight**: Separating these means the protocol can evolve independently from team structure.

```
┌─────────────────────────────────────────────────────────┐
│ Canonical Protocol (Template)                           │
│ - Consultation request/response schema                  │
│ - Phase sequencing structure                            │
│ - Handoff criteria format                               │
│ - Anti-patterns and domain authority concepts           │
│ [Changes here affect ALL orchestrators at once]         │
└─────────────────────────────────────────────────────────┘
                         │
                         │ Merge
                         ▼
┌─────────────────────────────────────────────────────────┐
│ Team Config (YAML)                                      │
│ - Specialist names and roles                            │
│ - Routing conditions                                    │
│ - Team-specific handoff criteria                        │
│ [Each team owns and evolves independently]              │
└─────────────────────────────────────────────────────────┘
                         │
                         │ Generate
                         ▼
┌─────────────────────────────────────────────────────────┐
│ Production Orchestrator (orchestrator.md)               │
│ [Ready to commit to git, use with swap-team.sh]         │
└─────────────────────────────────────────────────────────┘
```

## Why This Architecture?

### Problem: Hand-Written Inconsistency

Before templating, every orchestrator was hand-written:

**Risks**:
- Copy-paste errors in protocol sections
- Inconsistent formatting across teams
- Updating protocol meant editing 10+ files
- High risk of missing a team or introducing inconsistency
- Difficult to onboard new teams

**Example**: Protocol evolves → must update 10 files → risk of subtle differences → testing burden

### Solution: Template-Based Generation

Separate the parts that are the same from the parts that differ:

**Benefits**:
- Protocol updates happen once (in template)
- All teams automatically get improvements
- Team-specific data lives in YAML (durable, human-readable)
- Generation is deterministic (no surprises)
- Easy to onboard new teams (copy YAML, run generator)

## System Architecture

### Component 1: Template (`orchestrator-base.md.tpl`)

**Purpose**: Define the canonical protocol and structure

**What it contains**:
- Frontmatter template (name, role, tools, model, color)
- Core Purpose and Responsibilities
- Consultation Protocol (canonical, frozen)
- Domain Authority structure
- Handling Failures framework
- Anti-Patterns framework
- Placeholders for team-specific data

**Sections marked**:
- `<!-- CANONICAL: ... -->` - Never change (protocol is frozen)
- `<!-- STABLE: ... -->` - Can refine per team
- `<!-- EXTENSION: ... -->` - Teams can customize

**File location**: `/Users/tomtenuta/Code/roster/templates/orchestrator-base.md.tpl`

**Size**: ~140 lines

**Stability**: Stable. Changes roughly once per quarter when new patterns discovered.

### Component 2: YAML Schema

**Purpose**: Define valid structure for team configurations

**What it contains**:
- JSON Schema Draft 7 validation rules
- Field requirements and constraints
- Type definitions
- Real examples for 4 teams

**Constraints enforced**:
- Required fields: team, frontmatter, routing, workflow_position, handoff_criteria, skills
- Specialist count: 3-6 minimum/maximum
- Skill format: @skill-name with description
- Handoff criteria: must exist for each specialist

**File location**: `/Users/tomtenuta/Code/roster/schemas/orchestrator.yaml.schema.json`

**Status**: Complete. Not expected to change.

### Component 3: Generator (`orchestrator-generate.sh`)

**Purpose**: Merge template + config → production orchestrator

**What it does**:
1. Reads orchestrator.yaml (team config)
2. Reads workflow.yaml (specialist definitions)
3. Reads orchestrator-base.md.tpl (canonical template)
4. Validates specialist names exist in workflow.yaml
5. Substitutes placeholders in template
6. Writes orchestrator.md

**Process**:
```bash
orchestrator.yaml + workflow.yaml + template
    │
    ├─→ Parse YAML configs
    ├─→ Validate specialist names
    ├─→ Generate routing table markdown
    ├─→ Generate handoff criteria sections
    ├─→ Substitute all placeholders
    └─→ Write orchestrator.md
```

**File location**: `/Users/tomtenuta/Code/roster/templates/orchestrator-generate.sh`

**Status**: Production-ready (Phase 3 validated on all 11 teams)

### Component 4: Validator (`validate-orchestrator.sh`)

**Purpose**: Verify generated output is production-ready

**Validation rules** (10 total):
1. File exists and is readable
2. No unreplaced placeholders (all {{}} substituted)
3. Valid YAML frontmatter
4. All required sections present
5. Specialist names consistent throughout
6. No duplicate section headers
7. Handoff criteria in checkbox format
8. Consultation Protocol present and correct
9. Skill references in @skill-name format
10. Markdown syntax valid (balanced code blocks)

**Exit codes**:
- 0: Pass (file is production-ready)
- 1: Invalid arguments or file not found
- 2: Structural validation failed
- 3: Semantic validation failed

**File location**: `/Users/tomtenuta/Code/roster/templates/validate-orchestrator.sh`

**Status**: Production-ready (Phase 3 tested on all 11 teams)

## Data Flow

### Creating a New Orchestrator

```
Developer
    │
    ├─→ Create orchestrator.yaml (config)
    │   └─→ Define team metadata (name, domain, color)
    │   └─→ Define routing (specialists and conditions)
    │   └─→ Define handoff criteria
    │   └─→ Define skills references
    │
    ├─→ Run generator
    │   orchestrator-generate.sh {team}
    │   │
    │   ├─→ Read orchestrator.yaml
    │   ├─→ Read workflow.yaml
    │   ├─→ Validate specialist names
    │   ├─→ Read orchestrator-base.md.tpl
    │   ├─→ Substitute placeholders
    │   └─→ Write orchestrator.md
    │
    ├─→ Run validator
    │   validate-orchestrator.sh orchestrator.md
    │   │
    │   └─→ Verify 10 validation rules
    │   └─→ Exit 0 (success) or 2-3 (failure)
    │
    ├─→ Commit both files
    │   git add orchestrator.yaml orchestrator.md
    │   git commit -m "feat: add orchestrator"
    │
    └─→ Test with swap-team.sh
        ./swap-team.sh my-team
        └─→ Reads frontmatter from orchestrator.md
        └─→ Validates format
        └─→ Activates team
```

### Updating the Template

```
Discovery: All orchestrators should have [new pattern]
    │
    ├─→ Update template
    │   Edit orchestrator-base.md.tpl
    │   Add new section or improve existing
    │
    ├─→ Regenerate all teams
    │   for team in all_teams; do
    │     orchestrator-generate.sh $team
    │   done
    │
    ├─→ Validate all
    │   for team in all_teams; do
    │     validate-orchestrator.sh orchestrator.md
    │   done
    │
    ├─→ Review diffs
    │   git diff teams/*/agents/orchestrator.md
    │
    └─→ Commit template + regenerated files
        git add templates/orchestrator-base.md.tpl
        git add teams/*/agents/orchestrator.md
        git commit -m "refactor: update orchestrator template"
```

## Integration Points

### 1. swap-team.sh (Team Activation)

**What swap-team.sh does**:
- Reads frontmatter from orchestrator.md
- Extracts: name, role, description, model, color, tools
- Validates format
- Symlinks orchestrator.md into active agent position

**Dependency**: Frontmatter must follow specific format

```yaml
---
name: orchestrator
role: "Team role description"
description: "Team description"
tools: Read, Skill
model: claude-opus-4-5
color: team-color
---
```

**Generator responsibility**: Always produce valid frontmatter

**Validator responsibility**: Verify frontmatter parses correctly

**Impact if broken**: Cannot activate teams

### 2. workflow.yaml (Specialist Definitions)

**What workflow.yaml contains**:
- Phase sequence
- Specialist (agent) names
- Phase-specific complexity enums
- Phase descriptions

**Dependency**: orchestrator.yaml routing must reference specialists that exist in workflow.yaml

```yaml
# workflow.yaml defines:
phases:
  - agent: integration-researcher
  - agent: technology-scout
  - agent: prototype-engineer
  - agent: moonshot-architect

# orchestrator.yaml must use same names:
routing:
  integration-researcher: "..."
  technology-scout: "..."
  prototype-engineer: "..."
  moonshot-architect: "..."
```

**Generator responsibility**: Validate specialist names and report errors

**Impact if broken**: Generator exits with "Specialist not found" error

### 3. CEM Sync (File Distribution)

**What CEM does**:
- Distributes orchestrator.md across satellites
- Validates against standard agent checksum
- Tracks in AGENT_MANIFEST

**Dependency**: orchestrator.md is standard Markdown with valid YAML frontmatter

**Generator responsibility**: Produce valid Markdown and YAML

**Validator responsibility**: Verify Markdown syntax and frontmatter

**Impact if broken**: CEM sync fails, agents not available on satellites

### 4. AGENT_MANIFEST (Agent Registry)

**What AGENT_MANIFEST tracks**:
```json
{
  "agents": {
    "orchestrator": {
      "source": "team",
      "origin": "my-team",
      "model": "claude-opus-4-5"
    }
  }
}
```

**Dependency**: orchestrator.md must be discoverable by agent scanning

**Generator responsibility**: Produce valid agent file

**Impact if broken**: Orchestrator not registered, team can't be activated

## Evolution Strategy

### Phase 1: Foundation (Complete - Phase 4)

**What we have**:
- Template architecture with placeholders
- YAML schema defining team config structure
- Generator that merges template + config
- Validator ensuring quality
- All 11 teams produce valid orchestrators

**What we learned**:
- Template-based approach works (93.5% similarity achieved)
- YAML is readable and maintainable by teams
- Generation is deterministic and predictable
- Validation catches issues early

### Phase 2: Enhancement (Future - Phase 5)

**What we might add**:
- Template versioning for migration support
- CI integration to validate on commit
- Automated regression testing
- Schema enhancements (new fields)
- Diagram generation for complex teams

### Phase 3: Scale (Future - Phase 5+)

**What becomes possible**:
- Easy creation of new teams (copy YAML, run generator)
- Pattern discovery across teams (data-driven insights)
- Orchestrator composition (teams collaborating)
- Multi-team coordination (nested orchestrators)

## Durable Abstractions

### What Makes an Abstraction Durable?

An abstraction is durable when it **doesn't need to change when the implementation changes**.

**Example**:
```
Durable: "Team specifies routing table in YAML"
Not durable: "Team edits routing markdown table by hand"

Why? If we switch from markdown to YAML for routing, YAML-based teams don't break.
But hand-edited tables are hard to migrate.
```

### How We Achieve Durability

**1. Semantic Specification** (YAML)
- What are the facts? (Specialists, routing, criteria)
- YAML captures facts, not presentation

**2. Implementation Abstraction** (Generator)
- How do we present those facts? (Markdown tables, sections)
- Generator handles presentation
- Can change presentation without changing facts

**3. Validation** (Schema + Validator)
- What makes a specification valid?
- Schema validates facts (team, routing are present)
- Validator verifies presentation (markdown is valid)

**Result**: If generator changes (new markdown style, new template format), facts stay the same. Teams don't need to update YAML.

## Design Decisions

### Why YAML Instead of JSON?

**Tradeoff**:
- JSON: Fully standardized, strict validation
- YAML: Human-readable, indentation-based, flexible

**Decision**: YAML for team configs

**Rationale**:
- Teams read and edit by hand → YAML is more readable
- Fewer quotes and brackets → less error-prone
- Comments are natural in YAML → teams can explain choices
- Standard YAML tools (yq) available → easy validation

### Why Single Template for All Teams?

**Tradeoff**:
- One template: Consistency, single source of truth
- Multiple templates: Flexibility, team-specific sections

**Decision**: One template with team-specific YAML data

**Rationale**:
- Consultation protocol is identical → template captures it once
- Different teams have different specialists → YAML captures variation
- One template easier to maintain → less bug risk
- Easy to evolve pattern → benefits all teams automatically

### Why Not Full Automation (e.g., derive from workflow.yaml)?

**Tradeoff**:
- Manual YAML: Extra work, but explicit choices visible
- Auto-derived: Zero manual work, but less control

**Decision**: Manual YAML with generator

**Rationale**:
- orchestrator.yaml documents team design choices (why route to specialist X?)
- workflow.yaml describes phase mechanics (phase order, agent names)
- Different concerns require separate files
- Humans need to make routing decisions (can't automate "when to route")

## Failure Modes and Recovery

### If Template Breaks

**Symptom**: Generated orchestrator has missing sections or wrong format

**Recovery**:
1. Revert template to last known good version
2. Regenerate all teams
3. Validate all pass
4. Identify template issue
5. Fix and test on one team first

### If Generator Breaks

**Symptom**: Placeholder substitution fails or produces wrong output

**Recovery**:
1. Keep old orchestrator.md files (git has history)
2. Fix generator script
3. Regenerate
4. Validate output
5. Commit fix + regenerated files

### If YAML Config Is Invalid

**Symptom**: Generator rejects YAML or produces wrong routing

**Recovery**:
1. Validate YAML syntax: `yq . orchestrator.yaml`
2. Check specialist names against workflow.yaml
3. Fix YAML
4. Regenerate
5. Validate output

**Prevention**: Always validate YAML before generating.

## Roadmap

### Short Term (Phase 5)

- CI integration: Validate on every commit
- Schema enhancements: Add optional fields
- Error messages: Improve clarity for common issues

### Medium Term (Phase 6+)

- Versioning: Support orchestrator.yaml v1, v2, etc.
- Migration tools: Auto-convert old format to new
- Nested orchestrators: For teams with 7+ specialists
- Advanced diagrams: Better visualization of specialist flow

### Long Term (Phase 7+)

- Multi-team orchestrators: One orchestrator coordinates multiple teams
- Composition: Assembling orchestrators from modules
- Meta-programming: Orchestrators that generate orchestrators
- Cross-org coordination: Sharing patterns between organizations

## Key Metrics

**Current state** (Phase 4):
- Template size: 140 lines
- Generator complexity: ~200 lines of bash
- Validator rules: 10 comprehensive checks
- Teams supported: 11/11 (100%)
- Validation pass rate: 100% (110/110 rules × 11 teams)
- Time to create new team: 15-30 minutes
- Time to update pattern: 10-15 minutes

**Efficiency gains**:
- Before: 2 hours to update pattern across 10 teams
- After: 5 minutes (update template, regenerate, commit)
- Break-even: 12-18 updates (achievable in 1 quarter)

---

**Status**: Foundation complete, ready for evolution
**Last Updated**: 2025-12-29
**Next Phase**: Phase 5 (CI integration, enhancements)

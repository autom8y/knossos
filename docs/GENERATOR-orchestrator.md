# Orchestrator Template Generator

> Production-ready generator for orchestrator.md agent files from YAML configuration

**Last Updated**: 2025-12-29
**Status**: Production (Phase 2 Complete)
**Location**: `/roster/templates/orchestrator-generate.sh`

---

## Overview

The `orchestrator-generate.sh` generator creates production-ready orchestrator.md agent files for all 10 rites from declarative YAML configurations (`orchestrator.yaml`) and workflow definitions (`workflow.yaml`).

### What It Does

1. **Schema Validation**: Validates orchestrator.yaml against JSON schema
2. **Workflow Validation**: Ensures all specialists exist in workflow.yaml
3. **Template Processing**: Substitutes configuration values into canonical template
4. **Post-Generation Validation**: Runs `validate-orchestrator.sh` to ensure quality
5. **Batch Generation**: Processes all teams with rollback on first error

### What It Produces

- **orchestrator.md**: Production-ready agent file ready to commit to git
- **Validated**: Passes all 10 validation rules from validate-orchestrator.sh
- **Compatible**: Frontmatter parses by swap-rite.sh patterns
- **Integrated**: Specialist references validate against workflow.yaml

---

## Quick Start

### Generate Single Team

```bash
cd /roster

# Generate with validation
./templates/orchestrator-generate.sh rnd

# Preview output (no file written)
./templates/orchestrator-generate.sh rnd --dry-run

# Validate config without generating
./templates/orchestrator-generate.sh rnd --validate-only

# Regenerate existing team
./templates/orchestrator-generate.sh security --force
```

### Generate All Teams

```bash
# Validate all configs without generating
./templates/orchestrator-generate.sh --all --validate-only

# Batch generate all teams
./templates/orchestrator-generate.sh --all

# Regenerate all teams (overwrite existing)
./templates/orchestrator-generate.sh --all --force

# Preview all output without writing
./templates/orchestrator-generate.sh --all --dry-run
```

---

## Command Reference

### Syntax

```bash
./orchestrator-generate.sh <team-name> [options]
./orchestrator-generate.sh --all [options]
./orchestrator-generate.sh --help
```

### Arguments

| Argument | Purpose | Example |
|----------|---------|---------|
| `<team-name>` | Generate single team (lowercase with hyphens) | `rnd`, `security` |
| `--all` | Batch-generate all teams found in rites/ directory | `./orchestrator-generate.sh --all` |

### Options

| Flag | Purpose | Usage |
|------|---------|-------|
| `--validate-only` | Check configs and schema, skip generation | `./orchestrator-generate.sh rnd --validate-only` |
| `--dry-run` | Output to stdout instead of writing file | `./orchestrator-generate.sh rnd --dry-run` |
| `--force` | Overwrite existing orchestrator.md | `./orchestrator-generate.sh security --force` |
| `--help` | Show help message | `./orchestrator-generate.sh --help` |

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `ROSTER_HOME` | `$HOME/Code/roster` | Root directory for roster structure |

---

## Configuration Structure

### orchestrator.yaml

Each team has an `orchestrator.yaml` file that defines its orchestrator.md generation parameters.

**Location**: `/roster/rites/<team>/orchestrator.yaml`

**Structure**:

```yaml
# Team metadata
team:
  name: rnd                          # Must match directory name
  domain: technology exploration
  color: purple                           # Hex or named color

# Frontmatter for generated agent
frontmatter:
  role: "Coordinates technology exploration phases"
  description: |
    Multi-line description of team purpose and use cases.
    Use when: exploration spans multiple phases or requires coordination.
    Triggers: coordinate, orchestrate, R&D workflow, etc.

# Routing rules: when to invoke each specialist
routing:
  specialist-name: "Condition for routing"
  another-specialist: "When this specialist should be invoked"

# Position in larger ecosystem
workflow_position:
  upstream: "What feeds into orchestrator"
  downstream: "Where orchestrator output flows"

# Phase completion checklists
handoff_criteria:
  phase-name:
    - "Concrete criterion 1"
    - "Concrete criterion 2"

# Skills this orchestrator references
skills:
  - "@skill-name for description of use"
  - "@another-skill for relevance"

# OPTIONAL: Team-specific anti-patterns
antipatterns:
  - "Anti-pattern 1"
  - "Anti-pattern 2"

# OPTIONAL: Cross-team protocol (if hub team)
cross_team_protocol: "Description of how team coordinates with others"
```

**Example**: `/roster/rites/rnd/orchestrator.yaml`

### workflow.yaml

Each team also has a `workflow.yaml` that defines phases, specialists, and complexity levels.

**Location**: `/roster/rites/<team>/workflow.yaml`

**Key Sections**:
- `phases[]`: Array of phase names and agent names
- `complexity_levels[]`: Complexity enum values used in consultation protocol
- `entry_point`: Starting agent and artifact type

The generator uses workflow.yaml to:
- Extract specialist names for routing table
- Extract complexity levels for CONSULTATION_REQUEST schema
- Validate that orchestrator.yaml references exist in workflow

---

## Examples

### Example 1: Single Team Generation with Validation

```bash
$ cd /roster
$ ./templates/orchestrator-generate.sh docs

OK: All dependencies found
OK: All required files exist

INFO: Processing team: docs
OK: Schema validation passed: /roster/rites/docs/orchestrator.yaml
OK: Workflow references validated: /roster/rites/docs/orchestrator.yaml -> /roster/rites/docs/workflow.yaml
INFO: Generating: docs
OK: Generated: /roster/rites/docs/agents/orchestrator.md
```

### Example 2: Dry-Run Preview

```bash
$ ./templates/orchestrator-generate.sh security --dry-run 2>&1 | head -40

OK: All dependencies found
OK: All required files exist

INFO: Processing team: security
OK: Schema validation passed: ...
OK: Workflow references validated: ...
INFO: Generating: security

=== Generated Content (dry-run) ===
---
name: orchestrator
role: "Coordinates security initiatives"
description: "Coordinates security phases for security work. Routes tasks ..."
tools: Read, Skill
model: opus
color: red
---

# Orchestrator

Stateless advisor that receives context and returns structured directives.
Analyzes initiative state, decides which specialist acts next, and crafts
focused prompts. Does NOT execute work—the main agent controls all execution
via Task tool.

...
=== End Generated Content ===
```

### Example 3: Batch Validation

```bash
$ ./templates/orchestrator-generate.sh --all --validate-only

OK: All dependencies found
OK: All required files exist

INFO: Batch generating 11 teams

INFO: Processing team: 10x-dev
OK: Schema validation passed: ...
OK: Workflow references validated: ...
OK: Validation passed: 10x-dev

INFO: Processing team: debt-triage
OK: Schema validation passed: ...
OK: Workflow references validated: ...
OK: Validation passed: debt-triage

... (9 more teams) ...

INFO: Batch generation complete
OK: All teams generated successfully
```

### Example 4: Batch Generation

```bash
$ ./templates/orchestrator-generate.sh --all --force

OK: All dependencies found
OK: All required files exist

INFO: Batch generating 11 teams

... (validation for all teams) ...
INFO: Generating: rnd
OK: Generated: /roster/rites/rnd/agents/orchestrator.md

INFO: Generating: security
OK: Generated: /roster/rites/security/agents/orchestrator.md

... (9 more teams) ...

INFO: Batch generation complete
OK: All teams generated successfully
```

---

## Troubleshooting

### Missing Dependencies

**Error**:
```
ERROR: Missing required tools: yq jq
```

**Solution**:
```bash
# macOS
brew install yq jq

# Linux
apt-get install yq jq
```

### Config File Not Found

**Error**:
```
ERROR: Config file not found: /roster/rites/my-team/orchestrator.yaml
```

**Cause**: Team directory exists but no orchestrator.yaml file

**Solution**: Create `orchestrator.yaml` in team directory using schema template

### Schema Validation Failed

**Error**:
```
ERROR: Missing required field in orchestrator.yaml: routing
```

**Cause**: orchestrator.yaml missing required top-level field

**Solution**: Check orchestrator.yaml against schema at `/roster/schemas/orchestrator.yaml.schema.json`

### Specialist Not Found in Workflow

**Error**:
```
ERROR: Specialist 'bad-specialist-name' not found in workflow.yaml: /roster/rites/my-team/workflow.yaml
```

**Cause**: orchestrator.yaml references specialist that doesn't exist in workflow.yaml

**Solution**:
1. Check workflow.yaml `phases[].agent` names
2. Update routing table in orchestrator.yaml to match actual specialist names

### Post-Generation Validation Failed

**Error**:
```
ERROR: Post-generation validation failed: /var/folders/.../tmp.xyz
```

**Cause**: Generated file fails validate-orchestrator.sh checks

**Diagnosis**:
```bash
# Run validator directly to see what failed
./templates/validate-orchestrator.sh /roster/rites/my-team/agents/orchestrator.md
```

**Common Causes**:
- Unreplaced placeholders (template substitution failed)
- Missing required sections (template issue)
- Duplicate section headers (template duplicate)

### Output File Exists

**Error**:
```
ERROR: Output file exists: /roster/rites/rnd/agents/orchestrator.md (use --force to overwrite)
```

**Cause**: orchestrator.md already exists and --force not specified

**Solution**:
```bash
# Option 1: Use --force to overwrite
./templates/orchestrator-generate.sh rnd --force

# Option 2: Delete existing file first
rm /roster/rites/rnd/agents/orchestrator.md
./templates/orchestrator-generate.sh rnd
```

---

## Architecture & Integration Points

### Data Flow

```
orchestrator.yaml ─┐
                   ├─> Schema Validation ──┐
workflow.yaml  ───┤                        ├─> Template Processing ─> orchestrator.md
                   ├─> Workflow Validation ┤
template ──────────┘                        └─> Post-Validation
                                                  (validate-orchestrator.sh)
```

### Generated File Quality

All generated files pass:

1. **Schema Validation**: orchestrator.yaml conforms to JSON schema
2. **Workflow Validation**: All specialists exist in workflow.yaml
3. **Substitution Check**: No unreplaced placeholders remain
4. **Post-Generation Validation**: 10-rule validation from validate-orchestrator.sh
5. **Frontmatter Parsing**: Compatible with swap-rite.sh grep/sed patterns
6. **Markdown Syntax**: Valid Markdown structure

### Integration with Existing Tools

**swap-rite.sh**: Generated frontmatter parses cleanly without modifications

**CEM Sync**: Generated files treated as standard agent files (no special handling)

**validate-orchestrator.sh**: All generated files pass all 10 validation rules

---

## Advanced Usage

### Custom Colors

Update `team.color` in orchestrator.yaml:

```yaml
team:
  color: purple          # Named color
  # OR
  color: "#FF5733"       # Hex color
```

### Multiple Complexity Levels

Different teams have different complexity enums. The generator extracts these from workflow.yaml:

**rnd**: SPIKE, EVALUATION, MOONSHOT
**ecosystem**: PATCH, MODULE, SYSTEM, MIGRATION
**security**: PATCH, FEATURE, SYSTEM

The `{{COMPLEXITY_ENUM}}` placeholder is automatically populated.

### Team-Specific Anti-Patterns

Add custom anti-patterns in orchestrator.yaml:

```yaml
antipatterns:
  - "Treating PATCH as SYSTEM (different scope requires different phases)"
  - "Skipping threat modeling for 'simple' features"
  - "Accepting unmitigated CRITICAL vulnerabilities"
```

### Cross-Team Protocol

If your team coordinates with other teams:

```yaml
cross_team_protocol: "Ecosystem-pack acts as hub for infrastructure changes. When escalating to user, notify all active rite leads."
```

---

## Development & Testing

### Run Integration Tests

```bash
cd /roster

# Test file generation for 5 teams
bash << 'EOF'
for team in rnd security ecosystem docs 10x-dev; do
  echo "Testing $team..."
  ./templates/orchestrator-generate.sh "$team" --force
  ./templates/validate-orchestrator.sh "rites/$team/agents/orchestrator.md"
done
EOF
```

### Validate All Teams

```bash
./templates/orchestrator-generate.sh --all --validate-only
```

### Regenerate All Teams

```bash
./templates/orchestrator-generate.sh --all --force
```

---

## File Manifest

| File | Purpose | Status |
|------|---------|--------|
| `/roster/templates/orchestrator-generate.sh` | Production generator | Complete |
| `/roster/templates/orchestrator-base.md.tpl` | Canonical template | Complete |
| `/roster/templates/validate-orchestrator.sh` | Validation script (10 rules) | Complete |
| `/roster/schemas/orchestrator.yaml.schema.json` | JSON Schema | Complete |
| `/roster/rites/*/orchestrator.yaml` | Team configs (11 total) | Complete |
| `/roster/rites/*/agents/orchestrator.md` | Generated agents (11 total) | Generated |

---

## Compatibility & Support

**Supported Platforms**:
- macOS (BSD sed)
- Linux (GNU sed)

**Shell**: bash 4.0+

**Required Tools**: yq (v4.x), jq, sed, awk

**Backward Compatibility**:
- Generated files compatible with swap-rite.sh (verified)
- No CEM changes required
- No breaking changes to existing satellites

---

## Next Steps

### Phase 3 (Compatibility Tester)
- Satellite compatibility verification matrix
- Integration with CI/CD pipeline
- Automated regeneration on workflow.yaml changes

### Phase 5 (CI Integration)
- Build pipeline validation
- Drift detection (orchestrator.yaml vs generated .md)
- Automatic PR generation for regenerations

---

## See Also

- **Schema Reference**: `/roster/schemas/orchestrator.yaml.schema.json`
- **Validation Rules**: `/roster/templates/validate-orchestrator.sh`
- **Context Design**: `/skeleton_claude/docs/ecosystem/CONTEXT-orchestrator-template-schema.md`
- **Integration Points**: `/roster/schemas/INTEGRATION-POINTS.md`

---

**Generated**: 2025-12-29
**Generator Version**: 1.0 (Production)
**Template Version**: 1.0

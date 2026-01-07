# Orchestrator YAML Schema Reference

> Complete specification for orchestrator.yaml configuration files.

## File Locations

**Template location**: `rites/{rite-name}/orchestrator.yaml`

**Example location**: `rites/rnd-pack/orchestrator.yaml`

## Schema Overview

The schema enforces structure and consistency across all team configurations. Each field maps directly to template placeholders.

```yaml
team:                    # REQUIRED: Team metadata
  name: string
  domain: string
  color: string

frontmatter:             # REQUIRED: Agent file header
  role: string
  description: string

routing:                 # REQUIRED: Specialist routing table
  specialist-one: string
  specialist-two: string
  # ... 3-6 total specialists

workflow_position:       # REQUIRED: Workflow context
  upstream: string
  downstream: string

handoff_criteria:        # REQUIRED: Phase exit gates
  specialist-one:
    - string
    - string
  specialist-two:
    - string

skills:                  # REQUIRED: Skill references
  - string
  - string

antipatterns:            # OPTIONAL: Team-specific anti-patterns
  - string

cross_team_protocol:     # OPTIONAL: Hub responsibilities
  subject: string

extension_points:        # OPTIONAL: Team customizations
  key: string
```

## Field Reference

### team.name (REQUIRED)

**Type**: String
**Pattern**: `^[a-z0-9-]+$` (lowercase, numbers, hyphens only)
**Length**: 1-64 characters
**Example**: `rnd-pack`, `security-pack`, `my-new-team`

Maps to frontmatter `name:` field. Must be "orchestrator" (this is handled automatically).

```yaml
team:
  name: my-team  # Used internally; frontmatter always "orchestrator"
```

### team.domain (REQUIRED)

**Type**: String
**Purpose**: Describe what your team does in 1-2 sentences
**Example**: "Coordinates research, prototyping, and architectural planning"

Becomes part of the Consultation Role section.

```yaml
team:
  domain: "Technology evaluation and prototyping specialist team"
```

### team.color (REQUIRED)

**Type**: String (CSS color or named color)
**Format**: `#RRGGBB` (hex) or color name
**Examples**: `purple`, `#FF6B6B`, `blue`

Used for visual identification in UI and documentation.

```yaml
team:
  color: purple
```

### frontmatter.role (REQUIRED)

**Type**: String
**Length**: 40-80 characters (one line)
**Purpose**: One-sentence summary of orchestrator's role
**Template field**: `{{ROLE}}`

This appears in the agent description. Keep concise.

```yaml
frontmatter:
  role: "Coordinates rnd-pack phases for technology evaluation and prototyping"
```

Maps to generated orchestrator.md frontmatter:
```yaml
role: "Coordinates rnd-pack phases for technology evaluation and prototyping"
```

### frontmatter.description (REQUIRED)

**Type**: String (can be multi-line)
**Purpose**: 2-3 sentence description of when to use this team
**Template field**: `{{DESCRIPTION}}`

Explains the consultation role and triggers.

```yaml
frontmatter:
  description: "Coordinates rnd-pack phases for technology evaluation and prototyping. Use when: evaluating emerging technologies, building proof-of-concept, or designing long-term architecture. Triggers: evaluate, prototype, research, innovation."
```

### routing (REQUIRED)

**Type**: Object (map of specialist names to routing conditions)
**Min specialists**: 3
**Max specialists**: 6
**Keys**: Specialist names from workflow.yaml
**Values**: Routing conditions (when to route to this specialist)

**Template field**: `{{ROUTING_TABLE}}`

Each key must match an agent name in workflow.yaml phases. Each value is a conditional description.

```yaml
routing:
  integration-researcher: "Needs research on technology integration"
  technology-scout: "Needs evaluation of emerging tool"
  prototype-engineer: "Needs proof-of-concept implementation"
  moonshot-architect: "Needs long-term architectural design"
```

**Validation**:
- Generator verifies each specialist exists in workflow.yaml
- Generator creates routing table in Routing Decisions section
- Specialists appear in correct order in workflow diagram

### routing.{specialist-name} (REQUIRED)

**Type**: String
**Length**: 40-100 characters
**Purpose**: Condition for routing to this specialist

Explains when the orchestrator routes to this specialist.

```yaml
# Example:
integration-researcher: "Needs research on how technologies integrate with existing systems"
```

Maps to routing table in generated orchestrator.md:
```markdown
| When | Route To | Prerequisites |
|------|----------|---------------|
| Needs research on how technologies integrate | integration-researcher | Clear problem statement |
```

### workflow_position.upstream (REQUIRED)

**Type**: String
**Purpose**: Which team comes before yours in typical workflow
**Example**: "integration-researcher", "user-researcher", "None" if first

```yaml
workflow_position:
  upstream: "None (can start projects)"
```

Used in Position in Workflow section. Helps teams understand integration context.

### workflow_position.downstream (REQUIRED)

**Type**: String
**Purpose**: Which team comes after yours in typical workflow
**Example**: "moonshot-architect", "principal-engineer", "None" if final

```yaml
workflow_position:
  downstream: "Principal engineer or implementation team"
```

### handoff_criteria (REQUIRED)

**Type**: Object (map of specialist names to acceptance criteria arrays)
**Keys**: Must match routing specialist names
**Values**: Array of 1-5 criteria strings each

**Template field**: `{{HANDOFF_CRITERIA}}`

Defines phase exit gates. Each specialist has acceptance criteria their work must meet.

```yaml
handoff_criteria:
  integration-researcher:
    - "Root cause identified at file:line"
    - "Reproduction steps documented"
    - "Success criteria measurable"
    - "All artifacts verified via Read tool"

  technology-scout:
    - "Technology evaluated against selection criteria"
    - "Comparison matrix completed"
    - "Recommendation with rationale"
    - "All artifacts verified via Read tool"
```

**Rules**:
- Each specialist must have at least 1 criterion
- Criteria are rendered as checkboxes in generated orchestrator.md
- Always include "All artifacts verified via Read tool" for consistency

### handoff_criteria.{specialist-name} (REQUIRED)

**Type**: Array of strings
**Min items**: 1
**Max items**: 5
**Format**: Complete, actionable criteria

Each criterion becomes a checkbox in the generated orchestrator.

```yaml
handoff_criteria:
  analyst:
    - "Root cause at file:line identified"
    - "Reproduction confirmed"
    - "Success criteria defined"
    - "Artifacts verified via Read tool"
```

Generated as:
```markdown
## Handoff Criteria

### {specialist-name} → Next Phase

- [ ] Root cause at file:line identified
- [ ] Reproduction confirmed
- [ ] Success criteria defined
- [ ] Artifacts verified via Read tool
```

### skills (REQUIRED)

**Type**: Array of strings
**Min items**: 1
**Max items**: 10
**Format**: `@skill-name brief description`

**Template field**: `{{SKILLS_REFERENCE}}`

References to related skills and documentation.

```yaml
skills:
  - "@doc-rnd for artifact templates and checklists"
  - "@rnd-ref for technology scouting workflows"
  - "@prompting for invoking agents and teams"
```

**Rules**:
- Must start with `@` symbol
- Skill name is lowercase with hyphens
- Space after skill name, then brief description
- Description explains relevance to your team

**Validation**: No automated checking of skill existence (yet), but follow naming conventions.

### antipatterns (OPTIONAL)

**Type**: Array of strings
**Default**: Empty array (uses canonical anti-patterns only)
**Max items**: 10

Team-specific anti-patterns that augment the canonical list.

```yaml
antipatterns:
  - "Designing without prototyping—always build proof-of-concept"
  - "Adopting technology without evaluation—use scout first"
  - "Skipping long-term vision—architect designs for 2+ years"
```

**Pattern**:
- Describe what NOT to do
- Explain why it's a trap
- Keep to 1-2 sentences

If empty or omitted, generated orchestrator uses only canonical anti-patterns.

### cross_team_protocol (OPTIONAL)

**Type**: String
**Default**: Empty string (no cross-rite responsibilities)
**Purpose**: Responsibilities when other teams consult this orchestrator

Some orchestrators act as hubs that other teams consult. This documents that protocol.

```yaml
cross_team_protocol: |
  When other teams request technology scouting, route to technology-scout.
  When other teams request long-term architecture, route to moonshot-architect.
```

Appears in a dedicated section in generated orchestrator.

### extension_points (OPTIONAL)

**Type**: Object
**Default**: Empty object (no extensions)
**Purpose**: Team-specific customizations and additions

Allows teams to add custom content to generated orchestrator without editing template.

```yaml
extension_points:
  examples: |
    ### Example: Evaluating a New JavaScript Framework
    1. Technology Scout identifies candidate
    2. Integration Researcher studies integration requirements
    3. Prototype Engineer builds proof-of-concept
    4. Moonshot Architect evaluates long-term fit
```

Each extension point creates a new section in the generated orchestrator.

---

## Complete Schema (JSON Reference)

For validation purposes, the official schema is:

**Location**: `/roster/schemas/orchestrator.yaml.schema.json`

Key constraints:
- Required fields: team, frontmatter, routing, workflow_position, handoff_criteria, skills
- Routing must have 3-6 specialists
- All specialist names must exist in workflow.yaml
- Handoff criteria must match routing specialist names

---

## Common Configuration Patterns

### Pattern 1: Simple Linear Pipeline

4-phase workflow with no hubs or special protocols.

```yaml
team:
  name: my-team
  domain: "Specialist team for X"
  color: blue

frontmatter:
  role: "Coordinates my-team phases for X"
  description: "Handles X workflows. Use when: [conditions]. Triggers: [keywords]."

routing:
  specialist-one: "When [condition]"
  specialist-two: "When [condition]"
  specialist-three: "When [condition]"
  specialist-four: "When [condition]"

workflow_position:
  upstream: "None (can start projects)"
  downstream: "None (final phase)"

handoff_criteria:
  specialist-one:
    - "Criterion 1"
    - "Criterion 2"
    - "Artifacts verified via Read tool"
  specialist-two:
    - "Criterion 1"
    - "Criterion 2"
    - "Artifacts verified via Read tool"
  specialist-three:
    - "Criterion 1"
    - "Artifacts verified via Read tool"
  specialist-four:
    - "Criterion 1"
    - "Artifacts verified via Read tool"

skills:
  - "@skill-one for purpose"
  - "@skill-two for purpose"
```

**Example**: rnd-pack, security-pack, doc-rite-pack

### Pattern 2: Hub Coordination

5-6 specialists with one coordinator that routes to multiple paths.

```yaml
team:
  name: ecosystem-pack
  domain: "Coordinates infrastructure changes across satellites"
  color: green

routing:
  ecosystem-analyst: "Initial issue diagnosis"
  context-architect: "Design phase"
  integration-engineer: "Implementation"
  documentation-engineer: "Documentation"
  compatibility-tester: "Validation"

workflow_position:
  upstream: "Any team"
  downstream: "Rollout team"

handoff_criteria:
  ecosystem-analyst:
    - "Root cause identified"
    - "Gap analysis complete"
    - "Artifacts verified via Read tool"
  # ... continue for each specialist
```

**Example**: ecosystem-pack

### Pattern 3: Domain-Specific Complexity

Team with custom complexity levels matching their domain.

```yaml
team:
  name: security-pack
  domain: "Security threat modeling and compliance"
  color: red

# Complexity enum defined in workflow.yaml:
# PATCH | FEATURE | SYSTEM
# (automatically used in generated orchestrator)

routing:
  threat-modeler: "Identify threats"
  compliance-architect: "Design compliance controls"
  penetration-tester: "Test security"
  security-reviewer: "Final review"
```

**Generator automatically pulls complexity from workflow.yaml**, so no manual specification needed.

---

## Validation Rules

When you save your orchestrator.yaml, ensure:

| Rule | How to Check |
|------|--------------|
| Valid YAML | File parses with `yq` or standard YAML parser |
| All required fields | team, frontmatter, routing, workflow_position, handoff_criteria, skills |
| Specialist names | Each routing specialist exists in workflow.yaml agents |
| Handoff matches routing | Each routing specialist has handoff_criteria entry |
| 3-6 specialists | Check routing has at least 3 and at most 6 entries |
| Color is valid | Valid hex (#RRGGBB) or CSS color name |
| Descriptions not empty | role and description fields must have text |
| Skills format | Each skill starts with @ and has description |

**Automated check**: Run generator, then validator
```bash
orchestrator-generate.sh my-team
validate-orchestrator.sh .claude/rites/my-team/agents/orchestrator.md
```

---

## Template Field Mapping

Reference for internal template substitution:

| YAML Field | Template Placeholder | Generated Section |
|------------|---------------------|-------------------|
| team.name | (implicit) | Frontmatter name field |
| team.domain | Part of DESCRIPTION | Consultation Role |
| team.color | {{COLOR}} | Frontmatter color field |
| frontmatter.role | {{ROLE}} | Frontmatter role field |
| frontmatter.description | {{DESCRIPTION}} | Description in Consultation Role |
| routing | {{ROUTING_TABLE}} | Routing Decisions section |
| handoff_criteria | {{HANDOFF_CRITERIA}} | Handoff Criteria sections |
| skills | {{SKILLS_REFERENCE}} | Skills Reference section |

---

## Migration Path

### Extracting Configuration from Existing Orchestrator

If you have a hand-written orchestrator.md and want to extract it to orchestrator.yaml:

1. **Read routing table** from "Routing Decisions" section
2. **Copy specialist names** to routing object
3. **Copy routing conditions** as values
4. **Extract handoff criteria** from each specialist section
5. **Extract skills** from Skills Reference section
6. **Save as orchestrator.yaml** in proper location
7. **Regenerate** and compare to original

### Creating orchestrator.yaml from Scratch

1. **Copy an example**: Use doc-rite-pack or rnd-pack as template
2. **Update team metadata**: name, domain, color
3. **Update frontmatter**: role and description for your team
4. **Update routing**: Specialist names and conditions
5. **Update handoff_criteria**: Phase exit gates for each specialist
6. **Update skills**: Related skill references
7. **Run generator**: Verify output matches your needs
8. **Commit both files**: orchestrator.yaml and orchestrator.md

---

## Schema Evolution

### Backward Compatibility

Current schema supports all 11 teams without changes. Future evolution handled through:

1. **Optional fields**: New features added as optional (like extension_points)
2. **Schema versioning**: template_version field (future)
3. **Migration tools**: convert-orchestrator.yaml v1→v2 (future)

### Adding Custom Fields (Not Yet Supported)

If you need custom fields beyond schema:
1. **Use extension_points**: Add rite-specific configuration here
2. **Request feature**: Propose to tech team for next template evolution
3. **Don't edit template**: Keep modifications in YAML, not in base template

---

**Last Updated**: 2025-12-29
**Schema Version**: 1.0 (Phase 3-4)
**Validation**: Phase 3 verified all 11 teams

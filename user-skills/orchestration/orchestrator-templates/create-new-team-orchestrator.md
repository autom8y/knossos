# Step-by-Step: Create a New Team Orchestrator

> Follow this guide to create a production-ready orchestrator for your new agent team.

## Timeline

**Simple team** (linear pipeline): 15 minutes
**Complex team** (5+ specialists, custom protocol): 30-45 minutes

## Prerequisites

- [ ] You have a team directory created: `teams/my-team/`
- [ ] You have a workflow.yaml with your team's specialists: `teams/my-team/workflow.yaml`
- [ ] You know the team's: name, domain, color, and specialist roles
- [ ] You have access to generator and validator scripts

## Phase 1: Create orchestrator.yaml (10 minutes)

### Step 1.1: Create the File

Create `teams/my-team/orchestrator.yaml`:

```bash
touch teams/my-team/orchestrator.yaml
```

### Step 1.2: Add Team Metadata

Open the file and add basic team information:

```yaml
team:
  name: my-team           # Lowercase, hyphens only
  domain: "Brief description of team purpose"
  color: blue             # Valid hex or CSS color
```

**Example** (doc-team-pack):
```yaml
team:
  name: doc-team-pack
  domain: "Documentation and technical writing"
  color: "#4A90E2"        # Or just: blue
```

**Choose a color** that visually distinguishes your team. Use hex (#RRGGBB) or CSS names: red, blue, purple, green, orange, etc.

### Step 1.3: Add Frontmatter Section

```yaml
frontmatter:
  role: "One-line summary of what this orchestrator does"
  description: |
    Multi-line description that explains:
    1. What your team does
    2. When to use it
    3. Key triggers
```

**Example** (security-pack):
```yaml
frontmatter:
  role: "Coordinates security-pack threat modeling and compliance"
  description: |
    Coordinates security-pack phases for threat modeling, compliance architecture,
    and penetration testing. Use when: designing security controls, planning compliance,
    or assessing security vulnerabilities. Triggers: threat-model, security-review,
    compliance-planning, penetration-test.
```

**Keep description under 300 characters** (should fit in one paragraph).

### Step 1.4: Add Routing Table

```yaml
routing:
  specialist-one: "When [condition]"
  specialist-two: "When [condition]"
  specialist-three: "When [condition]"
```

**Important**: Specialist names MUST match workflow.yaml agent definitions.

**Check your workflow.yaml**:
```bash
grep "agent:" teams/my-team/workflow.yaml
```

This shows exact specialist names. Use those names, not variations.

**Example** (rnd-pack):
```yaml
routing:
  integration-researcher: "Needs research on technology integration paths"
  technology-scout: "Needs evaluation of emerging tools and frameworks"
  prototype-engineer: "Needs proof-of-concept implementation"
  moonshot-architect: "Needs long-term (2+ year) architectural design"
```

**Guidelines**:
- 3-6 specialists total (min/max enforced)
- Routing conditions are 40-100 characters
- Conditions describe WHEN, not HOW
- List in logical execution order

### Step 1.5: Add Workflow Position

```yaml
workflow_position:
  upstream: "Which team typically comes before yours"
  downstream: "Which team typically comes after yours"
```

**Example** (ecosystem-pack):
```yaml
workflow_position:
  upstream: "Any team (multi-phase coordination)"
  downstream: "Integration teams (for implementation)"
```

**For first team in organization**: `upstream: "None (can initiate projects)"`
**For last team in organization**: `downstream: "None (final phase)"`

### Step 1.6: Add Handoff Criteria

```yaml
handoff_criteria:
  specialist-one:
    - "Criterion 1"
    - "Criterion 2"
    - "Artifacts verified via Read tool"
  specialist-two:
    - "Criterion 1"
    - "Artifacts verified via Read tool"
  specialist-three:
    - "Criterion 1"
    - "Artifacts verified via Read tool"
```

**Important**: Each specialist in routing MUST have handoff_criteria entry.

**Example** (rnd-pack):
```yaml
handoff_criteria:
  integration-researcher:
    - "Integration requirements documented"
    - "Compatibility analysis complete"
    - "Tech stack recommendations with rationale"
    - "All artifacts verified via Read tool"

  technology-scout:
    - "Technologies evaluated against selection criteria"
    - "Comparison matrix completed"
    - "Top 3 recommendations with trade-off analysis"
    - "All artifacts verified via Read tool"

  prototype-engineer:
    - "Proof-of-concept implementation complete"
    - "Key integration points validated"
    - "Lessons learned documented"
    - "All artifacts verified via Read tool"

  moonshot-architect:
    - "2+ year architectural vision designed"
    - "Migration path from current state"
    - "Success metrics and evolution strategy"
    - "All artifacts verified via Read tool"
```

**Guidelines**:
- 2-4 criteria per specialist
- Always include "All artifacts verified via Read tool"
- Make criteria measurable and testable
- Each becomes a checkbox in generated orchestrator

### Step 1.7: Add Skills References

```yaml
skills:
  - "@skill-name brief description of relevance"
  - "@another-skill why this team uses it"
```

**Example** (security-pack):
```yaml
skills:
  - "@security-ref for threat modeling workflows and controls"
  - "@doc-security for security documentation templates"
  - "@prompting for invoking security agents"
```

**Rules**:
- Start with @
- Skill name lowercase with hyphens
- Space, then brief description (20-40 chars)
- 1-10 skills per team

## Phase 2: Generate orchestrator.md (3 minutes)

### Step 2.1: Run the Generator

```bash
cd $ROSTER_HOME
/roster/templates/orchestrator-generate.sh my-team
```

**Expected output**:
```
Generating orchestrator for my-team...
Template: /roster/templates/orchestrator-base.md.tpl
Config: teams/my-team/orchestrator.yaml
Output: teams/my-team/agents/orchestrator.md
Generation complete!
```

### Step 2.2: Check for Errors

If generator fails:

**Error: "Specialist not found in workflow.yaml"**
- Check spelling in orchestrator.yaml routing table
- Verify specialist exists in workflow.yaml
- Fix and retry

**Error: "Template file not found"**
- Verify template exists: `/roster/templates/orchestrator-base.md.tpl`
- Check file permissions
- Try again

**Error: "YAML parsing failed"**
- Check orchestrator.yaml for syntax errors
- Ensure proper YAML indentation
- Use online YAML validator if needed

## Phase 3: Validate Output (3 minutes)

### Step 3.1: Run Validation

```bash
/roster/templates/validate-orchestrator.sh \
  teams/my-team/agents/orchestrator.md
```

**Expected output**:
```
Validating orchestrator.md...
✓ File exists and is readable
✓ No unreplaced placeholders
✓ Valid YAML frontmatter
✓ All required sections present
✓ Specialist consistency verified
✓ No duplicate sections
✓ Handoff criteria properly formatted
✓ Consultation protocol present
✓ Skill references valid
✓ Markdown syntax valid

VALIDATION PASSED (exit code 0)
```

### Step 3.2: Address Validation Failures

**If validation fails**, review the specific error:

```
ERROR: Found 2 unreplaced placeholder(s)
  Line 45: {{WORKFLOW_DIAGRAM}}
  Line 112: {{ROUTING_TABLE}}
```

**Solution**: This indicates a generator bug. Options:
1. Re-run generator
2. Check orchestrator.yaml for syntax errors
3. Report issue with exact error message

## Phase 4: Review Generated Output (5 minutes)

### Step 4.1: Preview the File

```bash
head -80 teams/my-team/agents/orchestrator.md
```

**Check these sections**:

**Frontmatter** (lines 1-10):
```yaml
---
name: orchestrator
role: "Your role here"
description: "Your description here"
tools: Read, Skill
model: opus
color: your-color
---
```

**Consultation Role** (section heading):
- Describes your team's purpose
- References specialists by name
- Explains when to use

**Routing Decisions** (table):
- Lists all your specialists
- Shows routing conditions
- Correct specialist names from workflow.yaml

**Handoff Criteria** (sections):
- One section per specialist
- Checkbox format (- [ ])
- Your criteria from YAML

**Skills Reference** (list):
- All skills present
- @skill-name format
- Descriptions included

### Step 4.2: Compare Specialist Names

Verify specialist names match workflow.yaml exactly:

```bash
# Show specialists in YAML
echo "=== orchestrator.yaml routing ==="
yq '.routing | keys[]' teams/my-team/orchestrator.yaml | sort

# Show specialists in generated orchestrator
echo "=== Generated routing table ==="
grep "^|.*→" teams/my-team/agents/orchestrator.md | awk '{print $3}' | sort | uniq
```

**These should match.**

## Phase 5: Commit to Git (2 minutes)

### Step 5.1: Stage Both Files

```bash
cd $ROSTER_HOME
git add .claude/teams/my-team/orchestrator.yaml
git add .claude/teams/my-team/agents/orchestrator.md
```

### Step 5.2: Create Commit

```bash
git commit -m "feat: add my-team orchestrator configuration and generated agent"
```

**Commit message pattern**:
```
feat: add {team-name} orchestrator configuration and generated agent

- Creates orchestrator.yaml with team routing and handoff criteria
- Generates orchestrator.md from canonical template
- Validates against schema and structural rules
- Ready for team activation via swap-team.sh
```

### Step 5.3: Verify Commit

```bash
git log -1 --stat
```

Should show both files added:
```
.claude/teams/my-team/agents/orchestrator.md
.claude/teams/my-team/orchestrator.yaml
```

## Phase 6: Test Team Activation (2 minutes)

### Step 6.1: Activate Your Team

```bash
./swap-team.sh my-team
```

**Expected output**:
```
Switching to team: my-team
✓ Team artifacts found
✓ Orchestrator loaded
✓ CLAUDE.md updated
✓ Team activation complete

Current team: my-team
```

### Step 6.2: Verify Orchestrator Loads

```bash
echo "=== Current Orchestrator Role ==="
grep "^role:" .claude/agents/orchestrator.md

echo "=== Team Color ==="
grep "^color:" .claude/agents/orchestrator.md
```

Should show your team's configuration.

### Step 6.3: Check Specialist References

```bash
echo "=== Specialists in Routing Table ==="
grep "^|" .claude/agents/orchestrator.md | grep -v "^|---" | grep -v "^| When"
```

All your specialists should appear.

## Common Issues and Fixes

### Issue: Generator says "Specialist not found"

**Cause**: Your orchestrator.yaml specialist names don't match workflow.yaml

**Fix**:
```bash
# Check exact names in workflow.yaml
yq '.phases[] | .agent' teams/my-team/workflow.yaml

# Update orchestrator.yaml routing to match exactly
# Then regenerate
```

### Issue: Validation says "No placeholders replaced"

**Cause**: Generator encountered an error

**Fix**:
```bash
# Verify orchestrator.yaml syntax
yq . teams/my-team/orchestrator.yaml

# Try manual validation
head -20 teams/my-team/agents/orchestrator.md

# Regenerate
/roster/templates/orchestrator-generate.sh my-team
```

### Issue: Frontmatter in generated file is wrong

**Cause**: YAML config errors in team metadata

**Fix**:
```bash
# Review your orchestrator.yaml team section
yq '.team' teams/my-team/orchestrator.yaml

# All fields required and valid:
# - name: lowercase-hyphens
# - domain: non-empty string
# - color: valid CSS color
```

### Issue: Handoff criteria show as "- " instead of checkboxes

**Cause**: YAML formatting error

**Fix**:
```yaml
# WRONG: Missing dash before bracket
handoff_criteria:
  specialist:
    [ ] Criterion

# RIGHT: Dash before bracket
handoff_criteria:
  specialist:
    - [ ] Criterion  # Note: starts with dash and space
```

### Issue: Specialist names appear capitalized in generated file

**Cause**: This is normal behavior

**Solution**: Names from workflow.yaml appear as-is. If you want different capitalization:
1. Update names in workflow.yaml (if you control it)
2. Or use display_name extension point (future feature)
3. Accept current names as defined in workflow.yaml

## Success Criteria Checklist

Before considering orchestrator complete:

- [ ] orchestrator.yaml created with all required fields
- [ ] All specialists in routing exist in workflow.yaml
- [ ] Specialist count is 3-6
- [ ] Handoff criteria defined for each specialist
- [ ] Generator runs without errors
- [ ] Validator passes (exit code 0)
- [ ] All specialist names match workflow.yaml exactly
- [ ] Skills references use @skill-name format
- [ ] Color is valid CSS color or hex
- [ ] Both orchestrator.yaml and orchestrator.md committed to git
- [ ] Team activation works via swap-team.sh
- [ ] Frontmatter parses correctly

## Next Steps

After successful creation:

1. **Test with swap-team.sh**: Activate and verify
2. **Add to documentation**: Reference in team guide
3. **Coordinate with Phase 5**: CI/CD integration when ready
4. **Share with team**: Team members know how to use it

## Example: Complete doc-team-pack Walkthrough

For reference, here's a real example of doc-team-pack configuration:

**File**: `teams/doc-team-pack/orchestrator.yaml`

```yaml
team:
  name: doc-team-pack
  domain: "Documentation design and technical writing"
  color: "#4A90E2"

frontmatter:
  role: "Coordinates doc-team-pack documentation phases"
  description: |
    Coordinates doc-team-pack phases for documentation audit, information architecture,
    and technical writing. Use when: auditing documentation, reorganizing information,
    or writing comprehensive guides. Triggers: audit, documentation, rewrite, consolidate.

routing:
  doc-auditor: "Documentation needs review or has quality issues"
  information-architect: "Information structure needs redesign"
  tech-writer: "Documentation needs authoring or rewriting"
  doc-reviewer: "Documentation needs technical accuracy validation"

workflow_position:
  upstream: "Any team (can start documentation work)"
  downstream: "Integration teams (documentation enables implementation)"

handoff_criteria:
  doc-auditor:
    - "Audit completed with gap analysis"
    - "Staleness and quality issues identified"
    - "Recommendations prioritized"
    - "All artifacts verified via Read tool"

  information-architect:
    - "Information architecture designed"
    - "Content organization structure approved"
    - "Cross-references validated"
    - "All artifacts verified via Read tool"

  tech-writer:
    - "Documentation written and technically accurate"
    - "Examples tested and runnable"
    - "Scannability optimized"
    - "All artifacts verified via Read tool"

  doc-reviewer:
    - "Technical accuracy verified"
    - "Information architecture validated"
    - "No dead links or broken references"
    - "All artifacts verified via Read tool"

skills:
  - "@documentation for templates and standards"
  - "@doc-reviews for audit and architecture"
  - "@standards for naming and formatting conventions"
```

**Result**: Generates a complete orchestrator.md with all sections properly populated.

---

**Status**: Production guide
**Last Updated**: 2025-12-29
**Tested**: Phase 3 validation (all 11 teams)

# Troubleshooting Guide

> Solutions for common orchestrator templating issues.

## Quick Diagnosis

Start here to identify your problem:

| Symptom | Likely Cause | Go To Section |
|---------|--------------|---------------|
| Generator fails immediately | Configuration error | Section 1 |
| Generated file has placeholder like `{{ROUTING_TABLE}}` | Substitution failed | Section 2 |
| Validation fails | Output format issue | Section 3 |
| swap-team.sh can't parse frontmatter | Frontmatter corrupted | Section 4 |
| Specialist names don't match | Naming inconsistency | Section 5 |
| Generated file looks different from expected | Template mismatch | Section 6 |
| Handoff criteria are wrong | YAML parsing issue | Section 7 |

## Section 1: Configuration Errors

### Symptom: Generator exits immediately

**Error message**:
```
Error: orchestrator.yaml not found
Error: Invalid YAML syntax
Error: Required field missing
```

### Root Cause

Your orchestrator.yaml has syntax errors or missing required fields.

### Diagnosis

```bash
# Check file exists
ls -la teams/my-team/orchestrator.yaml

# Validate YAML syntax
yq . teams/my-team/orchestrator.yaml

# Check required fields
yq '.team.name' teams/my-team/orchestrator.yaml
yq '.frontmatter.role' teams/my-team/orchestrator.yaml
yq '.routing | keys' teams/my-team/orchestrator.yaml
```

### Solutions

**If file doesn't exist**:
```bash
# Create from template
cp teams/doc-team-pack/orchestrator.yaml \
   teams/my-team/orchestrator.yaml

# Edit with your values
nano teams/my-team/orchestrator.yaml
```

**If YAML syntax error**:
```bash
# Check for common issues:
# 1. Indentation (must be spaces, not tabs)
# 2. Missing colons after keys
# 3. Unclosed quotes
# 4. Invalid characters

# Use online YAML validator as backup
# https://www.yamllint.com/

# Fix and retry
yq . teams/my-team/orchestrator.yaml
```

**If required field missing**:
```bash
# Check what's missing
yq keys teams/my-team/orchestrator.yaml

# Required: team, frontmatter, routing, workflow_position, handoff_criteria, skills

# Add missing section
cat >> teams/my-team/orchestrator.yaml << 'EOF'
skills:
  - "@skill-one brief description"
EOF
```

## Section 2: Placeholder Substitution Failures

### Symptom: Generated file contains unreplaced placeholders

**Generated orchestrator.md contains**:
```markdown
{{ROUTING_TABLE}}
{{WORKFLOW_DIAGRAM}}
{{HANDOFF_CRITERIA}}
```

### Root Cause

Generator encountered error during template substitution.

### Diagnosis

```bash
# Check what placeholders remain
grep "{{" teams/my-team/agents/orchestrator.md

# Run generator with debugging
bash -x /roster/templates/orchestrator-generate.sh my-team 2>&1 | tail -50

# Check template file is readable
file /roster/templates/orchestrator-base.md.tpl
```

### Solutions

**Most common**: Template not found or not readable

```bash
# Verify template exists
ls -la /roster/templates/orchestrator-base.md.tpl

# Check permissions
stat /roster/templates/orchestrator-base.md.tpl | grep Access

# If not readable, fix permissions
chmod 644 /roster/templates/orchestrator-base.md.tpl
```

**If YAML parsing failed**:

```bash
# Validate orchestrator.yaml again
yq . teams/my-team/orchestrator.yaml > /dev/null
if [ $? -ne 0 ]; then
  echo "YAML is invalid"
  # Fix issues shown above in Section 1
fi
```

**If generator script has bug**:

```bash
# Delete bad output and retry
rm teams/my-team/agents/orchestrator.md

# Regenerate
/roster/templates/orchestrator-generate.sh my-team

# Validate
/roster/templates/validate-orchestrator.sh \
  teams/my-team/agents/orchestrator.md
```

## Section 3: Validation Failures

### Symptom 3a: No placeholders replaced

**Error**:
```
ERROR: Found 154 unreplaced placeholder(s)
[Lists all {{PLACEHOLDER}} instances]
```

**Root cause**: Generator didn't run or failed silently

**Solution**: See Section 2 above

### Symptom 3b: Invalid YAML frontmatter

**Error**:
```
ERROR: Invalid YAML frontmatter
File does not have proper --- delimiters
```

**Root cause**: Generated frontmatter is malformed

**Diagnosis**:
```bash
# Check frontmatter structure
head -15 teams/my-team/agents/orchestrator.md

# Should look like:
# ---
# name: orchestrator
# role: "..."
# ...
# ---
```

**Solution**:
```bash
# If frontmatter is missing delimiters, regenerate
rm teams/my-team/agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team

# If still broken, check orchestrator.yaml:
yq '.frontmatter' teams/my-team/orchestrator.yaml
```

### Symptom 3c: Missing required sections

**Error**:
```
ERROR: Missing required sections
Missing: Consultation Role
Missing: Routing Decisions
```

**Root cause**: Template is incomplete or outdated

**Diagnosis**:
```bash
# List sections in generated file
grep "^## " teams/my-team/agents/orchestrator.md

# Should have 9-10 sections including:
# - Consultation Role
# - Tool Access
# - Consultation Protocol
# - Routing Decisions
# - Handoff Criteria
# - Anti-Patterns
# - Skills Reference
```

**Solution**:
```bash
# Verify template is complete
wc -l /roster/templates/orchestrator-base.md.tpl
# Should be 130-150 lines

# Regenerate
rm teams/my-team/agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team
```

### Symptom 3d: Markdown syntax errors

**Error**:
```
ERROR: Markdown syntax error
Unbalanced code block fences
```

**Root cause**: Generated file has unmatched backticks or brackets

**Diagnosis**:
```bash
# Count backticks
grep -o '```' teams/my-team/agents/orchestrator.md | wc -l
# Should be even number

# Find problematic areas
grep -n "^" teams/my-team/agents/orchestrator.md | \
  grep -B2 -A2 '```' | head -20
```

**Solution**:
```bash
# Regenerate (usually fixes this)
rm teams/my-team/agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team

# If still broken, check template
grep '```' /roster/templates/orchestrator-base.md.tpl | wc -l
# Should be even
```

## Section 4: Frontmatter Parsing Issues

### Symptom: swap-team.sh fails to parse orchestrator

**Error**:
```
Error: Could not parse orchestrator name
Error: Could not parse role
```

### Root Cause

Frontmatter doesn't match format swap-team.sh expects.

### Diagnosis

```bash
# Show actual frontmatter
head -10 teams/my-team/agents/orchestrator.md

# Should be:
# ---
# name: orchestrator
# role: "Your role"
# description: "Your description"
# tools: Read, Skill
# model: opus
# color: your-color
# ---

# Test if grep can parse it
sed -n '/^---$/,/^---$/p' teams/my-team/agents/orchestrator.md | \
  grep "^name:" | sed 's/^name:[[:space:]]*//'
```

### Solutions

**If parsing fails**:

1. Regenerate clean file:
```bash
rm teams/my-team/agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team
```

2. If still fails, check orchestrator.yaml fields:
```yaml
frontmatter:
  role: "Must be present and quoted"
  description: "Multi-line is OK"
```

3. Verify no special characters in role/description that break grep:
```bash
yq '.frontmatter.role' teams/my-team/orchestrator.yaml | \
  grep -E '[$\`"'\''|]'
# If any matches, those characters need escaping
```

## Section 5: Specialist Name Inconsistencies

### Symptom: Specialist names don't match workflow.yaml

**Error from generator**:
```
Error: Specialist 'technology-scout' not found in workflow.yaml
```

**Or noticed after generation**:
- Routing table shows different names than workflow.yaml
- Team activation shows wrong specialists

### Root Cause

Specialist names in orchestrator.yaml don't match workflow.yaml exactly.

### Diagnosis

```bash
# Get specialist names from workflow.yaml
echo "=== workflow.yaml specialists ==="
yq '.phases[].agent' teams/my-team/workflow.yaml | sort

# Get specialist names from orchestrator.yaml
echo ""
echo "=== orchestrator.yaml specialists ==="
yq '.routing | keys[]' teams/my-team/orchestrator.yaml | sort

# Compare
echo ""
echo "=== Differences ==="
diff <(yq '.phases[].agent' workflow.yaml | sort) \
     <(yq '.routing | keys[]' orchestrator.yaml | sort)
```

### Solutions

**If specialist in orchestrator but not workflow**:

```yaml
# WRONG in orchestrator.yaml:
routing:
  tech-scout: "..."  # Doesn't exist in workflow.yaml

# RIGHT (check exact name in workflow.yaml):
routing:
  technology-scout: "..."  # Matches workflow.yaml
```

**If specialist in workflow but not orchestrator**:

```yaml
# Add missing specialist to orchestrator.yaml:
routing:
  [existing-specialist]: "Existing condition"
  [missing-specialist]: "New condition"

# Add handoff criteria:
handoff_criteria:
  [missing-specialist]:
    - "Criterion 1"
    - "Criterion 2"
    - "Artifacts verified via Read tool"
```

**To verify exact names**:

```bash
# Interactive check
team="my-team"
echo "Workflow agents:"
yq '.phases[] | {agent, description}' teams/$team/workflow.yaml

echo ""
echo "Current orchestrator routing:"
yq '.routing' teams/$team/orchestrator.yaml
```

## Section 6: Output Differs from Expected

### Symptom: Generated orchestrator looks wrong

Examples:
- Wrong team color in frontmatter
- Specialist names capitalized differently
- Different sections than expected
- Different skill references

### Root Cause

Mismatch between orchestrator.yaml configuration and template generation.

### Diagnosis

```bash
# Compare YAML config with generated output
echo "=== Team Name ==="
yq '.team.name' orchestrator.yaml
grep "^name:" agents/orchestrator.md

echo ""
echo "=== Role ==="
yq '.frontmatter.role' orchestrator.yaml
grep "^role:" agents/orchestrator.md

echo ""
echo "=== Color ==="
yq '.team.color' orchestrator.yaml
grep "^color:" agents/orchestrator.md
```

### Solutions

**If output doesn't match config**:

1. Verify YAML is valid:
```bash
yq . orchestrator.yaml > /dev/null && echo "Valid" || echo "Invalid"
```

2. Regenerate:
```bash
rm agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team
```

3. Compare again:
```bash
diff <(yq '.team.color' orchestrator.yaml) \
     <(grep "^color:" agents/orchestrator.md | sed 's/^color:[[:space:]]*//')
```

## Section 7: Handoff Criteria Issues

### Symptom: Handoff criteria are wrong or missing

**Problems**:
- Criteria show as blank
- Checkbox format is wrong (shows `[ ]` instead of `- [ ]`)
- Missing criteria for some specialists
- Extra criteria appearing

### Root Cause

handoff_criteria in orchestrator.yaml doesn't match routing specialists.

### Diagnosis

```bash
# Get specialists from routing
echo "=== Specialists in routing ==="
yq '.routing | keys[]' orchestrator.yaml | sort

# Get specialists with criteria
echo ""
echo "=== Specialists with criteria ==="
yq '.handoff_criteria | keys[]' orchestrator.yaml | sort

# Find mismatches
echo ""
echo "=== Mismatches ==="
comm -23 <(yq '.routing | keys[]' orchestrator.yaml | sort) \
         <(yq '.handoff_criteria | keys[]' orchestrator.yaml | sort) | \
  sed 's/^/Missing criteria for: /'
```

### Solutions

**If specialist missing from handoff_criteria**:

```yaml
# Add missing entry
handoff_criteria:
  specialist-name:
    - "Criterion 1"
    - "Criterion 2"
    - "Artifacts verified via Read tool"
```

**If criteria are blank in generated file**:

```bash
# Check YAML criteria are not empty
yq '.handoff_criteria' orchestrator.yaml | head -20

# Should show list items, not empty values
# If empty, add criteria as shown above

# Regenerate
rm agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team
```

**If checkbox format is wrong**:

Generated file should show:
```markdown
## Handoff Criteria

### specialist-name → Next Phase

- [ ] Criterion 1
- [ ] Criterion 2
```

If showing something else, this is likely a generator bug. Check:

```bash
# Verify template has correct format
grep -A5 "HANDOFF_CRITERIA" /roster/templates/orchestrator-base.md.tpl
```

## Section 8: Schema Validation Issues

### Symptom: Custom extension points not working

**Problem**: You added `extension_points` but it doesn't appear in generated file.

### Root Cause

extension_points is in YAML but template doesn't have placeholder for them.

### Diagnosis

```bash
# Check if extension_points defined
yq '.extension_points' orchestrator.yaml

# Check template has placeholder
grep "EXTENSION" /roster/templates/orchestrator-base.md.tpl
```

### Solutions

**extension_points is optional and not yet fully implemented in Phase 4.**

For now, if you need custom content:
1. Use custom section under "Skills Reference"
2. Or wait for Phase 5 extension point implementation
3. Or manually edit generated orchestrator.md (note: will be overwritten on regeneration)

## Section 9: Team Activation Issues

### Symptom: swap-team.sh rejects new orchestrator

**Error**:
```
Error: orchestrator not found
Error: Cannot activate team
```

### Root Cause

Generated orchestrator not in expected location or structure.

### Diagnosis

```bash
# Check file exists
ls -la teams/my-team/agents/orchestrator.md

# Check frontmatter
head -10 teams/my-team/agents/orchestrator.md

# Check name field
grep "^name:" teams/my-team/agents/orchestrator.md
# Must be exactly: name: orchestrator
```

### Solutions

**If name field is wrong**:

The generated file always has `name: orchestrator` (hardcoded in template). If it's different:
1. Regenerate
2. Don't manually edit name field

**If file location is wrong**:

```bash
# File must be here:
teams/{team-name}/agents/orchestrator.md

# Verify directory exists
mkdir -p teams/my-team/agents

# Regenerate
/roster/templates/orchestrator-generate.sh my-team
```

## Quick Fixes (TRY THESE FIRST)

Before detailed troubleshooting, try:

```bash
# 1. Delete and regenerate
rm teams/my-team/agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team

# 2. Validate output
/roster/templates/validate-orchestrator.sh \
  teams/my-team/agents/orchestrator.md

# 3. Check YAML syntax
yq . teams/my-team/orchestrator.yaml > /dev/null

# 4. Test team activation
./swap-team.sh my-team
grep "^role:" .claude/agents/orchestrator.md
./swap-team.sh previous-team  # Switch back
```

**Most issues resolve with these four steps.**

## Escalation

If none of these solutions work:

1. **Gather information**:
   - Full error message
   - Output of `git status` and `git log -1`
   - Contents of orchestrator.yaml
   - Output of generator with `-x` flag
   - Output of validator

2. **Document**:
   - What you tried
   - What output you got
   - What you expected

3. **Report**:
   - Post issue with information above
   - Reference this troubleshooting guide sections you already tried

---

**Status**: Production guide
**Last Updated**: 2025-12-29
**Covers**: 8 major issue categories, 25+ specific scenarios

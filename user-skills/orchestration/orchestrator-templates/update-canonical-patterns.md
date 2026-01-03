# Step-by-Step: Update Canonical Patterns

> Follow this guide when you discover new patterns all orchestrators should follow.

## When to Update

Update the canonical template when:

**Safe updates** (Low risk):
- Adding optional new section
- Improving wording in canonical sections
- Fixing typos or formatting

**Planned updates** (Medium risk):
- Changing required section structure
- Adding new required field
- Modifying Consultation Protocol schema

**Coordinated updates** (High risk):
- Removing sections
- Breaking changes to frontmatter
- Changes to integration points (swap-team.sh, workflow.yaml)

## Timeline

**Safe update**: 10-15 minutes (update, regenerate, commit)
**Planned update**: 30-45 minutes (update, regenerate, review diffs, test)
**Coordinated update**: 1-2 hours (requires team communication)

## Phase 1: Plan Your Change (5 minutes)

### Step 1.1: Identify the Pattern

Document what all orchestrators should do:

**Example 1: Add optional section**
```
Pattern: All orchestrators should have example usage section
Scope: Adding new "Examples" section after Anti-Patterns
Risk: LOW (optional section, doesn't affect existing content)
Impact: All 11 teams get new section with customizable examples
```

**Example 2: Improve wording**
```
Pattern: Core Purpose section wording is unclear
Scope: Rewrite first 3 sentences of Core Purpose
Risk: LOW (wording only, no structural change)
Impact: All 11 teams get improved clarity
```

**Example 3: Add required field**
```
Pattern: All orchestrators need "team_color" in frontmatter
Scope: Add color field to frontmatter section, add to schema
Risk: MEDIUM (new required field, all teams must have it)
Impact: All 11 teams need updated orchestrator.yaml
```

### Step 1.2: Assess Risk Level

Ask yourself:

1. **Does this change existing sections?** No = LOW risk. Yes = MEDIUM risk.
2. **Does this change the Consultation Protocol?** No = LOW risk. Yes = HIGH risk.
3. **Does this affect swap-team.sh parsing?** No = LOW risk. Yes = HIGH risk.
4. **Must existing teams update YAML?** No = LOW risk. Yes = MEDIUM risk.
5. **Can this be optional?** Yes = LOW risk. No = MEDIUM/HIGH risk.

**Risk score**:
- 0 YES answers = LOW risk → proceed
- 1-2 YES answers = MEDIUM risk → proceed with testing
- 3+ YES answers = HIGH risk → coordinate with team first

### Step 1.3: Get Alignment (For MEDIUM/HIGH Risk)

**Medium risk**: Post in team channel, 2-hour feedback window
**High risk**: Schedule discussion with team leads

**Questions to answer**:
- Does this benefit all orchestrators equally?
- Can teams opt out if needed?
- What's the rollback plan?
- When would we make this change? (avoid busy periods)

## Phase 2: Update the Template (5 minutes)

### Step 2.1: Locate the Template

**File**: `/roster/templates/orchestrator-base.md.tpl`

Open and familiarize yourself with structure:
- Lines 1-10: Frontmatter
- Lines 11-30: Core Purpose and Responsibilities
- Lines 31-80: Consultation Protocol (CANONICAL - FROZEN)
- Lines 81-120: Domain Authority and Routing
- Lines 121-150: Other sections and skills

### Step 2.2: Make Your Change

**For optional sections** (add after existing section):

```markdown
## [NEW SECTION]

[Content here with placeholders as needed]
```

**For wording improvements**:

Edit existing section text. Keep structure intact.

**For new required fields**:

1. Add to Consultation Protocol input/output
2. Update placeholders used in template
3. Update orchestrator.yaml schema

### Step 2.3: Validate Syntax

After editing, verify:

```bash
# Check template is valid Markdown
head -20 /roster/templates/orchestrator-base.md.tpl
tail -20 /roster/templates/orchestrator-base.md.tpl

# Count section headers
grep "^## " /roster/templates/orchestrator-base.md.tpl | wc -l
# Should be 9-10 sections

# Verify no unclosed code blocks
grep "^" /roster/templates/orchestrator-base.md.tpl | \
  grep -c '```' | xargs -I {} expr {} % 2
# Should output 0 (even count of backticks)
```

## Phase 3: Regenerate All Teams (5 minutes)

### Step 3.1: Backup Current State

Before regenerating, create a backup:

```bash
cd $ROSTER_HOME

# Backup all orchestrator.md files
mkdir -p /tmp/orchestrator-backup
find .claude/teams -name "orchestrator.md" -exec cp {} /tmp/orchestrator-backup/ \;

# Verify backup
ls -la /tmp/orchestrator-backup/ | head -20
# Should show orchestrator.md files from multiple teams
```

### Step 3.2: Regenerate All Teams

Use batch regeneration command:

```bash
# Regenerate all orchestrators
for team in .claude/teams/*/; do
  team_name=$(basename "$team")
  echo "Regenerating $team_name..."
  /roster/templates/orchestrator-generate.sh "$team_name"
done

# Verify all succeeded
echo "Checking for generation errors..."
find .claude/teams -name "orchestrator.md" -exec grep -l "{{" {} \;
# Should output nothing (no unreplaced placeholders)
```

### Step 3.3: Validate All Teams

```bash
# Run validator on each team's orchestrator.md
validation_failed=0
for team_md in .claude/teams/*/agents/orchestrator.md; do
  echo "Validating $team_md..."
  if ! /roster/templates/validate-orchestrator.sh "$team_md"; then
    validation_failed=$((validation_failed + 1))
    echo "FAILED: $team_md"
  fi
done

if [ $validation_failed -eq 0 ]; then
  echo "✓ All validators passed"
else
  echo "✗ $validation_failed validator(s) failed"
  exit 1
fi
```

## Phase 4: Review Diffs (10-20 minutes)

### Step 4.1: Generate Diff Report

```bash
# Show summary of changes
echo "=== Change Summary ==="
git diff --stat .claude/teams/*/agents/orchestrator.md | tail -5

# Show line count changes per team
echo ""
echo "=== Changes per Team ==="
git diff .claude/teams/*/agents/orchestrator.md | \
  grep "^diff --git" | sed 's|.*/\([^/]*\)/agents.*|\1|' | while read team; do
  lines=$(git diff .claude/teams/$team/agents/orchestrator.md | wc -l)
  echo "$team: $lines lines changed"
done
```

### Step 4.2: Review Diff by Team

**For low-risk updates**, spot-check 2-3 teams:

```bash
# Review one team's changes
git diff .claude/teams/rnd-pack/agents/orchestrator.md

# Review another team
git diff .claude/teams/security-pack/agents/orchestrator.md
```

**For medium-risk updates**, review all teams:

```bash
# Show all diffs side-by-side
git diff .claude/teams/*/agents/orchestrator.md
```

**What to look for**:
- Changes align with your template update
- No accidental deletions
- Specialists and skills still present
- Handoff criteria still correct

### Step 4.3: Spot-Check Frontmatter

```bash
# Verify frontmatter unchanged on several teams
for team in rnd-pack security-pack ecosystem-pack; do
  echo "=== $team ==="
  head -10 .claude/teams/$team/agents/orchestrator.md
done
```

Should show:
```yaml
---
name: orchestrator
role: "..."
[team-specific role]
...
---
```

### Step 4.4: Check Specialist Consistency

```bash
# Verify specialist names are consistent across all diffs
echo "Specialists in rnd-pack after update:"
grep "routing_specialist=" .claude/teams/rnd-pack/agents/orchestrator.md | head -3

echo ""
echo "Specialists in security-pack after update:"
grep "routing_specialist=" .claude/teams/security-pack/agents/orchestrator.md | head -3
```

All specialist names should match original configuration.

## Phase 5: Test Specific Cases

### Step 5.1: Test Team Activation

Activate a team and verify it still works:

```bash
# Save current team
current_team=$(cat .claude/ACTIVE_TEAM)

# Test activation with updated orchestrator
./swap-team.sh rnd-pack

# Verify frontmatter parsed
grep "^role:" .claude/agents/orchestrator.md

# Restore original team
./swap-team.sh "$current_team"
```

### Step 5.2: Verify Consultation Protocol (For Protocol Changes)

If you changed Consultation Protocol section:

```bash
# Check CONSULTATION_REQUEST still present
grep "CONSULTATION_REQUEST" .claude/teams/rnd-pack/agents/orchestrator.md

# Check CONSULTATION_RESPONSE still present
grep "CONSULTATION_RESPONSE" .claude/teams/rnd-pack/agents/orchestrator.md

# Check input/output structure preserved
grep "type:" .claude/teams/rnd-pack/agents/orchestrator.md
grep "directive:" .claude/teams/rnd-pack/agents/orchestrator.md
```

### Step 5.3: Verify Skills References

If you changed how skills are referenced:

```bash
# Check skill references format
grep "^- @" .claude/teams/*/agents/orchestrator.md | head -20

# Should all be: "- @skill-name description"
```

## Phase 6: Commit Changes (5 minutes)

### Step 6.1: Stage Files

```bash
cd $ROSTER_HOME

# Stage only orchestrator.md files (not YAML, which shouldn't change)
git add .claude/teams/*/agents/orchestrator.md

# Verify what's staged
git status
```

You should see all orchestrator.md files staged, with no orchestrator.yaml changes (unless you intentionally updated schema).

### Step 6.2: Create Commit Message

```bash
git commit -m "refactor: regenerate all orchestrators with updated template

- [Description of change]
- Regenerated all 11 orchestrators from updated orchestrator-base.md.tpl
- All validators pass (10 rules per team)
- Team activation tested with swap-team.sh
- No breaking changes to specialist routing or handoff criteria"
```

**Example commit message**:
```
refactor: regenerate all orchestrators with improved Consultation Role wording

- Clarified Core Purpose and Responsibilities sections
- Improved explanation of stateless advisor pattern
- Regenerated all 11 orchestrators from updated template
- All validators pass (10 rules per team × 11 teams = 110/110)
- Team activation tested with swap-team.sh
- No changes to Consultation Protocol or routing behavior
```

### Step 6.3: Verify Commit

```bash
git log -1 --stat | head -30

# Should show:
# - orchestrator.md file count matches team count (11)
# - orchestrator.yaml files NOT changed (unless intentional)
# - Commit message explains the change
```

## Phase 7: Communicate Changes (5 minutes, for medium/high risk)

### Step 7.1: Document in Release Notes

Create entry in team updates or changelog:

```markdown
## Orchestrator Template Update

**Date**: 2025-12-29
**Change**: [Your change description]
**Impact**: Affects all orchestrators

### What Changed
[Explain clearly what teams will see]

### What Stayed the Same
[Reassure about what didn't break]

### Action Required
- [ ] No action required (safe change)
- [ ] Teams may need to update orchestrator.yaml (for schema changes)
```

### Step 7.2: Post to Team Channel

For high-risk changes, post notification:

> All orchestrator.md files regenerated with [change description].
>
> Change is backward compatible. No action required.
>
> If you notice any issues: [how to report]

### Step 7.3: Update Related Docs

If template change affects documentation:
- Update schema-reference.md (if schema changed)
- Update create-new-team-orchestrator.md (if process changed)
- Update troubleshooting.md (if new common issues)

## Example Workflows

### Example 1: Safe Change - Improve Wording

**Scenario**: Core Purpose section is confusing. Rewrite it.

```
Timeline: 10 minutes
Risk: LOW
Steps:
1. Edit template Core Purpose (2 min)
2. Regenerate all teams (3 min)
3. Spot-check 2-3 diffs (3 min)
4. Commit (2 min)
```

**Command sequence**:
```bash
# Edit template
nano /roster/templates/orchestrator-base.md.tpl

# Regenerate
for team in .claude/teams/*/; do
  /roster/templates/orchestrator-generate.sh $(basename "$team")
done

# Review sample
git diff .claude/teams/rnd-pack/agents/orchestrator.md | head -50

# Commit
git add .claude/teams/*/agents/orchestrator.md
git commit -m "refactor: improve Core Purpose wording in all orchestrators"
```

### Example 2: Medium Risk - Add Optional Section

**Scenario**: Add new "Examples" section to all orchestrators.

```
Timeline: 20 minutes
Risk: MEDIUM
Steps:
1. Add section to template (3 min)
2. Regenerate all (3 min)
3. Review all diffs (8 min)
4. Test 2-3 team activations (4 min)
5. Commit (2 min)
```

**Command sequence**:
```bash
# Backup
mkdir -p /tmp/orch-backup
cp -r .claude/teams/*/agents/orchestrator.md /tmp/orch-backup/

# Edit template - add new section with placeholder
nano /roster/templates/orchestrator-base.md.tpl

# Regenerate all
for team in .claude/teams/*/; do
  /roster/templates/orchestrator-generate.sh $(basename "$team")
done

# Validate all
for md in .claude/teams/*/agents/orchestrator.md; do
  /roster/templates/validate-orchestrator.sh "$md" || exit 1
done

# Spot-check diffs
git diff --stat .claude/teams/*/agents/orchestrator.md
git diff .claude/teams/rnd-pack/agents/orchestrator.md | head -80

# Test activation
./swap-team.sh security-pack
grep "Examples" .claude/agents/orchestrator.md  # Should show new section
./swap-team.sh rnd-pack  # Restore

# Commit
git add .claude/teams/*/agents/orchestrator.md
git commit -m "feat: add Examples section to all orchestrators

- New optional section after Anti-Patterns
- Teams can customize with team-specific examples
- Regenerated all 11 orchestrators
- Validators pass: 110/110 rules
- Backward compatible, no YAML changes required"
```

### Example 3: High Risk - Schema Change

**Scenario**: Add required field to Consultation Protocol.

```
Timeline: 1+ hour
Risk: HIGH
Steps:
1. Coordinate with team (15 min)
2. Update schema and template (10 min)
3. Regenerate (3 min)
4. Update all YAML configs (20 min)
5. Review diffs extensively (15 min)
6. Test thoroughly (10 min)
7. Commit with documentation (5 min)
```

**This requires**:
- Team coordination
- Schema updates
- YAML updates for all teams
- Extended testing
- Clear communication

**Don't do this alone.** Coordinate with team first.

## Troubleshooting

### Issue: Some teams show "No changes" after template update

**Cause**: Those sections weren't generated (template placeholders)

**Example**: If you only changed canonical protocol, custom sections won't change.

**Solution**: This is normal. Validate that expected sections did change:

```bash
# Check rnd-pack for your change
git diff .claude/teams/rnd-pack/agents/orchestrator.md | grep "^+" | head -20
```

### Issue: Validation fails after regeneration

**Cause**: Template syntax error introduced

**Solution**:
```bash
# Review template changes
git diff /roster/templates/orchestrator-base.md.tpl

# Check for common errors:
# - Unclosed code blocks
# - Unmatched brackets
# - Missing colons in YAML

# Revert and try again
git checkout /roster/templates/orchestrator-base.md.tpl
```

### Issue: Specialist names changed in output

**Cause**: Template or workflow.yaml changed

**Solution**:
```bash
# Verify workflow.yaml unchanged
git diff .claude/teams/*/workflow.yaml

# Verify orchestrator.yaml unchanged
git diff .claude/teams/*/orchestrator.yaml

# Specialist names should come from orchestrator.yaml, not template
```

## Rollback Plan

If something goes wrong:

```bash
# Restore from backup
cp /tmp/orchestrator-backup/* .claude/teams/*/agents/orchestrator.md

# Revert template
git checkout /roster/templates/orchestrator-base.md.tpl

# Verify
for md in .claude/teams/*/agents/orchestrator.md; do
  /roster/templates/validate-orchestrator.sh "$md" || exit 1
done
```

Or restore git state:

```bash
git reset --hard HEAD~1  # Undo last commit
```

---

**Status**: Production guide
**Last Updated**: 2025-12-29
**Tested**: Phase 3 (all 11 teams regenerated without issues)

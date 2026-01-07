# Orchestrator Templating: Quick Start Guide

**For**: Team Leads and Engineers
**Time**: 5 minutes to understand, 15 minutes to adopt
**Goal**: Help your team migrate to orchestrator templating (optional)

## What's Happening?

Orchestrators can now be generated automatically from simple YAML configuration files. This means:

- **Before**: Manually edit orchestrator.md (2 hours to update)
- **After**: Edit orchestrator.yaml (2 minutes), generator creates orchestrator.md (3 minutes)
- **Benefit**: 24x faster updates, fewer manual mistakes, consistent patterns

## Is This Required?

**No.** Adoption is completely optional. Your team can:
- Keep your hand-written orchestrator.md as-is (no changes needed)
- Adopt templating when you're ready
- Switch back to manual if needed

## Quick Start: Adopt Templating (15 Minutes)

### Step 1: Check Your Team's YAML (1 min)

Look for your team's orchestrator.yaml:

```bash
ls /roster/rites/<YOUR-TEAM>/orchestrator.yaml
```

If it exists, you're already set up (skip to Step 3).

### Step 2: Migrate Existing Orchestrator (5 min)

If you only have orchestrator.md (no YAML yet), migrate it:

```bash
cd /roster

# Migrate your team
./scripts/orchestrator-migrate.sh <YOUR-TEAM>

# Review the diff
git diff rites/<YOUR-TEAM>/orchestrator.yaml

# Verify it looks correct
cat rites/<YOUR-TEAM>/orchestrator.yaml
```

The script extracts your orchestrator.yaml from the existing orchestrator.md and validates it produces the same output.

### Step 3: Install Pre-Commit Hook (2 min)

Optional but recommended: Install local validation:

```bash
# Copy hook to your repository
cp /roster/.githooks/pre-commit-orchestrator /roster/.git/hooks/pre-commit
chmod +x /roster/.git/hooks/pre-commit
```

Now when you commit, the hook automatically:
- Validates your YAML syntax
- Regenerates orchestrator.md
- Stages the changes

### Step 4: Make a Change (5 min)

Try updating your orchestrator. Example: Add a new specialist to routing.

**Old Way** (Manual):
1. Edit orchestrator.md
2. Find routing table section
3. Add new row
4. Update workflow diagram
5. Update any related sections
6. Test and commit

**New Way** (Templated):
1. Edit orchestrator.yaml:
   ```yaml
   routing:
     existing-specialist: "when to use"
     new-specialist: "when to use this specialist"
   ```

2. Regenerate:
   ```bash
   ./templates/orchestrator-generate.sh <YOUR-TEAM>
   ```

3. Review changes:
   ```bash
   git diff rites/<YOUR-TEAM>/agents/orchestrator.md
   ```

4. Commit:
   ```bash
   git add rites/<YOUR-TEAM>/
   git commit -m "Update routing for new specialist"
   ```

Done! The orchestrator.md was automatically updated.

## Understanding orchestrator.yaml

Your orchestrator.yaml has these main sections:

### 1. Team Metadata
```yaml
team:
  name: your-team-name           # Must match directory name
  domain: "your domain"           # What this team does
  color: purple                   # Team branding color
```

### 2. Frontmatter (appears at top of orchestrator.md)
```yaml
frontmatter:
  role: "One-line role summary"
  description: "Multi-line description of what you do"
```

### 3. Routing (when to invoke each specialist)
```yaml
routing:
  specialist-1: "When to invoke this specialist"
  specialist-2: "When to invoke this one"
  # Usually 3-5 specialists
```

### 4. Workflow Position
```yaml
workflow_position:
  upstream: "What feeds into your orchestrator"
  downstream: "Where your output goes"
```

### 5. Handoff Criteria (checklist for phase completion)
```yaml
handoff_criteria:
  phase-1:
    - "Concrete, verifiable criterion"
    - "Another criterion"
  phase-2:
    - "Phase 2 criteria"
```

### 6. Skills Reference
```yaml
skills:
  - "@skill-name for description"
  - "@another-skill for what it's used for"
```

### Optional Sections
```yaml
antipatterns:            # Team-specific anti-patterns (optional)
  - "Anti-pattern to avoid"

cross_team_protocol: ""  # If you coordinate with other teams (optional)
```

## Common Tasks

### Task 1: Update Routing Table

**Before** (Manual): Edit orchestrator.md, find routing table, add row, update diagram, commit.

**After** (Templated):
```bash
# 1. Edit orchestrator.yaml
nano rites/<YOUR-TEAM>/orchestrator.yaml

# 2. Add new specialist to routing:
# routing:
#   new-specialist: "When to use it"

# 3. Regenerate
./templates/orchestrator-generate.sh <YOUR-TEAM>

# 4. Review and commit
git diff
git add rites/<YOUR-TEAM>/
git commit -m "Add new specialist to routing"
```

### Task 2: Update Handoff Criteria

**Before** (Manual): Edit orchestrator.md, find handoff section, add checklist items.

**After** (Templated):
```bash
# 1. Edit orchestrator.yaml
# handoff_criteria:
#   phase-name:
#     - "New checklist item"

# 2. Regenerate
./templates/orchestrator-generate.sh <YOUR-TEAM>

# 3. Commit
git add rites/<YOUR-TEAM>/
git commit -m "Update handoff criteria for phase-name"
```

### Task 3: Stay Manual (Don't Adopt)

If you prefer to keep hand-writing your orchestrator:

```bash
# Delete the YAML file
cd /roster/rites/<YOUR-TEAM>
rm orchestrator.yaml

# Update your manifest to say "user" (not generated)
jq '.orchestrator.source = "user"' AGENT_MANIFEST.json > temp && mv temp AGENT_MANIFEST.json

# Commit
git commit -m "Revert to manual orchestrator"

# Your orchestrator.md stays as-is, no other changes needed
```

CI will skip validation for your team.

## Troubleshooting

### Q: My YAML validation fails

**Error**: `Schema validation failed: missing required field 'routing'`

**Fix**: Make sure orchestrator.yaml has all required sections. Compare to an example:

```bash
cat /roster/schemas/orchestrator.yaml.schema.json | jq '.examples[0]'
```

### Q: Generation produces different output than my current orchestrator.md

**Expected**: The first time you migrate, output may differ slightly (whitespace, formatting).

**Fix**: Review the diff and commit the generated version:

```bash
git diff rites/<YOUR-TEAM>/agents/orchestrator.md
git add rites/<YOUR-TEAM>/agents/orchestrator.md
git commit -m "Regenerate orchestrator with templates"
```

The generated version is valid and maintains the same content.

### Q: Pre-commit hook blocks my commit

**Error**: `orchestrator.md changed without orchestrator.yaml`

**Fix**: Either:
- **Option A** (Recommended): Update orchestrator.yaml, regenerate, re-commit
- **Option B** (Workaround): Skip hook with `git commit --no-verify`

### Q: I want to go back to manual

**Steps**:
```bash
# Delete orchestrator.yaml
rm rites/<YOUR-TEAM>/orchestrator.yaml

# Update manifest
jq '.orchestrator.source = "user"' AGENT_MANIFEST.json > temp && mv temp AGENT_MANIFEST.json

# Commit
git commit -m "Revert to manual orchestrator"

# Keep orchestrator.md as-is, no other changes needed
```

### Q: How do I add custom anti-patterns?

**Steps**:
```bash
# Edit orchestrator.yaml
nano orchestrator.yaml

# Add to antipatterns section:
# antipatterns:
#   - "Treating PATCH as SYSTEM (different scope)"
#   - "Your custom anti-pattern"

# Regenerate
./templates/orchestrator-generate.sh <YOUR-TEAM>
```

## Getting Help

**For Adoption Questions**:
- Skill: `@orchestrator-templates` (examples, FAQ)
- Email: [contact]
- Pairing: Schedule 1-hour session

**For Technical Issues**:
- Issue: File GitHub issue with error output
- Slack: Post in #orchestrator-templates

**For Feedback**:
- Survey: [link] (takes 5 minutes)
- Email: [contact]

## Next Steps

1. **Read**: Check out `@orchestrator-templates` skill for full details
2. **Decide**: Does your team want to adopt? (No pressure either way)
3. **Adopt** (Optional): Run migration or installation steps above
4. **Feedback**: Tell us how it goes!

---

**Timeline**:
- **Now**: Adoption is optional
- **Month 1**: Early teams adopt, we gather feedback
- **Month 3**: We'll check in with all teams
- **Month 6**: Most teams will have adopted (or made explicit choice)

**Questions?** Reach out to [contact] or post in [channel].

**Learn More**: `/roster/docs/@orchestrator-templates`

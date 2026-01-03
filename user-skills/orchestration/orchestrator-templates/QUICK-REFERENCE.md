# Quick Reference Card

## Creating a New Orchestrator (15 min)

```bash
# Step 1: Create config
cat > .claude/teams/my-team/orchestrator.yaml << 'EOF'
team:
  name: my-team
  domain: "What your team does"
  color: purple

frontmatter:
  role: "One-line role"
  description: |
    Multi-line description explaining
    what the team does and when to use it.

routing:
  specialist-one: "When [condition]"
  specialist-two: "When [condition]"
  specialist-three: "When [condition]"

workflow_position:
  upstream: "None or team name"
  downstream: "None or team name"

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

skills:
  - "@skill-one brief description"
  - "@skill-two brief description"
EOF

# Step 2: Generate orchestrator.md
/roster/templates/orchestrator-generate.sh my-team

# Step 3: Validate
/roster/templates/validate-orchestrator.sh \
  .claude/teams/my-team/agents/orchestrator.md

# Step 4: Commit
git add .claude/teams/my-team/orchestrator.yaml
git add .claude/teams/my-team/agents/orchestrator.md
git commit -m "feat: add my-team orchestrator"

# Step 5: Test
./swap-team.sh my-team
grep "^role:" .claude/agents/orchestrator.md
./swap-team.sh previous-team
```

## Updating All Orchestrators (20 min)

```bash
# Step 1: Edit template
nano /roster/templates/orchestrator-base.md.tpl

# Step 2: Regenerate all
for team in .claude/teams/*/; do
  /roster/templates/orchestrator-generate.sh $(basename "$team")
done

# Step 3: Validate all
validation_failed=0
for md in .claude/teams/*/agents/orchestrator.md; do
  /roster/templates/validate-orchestrator.sh "$md" || \
    validation_failed=$((validation_failed + 1))
done
[ $validation_failed -eq 0 ] && echo "All passed" || echo "$validation_failed failed"

# Step 4: Review diffs
git diff .claude/teams/*/agents/orchestrator.md | head -100

# Step 5: Commit
git add .claude/teams/*/agents/orchestrator.md
git commit -m "refactor: regenerate orchestrators with updated template"
```

## Troubleshooting (5-15 min)

### Generator Error: "Specialist not found"

```bash
# Check exact names in workflow.yaml
yq '.phases[].agent' .claude/teams/my-team/workflow.yaml

# Update orchestrator.yaml to match exactly
nano .claude/teams/my-team/orchestrator.yaml

# Regenerate
/roster/templates/orchestrator-generate.sh my-team
```

### Validation Error: "No placeholders replaced"

```bash
# Delete bad file, regenerate
rm .claude/teams/my-team/agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team

# If still fails, check YAML
yq . .claude/teams/my-team/orchestrator.yaml > /dev/null
```

### swap-team.sh Can't Parse

```bash
# Check frontmatter
head -10 .claude/teams/my-team/agents/orchestrator.md

# Regenerate if malformed
rm .claude/teams/my-team/agents/orchestrator.md
/roster/templates/orchestrator-generate.sh my-team

# Validate
/roster/templates/validate-orchestrator.sh \
  .claude/teams/my-team/agents/orchestrator.md
```

## YAML Quick Reference

```yaml
team:
  name: lowercase-hyphens              # Required
  domain: "Team description"           # Required
  color: purple or #RRGGBB            # Required

frontmatter:
  role: "One-line role"               # Required (40-80 chars)
  description: "Multi-line description" # Required

routing:                              # Required (3-6 specialists)
  specialist-name: "Routing condition"

workflow_position:                    # Required
  upstream: "Upstream team or None"
  downstream: "Downstream team or None"

handoff_criteria:                     # Required
  specialist-name:
    - "Criterion 1"
    - "Criterion 2"
    - "Artifacts verified via Read tool"

skills:                               # Required (1-10 skills)
  - "@skill-name brief description"

# Optional:
antipatterns:
  - "Anti-pattern description"

cross_team_protocol: "Hub protocol description"

extension_points:
  examples: |
    Custom content here
```

## File Locations

| Purpose | Location |
|---------|----------|
| Skill docs | `user-skills/orchestration/orchestrator-templates/` |
| Generator | `templates/orchestrator-generate.sh` |
| Validator | `templates/validate-orchestrator.sh` |
| Template | `templates/orchestrator-base.md.tpl` |
| Schema | `schemas/orchestrator.yaml.schema.json` |
| Team config | `teams/{team-name}/orchestrator.yaml` |
| Generated | `teams/{team-name}/agents/orchestrator.md` |

## Validation Rules (10)

```
✓ File exists and readable
✓ No unreplaced {{}} placeholders
✓ Valid YAML frontmatter
✓ All required ## sections present
✓ Specialist names consistent
✓ No duplicate sections
✓ Handoff criteria in - [ ] format
✓ CONSULTATION_REQUEST/RESPONSE present
✓ Skill references @skill-name format
✓ Markdown syntax valid
```

## Common Commands

```bash
# Validate single file
/roster/templates/validate-orchestrator.sh path/to/orchestrator.md

# Validate YAML
yq . path/to/orchestrator.yaml

# Generate with force
/roster/templates/orchestrator-generate.sh my-team --force

# Check specialist names match
diff <(yq '.phases[].agent' workflow.yaml | sort) \
     <(yq '.routing | keys[]' orchestrator.yaml | sort)

# Show routing table
grep "^|" .claude/teams/my-team/agents/orchestrator.md | grep -v "^|---"

# Test team activation
./swap-team.sh my-team && echo "OK" || echo "FAILED"
```

## Success Checklist

Before considering orchestrator complete:

- [ ] orchestrator.yaml created with all required fields
- [ ] All specialists in routing exist in workflow.yaml
- [ ] Specialist count is 3-6
- [ ] Handoff criteria defined for each specialist
- [ ] Generator runs without errors
- [ ] Validator passes (exit code 0)
- [ ] All specialist names match workflow.yaml exactly
- [ ] Skills use @skill-name format
- [ ] Both files committed to git
- [ ] Team activation works (swap-team.sh)

## FAQ

**Can I have a team without an orchestrator?**
- Orchestrators required for multi-phase teams. Single-phase teams don't need one.

**What if I edit orchestrator.md directly?**
- Changes lost on next regeneration. Always edit orchestrator.yaml instead.

**Can I customize specific sections?**
- Not by design. Use `extension_points` in YAML for custom content.

**Is orchestrator.md checked into git?**
- Yes. Commit both orchestrator.yaml AND orchestrator.md together.

**How do I add a new specialist?**
- Update orchestrator.yaml routing, re-run generator, validate, commit.

**How often should we update the template?**
- When you discover a pattern all orchestrators should follow. Typically quarterly.

## Where to Get Help

| Question | Reference |
|----------|-----------|
| "What is this?" | [SKILL.md](SKILL.md) |
| "How do I create?" | [create-new-team-orchestrator.md](create-new-team-orchestrator.md) |
| "How do I update template?" | [update-canonical-patterns.md](update-canonical-patterns.md) |
| "I have an error" | [troubleshooting.md](troubleshooting.md) |
| "What are the fields?" | [schema-reference.md](schema-reference.md) |
| "How does it work?" | [architecture-overview.md](architecture-overview.md) |
| "Should I migrate?" | [migration-guide.md](migration-guide.md) |

---

**Quick Reference Card** - Print, tape to monitor, save locally.

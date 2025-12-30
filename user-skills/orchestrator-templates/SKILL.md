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
Your team discovered a pattern all orchestrators should follow. Update `orchestrator-base.md.tpl` and regenerate all teams:
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
- How templating integrates with CEM, swap-team.sh, and workflow.yaml

**Time**: 10-15 minutes for solid understanding.

## Core Concept: Durable Abstractions

### The Problem

Hand-written orchestrators encode both **unchanging protocol** and **team-specific details**:

```markdown
# Orchestrator

## Consultation Protocol
[50 lines of canonical protocol - IDENTICAL across all orchestrators]

## Routing Decisions
[20 lines of team-specific routing - DIFFERENT per team]
```

When consultation protocol evolves, you must update 10+ files manually. Risk of inconsistency is high.

### The Solution

Separate **canonical protocol** (template) from **team specifics** (YAML data):

```yaml
# orchestrator.yaml
team:
  name: rnd-pack
routing:
  integration-researcher: "Needs research on technologies"
  technology-scout: "Evaluates emerging tools"
  prototype-engineer: "Builds working demo"
  moonshot-architect: "Designs long-term architecture"
```

The template contains the protocol once. YAML contains the variation. Generator merges them.

### Why This Matters

1. **One place to update**: Change consultation protocol once in template
2. **Consistent contracts**: All orchestrators have identical protocol section
3. **No copy-paste**: Reduces maintenance burden by 80%
4. **Team specialization**: Each team focuses on their routing logic, not protocol details
5. **Evolution-safe**: Adding new section works for all teams simultaneously

## Workflow

### Phase 1: Design Your Team's Configuration

Create `/Users/tomtenuta/Code/skeleton_claude/.claude/teams/my-team/orchestrator.yaml`:

```yaml
team:
  name: my-team
  domain: "Describe what your team does"
  color: purple

frontmatter:
  role: "One-line role summary"
  description: "Multi-line description of consultation role"

routing:
  specialist-one: "When to route to specialist one"
  specialist-two: "When to route to specialist two"
  specialist-three: "When to route to specialist three"

workflow_position:
  upstream: "Which team comes before yours"
  downstream: "Which team comes after yours"

handoff_criteria:
  specialist-one:
    - "Criterion 1"
    - "Criterion 2"
  specialist-two:
    - "Criterion 1"

skills:
  - "@skill-one brief description"
  - "@skill-two brief description"
```

### Phase 2: Run the Generator

```bash
# Generate orchestrator.md from your YAML config
/Users/tomtenuta/Code/roster/templates/orchestrator-generate.sh my-team

# Validate the generated file
/Users/tomtenuta/Code/roster/templates/validate-orchestrator.sh \
  /Users/tomtenuta/Code/skeleton_claude/.claude/teams/my-team/agents/orchestrator.md
```

### Phase 3: Commit Both Files

```bash
git add teams/my-team/orchestrator.yaml
git add teams/my-team/agents/orchestrator.md
git commit -m "feat: add my-team orchestrator configuration and generated agent"
```

### Phase 4: Swap Team to Verify

```bash
# Activate your team
./swap-team.sh my-team

# Verify orchestrator loads correctly
echo "Testing orchestrator role..."
grep "^role:" .claude/agents/orchestrator.md
```

## File Structure

### Input Files

**orchestrator.yaml** (your team's configuration)
```
Location: teams/{team-name}/orchestrator.yaml
Purpose: Team-specific configuration (durable spec)
Tracked: Yes (check into git)
```

**orchestrator-base.md.tpl** (canonical template)
```
Location: /Users/tomtenuta/Code/roster/templates/orchestrator-base.md.tpl
Purpose: Shared protocol and structure
Tracked: Yes (part of roster)
```

**workflow.yaml** (phase and specialist definitions)
```
Location: teams/{team-name}/workflow.yaml
Purpose: Phase sequence and specialist details (read-only reference)
Tracked: Yes
```

### Output Files

**orchestrator.md** (generated agent)
```
Location: teams/{team-name}/agents/orchestrator.md
Purpose: Production-ready orchestrator agent
Tracked: Yes (check into git for consistency)
Generated: By orchestrator-generate.sh
```

## Generator Behavior

### What the Generator Does

1. **Reads** orchestrator.yaml (your configuration)
2. **Reads** workflow.yaml (specialist and phase definitions)
3. **Reads** orchestrator-base.md.tpl (canonical template)
4. **Substitutes** placeholders in template with your data
5. **Validates** output (no unreplaced placeholders)
6. **Writes** orchestrator.md

### What the Generator Doesn't Do

- Edit YAML files
- Remove old orchestrator.md versions
- Automatically commit to git
- Update AGENT_MANIFEST
- Run validation (that's a separate step)

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

### Pattern 1: Linear Pipeline (4-5 Specialists)

Most teams follow a linear sequence:

```yaml
routing:
  analyst: "Diagnoses root cause"
  designer: "Creates solution plan"
  implementer: "Executes implementation"
  reviewer: "Validates completed work"
```

**Example teams**: rnd-pack, security-pack, doc-team-pack

### Pattern 2: Hub Coordination (5+ Specialists)

Some teams have a hub that coordinates multiple paths:

```yaml
routing:
  analyzer: "Identifies gaps and inconsistencies"
  architect: "Designs solution"
  engineer: "Implements across ecosystem"
  reviewer: "Validates solution quality"
  tester: "Tests compatibility"
```

**Example teams**: ecosystem-pack

### Pattern 3: Domain-Specific Complexity

Security teams use different complexity enums than documentation teams:

```yaml
# In orchestrator.yaml, complexity comes from workflow.yaml
# workflow.yaml defines PATCH | FEATURE | SYSTEM for security
# workflow.yaml defines PAGE | SECTION | SITE for documentation
```

Generator automatically pulls correct complexity values from your workflow.yaml.

## Integration Points

### swap-team.sh (Team Activation)

When you run `swap-team.sh my-team`, the script:
1. Reads frontmatter from orchestrator.md
2. Extracts name, role, description, model
3. Symlinks orchestrator.md into active agent position

**Requirement**: Generated frontmatter must parse cleanly via grep/sed. Validation ensures this.

### workflow.yaml (Specialist Definitions)

The generator validates that each specialist in your routing table exists in workflow.yaml:

```yaml
# orchestrator.yaml
routing:
  integration-researcher: "..."  # Must exist in workflow.yaml phases
```

If specialist doesn't exist, generator exits with error.

### CEM Sync (File Distribution)

Generated orchestrator.md is treated like any other agent file:
- Standard Markdown format
- Valid YAML frontmatter
- Passes all CEM validation
- No special handling required

**Optional**: CEM can track `generated: true` in AGENT_MANIFEST for auditing (Phase 5).

### AGENT_MANIFEST (Agent Registry)

The AGENT_MANIFEST automatically tracks orchestrator.md:
- Source: "team"
- Origin: Your team name
- Model: opus

**Note**: No manual edits needed. Standard agent discovery handles it.

## Troubleshooting

### Issue: Generator fails with "Specialist not found"

**Cause**: orchestrator.yaml references specialist not in workflow.yaml

**Solution**:
1. Open your workflow.yaml
2. Check phases[] for correct agent name
3. Update orchestrator.yaml routing to match
4. Re-run generator

**Example**:
```yaml
# WRONG: workflow.yaml has "technology-scout"
routing:
  tech-scout: "..."

# RIGHT:
routing:
  technology-scout: "..."
```

### Issue: Generated file has unreplaced placeholders

**Cause**: Generator script encountered error during substitution

**Solution**:
1. Check error output from generator
2. Verify orchestrator.yaml is valid YAML
3. Verify all required fields present
4. Try regenerating with verbose flag (if available)

**Prevention**: Run validation immediately after generation.

### Issue: Validation fails with "No placeholders replaced"

**Cause**: Template file not found or not readable

**Solution**:
1. Verify template exists: `/Users/tomtenuta/Code/roster/templates/orchestrator-base.md.tpl`
2. Check file permissions (should be readable)
3. Verify generator has correct path to template
4. Try manual generation test

### Issue: Frontmatter parsing fails in swap-team.sh

**Cause**: Generated frontmatter doesn't match swap-team.sh expectations

**Solution**:
1. Run validation: `validate-orchestrator.sh orchestrator.md`
2. Check frontmatter section manually:
   ```bash
   head -10 orchestrator.md
   ```
3. Ensure lines match format:
   ```
   ---
   name: orchestrator
   role: "..."
   description: "..."
   tools: Read, Skill
   model: opus
   color: purple
   ---
   ```

### Issue: Diff shows 30-40% changes after update

**Cause**: Template evolved; all teams need regeneration

**Is this normal?** Yes, if you updated:
- Canonical sections (Consultation Protocol)
- Required section structure
- Placeholder content

**Solution**:
1. Review diff to understand changes
2. Regenerate all teams: `orchestrator-generate.sh --all`
3. Spot-check 2-3 teams
4. Commit all changes together

**Prevention**: Document template changes in commit message.

### Issue: My team-specific examples disappeared

**Cause**: Examples live in template, not YAML config

**Solution**: Use `extension_points` in orchestrator.yaml to add team-specific content:

```yaml
extension_points:
  examples: |
    ### Team-Specific Example
    [Your custom example here]
```

### Issue: How do I customize anti-patterns for my team?

**Solution**: Add to orchestrator.yaml:

```yaml
antipatterns:
  - "Team-specific anti-pattern 1"
  - "Team-specific anti-pattern 2"
```

These are appended to canonical anti-patterns.

### Issue: Can I rollback to hand-written orchestrator?

**Solution**: Yes, but not recommended:

1. Delete orchestrator.yaml
2. Keep orchestrator.md (or restore from git history)
3. Update AGENT_MANIFEST source from "generated" to "manual"
4. Test with swap-team.sh

**Better approach**: Fix the YAML config and regenerate instead.

## Evolution Path (Phase 5 and Beyond)

### Template Versioning (Future)

When orchestrator pattern evolves, we may add versioning:

```yaml
# orchestrator.yaml
template_version: 1  # Enables migrations to v2, v3, etc.
```

### Migration Strategy (Phase 5)

When migration needed:
```bash
./migrate-orchestrator.yaml v1 v2 teams/my-team
./orchestrator-generate.sh my-team
```

### CI Integration (Phase 5)

Build pipeline will:
1. Validate all orchestrator.yaml files against schema
2. Regenerate all orchestrator.md files
3. Verify no changes (developer didn't commit modified .md directly)
4. Fail build if inconsistencies detected

## Related Skills

- `standards` - Naming conventions and code style
- `agent-prompt-engineering` - Writing effective agent prompts (available with forge-pack)
- Team-specific skills - Each team defines its own documentation templates and workflow protocols

## Quick Reference

| Task | Command |
|------|---------|
| Create new team | `cp -r .claude/teams/doc-team-pack .claude/teams/my-team && edit orchestrator.yaml` |
| Generate orchestrator.md | `/Users/tomtenuta/Code/roster/templates/orchestrator-generate.sh my-team` |
| Validate output | `/Users/tomtenuta/Code/roster/templates/validate-orchestrator.sh .claude/teams/my-team/agents/orchestrator.md` |
| Update template | Edit `/Users/tomtenuta/Code/roster/templates/orchestrator-base.md.tpl` then regenerate all teams |
| View current team | `cat .claude/ACTIVE_TEAM` |
| Switch teams | `./swap-team.sh my-team` |

## FAQ

**Q: Can I have a team without an orchestrator?**
A: Orchestrators are required for multi-phase teams. Single-phase teams don't need one.

**Q: What happens if I edit orchestrator.md directly instead of updating YAML?**
A: Changes will be lost on next regeneration. Always edit orchestrator.yaml instead.

**Q: Can I customize specific sections while keeping others generated?**
A: Not with this system (by design). If you need customization, use extension_points in YAML.

**Q: Is the generated orchestrator.md checked into git?**
A: Yes. Both orchestrator.yaml AND orchestrator.md should be committed together.

**Q: What if I need to add a new specialist halfway through?**
A: Update orchestrator.yaml routing table, re-run generator, validate, and commit.

**Q: How often should we update the template?**
A: Whenever you discover a pattern all orchestrators should follow. Typically once per quarter.

---

**Last Updated**: 2025-12-29
**Status**: Production (Phase 4 complete)
**Maintenance**: Coordinate template changes through orchestrator-templates skill

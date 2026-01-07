# Migration Guide: Adopting Orchestrator Templating

> Roadmap for migrating existing hand-written orchestrators to generated configuration.

## Overview

Currently, orchestrators are hand-written. We're introducing an optional templating system that allows orchestrators to be generated from YAML configuration. This guide helps teams transition at their own pace.

**Key principle**: Adoption is optional and low-risk. Teams can transition incrementally.

## Timeline

**Per team**: 15-30 minutes of work spread over 1-2 sprints

**Organization**: Phase 5+ (after all infrastructure is in place)

## Readiness Criteria

Before migrating a team, ensure:

- [ ] All 11 teams have validated orchestrators (Phase 3 complete)
- [ ] Generation system is production-tested
- [ ] Validation system is proven
- [ ] Team has time to review changes
- [ ] Rollback plan is documented

**Status**: Phases 1-3 complete. Ready for Phase 5 adoption.

## Benefits of Migration

| Benefit | When You'll See It |
|---------|-------------------|
| Easier team creation | Next new team (day 1 of team) |
| Faster pattern updates | Next protocol change (next quarter?) |
| Clearer specialization | Day 1 (YAML captures design decisions) |
| Reduced maintenance burden | Over time (fewer hand-edits) |
| Better consistency | Day 1 (all teams use same protocol) |
| Easier onboarding | For new team members |

**Cost**: None if you don't adopt. Optional for each team.

## Migration Phases

### Phase 1: Readiness (1 week before migration)

**Activity**: Prepare for migration without any changes

**Checklist**:
- [ ] Review this guide
- [ ] Read architecture-overview.md
- [ ] Confirm team understands YAML format
- [ ] Schedule migration (avoid busy periods)

**Time commitment**: 30 minutes reading

### Phase 2: Extraction (Day 1 - 15 minutes)

**Activity**: Extract configuration from existing orchestrator.md

**Steps**:

1. **Read current orchestrator**
   ```bash
   cat rites/my-team/agents/orchestrator.md | head -80
   ```

2. **Extract team metadata**
   - Look for name, role, color, domain
   - Write down in notes

3. **Extract routing table**
   - Find "Routing Decisions" section
   - List specialists and routing conditions
   - Copy exactly as-is

4. **Extract handoff criteria**
   - Find handoff sections (one per specialist)
   - List criteria with exact wording
   - Note any rite-specific criteria

5. **Extract skills**
   - Find "Skills Reference" section
   - List all @skill-name references
   - Keep descriptions

6. **Create orchestrator.yaml**
   ```bash
   # Use create-new-rite-orchestrator guide
   # Or copy from example team and modify
   cp rites/doc-rite-pack/orchestrator.yaml \
      rites/my-team/orchestrator.yaml

   # Edit to match your extracted configuration
   nano rites/my-team/orchestrator.yaml
   ```

**Output**: `rites/my-team/orchestrator.yaml` (ready to generate from)

### Phase 3: Generation (Day 2 - 5 minutes)

**Activity**: Generate orchestrator.md from YAML config

**Steps**:

1. **Run generator**
   ```bash
   /roster/templates/orchestrator-generate.sh my-team
   ```

2. **Run validator**
   ```bash
   /roster/templates/validate-orchestrator.sh \
     rites/my-team/agents/orchestrator.md
   ```

3. **Verify success**
   ```
   VALIDATION PASSED (exit code 0)
   ```

**Output**: `rites/my-team/agents/orchestrator.md` (regenerated)

### Phase 4: Comparison (Day 2-3 - 15 minutes)

**Activity**: Compare generated to original

**Steps**:

1. **Backup original**
   ```bash
   cp rites/my-team/agents/orchestrator.md \
      rites/my-team/agents/orchestrator.md.original
   ```

2. **Review diff**
   ```bash
   git diff rites/my-team/agents/orchestrator.md.original \
            rites/my-team/agents/orchestrator.md
   ```

3. **What to look for**:
   - Are specialist names the same?
   - Are routing conditions preserved?
   - Are handoff criteria intact?
   - Are skills references correct?
   - Is formatting similar?

4. **Accept differences** that are cosmetic:
   - Whitespace changes (OK)
   - Section reordering (OK if content is same)
   - Wording improvements (OK if meaning preserved)
   - Unicode characters vs ASCII (OK)

5. **Flag substantive differences**:
   - Missing specialist
   - Wrong routing condition
   - Incomplete handoff criteria
   - Missing skills

**Decision**:
- **Generated looks correct**: Move to Phase 5 (Commit)
- **Generated has issues**: Return to Phase 2 (re-extract YAML)

### Phase 5: Decision Point

**Choose one path**:

#### Path A: Adopt Generated Version (Recommended)

**Benefit**: Team is now templated. Future updates are easier.

**Steps**:
1. Delete original backup: `rm orchestrator.md.original`
2. Proceed to Phase 6 (Commit)

#### Path B: Keep Hand-Written Version (Conservative)

**Benefit**: Zero risk. Team stays in control of every word.

**Steps**:
1. Restore original: `cp orchestrator.md.original orchestrator.md`
2. Delete YAML: `rm orchestrator.yaml`
3. Skip Phase 6 (no changes to commit)

**Note**: You can migrate later. Both approaches are supported.

### Phase 6: Commit (Day 3 - 5 minutes)

**Activity**: Commit both config and generated files

**Only if you chose Path A above.**

```bash
cd $KNOSSOS_HOME

# Stage both files
git add .claude/rites/my-team/orchestrator.yaml
git add .claude/rites/my-team/agents/orchestrator.md

# Create commit
git commit -m "refactor: migrate my-team to templated orchestrator

- Extract team routing and handoff criteria to orchestrator.yaml
- Generate orchestrator.md from canonical template
- Functionality and behavior unchanged
- Future template updates will apply automatically"

# Verify
git log -1 --stat
```

### Phase 7: Testing (Day 3-4 - 5 minutes)

**Activity**: Verify team activation and usage

```bash
# Test 1: Activate team
./swap-rite.sh my-team

# Test 2: Verify frontmatter
grep "^role:" .claude/agents/orchestrator.md

# Test 3: Verify specialists
grep "specialist-name" .claude/agents/orchestrator.md

# Test 4: Restore previous team
./swap-rite.sh previous-team

# Test 5: Quick smoke test with real usage
# (e.g., did orchestrator receive consultation request?)
```

## Rollback Plan

If anything goes wrong during migration:

### Immediate Rollback (Within same day)

```bash
# 1. Restore original orchestrator.md
cp rites/my-team/agents/orchestrator.md.original \
   rites/my-team/agents/orchestrator.md

# 2. Remove YAML config
rm rites/my-team/orchestrator.yaml

# 3. Verify restoration
head -10 rites/my-team/agents/orchestrator.md

# 4. Don't commit anything
# (your original .md file is unchanged in git)
```

### Git-Level Rollback (If already committed)

```bash
# Revert commit
git revert HEAD

# Or reset to before migration
git reset --hard HEAD~1
```

## Migration Checklist per Team

For each team migrating, use this checklist:

### Pre-Migration
- [ ] Team lead reviewed architecture-overview.md
- [ ] Team understands benefits and tradeoffs
- [ ] Backups created (git history is backup)
- [ ] Time block scheduled (30 minutes)

### Extraction
- [ ] Metadata extracted (name, domain, color)
- [ ] Routing table extracted (all specialists)
- [ ] Handoff criteria extracted (all specialists)
- [ ] Skills extracted
- [ ] orchestrator.yaml created and syntax validated

### Generation
- [ ] Generator runs without errors
- [ ] Validator passes all 10 rules
- [ ] No unreplaced placeholders

### Comparison
- [ ] Diff reviewed (no surprises)
- [ ] Specialist names match
- [ ] Routing conditions preserved
- [ ] Handoff criteria intact

### Decision
- [ ] Team agreed to adopt generated version OR keep hand-written
- [ ] If adopting: Proceed to commit
- [ ] If not adopting: Delete YAML, restore original

### Commit
- [ ] Both orchestrator.yaml and orchestrator.md staged
- [ ] Commit message written and reviewed
- [ ] Commit created successfully

### Testing
- [ ] Team activation works (swap-rite.sh)
- [ ] Frontmatter parses correctly
- [ ] Specialists appear in routing
- [ ] Smoke test with real usage

## FAQ

### Q: Do I have to migrate?

**A**: No. Adoption is optional and gradual. You can migrate some teams now and others later.

### Q: What happens if I don't migrate?

**A**: Your orchestrator stays hand-written. You manually edit it. You don't benefit from template updates. That's OK.

### Q: Can I migrate partially (some teams, not all)?

**A**: Yes. Each team can adopt independently. Organization doesn't have to coordinate globally.

### Q: If I migrate, do I need to migrate others?

**A**: No. Migrated and hand-written orchestrators work together fine.

### Q: Can I migrate back to hand-written?

**A**: Yes. Delete orchestrator.yaml, keep orchestrator.md, update AGENT_MANIFEST if needed.

### Q: What if I want to migrate but not today?

**A**: Document your intent in team notes. Migrate when ready. Process and tools stay the same.

### Q: Who can migrate a team?

**A**: Any team member with git access. No special permissions needed.

### Q: Do I need approval to migrate?

**A**: Check with your team lead. Migration changes git history (commits) but doesn't change behavior.

### Q: What if generated version differs from hand-written?

**A**: Expected. They're semantically equivalent but formatted differently. Accept differences if content is correct.

### Q: Will migration affect my users?

**A**: No. Orchestrator behavior is identical. Users won't notice anything changed.

## Extraction Examples

### Example 1: Extracting rnd-pack

**Original orchestrator.md has**:

```markdown
| When | Route To | Prerequisites |
|------|----------|---------------|
| Needs research on technology integration | integration-researcher | ... |
| Needs evaluation of emerging tools | technology-scout | ... |
| Needs proof-of-concept implementation | prototype-engineer | ... |
| Needs long-term architecture | moonshot-architect | ... |
```

**Extract to YAML**:

```yaml
routing:
  integration-researcher: "Needs research on technology integration"
  technology-scout: "Needs evaluation of emerging tools"
  prototype-engineer: "Needs proof-of-concept implementation"
  moonshot-architect: "Needs long-term architecture"
```

### Example 2: Extracting security-pack

**Original has**:
```markdown
## Anti-Patterns

- Designing without threat modeling—always model threats first
- Over-trusting compliance frameworks—verify controls work
```

**Extract to YAML**:
```yaml
antipatterns:
  - "Designing without threat modeling—always model threats first"
  - "Over-trusting compliance frameworks—verify controls work"
```

### Example 3: Extracting Skills

**Original has**:
```markdown
## Skills Reference

`@security-ref` (threat modeling workflows)
`@doc-security` (security documentation)
`@prompting` (agent invocation)
```

**Extract to YAML**:
```yaml
skills:
  - "@security-ref for threat modeling workflows"
  - "@doc-security for security documentation"
  - "@prompting for agent invocation"
```

## Timeline by Team Size

| Team Characteristics | Extraction Time | Review Time | Total |
|-----|----------|-----------|---------|
| Simple (4 specialists, standard pattern) | 10 min | 5 min | 15 min |
| Standard (4-5 specialists, custom details) | 15 min | 10 min | 25 min |
| Complex (6 specialists, extensive criteria) | 20 min | 15 min | 35 min |

## Common Gotchas

### Gotcha 1: Specialist Names Don't Match Exactly

**Problem**: orchestrator.yaml has `tech-scout` but workflow.yaml has `technology-scout`

**Solution**: Generator will error. Use exact names from workflow.yaml.

**Prevention**: Compare before generating:
```bash
echo "workflow.yaml:"
yq '.phases[].agent' workflow.yaml
echo ""
echo "orchestrator.yaml:"
yq '.routing | keys' orchestrator.yaml
```

### Gotcha 2: Lost Custom Wording

**Problem**: Generated version uses template wording. You prefer your original wording.

**Solution**: Two options:
1. Accept generated (template improves consistency)
2. Keep hand-written (you own wording)

Don't migrate if wording matters more than consistency.

### Gotcha 3: Regeneration Changes Everything

**Problem**: Next time you regenerate, handoff criteria get reformatted

**Solution**: This is expected. YAML is source of truth. If you want different formatting, edit YAML, not .md file.

### Gotcha 4: Git History Gets Messy

**Problem**: Migration commit shows 40% of file changed

**Solution**: This is normal. You're switching from hand-written to generated.

**Mitigation**: Clean commit message explains why.

## Success Metrics

After migration, verify:

- [ ] Team continues to function (no behavior change)
- [ ] Orchestrator can be activated (swap-rite.sh works)
- [ ] Consultation requests work as before
- [ ] Routing works as before
- [ ] Team members understand YAML location
- [ ] Team knows how to regenerate if needed

## Communication Template

**To announce migration to team**:

> We're migrating [rite-name] orchestrator to a templated system. This means:
>
> **What changes**: Orchestrator configuration moves to YAML file (orchestrator.yaml)
> **What stays same**: Behavior and functionality are identical
> **When**: [Day/Sprint]
> **Who**: [Person will do work]
> **Risk**: Low (easy rollback if needed)
> **Benefit**: Future updates easier, consistent with other teams
>
> Timeline: 30-45 minutes total
> Review time: Needed from team lead
> Questions: [Slack channel or person]

## Next Steps After Migration

Once your team is migrated:

1. **Document**: Add to team wiki "We're using templated orchestrators"
2. **Learn**: Read SKILL.md for how to use system
3. **Update**: If handoff criteria change, edit orchestrator.yaml and regenerate
4. **Share**: Help other teams with migration if they ask

## Escalation

If migration has issues:

1. **Immediate**: Try rollback (restore original, delete YAML)
2. **Diagnostic**: Share error message and git status
3. **Help**: Post in team channel or contact tech team
4. **Revert**: Always can migrate back to hand-written (no permanent change)

---

**Status**: Ready for Phase 5 rollout
**Last Updated**: 2025-12-29
**Risk Level**: LOW (easy rollback, no behavior changes)
**Team readiness**: All 11 teams validated in Phase 3

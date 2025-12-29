# Orchestrator Templating: Rollout Plan

## Initiative Overview

**Goal**: Transition all teams from hand-written orchestrators to generated-from-config orchestrators

**Timeline**: 6 months, phased adoption (opt-in to forced, never mandatory)

**Expected Impact**:
- Orchestrator updates: 2 hours → 5 minutes (24x improvement)
- Maintenance burden: Shared orchestrator template vs 10 independent hand-written versions
- Consistency: Enforced by schema validation
- Evolution tracking: Visible in orchestrator.yaml changes, not markdown churn

## Phase A: Announcement & Setup (Week 1)

**Goal**: Make templating available, explain benefits, no pressure to adopt

**Activities**:

1. **Internal Announcement**
   - Message: "Orchestrator templating now available. Adoption optional. Learn more at [link]"
   - Audience: All team leads
   - Format: Slack announcement + email
   - Include: Quick start guide, FAQ, timeline

2. **Documentation Release**
   - Publish `@orchestrator-templates` skill
   - Create adoption guide with worked examples
   - Provide migration script for teams ready to adopt

3. **Enable Infrastructure**
   - Commit `orchestrator-generate.sh` and `validate-orchestrator.sh` to production
   - Deploy `.github/workflows/validate-orchestrators.yml`
   - Make `.githooks/pre-commit-orchestrator` available (optional for early adopters)

4. **Team Communication**
   - Team lead email with benefits and how to get started
   - Offer pairing session for first team adopting
   - Set expectation: "Available now, no deadline for adoption"

**Success Criteria**:
- [ ] Announcement posted in main communication channels
- [ ] Generator/validator scripts available in production
- [ ] CI workflow deployed and tested
- [ ] Documentation published and reviewed
- [ ] At least 1 team expresses interest in adopting

**Rollback Risk**: Low (everything is opt-in)

---

## Phase B: Opt-In Adoption (Weeks 2-4)

**Goal**: Enable teams to voluntarily migrate to templated orchestrators

**Activities**:

1. **Team-Driven Migration**
   - Teams that want to adopt run:
     ```bash
     cd /Users/tomtenuta/Code/roster
     ./scripts/orchestrator-migrate.sh <team-name>
     ```
   - Script extracts orchestrator.yaml from existing orchestrator.md
   - Validates generated output matches original
   - Team reviews diff and commits

2. **Pairing Sessions**
   - Offer 1-hour session with first adopter
   - Document their feedback and integration points
   - Use insights to improve documentation

3. **CI Integration for Adopters**
   - Teams with orchestrator.yaml automatically validated in CI
   - Non-adopters continue as before (hand-written orchestrators)
   - No breaking changes for teams not yet adopted

4. **Feedback Loop**
   - Weekly sync with early adopters
   - Gather feedback on usability, migration process
   - Update documentation based on common questions

**Success Criteria**:
- [ ] 2-3 teams have successfully migrated
- [ ] Feedback incorporated into documentation
- [ ] No blocking issues discovered
- [ ] CI validation working correctly for adopting teams

**Target Outcome**: 20-30% adoption by end of phase

**Rollback Risk**: Low (can delete orchestrator.yaml, revert to hand-written)

---

## Phase C: Infrastructure Update (Months 2-3)

**Goal**: Generate orchestrator.yaml for all teams (even those not yet opted in)

**Activities**:

1. **Bulk Generation**
   - Generate orchestrator.yaml for all 10 teams
   - Extract from current orchestrator.md files
   - Commit all in single PR with clear message

2. **CI Validation Enforcement**
   - CI begins validating ALL teams going forward
   - Validation is "soft" (warns about drift, doesn't fail merge)
   - Teams not opted in: can choose to stay manual

3. **Team Choice Point**
   - Each team makes explicit choice:
     - **Option A**: Adopt templating (update orchestrator.yaml for future changes)
     - **Option B**: Stay manual (delete orchestrator.yaml, keep hand-written)
   - No team is forced; both options are supported

4. **Documentation & Support**
   - Publish updated skill with guidance for both options
   - Create runbook for teams choosing to stay manual
   - Offer support for any team needing help

**Success Criteria**:
- [ ] orchestrator.yaml files generated for all teams
- [ ] CI validation updated to support both manual and templated
- [ ] All teams have made explicit choice (adopt or stay manual)
- [ ] Documentation updated for both paths

**Target Outcome**: 50-70% adoption (infrastructure in place, choice made)

**Rollback Risk**: Medium (easy to revert PR, but teams may have started using YAML)

---

## Phase D: Adoption Reminders & Support (Months 3-4)

**Goal**: Help remaining teams adopt, document benefits

**Activities**:

1. **Adoption Reminders**
   - Target teams that haven't adopted yet
   - Send personalized email with benefits specific to their team
   - Offer pairing session or support
   - Timeline: "Consider adopting in Q1"

2. **Case Studies**
   - Document teams that adopted successfully
   - Share feedback and results
   - Highlight time savings and maintenance benefits

3. **Refinement**
   - Incorporate feedback from Phase B-C adopters
   - Improve generator based on edge cases discovered
   - Update orchestrator template based on usage patterns

4. **Automation Expansion** (Optional)
   - Consider similar templating for specialist agents
   - Document pattern for future agent template architecture

**Success Criteria**:
- [ ] 70-80% adoption achieved
- [ ] Case studies documented and shared
- [ ] Remaining teams aware of benefits
- [ ] All edge cases handled

**Target Outcome**: 70-80% adoption, clear path for remaining teams

---

## Phase E: Sunset Hand-Written Orchestrators (Month 5-6)

**Goal**: Complete migration for remaining teams

**Activities**:

1. **Final Migration Push**
   - Contact remaining non-adopter teams
   - Offer migration assistance
   - Set soft deadline: "Full migration by end of Q1"

2. **Template Evolution**
   - Make breaking changes to orchestrator template (if needed)
   - Non-adopters must migrate or manually update
   - Adopters auto-benefit from improvements

3. **Deprecation Notice**
   - Document that hand-written orchestrators are deprecated
   - Recommend all new teams use templating by default
   - Legacy teams can stay manual if they choose

4. **Metrics & Celebration**
   - Calculate maintenance hours saved
   - Share ROI: "24x faster updates, 10 teams standardized"
   - Celebrate adoption milestone

**Success Criteria**:
- [ ] 95%+ adoption achieved or explicit "stay manual" choice made
- [ ] All new teams default to templating
- [ ] Maintenance burden demonstrably reduced
- [ ] System stable and production-ready

---

## Communication Templates

### Phase A: Announcement Email

```
Subject: Orchestrator Templating Now Available (Optional)

Hi Team Leads,

Great news: orchestrator templating is now available for your teams.

WHAT'S CHANGING:
Orchestrators can now be generated from a simple YAML configuration instead of
hand-written. Benefits:
- 24x faster to update (5 minutes vs 2 hours)
- Consistent pattern across all teams
- Schema validation prevents drift
- Less maintenance burden

WHAT YOU NEED TO DO:
Nothing right now. Adoption is completely optional. Your current orchestrators
continue to work as-is.

WANT TO ADOPT?
If your team is interested, we have:
- Migration script (5-minute process)
- Pairing session available
- Updated documentation and examples

Learn more: @orchestrator-templates skill

Questions? Reach out to [contact]

Thanks,
Integration Team
```

### Phase B: Early Adopter Feedback Request

```
Subject: How's the orchestrator migration going?

Hi [Team],

I see you've adopted the new orchestrator templating—thanks for being an early adopter!

Could you spare 15 minutes to give feedback on:
- Was the migration process smooth?
- Any confusing parts in the documentation?
- Anything we should improve?

This helps us make sure other teams have a great experience too.

Pairing session available if you hit any blockers.

Thanks,
Integration Team
```

### Phase C: Explicit Choice Request

```
Subject: Orchestrator Templating: Your Team's Choice

Hi [Team],

We've generated orchestrator.yaml for all teams. Your team can now make an
explicit choice:

OPTION A: Adopt Templating (Recommended)
- Update orchestrator.yaml for future changes
- Generator auto-produces orchestrator.md
- Less maintenance, faster updates

OPTION B: Stay Manual
- Keep hand-written orchestrator.md
- Delete the generated orchestrator.yaml
- Same workflow as before

No pressure either way. Both options are fully supported.

Make your choice by [DATE]:
Reply with "ADOPT" or "STAY MANUAL"

Questions? [contact]
```

### Phase D: Adoption Reminder

```
Subject: Orchestrator Templating: Final Adoption Window

Hi [Team],

We're in the final adoption window for orchestrator templating. Here's why
your team should consider adopting:

[Personalized benefit based on team's domain]

We've made it easier than ever:
- Simple YAML structure
- Full documentation and examples
- Pairing session available
- < 15 minutes to migrate

Let's talk: [schedule pairing session link]

Or reply "ADOPT" to get started today.

Thanks,
Integration Team
```

## Success Metrics

| Metric | Week 2 | Week 4 | Month 3 | Month 5 | Goal |
|--------|--------|--------|---------|---------|------|
| Teams Adopted | 1-2 | 2-3 | 5-7 | 9-10 | 10 |
| Adoption % | 10-20% | 20-30% | 50-70% | 90-100% | 100% |
| CI Failures | < 2/week | < 1/week | 0/week | 0/week | 0 |
| Support Requests | 5-10 | 3-5 | 1-2 | 0 | 0 |
| Documentation Hit Rate | 30% | 50% | 70% | 90% | 90%+ |
| Manual Edits Prevented | 0 | 5-10 | 20+ | 50+ | Trending up |

## Rollback Procedures

### If Phase A fails (tools not working)
- Revert `.github/workflows/validate-orchestrators.yml`
- Keep scripts in place but skip CI integration
- No teams affected (nothing was mandatory)

### If Phase B shows blocking issues
- Pause new adoptions
- Fix issues with generation or validation
- Offer additional support to current adopters
- Document workarounds

### If Phase C infrastructure breaks
- Disable CI validation temporarily
- Revert generator changes
- Keep orchestrator.yaml files (useful for recovery)
- Communicate status to all teams

### If team wants to unadopt
1. Delete `orchestrator.yaml`
2. Keep `orchestrator.md` as-is
3. Update `AGENT_MANIFEST.json` source field to "user"
4. Team continues with hand-written version
5. CI validation skips that team

## Contingency Plans

### Generator Bug Discovered Post-Adoption
1. File issue with detailed reproduction
2. Revert broken PR
3. Provide interim manual fix for affected teams
4. Fix and re-release generator
5. Notify all teams with updated script

### CI Validation Too Strict
1. Loosen validation rules
2. Add `skip-ci-orchestrator-check` override option
3. Document exception process
4. Gather feedback from teams hitting false positives
5. Update rules based on feedback

### Team Loses Manual Changes During Migration
1. Keep backup of original orchestrator.md
2. Document exact changes that were lost
3. Help team recover changes to orchestrator.yaml
4. Improve migration script to prevent recurrence

## Exit Criteria

Phase 5 is complete when:

- [ ] 100% of teams have made explicit choice (adopt or stay manual)
- [ ] CI validation active for all templated orchestrators
- [ ] Pre-commit hook available and documented
- [ ] Zero blocking issues in production
- [ ] Documentation complete and team-tested
- [ ] Rollback procedures tested
- [ ] All stakeholders understand expectations
- [ ] Maintenance burden measurably reduced
- [ ] Team leads report satisfaction with new process

## Related Documents

- `PHASE-5-CI-VALIDATION-STRATEGY.md` - Technical CI/CD details
- `ORCHESTRATOR-CI-IMPLEMENTATION.md` - GitHub Actions workflow details
- `@orchestrator-templates` - Skill documentation with examples

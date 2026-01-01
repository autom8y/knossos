# Orchestrator Templates Skill - Complete Documentation Index

> Master the durable abstraction layer for agent orchestrator generation. All documentation for creating, updating, troubleshooting, and understanding orchestrator templating.

## Quick Navigation

**First time?** Start here:
1. Read [SKILL.md](SKILL.md) - 10 minutes to understand what this is and when to use it
2. Look at [architecture-overview.md](architecture-overview.md) - 15 minutes for big picture understanding
3. Follow [create-new-team-orchestrator.md](create-new-team-orchestrator.md) if you're building a new team

**Creating a new orchestrator?** Follow:
→ [create-new-team-orchestrator.md](create-new-team-orchestrator.md) (complete step-by-step guide)

**Updating the template?** Follow:
→ [update-canonical-patterns.md](update-canonical-patterns.md) (when all orchestrators should change)

**Something broken?** Consult:
→ [troubleshooting.md](troubleshooting.md) (25+ scenarios with solutions)

**Migrating existing orchestrator?** Follow:
→ [migration-guide.md](migration-guide.md) (optional adoption path for Phase 5+)

**Need API docs?** See:
→ [schema-reference.md](schema-reference.md) (complete YAML field reference)

## Documentation Files

### 1. SKILL.md (Main Entry Point)
**Length**: ~1500 words | **Time**: 10-15 minutes | **Audience**: Everyone

What is orchestrator templating, when to use it, core concepts, and common workflows.

**Covers**:
- What it is and why it exists
- When to create, update, or troubleshoot orchestrators
- Core concept: durable abstractions (YAML specs separate from generation)
- Three main workflows with examples
- Integration with swap-team.sh, workflow.yaml, CEM, AGENT_MANIFEST
- Common patterns (linear, hub, domain-specific)
- Troubleshooting quick reference
- FAQ and next steps

**Best for**: First-time users, understanding purpose and scope

### 2. schema-reference.md (API Documentation)
**Length**: ~1800 words | **Time**: 10-20 minutes | **Audience**: Implementation

Complete specification of orchestrator.yaml configuration format.

**Covers**:
- Full field reference (team, frontmatter, routing, workflow_position, handoff_criteria, skills, optional fields)
- Type constraints and validation rules
- Template field mapping (how YAML maps to generated markdown)
- Common configuration patterns (simple, hub coordination, domain-specific)
- Validation checklist
- Migration path (extracting config from existing orchestrators)
- Schema evolution and backward compatibility

**Best for**: Creating or editing orchestrator.yaml, understanding structure

### 3. create-new-team-orchestrator.md (Step-by-Step Guide)
**Length**: ~2500 words | **Time**: 15-30 minutes (execution) | **Audience**: Team leads creating new teams

Complete walkthrough from design through testing.

**Covers**:
- Phase 1: Create orchestrator.yaml (team metadata, routing, criteria)
- Phase 2: Run generator (3 minutes)
- Phase 3: Validate output (3 minutes)
- Phase 4: Review generated file (5 minutes)
- Phase 5: Commit to git (2 minutes)
- Phase 6: Test team activation (2 minutes)
- Common issues and fixes
- Success criteria checklist
- Real example (doc-team-pack walkthrough)

**Best for**: Actually creating a new orchestrator for your team

### 4. update-canonical-patterns.md (Evolution Guide)
**Length**: ~2200 words | **Time**: 20-45 minutes (execution) | **Audience**: Infrastructure teams, tech leads

How to update the template when all orchestrators should change.

**Covers**:
- When to update (safe changes, planned updates, coordinated changes)
- Phase 1: Planning your change (risk assessment)
- Phase 2: Updating the template (5 minutes)
- Phase 3: Regenerating all teams (batch process)
- Phase 4: Reviewing diffs (spotting issues)
- Phase 5: Testing specific cases (team activation, protocol validation)
- Phase 6: Committing changes (clean commit practices)
- Phase 7: Communication (documenting breaking changes)
- Example workflows (safe, medium-risk, high-risk)
- Troubleshooting (generation failures, unexpected diffs)
- Rollback procedures

**Best for**: Making improvements that benefit all teams, coordinating large template changes

### 5. troubleshooting.md (Problem Solver)
**Length**: ~2800 words | **Time**: 5-30 minutes depending on issue | **Audience**: Anyone encountering problems

Solutions for 25+ common scenarios organized by symptom.

**Covers**:
- Quick diagnosis (symptom → root cause → solution)
- 8 major issue categories:
  1. Configuration errors (YAML invalid)
  2. Placeholder substitution failures (generator bugs)
  3. Validation failures (output quality issues)
  4. Frontmatter parsing issues (swap-team.sh failures)
  5. Specialist name inconsistencies (routing problems)
  6. Output differs from expected (configuration mismatches)
  7. Handoff criteria issues (phase gates wrong)
  8. Schema validation (extension points, custom fields)
  9. Team activation issues (swap-team.sh rejection)
- Quick fixes (4-step recovery)
- Escalation procedures (when to ask for help)

**Best for**: Debugging generation, validation, or activation failures

### 6. architecture-overview.md (Design Documentation)
**Length**: ~2200 words | **Time**: 15-20 minutes | **Audience**: Tech leads, architects, Phase 5+ integration engineers

Why this architecture exists, how it evolves, and how it integrates.

**Covers**:
- Big picture: what's durable (YAML), what's implementation (generation)
- Problem it solves (hand-written inconsistency)
- System architecture (4 components: template, schema, generator, validator)
- Data flow (creating orchestrators, updating templates)
- Integration points (swap-team.sh, workflow.yaml, CEM, AGENT_MANIFEST)
- Durable abstractions (why separation of concerns matters)
- Design decisions (why YAML instead of JSON, single template vs many, etc.)
- Failure modes and recovery (template breaks, generator breaks, invalid YAML)
- Roadmap (Phase 5 enhancements, Phase 6+ scaling)
- Key metrics (efficiency gains, time savings)

**Best for**: Understanding why this system exists, long-term planning, design decisions

### 7. migration-guide.md (Adoption Roadmap)
**Length**: ~2200 words | **Time**: 15-30 minutes per team | **Audience**: Teams considering adoption

How to migrate existing hand-written orchestrators to templated system.

**Covers**:
- Benefits of migration (easier team creation, faster updates, better consistency)
- 7 migration phases:
  1. Readiness (prep work)
  2. Extraction (pull config from existing orchestrator)
  3. Generation (run generator)
  4. Comparison (review differences)
  5. Decision point (adopt or stay hand-written)
  6. Commit (check in both files)
  7. Testing (verify team activation)
- Migration checklist per team
- Rollback plan (immediate rollback, git rollback)
- FAQ (optional adoption, partial migration, reverting, etc.)
- Extraction examples (real code)
- Timeline by team size
- Common gotchas (specialist name mismatches, lost wording, etc.)
- Success metrics
- Communication template

**Best for**: Teams planning to adopt templating (Phase 5+), understanding migration risk/benefit

### 8. integration-diagram.txt (Visual Reference)
**Length**: ~500 lines (ASCII diagrams) | **Time**: 5-10 minutes | **Audience**: Visual learners

System architecture and data flow as ASCII diagrams.

**Covers**:
- Developer workflow (6 steps from YAML to activated team)
- System architecture (components and interactions)
- Integration points (where orchestrator.md touches other systems)
- Team perspective (what team does at different times)
- Ecosystem integration (how orchestrator fits with other agents)
- Data flow summary
- Key principles
- Evolution roadmap
- File locations summary

**Best for**: Quick visual understanding, presentations, architectural discussions

## By Role

### I'm a Team Lead (Creating a New Team)
1. Read [SKILL.md](SKILL.md) - Understand the system (10 min)
2. Follow [create-new-team-orchestrator.md](create-new-team-orchestrator.md) - Create orchestrator (30 min)
3. Keep [troubleshooting.md](troubleshooting.md) handy - Resolve issues (as needed)

### I'm an Infrastructure/Tech Lead (Updating Patterns)
1. Read [architecture-overview.md](architecture-overview.md) - Understand design (20 min)
2. Follow [update-canonical-patterns.md](update-canonical-patterns.md) - Update template (30-60 min)
3. Use [troubleshooting.md](troubleshooting.md) - Debug if needed (as needed)

### I'm Adopting Templating (Migration)
1. Read [SKILL.md](SKILL.md) - Quick overview (10 min)
2. Follow [migration-guide.md](migration-guide.md) - Migrate team (30-45 min)
3. Consult [troubleshooting.md](troubleshooting.md) - Resolve issues (as needed)

### I'm Building/Modifying YAML Config
1. Reference [schema-reference.md](schema-reference.md) - Field documentation (as needed)
2. Follow [create-new-team-orchestrator.md](create-new-team-orchestrator.md) - Step-by-step (15-30 min)
3. Check examples in [create-new-team-orchestrator.md](create-new-team-orchestrator.md) - Real code (5 min)

### I'm Debugging an Issue
1. Quick scan [troubleshooting.md](troubleshooting.md) - Find your symptom (2 min)
2. Follow prescribed solution - Try fix (5-15 min)
3. If stuck: [troubleshooting.md](troubleshooting.md) escalation section - Get help (5 min)

### I'm Designing the Next Phase
1. Read [architecture-overview.md](architecture-overview.md) - Current design (20 min)
2. Scan [update-canonical-patterns.md](update-canonical-patterns.md) - What changes look like (10 min)
3. Review [migration-guide.md](migration-guide.md) - Adoption considerations (15 min)

## File Structure

```
orchestrator-templates/
├── SKILL.md                          # Main entry point
├── INDEX.md                          # This file
├── schema-reference.md               # YAML field documentation
├── create-new-team-orchestrator.md   # Creation guide
├── update-canonical-patterns.md      # Template evolution guide
├── troubleshooting.md                # Problem solutions
├── architecture-overview.md          # Design rationale
├── migration-guide.md                # Adoption roadmap
├── integration-diagram.txt           # ASCII diagrams
└── examples/
    ├── example-1-doc-team-pack.md       # Simple team example
    ├── example-2-10x-dev-pack.md        # Complex team example
    └── example-3-security-pack.md       # Domain-specific example
```

## Key Concepts Glossary

**Orchestrator**: Agent that coordinates multi-phase work by routing consultation requests to specialists

**Template** (orchestrator-base.md.tpl): Canonical markdown template containing protocol and structure, shared across all teams

**YAML Config** (orchestrator.yaml): Team-specific configuration file defining specialists, routing, and handoff criteria

**Generator** (orchestrator-generate.sh): Bash script that merges template + YAML config → orchestrator.md

**Validator** (validate-orchestrator.sh): Bash script that verifies generated orchestrator.md is production-ready

**Durable Abstraction**: Separation where YAML (semantic spec) survives changes to generation (implementation)

**Consultation Protocol**: Standard request/response schema all orchestrators use to interact with main agents

**Routing Table**: Decision table mapping conditions → specialists (which specialist to invoke when)

**Handoff Criteria**: Checkpoints defining when a specialist's work is complete and phase can transition

**Swap Team**: Command (swap-team.sh) that activates a team by parsing frontmatter and loading orchestrator

## Common Tasks Quick Links

| Task | Go To | Time |
|------|-------|------|
| Create new team orchestrator | [create-new-team-orchestrator.md](create-new-team-orchestrator.md) | 15-30 min |
| Update all teams with new pattern | [update-canonical-patterns.md](update-canonical-patterns.md) | 20-60 min |
| Understand why this exists | [architecture-overview.md](architecture-overview.md) | 15-20 min |
| Fix generation error | [troubleshooting.md](troubleshooting.md) | 5-15 min |
| Adopt templating for my team | [migration-guide.md](migration-guide.md) | 30-45 min |
| Find field definition | [schema-reference.md](schema-reference.md) | 2-5 min |
| See system architecture | [integration-diagram.txt](integration-diagram.txt) | 5-10 min |

## Success Criteria

After reading this documentation, you should be able to:

- [ ] Explain what orchestrator templating is and why it exists
- [ ] Create a new orchestrator from YAML config in 15-30 minutes
- [ ] Understand how generator, validator, and template work together
- [ ] Update the canonical template when all teams should change
- [ ] Troubleshoot common generation and validation failures
- [ ] Migrate an existing hand-written orchestrator to templated system
- [ ] Reference YAML schema when creating team configurations
- [ ] Explain integration points (swap-team.sh, CEM, workflow.yaml, AGENT_MANIFEST)

## Related Skills

- `@documentation` - Templates for PRDs, TDDs, ADRs (for documentation artifacts)
- `@standards` - Naming conventions and code style (for consistency)
- `@10x-workflow` - Phase coordination and handoff protocols (related workflows)
- `@agent-prompt-engineering` - Writing effective agent prompts (for understanding agents)

## Support and Escalation

### Getting Help

1. **Quick question**: Check [troubleshooting.md](troubleshooting.md) FAQ
2. **Configuration question**: See [schema-reference.md](schema-reference.md)
3. **Process question**: Refer to relevant guide (create, update, migrate)
4. **Stuck anyway**: Post error message + what you tried to team channel

### Known Limitations

- forge-pack diagram formatting is suboptimal for 6-agent teams (visual only, doesn't affect function)
- extension_points in YAML not fully implemented yet (Phase 5 feature)
- CI integration not yet in place (Phase 5 work)

### Future Enhancements

- Phase 5: CI integration, error message improvements
- Phase 6+: Nested orchestrators, template versioning, multi-team coordination

## Version Information

**Documentation Version**: 1.0 (Phase 4)
**Skill Status**: Production
**Last Updated**: 2025-12-29
**Tested**: All 11 teams validated in Phase 3
**Compatibility**: Backward compatible with existing hand-written orchestrators

---

## Navigation

**Start here**: [SKILL.md](SKILL.md)
**Creating orchestrator**: [create-new-team-orchestrator.md](create-new-team-orchestrator.md)
**Updating template**: [update-canonical-patterns.md](update-canonical-patterns.md)
**Having issues**: [troubleshooting.md](troubleshooting.md)
**Full reference**: [schema-reference.md](schema-reference.md)
**Understanding why**: [architecture-overview.md](architecture-overview.md)
**Adopting later**: [migration-guide.md](migration-guide.md)

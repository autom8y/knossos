# Orchestrator Templates

> Durable abstraction layer for agent orchestrator generation.

## Quick Start

**First time?** Read [SKILL.md](SKILL.md) (10 min) to understand what this is.

**Creating orchestrator?** Follow [create-new-rite-orchestrator.md](create-new-rite-orchestrator.md) (15-30 min).

**Updating template?** Follow [update-canonical-patterns.md](update-canonical-patterns.md) (20-45 min).

**Something broken?** Consult [troubleshooting.md](troubleshooting.md) (5-15 min).

**Need field reference?** See [schema-reference.md](schema-reference.md).

## What It Is

Orchestrator templating separates **semantic specifications** (YAML) from **implementation** (generation):

```
orchestrator.yaml (team config, durable)
         ↓
orchestrator-generate.sh (bash, replaceable)
         ↓
orchestrator.md (production output, regenerable)
```

**Key insight**: Template is durable. Generation is an implementation detail.

## Documentation Map

| File | Purpose | Time | Audience |
|------|---------|------|----------|
| [SKILL.md](SKILL.md) | Main entry point | 10-15 min | Everyone |
| [create-new-rite-orchestrator.md](create-new-rite-orchestrator.md) | Step-by-step creation | 15-30 min | Team leads |
| [update-canonical-patterns.md](update-canonical-patterns.md) | Template evolution | 20-45 min | Infrastructure |
| [troubleshooting.md](troubleshooting.md) | Problem solving | 5-30 min | Debug issues |
| [schema-reference.md](schema-reference.md) | YAML field docs | As needed | Implementation |
| [architecture-overview.md](architecture-overview.md) | Design rationale | 15-20 min | Architects |
| [migration-guide.md](migration-guide.md) | Adoption roadmap | 15-30 min | Migration |
| [QUICK-REFERENCE.md](QUICK-REFERENCE.md) | Cheat sheet | 2 min | Quick lookup |
| [references/consultation-protocol.md](references/consultation-protocol.md) | Request/response schemas | As needed | Reference |

## By Task

### Creating a New Orchestrator
1. Read [SKILL.md](SKILL.md) for context
2. Follow [create-new-rite-orchestrator.md](create-new-rite-orchestrator.md)
3. Use [QUICK-REFERENCE.md](QUICK-REFERENCE.md) for commands

### Updating the Template
1. Review [architecture-overview.md](architecture-overview.md) for design
2. Follow [update-canonical-patterns.md](update-canonical-patterns.md)
3. Reference [troubleshooting.md](troubleshooting.md) if issues arise

### Debugging Issues
1. Scan [troubleshooting.md](troubleshooting.md) for symptoms
2. Check [schema-reference.md](schema-reference.md) for field validation
3. Use [QUICK-REFERENCE.md](QUICK-REFERENCE.md) for validation commands

### Migrating Existing Orchestrator
1. Read [SKILL.md](SKILL.md) for overview
2. Follow [migration-guide.md](migration-guide.md)
3. Reference [schema-reference.md](schema-reference.md) for extraction

## File Locations

| Component | Location |
|-----------|----------|
| Skill docs | `user-skills/orchestration/orchestrator-templates/` |
| Generator | `/roster/templates/orchestrator-generate.sh` |
| Validator | `/roster/templates/validate-orchestrator.sh` |
| Template | `/roster/templates/orchestrator-base.md.tpl` |
| Team config | `.claude/rites/{team}/orchestrator.yaml` |
| Generated | `.claude/rites/{team}/agents/orchestrator.md` |

## Common Commands

```bash
# Generate orchestrator
/roster/templates/orchestrator-generate.sh my-team

# Validate output
/roster/templates/validate-orchestrator.sh .claude/rites/my-team/agents/orchestrator.md

# Activate team
./swap-rite.sh my-team
```

See [QUICK-REFERENCE.md](QUICK-REFERENCE.md) for complete command reference.

## Success Criteria

After reading this documentation, you should be able to:

- [ ] Explain what orchestrator templating is and why it exists
- [ ] Create a new orchestrator from YAML in 15-30 minutes
- [ ] Update the canonical template when patterns evolve
- [ ] Troubleshoot common generation failures
- [ ] Reference YAML schema when creating configurations

## Related Skills

- `orchestration` - Phase coordination and handoff protocols
- `documentation` - Templates for PRDs, TDDs, ADRs
- `standards` - Naming conventions and code style

---

**Documentation Version**: 1.0 (Phase 4)
**Status**: Production
**Last Updated**: 2025-12-29

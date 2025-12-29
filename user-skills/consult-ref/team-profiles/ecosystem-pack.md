# ecosystem-pack

> Ecosystem infrastructure lifecycle: diagnose, design, implement, document, and validate CEM/skeleton/roster changes

## Overview

The ecosystem infrastructure team maintains the foundational systems that power Claude Code's multi-project context architecture. This team handles CEM sync issues, skeleton template updates, roster schema changes, and hook/skill pattern development through systematic diagnosis and backward-compatible implementation.

## Switch Command

```bash
/ecosystem
```

## Agents

| Agent | Model | Role |
|-------|-------|------|
| **ecosystem-analyst** | opus | Diagnoses ecosystem problems, traces root causes |
| **context-architect** | opus | Designs hooks, skills, schemas, CEM behavior |
| **integration-engineer** | sonnet | Implements CEM/skeleton/roster code changes |
| **documentation-engineer** | sonnet | Writes migration runbooks, compatibility matrices |
| **compatibility-tester** | opus | Validates across satellite matrix, tests upgrades |

## Workflow

```
analysis → design → implementation → documentation → validation
   │         │            │                │              │
   ▼         ▼            ▼                ▼              ▼
  Gap     Context     Working        Migration      Compatibility
Analysis  Design   Implementation    Runbook           Report
```

## Complexity Levels

| Level | When to Use | Phases |
|-------|-------------|--------|
| **PATCH** | Single file/config, no schema impact | analysis, implementation, validation |
| **MODULE** | Single system (CEM or skeleton or roster) | All 5 phases |
| **SYSTEM** | Multi-system change (CEM + skeleton + roster) | All 5 phases |
| **MIGRATION** | Cross-satellite rollout, coordinated upgrade | All 5 phases + extended validation |

## Best For

- Debugging CEM sync failures
- Designing new hook patterns
- Creating new skill patterns
- Updating settings schemas
- Migrating satellites to new ecosystem versions
- Testing backward compatibility
- Infrastructure changes requiring satellite validation

## Not For

- Product feature development → use 10x-dev-pack
- Satellite-specific code → use 10x-dev-pack in that satellite
- Documentation without ecosystem changes → use doc-team-pack
- Quick fixes that don't affect multiple satellites → use /hotfix

## Quick Start

```bash
/ecosystem                        # Switch to team
/task "Fix CEM sync conflicts"   # Start infrastructure task
# Work through phases...
/wrap                            # Finalize
```

## Common Patterns

### Debug CEM Sync Issue

```bash
/cem-debug                       # Fast-track diagnostic
# Ecosystem Analyst produces Gap Analysis with root cause
```

### Design New Hook Pattern

```bash
/ecosystem
/task "Add pre-commit hook support" --complexity=MODULE
# Analyst → Architect → Engineer → Docs → Tester
```

### Update Skeleton Template

```bash
/ecosystem
/task "Add new session lifecycle hook" --complexity=SYSTEM
# Full workflow with satellite validation
```

### Migrate Satellites to New Schema

```bash
/ecosystem
/task "Migrate to settings v2 schema" --complexity=MIGRATION
# Extended validation across all registered satellites
```

## Related Commands

- `/ecosystem` - Full pipeline (all agents)
- `/cem-debug` - Fast diagnostic for CEM issues
- `/ecosystem-analyze` - Analysis phase only
- `/ecosystem-design` - Design phase only
- `/ecosystem-implement` - Implementation phase only
- `/ecosystem-document` - Documentation phase only
- `/ecosystem-validate` - Validation phase only

## Key Differences from 10x-dev-pack

| Aspect | 10x-dev-pack | ecosystem-pack |
|--------|--------------|----------------|
| **Entry artifact** | PRD (user requirements) | Gap Analysis (diagnostic report) |
| **Design output** | TDD (technical design) | Context Design (hook/skill/schema) |
| **Focus** | Product features | Infrastructure enablers |
| **Documentation** | Optional ADRs | Required migration runbooks |
| **Validation** | Feature testing | Satellite matrix compatibility |
| **Success criteria** | Feature works in project | Works across all satellites |

## Success Criteria

Ecosystem changes succeed when:

- [ ] CEM sync completes without conflicts across test satellites
- [ ] Skeleton template application succeeds for new satellite init
- [ ] Hook/skill/agent registration works without manual intervention
- [ ] Migration runbooks execute successfully
- [ ] Compatibility matrix reflects actual tested combinations
- [ ] Breaking changes have documented upgrade paths
- [ ] No regressions in existing satellite functionality

## Related Teams

- [10x-dev-pack](10x-dev-pack.md) - Escalates satellite-specific issues to ecosystem-pack when root cause is infrastructure
- [doc-team-pack](doc-team-pack.md) - Uses migration runbooks as input for user-facing guides

---
name: complexity-levels
description: "Reference for task complexity classification and phase requirements"
---

# Complexity Levels Reference

> Canonical guide to task complexity classification, phase requirements, and team-specific variations.

## Overview

Complexity levels determine workflow phases, documentation requirements, and agent coordination. Standard levels (SCRIPT | MODULE | SERVICE | PLATFORM) apply to general development workflows, while specialized teams like ecosystem-pack use domain-specific levels.

---

## Standard Complexity Levels

Used by most teams (10x-dev-pack, hygiene-pack, etc.) for general software development tasks.

### SCRIPT

**Definition**: Single file change, under 200 lines of code, no external dependencies.

**Scope**:
- Bug fixes in isolated files
- Configuration tweaks
- Simple utility scripts
- Documentation updates

**Phases Required**:
1. **Implementation** (Principal Engineer)
2. **Validation** (QA Adversary)

**Artifacts**:
- Code changes (always)
- Tests (if applicable)
- Test Plan (lightweight)

**When to Use**:
- Change is confined to one file
- No architectural decisions needed
- Requirements are crystal clear
- Minimal risk and impact

**Skip Flags**:
- `--skip-prd`: Almost always appropriate for SCRIPT complexity
- `--skip-tdd`: Automatically skipped (no TDD needed)

---

### MODULE

**Definition**: Multiple files with clear interfaces, single logical component or subsystem.

**Scope**:
- New feature within existing service
- Refactoring a component
- Adding API endpoints
- Database schema changes

**Phases Required**:
1. **Requirements** (Requirements Analyst) - produces PRD
2. **Design** (Architect) - produces TDD + ADRs
3. **Implementation** (Principal Engineer)
4. **Validation** (QA Adversary)

**Artifacts**:
- PRD (requirements)
- TDD (technical design)
- ADRs (architecture decisions, as needed)
- Code changes
- Unit/integration tests
- Test Plan

**When to Use**:
- Multi-file changes with interface boundaries
- New component in existing architecture
- Requires design decisions
- Integration points need documentation

**Skip Flags**:
- `--skip-prd`: Not recommended (MODULE needs clear requirements)
- `--skip-tdd`: Use cautiously (design usually helpful)

---

### SERVICE

**Definition**: Multiple modules, APIs, persistence layer, external integrations.

**Scope**:
- New microservice
- Major service refactoring
- Multi-system integration
- API versioning changes

**Phases Required**:
1. **Requirements** (Requirements Analyst) - produces PRD
2. **Design** (Architect) - produces TDD + ADRs
3. **Implementation** (Principal Engineer)
4. **Validation** (QA Adversary)

**Artifacts**:
- PRD (detailed requirements with use cases)
- TDD (system architecture, API contracts, data models)
- ADRs (significant decisions documented)
- Code changes
- Integration tests
- Test Plan (comprehensive)

**When to Use**:
- Building new service from scratch
- Major architectural changes
- Multi-module coordination needed
- External API contracts involved

**Additional Considerations**:
- Infrastructure requirements (deployment, monitoring)
- Data migration strategies
- Backward compatibility planning
- Performance benchmarking

---

### PLATFORM

**Definition**: Multiple services, infrastructure, cross-cutting concerns, organizational impact.

**Scope**:
- Multi-service platform features
- Infrastructure-level changes
- Cross-team coordination
- System-wide refactoring

**Phases Required**:
1. **Requirements** (Requirements Analyst) - produces PRD
2. **Design** (Architect) - produces TDD + ADRs
3. **Implementation** (Principal Engineer)
4. **Validation** (QA Adversary)

**Artifacts**:
- PRD (extensive, with stakeholder analysis)
- TDD (multi-service design, deployment strategy, infrastructure)
- ADRs (all major decisions)
- Code changes (across services)
- Integration and E2E tests
- Test Plan (platform-level validation)
- Migration runbooks

**When to Use**:
- Changes span multiple services
- Infrastructure or tooling changes
- Org-wide impact
- Requires coordination across teams

**Additional Considerations**:
- Phased rollout strategy
- Feature flags and gradual deployment
- Monitoring and alerting updates
- Documentation for multiple audiences
- Stakeholder communication plan

---

## Team-Specific Variations

### Ecosystem Pack (CEM/Skeleton/Roster Infrastructure)

Used by ecosystem-pack for infrastructure work affecting the agent ecosystem itself.

#### PATCH

**Definition**: Single file or config change, no schema impact.

**Scope**:
- Typo fix in hook script
- Config value adjustment
- Documentation correction
- Single bash function change

**Phases**:
1. **Analysis** (Ecosystem Analyst) - lightweight gap analysis
2. **Implementation** (Integration Engineer)
3. **Validation** (Compatibility Tester) - verify no regressions

**When to Use**:
- Isolated change with no downstream effects
- No schema or interface changes
- Quick fix or improvement

---

#### MODULE

**Definition**: Single system change (CEM **or** skeleton **or** roster).

**Scope**:
- New hook type in skeleton
- CEM sync logic enhancement
- Roster schema addition
- Single-system feature

**Phases**:
1. **Analysis** (Ecosystem Analyst) - produces gap-analysis
2. **Design** (Context Architect) - produces context-design
3. **Implementation** (Integration Engineer)
4. **Documentation** (Documentation Engineer) - produces migration-runbook
5. **Validation** (Compatibility Tester) - produces compatibility-report

**When to Use**:
- Change affects one of CEM/skeleton/roster
- Schema or interface changes within one system
- Needs design to prevent downstream issues

---

#### SYSTEM

**Definition**: Multi-system change affecting CEM + skeleton + roster coordination.

**Scope**:
- New lifecycle phase affecting all three systems
- Cross-system schema coordination
- Hook registration flow changes
- Settings merge behavior updates

**Phases**:
1. **Analysis** (Ecosystem Analyst) - produces gap-analysis
2. **Design** (Context Architect) - produces context-design
3. **Implementation** (Integration Engineer)
4. **Documentation** (Documentation Engineer) - produces migration-runbook
5. **Validation** (Compatibility Tester) - produces compatibility-report

**When to Use**:
- Changes require coordination across CEM, skeleton, and roster
- Schema changes affect multiple systems
- Integration points between systems change

---

#### MIGRATION

**Definition**: Cross-satellite rollout requiring coordination, breaking changes.

**Scope**:
- Breaking schema changes
- New satellite onboarding requirements
- CEM sync behavior changes affecting all satellites
- Roster-wide pattern changes

**Phases**:
1. **Analysis** (Ecosystem Analyst) - produces gap-analysis
2. **Design** (Context Architect) - produces context-design
3. **Implementation** (Integration Engineer)
4. **Documentation** (Documentation Engineer) - produces migration-runbook
5. **Validation** (Compatibility Tester) - produces compatibility-report

**When to Use**:
- Breaking changes requiring satellite updates
- Rollout needs phasing across satellites
- Backward compatibility concerns
- Coordination with satellite maintainers needed

**Additional Artifacts**:
- Migration runbook (detailed rollout steps)
- Satellite compatibility matrix
- Rollback procedures
- Communication plan

---

## Decision Criteria

### How to Choose the Right Complexity Level

Use this decision tree to select appropriate complexity:

```
START
  |
  ├─ Single file, < 200 LOC, no interfaces changed?
  |    └─ YES → SCRIPT (standard) or PATCH (ecosystem)
  |
  ├─ Multiple files, single component/system?
  |    └─ YES → MODULE
  |         |
  |         ├─ Standard development? → MODULE (standard)
  |         └─ CEM/skeleton/roster infrastructure? → MODULE (ecosystem)
  |
  ├─ Multiple components, APIs, persistence?
  |    └─ YES → SERVICE (standard) or SYSTEM (ecosystem)
  |
  └─ Multiple services, infrastructure, cross-team?
       └─ YES → PLATFORM (standard) or MIGRATION (ecosystem)
```

### Key Questions

Ask yourself:

1. **Scope**: How many files/systems affected?
   - 1 file → SCRIPT/PATCH
   - 1 system/module → MODULE
   - Multiple systems → SERVICE/SYSTEM
   - Platform-wide → PLATFORM/MIGRATION

2. **Design Decisions**: How many architectural choices?
   - 0 decisions → SCRIPT/PATCH
   - 1-2 decisions → MODULE
   - 3+ decisions → SERVICE or higher

3. **Integration Points**: How many external interfaces?
   - 0 integrations → SCRIPT/MODULE
   - 1-2 integrations → MODULE/SERVICE
   - 3+ integrations → SERVICE/PLATFORM

4. **Impact Radius**: How many teams/satellites affected?
   - None → SCRIPT/PATCH
   - Single team → MODULE/SERVICE
   - Multiple teams → PLATFORM/MIGRATION

5. **Breaking Changes**: Any backward incompatibility?
   - No → Match scope (SCRIPT to SERVICE)
   - Yes → SERVICE or higher (ecosystem: MIGRATION)

### Complexity Escalation

Start conservative, escalate if needed:

- **Start**: Initial complexity estimate
- **During Design**: Architect may escalate if complexity underestimated
- **During Implementation**: Engineer may escalate if scope grows
- **User Confirmation**: Always confirm before escalating to higher complexity

### Common Anti-Patterns

**Over-complexifying**:
- Don't use PLATFORM for simple multi-file changes
- Don't require TDD for documentation updates
- Don't create PRD for typo fixes

**Under-complexifying**:
- Don't use SCRIPT for multi-system changes
- Don't skip design for new integrations
- Don't use PATCH for schema changes (ecosystem)

---

## Phase Requirements by Complexity

### Standard Complexity Phase Matrix

| Complexity | Requirements | Design | Implementation | Validation |
|------------|--------------|--------|----------------|------------|
| SCRIPT     | Optional*    | Skip   | Required       | Required   |
| MODULE     | Required     | Required | Required     | Required   |
| SERVICE    | Required     | Required | Required     | Required   |
| PLATFORM   | Required     | Required | Required     | Required   |

*Use `--skip-prd` for SCRIPT if requirements are obvious.

### Ecosystem Complexity Phase Matrix

| Complexity | Analysis | Design | Implementation | Documentation | Validation |
|------------|----------|--------|----------------|---------------|------------|
| PATCH      | Light    | Skip   | Required       | Skip          | Required   |
| MODULE     | Required | Required | Required     | Required      | Required   |
| SYSTEM     | Required | Required | Required     | Required      | Required   |
| MIGRATION  | Required | Required | Required     | Required      | Required   |

---

## Artifact Requirements by Complexity

### Standard Complexity Artifacts

| Artifact | SCRIPT | MODULE | SERVICE | PLATFORM |
|----------|--------|--------|---------|----------|
| PRD      | Optional | Required | Required | Required |
| TDD      | No     | Required | Required | Required |
| ADRs     | No     | As needed | Required | Required |
| Code     | Required | Required | Required | Required |
| Tests    | As needed | Required | Required | Required |
| Test Plan | Light | Required | Required | Required |
| Migration Runbook | No | No | As needed | Required |

### Ecosystem Complexity Artifacts

| Artifact | PATCH | MODULE | SYSTEM | MIGRATION |
|----------|-------|--------|--------|-----------|
| Gap Analysis | Light | Required | Required | Required |
| Context Design | No | Required | Required | Required |
| Implementation | Required | Required | Required | Required |
| Migration Runbook | No | Required | Required | Required |
| Compatibility Report | Required | Required | Required | Required |

---

## Examples by Complexity

### Standard Levels

**SCRIPT**:
- Fix typo in error message
- Update configuration value
- Add logging statement
- Rename variable

**MODULE**:
- Add user authentication to existing app
- Refactor data access layer
- Implement new API endpoint
- Add caching layer

**SERVICE**:
- Build new notification service
- Migrate database to new schema
- Add GraphQL API alongside REST
- Implement event-driven architecture

**PLATFORM**:
- Multi-region deployment capability
- Centralized logging/monitoring
- CI/CD pipeline overhaul
- Multi-tenant data isolation

### Ecosystem Levels

**PATCH**:
- Fix typo in hook comment
- Update hook registration path
- Correct skill metadata

**MODULE**:
- Add new lifecycle hook type
- Implement settings merge strategy
- Add skill category to roster

**SYSTEM**:
- Lifecycle phase coordination (CEM + skeleton + roster)
- Hook registration flow redesign
- Settings schema versioning

**MIGRATION**:
- Categorical resource organization rollout
- YAML-based hook registration migration
- Breaking schema changes across satellites

---

## Related Documentation

- **Session Context Schema**: `/Users/tomtenuta/Code/roster/user-skills/session-common/session-context-schema.md`
- **Task Reference**: `/Users/tomtenuta/Code/roster/user-skills/orchestration/task-ref/SKILL.md`
- **Ecosystem Workflow**: `/Users/tomtenuta/Code/roster/.claude/ACTIVE_WORKFLOW.yaml`
- **Agent Coordination**: See `orchestration` skill
- **Standards**: See `standards` skill

---

## Notes

- Complexity is **estimated upfront** but can **escalate during workflow**
- **Architect has authority** to escalate complexity during design phase
- **Always confirm with user** before escalating to higher complexity
- **Err on the side of lower complexity** initially (SCRIPT/MODULE over SERVICE/PLATFORM)
- **Ecosystem pack** uses different levels because infrastructure has different failure modes than application code

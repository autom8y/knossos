# Skeleton Claude Resources Inventory

> Complete catalog of skeleton_claude `.claude/` resources for migration planning

**Generated**: 2026-01-03
**Source**: `/Users/tomtenuta/Code/skeleton_claude/.claude/`
**Target**: `/Users/tomtenuta/Code/roster/.claude/`

**Note**: This document captures the historical skeleton_claude structure at the time of migration.
Location references (e.g., `.claude/user-agents/`) describe where resources existed in skeleton, NOT their
final destinations in roster. For the correct roster artifact architecture, see [INTEGRATION.md](../INTEGRATION.md).

After migration, the correct locations are:
- User-level content: `roster/user-*/` syncs to `~/.claude/*/`
- Team content: `roster/rites/{pack}/` syncs to `.claude/` (project-level)

---

## Summary

| Resource Type | Skeleton Count | Roster Count | Unique to Skeleton | Migration Needed |
|---------------|----------------|--------------|--------------------|--------------------|
| Agents | 5 | 5 | 0 | No (roster superior) |
| Commands | 2 | 2 | 0 | No (identical) |
| Hooks (main) | 11 | 12 | 2 | Yes |
| Hooks (lib/) | 10 | 13 | 2 | Evaluate |
| Skills | 18 | 25 | 11 | Yes |
| User-Agents | 7 | 0 | 7 | Yes |
| User-Commands | 38 | 0 | 38 | Yes |
| User-Skills | 2 | 0 | 2 | Yes |
| Schemas | 1 | 0 | 1 | Evaluate |

---

## 1. Agents

**Location**: `.claude/agents/`

Both skeleton and roster have the same 5 agents. **Roster versions are superior** (contain enhancements like security consultation, entry point selection, documentation impact assessment, and impact assessment sections).

| Resource | Skeleton Path | Purpose | Roster Equivalent | Action |
|----------|---------------|---------|-------------------|--------|
| architect.md | agents/architect.md | System design, TDD/ADR production | agents/architect.md | **KEEP ROSTER** (has security-pack consultation) |
| orchestrator.md | agents/orchestrator.md | Work breakdown, agent coordination | agents/orchestrator.md | **KEEP ROSTER** (has entry point selection) |
| principal-engineer.md | agents/principal-engineer.md | Implementation, code production | agents/principal-engineer.md | **KEEP ROSTER** (identical) |
| qa-adversary.md | agents/qa-adversary.md | Testing, validation | agents/qa-adversary.md | **KEEP ROSTER** (has doc impact, security/SRE handoff) |
| requirements-analyst.md | agents/requirements-analyst.md | PRD production | agents/requirements-analyst.md | **KEEP ROSTER** (has impact assessment) |

**Migration Recommendation**: DEPRECATE skeleton agents. Roster versions are enhanced.

---

## 2. Commands

**Location**: `.claude/commands/`

| Resource | Skeleton Path | Purpose | Roster Equivalent | Action |
|----------|---------------|---------|-------------------|--------|
| pr.md | commands/pr.md | Create GitHub PR | commands/pr.md | **KEEP ROSTER** (identical) |
| spike.md | commands/spike.md | Time-boxed research | commands/spike.md | **KEEP ROSTER** (identical) |

**Migration Recommendation**: DEPRECATE skeleton commands. Files are identical.

---

## 3. Hooks (Main Level)

**Location**: `.claude/hooks/`

### Skeleton Hooks

| Resource | Skeleton Path | Purpose | Roster Equivalent | Action |
|----------|---------------|---------|-------------------|--------|
| artifact-tracker.sh | hooks/artifact-tracker.sh | Track PRD/TDD/ADR creation | hooks/tracking/artifact-tracker.sh | **KEEP ROSTER** (uses hooks-init.sh) |
| auto-park.sh | hooks/auto-park.sh | Auto-park on session stop | hooks/session-guards/auto-park.sh | **KEEP ROSTER** (uses hooks-init.sh) |
| coach-mode.sh | hooks/coach-mode.sh | Coach pattern enforcement | hooks/context-injection/coach-mode.sh | **KEEP ROSTER** (reorganized) |
| command-validator.sh | hooks/command-validator.sh | Validate slash commands | hooks/validation/command-validator.sh | **KEEP ROSTER** (reorganized) |
| commit-tracker.sh | hooks/commit-tracker.sh | Track git commits | hooks/tracking/commit-tracker.sh | **KEEP ROSTER** (has lazy init optimization) |
| delegation-check.sh | hooks/delegation-check.sh | Check Task tool delegation | hooks/validation/delegation-check.sh | **KEEP ROSTER** (has caching optimization) |
| session-audit.sh | hooks/session-audit.sh | Audit session operations | hooks/tracking/session-audit.sh | **KEEP ROSTER** (reorganized) |
| session-context.sh | hooks/session-context.sh | SessionStart context injection | hooks/context-injection/session-context.sh | **KEEP ROSTER** (has verbose mode) |
| session-write-guard.sh | hooks/session-write-guard.sh | Guard SESSION_CONTEXT.md writes | hooks/session-guards/session-write-guard.sh | **KEEP ROSTER** (uses hooks-init.sh) |
| start-preflight.sh | hooks/start-preflight.sh | Pre-flight checks on session start | hooks/session-guards/start-preflight.sh | **KEEP ROSTER** (reorganized) |
| team-validator.sh | hooks/team-validator.sh | Validate rite | None | **EVALUATE** - may be useful |
| workflow-validator.sh | hooks/workflow-validator.sh | Validate workflow.yaml | None | **EVALUATE** - may be useful |

### Roster-Only Hooks (Main Level)

| Resource | Path | Purpose | Notes |
|----------|------|---------|-------|
| orchestrator-bypass-check.sh | hooks/validation/ | Check orchestrator bypass | New in roster |
| orchestrator-router.sh | hooks/validation/ | Route orchestrated work | New in roster |
| orchestrated-mode.sh | hooks/context-injection/ | Orchestrated mode context | New in roster |
| base_hooks.yaml | hooks/ | Base hook configuration | New in roster |

**Migration Recommendation**:
- KEEP all roster hooks (reorganized structure, optimizations)
- EVALUATE team-validator.sh and workflow-validator.sh for migration (unique to skeleton)

---

## 4. Hooks Library

**Location**: `.claude/hooks/lib/`

### Skeleton lib/ Files

| Resource | Skeleton Path | Purpose | Roster Equivalent | Action |
|----------|---------------|---------|-------------------|--------|
| artifact-validation.sh | hooks/lib/artifact-validation.sh | Validate artifact existence | None | **EVALUATE** - may be useful |
| config.sh | hooks/lib/config.sh | Configuration constants | hooks/lib/config.sh | **KEEP ROSTER** |
| handoff-validator.sh | hooks/lib/handoff-validator.sh | Validate agent handoffs | None | **EVALUATE** - may be useful |
| logging.sh | hooks/lib/logging.sh | Logging utilities | hooks/lib/logging.sh | **KEEP ROSTER** |
| primitives.sh | hooks/lib/primitives.sh | Core utilities | hooks/lib/primitives.sh | **KEEP ROSTER** |
| session-core.sh | hooks/lib/session-core.sh | Session core functions | hooks/lib/session-core.sh | **KEEP ROSTER** |
| session-manager.sh | hooks/lib/session-manager.sh | Session management | hooks/lib/session-manager.sh | **KEEP ROSTER** |
| session-state.sh | hooks/lib/session-state.sh | Session state handling | hooks/lib/session-state.sh | **KEEP ROSTER** |
| session-utils.sh | hooks/lib/session-utils.sh | Session utilities | hooks/lib/session-utils.sh | **KEEP ROSTER** |
| worktree-manager.sh | hooks/lib/worktree-manager.sh | Git worktree management | hooks/lib/worktree-manager.sh | **KEEP ROSTER** |

### Roster-Only lib/ Files

| Resource | Path | Purpose | Notes |
|----------|------|---------|-------|
| hooks-init.sh | hooks/lib/ | Hook initialization (ADR-0002) | New in roster - core infrastructure |
| orchestration-audit.sh | hooks/lib/ | Orchestration auditing | New in roster |
| session-fsm.sh | hooks/lib/ | Session finite state machine | New in roster |
| session-migrate.sh | hooks/lib/ | Session migration utilities | New in roster |
| rite-context-loader.sh | hooks/lib/ | Load rite context | New in roster |

**Migration Recommendation**:
- KEEP all roster lib files (newer infrastructure)
- EVALUATE artifact-validation.sh and handoff-validator.sh for migration (unique to skeleton)

---

## 5. Skills

**Location**: `.claude/skills/`

### Skills in Both (7 common)

| Resource | Purpose | Action |
|----------|---------|--------|
| 10x-ref | Quick switch to dev workflow | Compare versions |
| 10x-workflow | Development workflow reference | Compare versions |
| architect-ref | Architect skill reference | Compare versions |
| atuin-desktop | Atuin desktop integration | Compare versions |
| build-ref | Build reference | Compare versions |
| doc-artifacts | PRD/TDD/ADR templates | Compare versions |
| justfile | Task runner automation | Compare versions |

### Skills Unique to Skeleton (11 - MUST MIGRATE)

| Resource | Skeleton Path | Purpose | Action |
|----------|---------------|---------|--------|
| commit-ref | skills/commit-ref/ | AI-assisted commits with session tracking | **MIGRATE** |
| documentation | skills/documentation/ | Doc standards routing hub | **MIGRATE** |
| hotfix-ref | skills/hotfix-ref/ | Rapid fix workflow | **MIGRATE** |
| pr-ref | skills/pr-ref/ | PR creation workflow | **MIGRATE** |
| qa-ref | skills/qa-ref/ | QA validation workflow | **MIGRATE** |
| review | skills/review/ | Code review workflow | **MIGRATE** |
| spike-ref | skills/spike-ref/ | Time-boxed research reference | **MIGRATE** |
| sprint-ref | skills/sprint-ref/ | Multi-task sprint orchestration | **MIGRATE** |
| state-mate | skills/state-mate/ | Centralized state mutation | **MIGRATE** |
| task-ref | skills/task-ref/ | Full lifecycle task execution | **MIGRATE** |
| worktree-ref | skills/worktree-ref/ | Git worktree management | **MIGRATE** |

### Skills Unique to Roster (17 - already present)

| Resource | Purpose | Notes |
|----------|---------|-------|
| consult-ref | Ecosystem guidance | Already in roster |
| cross-rite | Cross-team routing | Already in roster |
| cross-rite-handoff | Cross-team handoffs | Already in roster |
| file-verification | Anti-hallucination verification | Already in roster |
| handoff-ref | Agent handoff reference | Already in roster |
| initiative-scoping | Session -1/0 protocols | Already in roster |
| orchestration | Workflow coordination | Already in roster |
| park-ref | Session parking | Already in roster |
| prompting | Agent invocation patterns | Already in roster |
| resume | Session resumption | Already in roster |
| session-lifecycle | Session lifecycle management | Already in roster |
| shared-templates | Multi-team templates | Already in roster |
| smell-detection | Code smell detection | Already in roster |
| standards | Code conventions | Already in roster |
| start-ref | Session start reference | Already in roster |
| team-ref | Team switching reference | Already in roster |
| wrap-ref | Session completion | Already in roster |

**Migration Recommendation**: MIGRATE all 11 skeleton-unique skills.

---

## 6. User-Agents

**Location**: `.claude/user-agents/`

Skeleton has 7 user-agents; roster has none.

| Resource | Skeleton Path | Purpose | Action |
|----------|---------------|---------|--------|
| agent-curator.md | user-agents/agent-curator.md | Integration specialist for team deployment | **MIGRATE** |
| agent-designer.md | user-agents/agent-designer.md | Team design from use cases | **MIGRATE** |
| consultant.md | user-agents/consultant.md | Ecosystem navigation guidance | **MIGRATE** |
| eval-specialist.md | user-agents/eval-specialist.md | Agent evaluation testing | **MIGRATE** |
| platform-engineer.md | user-agents/platform-engineer.md | Infrastructure/platform work | **MIGRATE** |
| prompt-architect.md | user-agents/prompt-architect.md | Agent prompt creation | **MIGRATE** |
| workflow-engineer.md | user-agents/workflow-engineer.md | Workflow wiring and orchestration | **MIGRATE** |

**Migration Recommendation**: MIGRATE all 7 user-agents (Forge team for meta-operations).

---

## 7. User-Commands

**Location**: `.claude/user-commands/`

Skeleton has 38 user-commands; roster has none.

| Resource | Purpose | Category | Action |
|----------|---------|----------|--------|
| 10x.md | Switch to 10x dev pack | Team Switching | **MIGRATE** |
| architect.md | Design-only session | Workflow | **MIGRATE** |
| build.md | Build project | Development | **MIGRATE** |
| cem-debug.md | CEM debugging | Debug | **MIGRATE** |
| code-review.md | Code review | Development | **MIGRATE** |
| commit.md | AI-assisted commit | Development | **MIGRATE** |
| consolidate.md | Consolidate sessions | Session Mgmt | **MIGRATE** |
| consult.md | Ecosystem guidance | Navigation | **MIGRATE** |
| continue.md | Continue session | Session Mgmt | **MIGRATE** |
| debt.md | Tech debt | Specialized | **MIGRATE** |
| docs.md | Documentation | Specialized | **MIGRATE** |
| ecosystem.md | Ecosystem info | Navigation | **MIGRATE** |
| eval-agent.md | Evaluate agent | Forge | **MIGRATE** |
| forge.md | Team factory | Forge | **MIGRATE** |
| handoff.md | Agent handoff | Workflow | **MIGRATE** |
| hotfix.md | Rapid fix | Development | **MIGRATE** |
| hygiene.md | Code hygiene | Specialized | **MIGRATE** |
| intelligence.md | A/B testing | Specialized | **MIGRATE** |
| new-team.md | Create new team | Forge | **MIGRATE** |
| park.md | Park session | Session Mgmt | **MIGRATE** |
| pr.md | Create PR | Development | **MIGRATE** |
| qa.md | QA validation | Development | **MIGRATE** |
| rnd.md | R&D exploration | Specialized | **MIGRATE** |
| security.md | Security review | Specialized | **MIGRATE** |
| sessions.md | Session management | Session Mgmt | **MIGRATE** |
| spike.md | Research spike | Development | **MIGRATE** |
| sprint.md | Sprint orchestration | Workflow | **MIGRATE** |
| sre.md | SRE operations | Specialized | **MIGRATE** |
| start.md | Start session | Session Mgmt | **MIGRATE** |
| strategy.md | Strategic planning | Specialized | **MIGRATE** |
| sync.md | Sync state | Session Mgmt | **MIGRATE** |
| task.md | Single task lifecycle | Development | **MIGRATE** |
| team.md | Team management | Team Switching | **MIGRATE** |
| validate-team.md | Validate rite | Forge | **MIGRATE** |
| worktree.md | Git worktree | Development | **MIGRATE** |
| wrap.md | Wrap session | Session Mgmt | **MIGRATE** |

**Migration Recommendation**: MIGRATE all 38 user-commands.

---

## 8. User-Skills

**Location**: `.claude/user-skills/`

| Resource | Skeleton Path | Purpose | Action |
|----------|---------------|---------|--------|
| consult-ref | user-skills/consult-ref/ | Consultant reference documentation | **MIGRATE** (may overlap with roster consult-ref) |
| forge-ref | user-skills/forge-ref/ | Forge team reference documentation | **MIGRATE** |

**Migration Recommendation**: MIGRATE both user-skills.

---

## 9. Schemas

**Location**: `.claude/schemas/`

| Resource | Skeleton Path | Purpose | Action |
|----------|---------------|---------|--------|
| workflow.schema.json | schemas/workflow.schema.json | JSON schema for workflow.yaml validation | **EVALUATE** - useful for validation |

**Migration Recommendation**: EVALUATE for migration.

---

## 10. Other Resources

### Skeleton-Only Directories

| Directory | Purpose | Action |
|-----------|---------|--------|
| .claude/forge/ | Team factory resources | Empty in skeleton |
| .claude/tests/ | Hook tests | Has 1 file (verify-commit-attribution.sh) - **EVALUATE** |
| .claude/designs/ | Design documents | Internal to skeleton |
| .claude/audits/ | Audit logs | Internal to skeleton |

### Configuration Files

| File | Skeleton | Roster | Action |
|------|----------|--------|--------|
| ACTIVE_RITE | Present | Present | Project-specific |
| ACTIVE_WORKFLOW.yaml | Present | Present | Project-specific |
| AGENT_MANIFEST.json | Present | Present | Project-specific |
| CLAUDE.md | Present | Present | Project-specific |
| settings.local.json | Present | Present | Project-specific |
| .shared-skills | N/A | Present | Roster-only (shared skills feature) |

---

## Migration Priority Summary

### Critical (Must Migrate)

1. **Skills (11)**: commit-ref, documentation, hotfix-ref, pr-ref, qa-ref, review, spike-ref, sprint-ref, state-mate, task-ref, worktree-ref
2. **User-Commands (38)**: All session, workflow, and specialized commands
3. **User-Agents (7)**: Forge team agents for meta-operations
4. **User-Skills (2)**: consult-ref, forge-ref

### Evaluate (May Be Useful)

1. **Hooks**: team-validator.sh, workflow-validator.sh
2. **Lib**: artifact-validation.sh, handoff-validator.sh
3. **Schemas**: workflow.schema.json
4. **Tests**: verify-commit-attribution.sh

### Deprecate (Roster Superior)

1. **Agents (5)**: All skeleton agents (roster versions enhanced)
2. **Commands (2)**: All skeleton commands (identical to roster)
3. **Hooks (11)**: All main hooks (roster versions reorganized + optimized)
4. **Lib (10)**: Common lib files (roster versions newer)

---

## File Counts

```
Skeleton .claude/ Contents:
- agents/: 5 files
- commands/: 2 files + 1 marker
- hooks/: 11 scripts + lib/ (10 files) + tests/ (1 file)
- skills/: 18 directories
- user-agents/: 7 files
- user-commands/: 38 files
- user-skills/: 2 directories
- schemas/: 1 file

Roster .claude/ Contents:
- agents/: 5 files
- commands/: 2 files + 1 marker
- hooks/: 12 scripts + lib/ (13 files) + subdirs (4)
- skills/: 25 directories
- (no user-* directories)
```

---

## Next Steps

1. **Phase 1**: Migrate 11 skeleton-unique skills to roster
2. **Phase 2**: Migrate 38 user-commands to roster (or convert to shared skills)
3. **Phase 3**: Migrate 7 user-agents to roster
4. **Phase 4**: Evaluate and selectively migrate hooks/lib/schemas
5. **Phase 5**: Archive skeleton_claude once migration complete

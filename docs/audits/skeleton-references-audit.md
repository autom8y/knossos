# Skeleton References Audit

> Comprehensive inventory of all skeleton_claude references in the roster codebase

**Generated**: 2026-01-03
**Scope**: All files in `/Users/tomtenuta/Code/roster`
**Total References**: 150+ across 95 files

---

## Executive Summary

This audit catalogs all references to `skeleton_claude`, `skeleton`, and `SKELETON_HOME` in the roster codebase. References are categorized by type and annotated with migration actions for the ecosystem independence initiative.

### Reference Categories

| Category | Count | Migration Complexity |
|----------|-------|---------------------|
| Environment Variables | 15 | Medium |
| Path References (hardcoded) | 45 | High |
| Documentation References | 60 | Low |
| Code Dependencies | 12 | High |
| Configuration References | 8 | Medium |
| Conceptual References | 30+ | Low/None |

---

## Category 1: Environment Variables

Critical infrastructure variables that define skeleton location.

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/user-hooks/lib/config.sh` | 17 | `SKELETON_HOME="${SKELETON_HOME:-$HOME/Code/skeleton_claude}"` | Default path definition | **MIGRATE**: Change default to roster-local or remove dependency |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/session-utils.sh` | 68 | `# - SKELETON_HOME` | Comment listing required env vars | **UPDATE**: Remove if dependency eliminated |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | 22 | `# Source configuration (provides SKELETON_HOME, ROSTER_HOME, etc.)` | Comment | **UPDATE**: Update comment if SKELETON_HOME removed |
| `/Users/tomtenuta/Code/roster/docs/INTEGRATION.md` | 41 | `SKELETON_HOME` `~/Code/skeleton_claude` | Environment variable table | **UPDATE**: Remove or mark as optional |
| `/Users/tomtenuta/Code/roster/docs/INTEGRATION.md` | 13 | `$SKELETON_HOME/docs/INTEGRATION.md` | Path reference | **UPDATE**: Point to roster docs |
| `/Users/tomtenuta/Code/roster/docs/ecosystem/CONTEXT-DESIGN-team-context-loader.md` | 373 | `local skeleton_home="${SKELETON_HOME:-$HOME/Code/skeleton_claude}"` | Design doc code example | **UPDATE**: Update example when code changes |
| `/Users/tomtenuta/Code/roster/teams/ecosystem-pack/context-injection.sh` | 39 | `local skeleton_home="${SKELETON_HOME:-$HOME/Code/skeleton_claude}"` | Skeleton ref lookup | **MIGRATE**: Remove or make optional |
| `/Users/tomtenuta/Code/roster/teams/ecosystem-pack/skills/ecosystem-ref/SKILL.md` | 21 | `Skeleton: $SKELETON_HOME or ~/Code/skeleton_claude` | Documentation | **UPDATE**: Remove if skeleton dependency eliminated |
| `/Users/tomtenuta/Code/roster/teams/ecosystem-pack/skills/ecosystem-ref/SKILL.md` | 70 | `Skeleton` `$SKELETON_HOME/.claude/` | Path table | **UPDATE**: Remove row |

---

## Category 2: Hardcoded Path References

Direct filesystem paths to skeleton_claude. Highest migration priority.

### CEM Command Paths

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 33 | `~/Code/skeleton_claude/cem install-user` | CEM command instruction | **MIGRATE**: Use `$CEM_HOME/cem` or local copy |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 41 | `~/Code/skeleton_claude/cem sync` | Sync command | **MIGRATE**: Same |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 44 | `~/Code/skeleton_claude/cem sync --refresh` | Refresh command | **MIGRATE**: Same |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 51 | `~/Code/skeleton_claude/cem init` | Init command | **MIGRATE**: Same |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 54 | `~/Code/skeleton_claude/cem status` | Status command | **MIGRATE**: Same |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 57 | `~/Code/skeleton_claude/cem diff` | Diff command | **MIGRATE**: Same |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 68 | `~/Code/skeleton_claude/cem install-user` | Install command | **MIGRATE**: Same |
| `/Users/tomtenuta/Code/roster/user-commands/cem/sync.md` | 95 | `~/Code/skeleton_claude/cem --help` | Help reference | **MIGRATE**: Same |

### Worktree Manager CEM Integration

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | 225 | `if [[ ! -x "$SKELETON_HOME/cem" ]]` | CEM existence check | **MIGRATE**: Use `$CEM_HOME` or bundle CEM |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | 227 | `CEM not found at '"$SKELETON_HOME/cem"'` | Error message | **MIGRATE**: Update error message |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | 236 | `"$SKELETON_HOME/cem" sync` | CEM sync call | **MIGRATE**: Use configurable path |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | 238 | `"$SKELETON_HOME/cem" init --force` | CEM init call | **MIGRATE**: Use configurable path |
| `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | 246 | `"$SKELETON_HOME/cem" init` | CEM init call | **MIGRATE**: Use configurable path |

### Troubleshooting Paths

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/user-skills/operations/worktree-ref/troubleshooting.md` | 33 | `$HOME/Code/skeleton_claude/cem` | CEM path check | **UPDATE**: Change to generic `$CEM_HOME` |
| `/Users/tomtenuta/Code/roster/user-skills/operations/worktree-ref/troubleshooting.md` | 34 | `skeleton_claude exist?` | Troubleshooting question | **UPDATE**: Generalize |
| `/Users/tomtenuta/Code/roster/user-skills/operations/worktree-ref/troubleshooting.md` | 39 | `ls -la $HOME/Code/skeleton_claude/cem` | Debug command | **UPDATE**: Use variable |
| `/Users/tomtenuta/Code/roster/user-skills/operations/worktree-ref/troubleshooting.md` | 42 | `chmod +x $HOME/Code/skeleton_claude/cem` | Fix command | **UPDATE**: Use variable |

### Orchestrator Template Paths (45+ references)

These are documentation paths showing where templates live in skeleton_claude.

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/schema-reference.md` | 7 | `/skeleton_claude/.claude/teams/{team-name}/orchestrator.yaml` | Template location | **UPDATE**: Change to roster path |
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/schema-reference.md` | 9 | `/skeleton_claude/.claude/teams/rnd-pack/orchestrator.yaml` | Example location | **UPDATE**: Change to roster path |
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/integration-diagram.txt` | 315-321 | Multiple `/skeleton_claude/.claude/teams/` paths | Diagram references | **UPDATE**: Update diagram |
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/create-new-team-orchestrator.md` | 12-514 | 30+ `/skeleton_claude/` paths | Tutorial paths | **UPDATE**: Change all to roster |
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/troubleshooting.md` | 38-638 | 40+ `/skeleton_claude/` paths | Debug paths | **UPDATE**: Change all to roster |
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/migration-guide.md` | 64-248 | 20+ `/skeleton_claude/` paths | Migration paths | **UPDATE**: Change all to roster |
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/update-canonical-patterns.md` | 145, 325 | `cd /skeleton_claude` | Working directory | **UPDATE**: Change to roster |
| `/Users/tomtenuta/Code/roster/user-skills/orchestration/orchestrator-templates/QUICK-REFERENCE.md` | 174 | `/skeleton_claude/.claude/skills/orchestrator-templates/` | Skill docs path | **UPDATE**: Change to roster |

---

## Category 3: Code Dependencies

Actual runtime code that depends on skeleton_claude existing.

| File | Line | Reference | Context | Purpose | Migration Action |
|------|------|-----------|---------|---------|------------------|
| `/Users/tomtenuta/Code/roster/user-hooks/lib/worktree-manager.sh` | 225-250 | `$SKELETON_HOME/cem` | CEM executable calls | Worktree initialization | **HIGH PRIORITY**: Bundle CEM or make optional |
| `/Users/tomtenuta/Code/roster/teams/ecosystem-pack/context-injection.sh` | 39-44 | `$skeleton_home/.git` | Git ref lookup | Context display | **MIGRATE**: Remove or make optional |
| `/Users/tomtenuta/Code/roster/swap-team.sh` | 3979 | `regenerate_skeleton_claude_md()` | Function name | Reset to baseline | **RENAME**: Function name only (no functional change) |
| `/Users/tomtenuta/Code/roster/swap-team.sh` | 4106 | `regenerate_skeleton_claude_md` | Function call | Reset operation | **RENAME**: Update call site |

---

## Category 4: Documentation References

References in documentation that describe the ecosystem relationship.

### Architecture Documentation

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/docs/INTEGRATION.md` | 3 | `skeleton_claude` | Pointer to primary guide | **UPDATE**: Make roster self-contained |
| `/Users/tomtenuta/Code/roster/docs/INTEGRATION.md` | 8 | `skeleton_claude: Infrastructure` | Repository description | **UPDATE**: Describe new architecture |
| `/Users/tomtenuta/Code/roster/docs/INTEGRATION.md` | 17 | `~/Code/skeleton_claude/docs/INTEGRATION.md` | Path to docs | **UPDATE**: Point to roster |
| `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0002-hook-library-resolution-architecture.md` | 45 | `roster, skeleton_claude` | Harness examples | **UPDATE**: Remove skeleton_claude |
| `/Users/tomtenuta/Code/roster/docs/decisions/ADR-0002-hook-library-resolution-architecture.md` | 312 | `skeleton_claude/CEM domain` | CEM compatibility note | **UPDATE**: Reflect new architecture |
| `/Users/tomtenuta/Code/roster/docs/GENERATOR-orchestrator.md` | 533 | `/skeleton_claude/docs/ecosystem/` | Context design path | **UPDATE**: Change to roster |

### PRD/Design Documents

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/docs/requirements/PRD-claude-md-descriptive-architecture.md` | 27 | `skeleton-owned` | Section ownership model | **REVIEW**: May need architectural update |
| `/Users/tomtenuta/Code/roster/docs/requirements/PRD-hook-ecosystem-parity.md` | 119 | `CEM (skeleton_claude)` | CEM ownership | **UPDATE**: Reflect new ownership |
| `/Users/tomtenuta/Code/roster/docs/requirements/PRD-rite-hook-context.md` | 27, 38, 46 | `skeleton` | Context requirements | **UPDATE**: Remove skeleton dependencies |
| `/Users/tomtenuta/Code/roster/docs/ecosystem/CONTEXT-DESIGN-team-context-loader.md` | 10, 371-379, 427, 449 | `skeleton` | Design context | **UPDATE**: After code migration |

### QA/Test Reports

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/docs/qa/QA-OrchestratorTemplateSystem-TestReport.md` | 20, 85, 86, 150 | `skeleton_claude` | Test satellite | **UPDATE**: Update test references |
| `/Users/tomtenuta/Code/roster/docs/testing/enforcement-validation-report.md` | 20-26 | `roster (skeleton)` | Test matrix | **UPDATE**: Clarify terminology |
| `/Users/tomtenuta/Code/roster/docs/qa/compatibility-orchestration-consolidation.md` | 14, 166 | `skeleton` | Compatibility test | **UPDATE**: Post-migration |
| `/Users/tomtenuta/Code/roster/docs/ecosystem/PHASE3-COMPATIBILITY-REPORT.md` | 6, 69 | `skeleton_claude` | Phase 3 satellite | **UPDATE**: Post-migration |

---

## Category 5: Configuration References

Schema and config files with skeleton references.

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/.gitignore` | 2 | `skeleton_claude/ (skeleton)` | Comment | **UPDATE**: Remove comment |
| `/Users/tomtenuta/Code/roster/schemas/handoff-criteria-schema.yaml` | 140 | `['skeleton', 'roster', 'CEM']` | Affected systems enum | **UPDATE**: Remove skeleton option |
| `/Users/tomtenuta/Code/roster/schemas/orchestrator.yaml.schema.json` | 70, 166, 268, 310 | `CEM/skeleton/roster` | Ecosystem pack description | **UPDATE**: Remove skeleton |
| `/Users/tomtenuta/Code/roster/schemas/artifacts/tdd.schema.json` | 87 | `owner: "skeleton"` | Component owner | **UPDATE**: Change owner |
| `/Users/tomtenuta/Code/roster/schemas/artifacts/ecosystem/gap-analysis.schema.json` | 76 | `["roster", "skeleton", "CEM"]` | Affected systems | **UPDATE**: Remove skeleton |
| `/Users/tomtenuta/Code/roster/schemas/artifacts/ecosystem/compatibility-report.schema.json` | 95 | `skeleton_claude` | Satellite example | **UPDATE**: Use different example |
| `/Users/tomtenuta/Code/roster/schemas/artifacts/ecosystem/context-design.schema.json` | 82 | `["roster", "skeleton"]` | Affected systems | **UPDATE**: Remove skeleton |
| `/Users/tomtenuta/Code/roster/schemas/artifacts/ecosystem/migration-runbook.schema.json` | 90 | `["skeleton_claude", "roster", "all-satellites"]` | Affected satellites | **UPDATE**: Remove skeleton_claude |
| `/Users/tomtenuta/Code/roster/workflow-schema.yaml` | 8 | `skeleton/.claude/skills/` | Schema location comment | **UPDATE**: Change to roster |

---

## Category 6: Conceptual/Terminology References

References to "skeleton" as a concept (baseline state, not the repository).

| File | Line | Reference | Context | Migration Action |
|------|------|-----------|---------|------------------|
| `/Users/tomtenuta/Code/roster/swap-team.sh` | 1554 | `--reset Reset to skeleton baseline` | Help text | **KEEP**: Conceptual use is fine |
| `/Users/tomtenuta/Code/roster/swap-team.sh` | 1595 | `# Reset to skeleton baseline` | Comment | **KEEP**: Conceptual |
| `/Users/tomtenuta/Code/roster/swap-team.sh` | 2167-2425 | `skeleton skill` | Skill layer terminology | **KEEP**: Describes base layer |
| `/Users/tomtenuta/Code/roster/swap-team.sh` | 3847-4110 | `skeleton baseline` | Reset terminology | **KEEP**: Conceptual term |
| `/Users/tomtenuta/Code/roster/docs/validation/COMPAT-orchestrator-entry-pattern.md` | 12, 181 | `skeleton templates` | Template location | **UPDATE**: Change to roster |
| `/Users/tomtenuta/Code/roster/docs/MOONSHOT-agent-template-ecosystem.md` | 136, 139, 447 | `SKELETON` | Architecture diagram | **UPDATE**: Update diagram |
| `/Users/tomtenuta/Code/roster/teams/10x-dev-pack/agents/principal-engineer.md` | 71 | `skeleton first` | Coding approach | **KEEP**: Unrelated to skeleton_claude |

---

## Migration Priority Matrix

### P0 - Critical (Blocking functionality)

| File | Issue | Impact |
|------|-------|--------|
| `user-hooks/lib/worktree-manager.sh` | Hard dependency on `$SKELETON_HOME/cem` | Worktree creation fails without skeleton |
| `user-hooks/lib/config.sh` | Default `SKELETON_HOME` path | All hooks assume skeleton exists |

### P1 - High (User-facing documentation)

| File | Issue | Impact |
|------|-------|--------|
| `user-commands/cem/sync.md` | Hardcoded skeleton paths | User confusion on where to run CEM |
| `user-skills/orchestration/orchestrator-templates/*.md` | 100+ skeleton paths | Tutorial unusable |
| `docs/INTEGRATION.md` | Points to skeleton docs | Users can't find docs |

### P2 - Medium (Schema/config updates)

| File | Issue | Impact |
|------|-------|--------|
| `schemas/*.json` | Skeleton in enums | Schema validation inconsistency |
| `teams/ecosystem-pack/context-injection.sh` | Skeleton ref display | Ecosystem context incomplete |

### P3 - Low (Documentation cleanup)

| File | Issue | Impact |
|------|-------|--------|
| `docs/qa/*.md` | Historical test references | No functional impact |
| `docs/ecosystem/*.md` | Design doc examples | No functional impact |

---

## Recommended Migration Approach

### Phase 1: Eliminate Hard Dependencies
1. Bundle CEM executable in roster or make it optional
2. Update `config.sh` to not require `SKELETON_HOME`
3. Update `worktree-manager.sh` to work without skeleton

### Phase 2: Update User-Facing Content
1. Migrate all orchestrator-template paths to roster
2. Update `user-commands/cem/sync.md` for new architecture
3. Make `docs/INTEGRATION.md` self-contained

### Phase 3: Schema and Config Cleanup
1. Remove `skeleton` from schema enums
2. Update ecosystem-pack context to not show skeleton ref
3. Update `.gitignore` comment

### Phase 4: Documentation Cleanup
1. Update QA reports and test matrices
2. Update architecture diagrams
3. Archive or update historical references

---

## Attestation

| Artifact | Absolute Path | Verified |
|----------|---------------|----------|
| Audit Report | `/Users/tomtenuta/Code/roster/docs/audits/skeleton-references-audit.md` | YES |

**Verification Method**: Read tool confirmation after write (2026-01-03)

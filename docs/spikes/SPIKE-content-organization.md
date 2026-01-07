# SPIKE: Content Organization

**Date**: 2026-01-07
**Initiative**: knossos-finalization
**Phase**: R&D (research only)
**Session**: session-20260107-164631-8dd6f03a
**Upstream**: SPIKE-knossos-consolidation-architecture.md (Section 8.2)

---

## Executive Summary

This SPIKE analyzes how rites, skills, and agents interconnect across the roster codebase. The analysis reveals a multi-layered content architecture with five distinct content locations, complex cross-referencing patterns via `@skill-name` syntax, and an existing shared-skills mechanism that enables skill reuse across rites. The target structure (`rites/{name}/skills/`, `rites/{name}/agents/`) is validated as sound, with recommendations for manifest schema design and migration approach.

**Key Finding**: The current architecture already implements the target pattern in `rites/` but maintains a parallel "active" copy in `.claude/` that is populated via `swap-rite.sh`. The consolidation should preserve this materialization model while making the source-of-truth locations canonical.

---

## 1. Current Content Structure Audit

### 1.1 Content Location Inventory

| Location | Purpose | File Count (approx) | Status |
|----------|---------|---------------------|--------|
| `.claude/skills/` | Active project skills (materialized) | 30+ directories | Runtime copy |
| `.claude/agents/` | Active rite agents (materialized) | 6 files | Runtime copy |
| `rites/*/skills/` | Rite-specific skills (source) | ~20 skill sets | Canonical source |
| `rites/*/agents/` | Rite-specific agents (source) | ~60 agents | Canonical source |
| `rites/shared/skills/` | Cross-rite shared skills | 4 skill sets | Special case |
| `user-skills/` | User-level skill overrides | 30+ directories | Override layer |
| `user-agents/` | User-level agents (Moirai) | 7 files | Override layer |

### 1.2 Rite Inventory

```
rites/
+-- 10x-dev/       # Full development lifecycle
+-- debt-triage/   # Technical debt management
+-- docs/          # Documentation workflows
+-- ecosystem/     # Build and ecosystem maintenance
+-- forge/         # Agent/rite development
+-- hygiene/       # Code cleanup and refactoring
+-- intelligence/  # Analytics and experimentation
+-- rnd/           # Research and development (ACTIVE)
+-- security/      # Security review workflows
+-- shared/        # Cross-rite primitives (special)
+-- sre/           # Site reliability engineering
+-- strategy/      # Business strategy and planning
```

**Notable**: `rites/shared/` is not a rite but a skill-only infrastructure provider.

### 1.3 Skill Naming Patterns

| Pattern | Example | Meaning |
|---------|---------|---------|
| `{skill-name}` | `prompting` | General skill |
| `{skill-name}-ref` | `commit-ref` | Reference/command skill |
| `doc-{domain}` | `doc-rnd`, `doc-sre` | Domain documentation templates |
| `{domain}-ref` | `rnd-ref`, `forge-ref` | Domain reference skill |

### 1.4 Agent Naming Patterns

| Pattern | Example | Produces |
|---------|---------|----------|
| `orchestrator` | Every rite has one | Work breakdown |
| `{role}-{noun}` | `technology-scout`, `debt-collector` | Domain artifact |
| `{adjective}-{noun}` | `principal-engineer`, `qa-adversary` | Domain artifact |

---

## 2. Dependency Mapping

### 2.1 Skill Reference Syntax

Agents reference skills using `@skill-name` syntax:

```markdown
Produce Integration Map using `@doc-rnd#integration-map-template`.
```

**Pattern breakdown**:
- `@skill-name` - Reference to skill
- `#anchor` - Deep link to section within skill
- `@skill-name#anchor-template` - Common pattern for template references

### 2.2 Cross-Rite Skill Dependencies

| Rite | Skills Referenced | Notes |
|------|-------------------|-------|
| rnd | `@doc-rnd`, `@standards`, `@cross-rite` | Rite-specific + shared |
| 10x-dev | `@doc-artifacts`, `@standards`, `@cross-rite` | Rite-specific + shared |
| hygiene | `@doc-ecosystem`, `@standards`, `@cross-rite` | Uses ecosystem templates |
| security | `@doc-security`, `@standards` | Rite-specific |
| sre | `@doc-sre`, `@standards` | Rite-specific |
| docs | `@doc-reviews`, `@standards` | Rite-specific |
| strategy | `@doc-strategy`, `@standards` | Rite-specific |
| intelligence | `@doc-intelligence`, `@doc-sre` | Cross-domain (tracking) |
| debt-triage | `@shared-templates`, `@standards`, `@cross-rite` | Uses shared templates |

### 2.3 Shared Skills Usage

From `rites/shared/README.md`:

```
rites/shared/
+-- skills/
    +-- cross-rite-handoff/    # HANDOFF artifact schema
    +-- shared-templates/      # Debt ledger, risk matrix, sprint package
    +-- smell-detection/       # Code smell taxonomy
```

**Sync Behavior** (from swap-rite.sh):
1. Shared skills sync to `.claude/skills/` alongside rite-specific skills
2. Rite-specific skills take precedence (override shared)
3. No merge or conflict resolution occurs

### 2.4 User-Level Override Layer

`user-skills/` provides user-level overrides organized by category:

```
user-skills/
+-- documentation/     # doc-artifacts, standards, etc.
+-- guidance/          # prompting, cross-rite, rite-ref
+-- session-lifecycle/ # wrap-ref, park-ref, start-ref, etc.
+-- operations/        # pr-ref, qa-ref, commit-ref, etc.
+-- orchestration/     # orchestrator-templates, task-ref, etc.
```

`user-agents/` provides session lifecycle agents:
- `moirai.md` (composite)
- `clotho.md`, `lachesis.md`, `atropos.md` (the Fates)
- `consultant.md`, `context-engineer.md`

### 2.5 Hidden Dependencies (Not Documented)

| Dependency | Impact | Discovery Method |
|------------|--------|------------------|
| Agent frontmatter references skill | Agent won't work without skill | Grep for `@` in agents |
| Skill cross-references other skills | Broken links if skill missing | Grep for `@` in skills |
| Session lifecycle agents | Moirai required for state mutations | CLAUDE.md state management section |
| Orchestrator templates | All orchestrators reference same schemas | Grep for `orchestrator-templates` |
| Standards skill | Nearly all agents reference it | Grep for `@standards` |

---

## 3. Target Structure Validation

### 3.1 Proposed Structure (from upstream SPIKE)

```
rites/{rite-name}/
+-- skills/         # Rite-specific skills
+-- agents/         # Rite-specific agents
+-- manifest.yaml   # Rite metadata and configuration
+-- README.md       # Rite documentation
+-- workflow.yaml   # (optional) Workflow definition
```

### 3.2 Validation Against Requirements

| Requirement | Current | Target | Valid? |
|-------------|---------|--------|--------|
| Rite-centric organization | Yes (rites/*/) | Yes | PASS |
| Self-contained rites | Mostly (some cross-refs) | Yes with resolution | PASS |
| Shared skill support | Yes (rites/shared/) | Yes | PASS |
| User-level overrides | Yes (user-skills/) | Yes | PASS |
| `.claude/` generated | Partially (swap-rite.sh) | Fully generated | ENHANCEMENT |

### 3.3 Does Target Support Shared Skills?

**Yes.** The `rites/shared/` pattern continues to work:

1. `rites/shared/skills/` contains cross-rite skills
2. `sync` command copies shared skills to `.claude/skills/`
3. Rite-specific skills override shared skills (precedence rules)

**Recommendation**: Formalize in manifest schema:

```yaml
# rites/shared/manifest.yaml
type: shared  # Special type, not a rite
skills:
  - cross-rite-handoff
  - shared-templates
  - smell-detection
# No agents section - shared is skills-only
```

### 3.4 Does Target Support User-Level Overrides?

**Yes, with clarification needed.**

Current model:
- `user-skills/` and `user-agents/` exist at repo root
- These are synced to `~/.claude/skills/` and `~/.claude/agents/` for user-level availability

Target model options:

**Option A: Keep separate user-* directories**
```
roster/
+-- rites/
+-- user-skills/      # User-level overrides
+-- user-agents/      # User-level agents
```
- Pros: Clear separation, existing pattern
- Cons: Duplication of structure

**Option B: user/ as special rite**
```
roster/
+-- rites/
    +-- user/
        +-- skills/
        +-- agents/
```
- Pros: Consistent with rite pattern
- Cons: user isn't really a rite

**Recommendation**: Option A - keep `user-skills/` and `user-agents/` separate since they serve a fundamentally different purpose (user-level defaults, not rite-specific content).

---

## 4. Manifest Schema Design

### 4.1 Current State

From `TDD-ariadne-manifest.md`, a team manifest schema exists:

```yaml
version: "1.0"
name: 10x-dev
description: Full development lifecycle
workflow:
  type: sequential
  entry_point: requirements-analyst
agents:
  - name: architect
    file: agents/architect.md
    role: Evaluates tradeoffs
    produces: TDD
skills:
  - commit-ref
  - pr-ref
hooks:
  - session-guards/auto-park.sh
```

### 4.2 Proposed Enhancements

```yaml
# rites/{name}/manifest.yaml
version: "1.1"
type: rite              # rite | shared
name: rnd
description: Technology scouting and research

# Skill configuration
skills:
  include:              # Skills to include
    - doc-rnd
    - rnd-ref
  shared:               # Shared skills to import
    - cross-rite-handoff
    - smell-detection
  exclude: []           # Explicit exclusions

# Agent configuration
agents:
  - name: technology-scout
    file: agents/technology-scout.md
    role: Technology horizon scanning
    produces: tech-assessment
    depends_on: []      # Agent dependencies
  - name: integration-researcher
    file: agents/integration-researcher.md
    role: Integration analysis
    produces: integration-map
    depends_on:
      - technology-scout  # Receives from scout

# Workflow configuration
workflow:
  type: sequential      # sequential | parallel | hybrid
  entry_point: technology-scout
  phases:
    - scouting
    - integration-analysis
    - prototyping
    - future-architecture

# Cross-rite handoff configuration
handoffs:
  outgoing:
    - target: 10x-dev
      trigger: "prototype validated"
      artifact: TRANSFER
  incoming:
    - source: hygiene
      receives: smell-report

# Complexity levels
complexity:
  levels:
    - name: SPIKE
      description: Quick feasibility check
      phases: [scouting]
    - name: EVALUATION
      description: Full technology evaluation
      phases: [scouting, integration-analysis, prototyping]
    - name: MOONSHOT
      description: Paradigm shift exploration
      phases: [scouting, integration-analysis, prototyping, future-architecture]
```

### 4.3 Shared Manifest Schema

```yaml
# rites/shared/manifest.yaml
version: "1.0"
type: shared
name: shared
description: Cross-rite primitives

skills:
  - name: cross-rite-handoff
    path: skills/cross-rite-handoff/
    description: HANDOFF artifact schema
  - name: shared-templates
    path: skills/shared-templates/
    description: Debt ledger, risk matrix templates
  - name: smell-detection
    path: skills/smell-detection/
    description: Code smell taxonomy

# No agents section - shared is skills-only
# No workflow section - not a workflow provider
```

### 4.4 Manifest JSON Schema

See `TDD-ariadne-manifest.md` Section 7.2 for the existing JSON Schema. Proposed additions:

```json
{
  "properties": {
    "type": {
      "type": "string",
      "enum": ["rite", "shared"],
      "default": "rite"
    },
    "skills": {
      "type": "object",
      "properties": {
        "include": { "type": "array", "items": { "type": "string" } },
        "shared": { "type": "array", "items": { "type": "string" } },
        "exclude": { "type": "array", "items": { "type": "string" } }
      }
    },
    "handoffs": {
      "type": "object",
      "properties": {
        "outgoing": { "type": "array" },
        "incoming": { "type": "array" }
      }
    },
    "complexity": {
      "type": "object",
      "properties": {
        "levels": { "type": "array" }
      }
    }
  }
}
```

---

## 5. Migration Path

### 5.1 Current State Summary

```
CURRENT:
.claude/skills/          <- Materialized (swap-rite.sh copies here)
.claude/agents/          <- Materialized (swap-rite.sh copies here)
rites/*/skills/          <- Source of truth
rites/*/agents/          <- Source of truth
rites/shared/skills/     <- Shared source
user-skills/             <- User overrides (synced to ~/.claude/)
user-agents/             <- User agents (synced to ~/.claude/)
```

### 5.2 Target State

```
TARGET:
rites/*/skills/          <- Canonical source (unchanged)
rites/*/agents/          <- Canonical source (unchanged)
rites/*/manifest.yaml    <- NEW: Formal manifest
rites/shared/skills/     <- Shared source (unchanged)
rites/shared/manifest.yaml <- NEW: Shared manifest
user-skills/             <- User overrides (unchanged)
user-agents/             <- User agents (unchanged)
.claude/                 <- FULLY GENERATED (gitignored)
```

### 5.3 Migration Steps

#### Phase 1: Add Manifests (Non-Breaking)

1. Create `manifest.yaml` for each rite
2. Create `manifest.yaml` for shared
3. Validate manifests with `ari manifest validate`
4. No behavior change yet

**Effort**: 2-3 days
**Risk**: LOW (additive only)

#### Phase 2: Update swap-rite.sh to Use Manifests

1. Modify `swap-rite.sh` to read from manifest
2. Use manifest for agent list instead of directory scan
3. Use manifest for skill inclusion rules
4. Preserve existing sync behavior

**Effort**: 3-5 days
**Risk**: MEDIUM (behavior change)
**Rollback**: Revert to directory scan

#### Phase 3: Add Skill Reference Validation

1. Create validation for `@skill-name` references in agents
2. Warn on missing skill references
3. Fail on broken references (optional --strict mode)

**Effort**: 2-3 days
**Risk**: LOW (validation only)

#### Phase 4: Migrate to Go-based Sync

1. Replace `swap-rite.sh` sync logic with `ari sync`
2. Use manifest as authoritative source
3. Generate `.claude/` from templates + rite content

**Effort**: 5-8 days (part of larger ari migration)
**Risk**: HIGH (complete replacement)
**Rollback**: Keep swap-rite.sh as fallback

### 5.4 File Movement Mapping

| Source | Destination | Action |
|--------|-------------|--------|
| `rites/*/skills/*` | Unchanged | Keep |
| `rites/*/agents/*` | Unchanged | Keep |
| `rites/*/README.md` | Unchanged | Keep |
| `rites/*/workflow.yaml` | Unchanged | Keep |
| `.claude/skills/*` | Generated | Gitignore |
| `.claude/agents/*` | Generated | Gitignore |

**Key Insight**: Most content stays in place. The migration is about formalizing the manifest and making `.claude/` fully generated.

---

## 6. Effort Estimation

| Task | Effort | Confidence | Assumptions |
|------|--------|------------|-------------|
| Create rite manifests (12 rites) | 2-3 days | HIGH | Manual creation, no automation |
| Manifest schema design | 1 day | HIGH | Building on existing TDD |
| swap-rite.sh manifest integration | 3-5 days | MEDIUM | Depends on shell complexity |
| Skill reference validation | 2-3 days | HIGH | Grep + validation logic |
| ari sync implementation | 5-8 days | MEDIUM | Part of larger migration |
| Documentation updates | 1-2 days | HIGH | Clear scope |

**Total Estimated Effort**: 14-22 days

---

## 7. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Skill reference breakage | MEDIUM | HIGH | Validation before migration |
| swap-rite.sh complexity | HIGH | MEDIUM | Incremental refactoring |
| User-skill sync issues | LOW | MEDIUM | Test extensively |
| Manifest schema evolution | MEDIUM | LOW | Version field in schema |
| Orphaned skills after migration | LOW | LOW | Audit script |

---

## 8. Recommendations

### 8.1 Immediate Actions

1. **Create manifest.yaml for each rite** - Start with rnd (current active rite) as proof of concept
2. **Document skill reference syntax** - Add to CLAUDE.md or skills documentation
3. **Add manifest validation** - Use existing `ari manifest validate` infrastructure

### 8.2 Short-Term Actions

4. **Create skill reference linter** - Validate `@skill-name` references exist
5. **Update swap-rite.sh to read manifest** - Gradual adoption
6. **Add shared manifest** - Formalize rites/shared/

### 8.3 Medium-Term Actions

7. **Gitignore .claude/** - Make fully generated
8. **Migrate sync to ari** - Replace shell with Go
9. **Add rite discovery** - `ari rite list` reads manifests

---

## 9. Verification Attestation

| File/Directory | Verified Via | Finding |
|----------------|--------------|---------|
| `.claude/skills/` | Glob | 30+ directories, 90+ files |
| `.claude/agents/` | Glob | 6 files (rnd active) |
| `rites/*/skills/` | Glob | 12 rites, ~20 skill sets |
| `rites/*/agents/` | Glob | 12 rites, ~60 agents |
| `rites/shared/` | Read | Skills-only, no agents |
| `user-skills/` | Glob | 30+ directories |
| `user-agents/` | Glob | 7 files |
| `@skill-name` references | Grep | 50+ occurrences in agents |
| Existing manifest TDD | Read | TDD-ariadne-manifest.md |

---

## Appendix A: Skill Reference Map

Top skills referenced by agents:

| Skill | References | Rites Using |
|-------|------------|-------------|
| `@standards` | 15+ | All |
| `@cross-rite` | 10+ | All with cross-rite work |
| `@doc-rnd` | 6 | rnd |
| `@doc-ecosystem` | 5 | hygiene, ecosystem |
| `@doc-sre` | 5 | sre, intelligence |
| `@doc-artifacts` | 4 | 10x-dev |
| `@doc-security` | 4 | security |
| `@shared-templates` | 3 | debt-triage |

---

## Appendix B: Rite-to-Skill Mapping

| Rite | Primary Skills | Shared Skills |
|------|----------------|---------------|
| 10x-dev | doc-artifacts, 10x-workflow | cross-rite-handoff |
| rnd | doc-rnd, rnd-ref | cross-rite-handoff |
| hygiene | doc-ecosystem, smell-detection | cross-rite-handoff, smell-detection |
| docs | doc-reviews, doc-consolidation | - |
| sre | doc-sre | - |
| security | doc-security | - |
| strategy | doc-strategy | - |
| intelligence | doc-intelligence | - |
| debt-triage | shared-templates | shared-templates, smell-detection |
| ecosystem | doc-ecosystem, claude-md-architecture | - |
| forge | forge-ref, agent-prompt-engineering, rite-development | - |

---

**Document Status**: SPIKE COMPLETE
**Next Step**: Draft ADR-content-organization.md
**Handoff**: This document serves as input for ADR creation

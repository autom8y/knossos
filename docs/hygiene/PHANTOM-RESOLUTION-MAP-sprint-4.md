# Sprint 4 Phantom Resolution Map

**Initiative**: Knossos Code Hygiene - Phantom Reference Resolution
**Phase**: Sprint 4 - Markdown Ecosystem Integrity
**Date**: 2026-02-06
**Architect**: architect-enforcer

## Executive Summary

Systematic audit of all `@reference` patterns across 56 agent files in 12 rites reveals **22 distinct phantom references** and **4 phantom fragment anchors**. The single highest-impact finding is `@orchestrator-templates` -- referenced from every orchestrator agent (11 rites) but never created as a mena skill. The remaining phantoms cluster into two categories: (1) orchestrator-specific skill references that were never built, and (2) cross-rite template references that don't resolve in the consuming rite's materialization context.

**Recommendation**: Collapse batches 19-22 into two phases (A and B). Phase A creates `orchestrator-templates` and the missing `audit-report-template`, fixing the highest-impact phantoms. Phase B updates or removes the remaining phantom skill references in orchestrator Skills Reference sections.

---

## Methodology

### Sources of Truth

- **Global mena** (`mena/`): Skills/commands available to ALL rites
- **Rite-local mena** (`rites/<rite>/mena/`): Skills available only within that rite
- **Shared rite mena** (`rites/shared/mena/`): Skills available to rites declaring `dependencies: [shared]`
- **Materialized output** (`.claude/skills/`, `.claude/commands/`): What actually resolves at runtime

### Search Method

1. Extracted all `@reference` patterns from `rites/*/agents/*.md` (56 files)
2. Extracted all `@skill#fragment-anchor` patterns from same files (35 unique anchors)
3. Cross-referenced against all INDEX.lego.md and INDEX.dro.md files in mena/ and rites/*/mena/
4. Verified fragment anchors exist as headings with `{#anchor}` syntax in target files
5. Checked materialization output to confirm runtime resolution

### Available Skills Registry

**Global (available to all rites):**
cross-rite, file-verification, prompting, rite-discovery, standards, doc-artifacts, documentation, atuin-desktop, justfile

**Rite-local (available only within owning rite):**

| Rite | Skills |
|------|--------|
| 10x-dev | 10x-ref, 10x-workflow, architect-ref, build-ref, doc-artifacts |
| debt-triage | debt-ref |
| docs | doc-consolidation, doc-reviews, docs-ref |
| ecosystem | claude-md-architecture, doc-ecosystem, ecosystem-ref |
| forge | agent-prompt-engineering, forge-ref, rite-development |
| hygiene | hygiene-ref |
| intelligence | doc-intelligence, intelligence-ref |
| rnd | doc-rnd, rnd-ref |
| security | doc-security, security-ref |
| shared | cross-rite-handoff, shared-templates, smell-detection |
| sre | doc-sre, sre-ref |
| strategy | doc-strategy, strategy-ref |

---

## Issue Catalog

### Category A: Missing Skill -- @orchestrator-templates (CRITICAL)

This is the single most impactful phantom. Every orchestrator agent references `@orchestrator-templates/schemas/consultation-request.md` and `@orchestrator-templates/schemas/consultation-response.md`. The skill does not exist anywhere in the codebase.

#### Issue 1: @orchestrator-templates phantom in strategy/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/strategy/agents/orchestrator.md` lines 62, 68
**Reference**: `@orchestrator-templates/schemas/consultation-request.md`, `@orchestrator-templates/schemas/consultation-response.md`
**Resolution**: CREATE
**Target**: New shared legomena at `rites/shared/mena/orchestrator-templates/` with consultation schemas
**Batch**: 19

#### Issue 2: @orchestrator-templates phantom in intelligence/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/intelligence/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 3: @orchestrator-templates phantom in 10x-dev/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/10x-dev/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 4: @orchestrator-templates phantom in rnd/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/rnd/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 5: @orchestrator-templates phantom in debt-triage/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/debt-triage/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 6: @orchestrator-templates phantom in sre/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/sre/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 7: @orchestrator-templates phantom in docs/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/docs/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 8: @orchestrator-templates phantom in forge/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/forge/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 9: @orchestrator-templates phantom in hygiene/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/hygiene/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 10: @orchestrator-templates phantom in ecosystem/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/ecosystem/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

#### Issue 11: @orchestrator-templates phantom in security/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/security/agents/orchestrator.md` lines 62, 68
**Reference**: Same as Issue 1
**Resolution**: CREATE (same target as Issue 1)
**Batch**: 19

**Batch 19 Contract**: Create `rites/shared/mena/orchestrator-templates/INDEX.lego.md` with two schema sub-files: `schemas/consultation-request.md` and `schemas/consultation-response.md`. Add `orchestrator-templates` to the shared rite manifest. The schemas should document the CONSULTATION_REQUEST and CONSULTATION_RESPONSE YAML formats that orchestrators already describe inline. This is a documentation extraction, not new behavior.

---

### Category B: Phantom Orchestrator Skills References

Every orchestrator has a "Skills Reference" section listing 2-3 phantom skills. These were placeholder names generated during rite creation that never became actual skills.

#### Issue 12: @agent-design, @workflow-design, @platform-integration in forge/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/forge/agents/orchestrator.md` lines 189-191
**References**: `@agent-design`, `@workflow-design`, `@platform-integration`
**Resolution**: UPDATE -- replace with actual forge skills
**Target**: `@agent-prompt-engineering` (exists), `@rite-development` (exists), `@forge-ref` (exists)
**Batch**: 20

#### Issue 13: @market-research, @business-modeling, @strategy in strategy/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/strategy/agents/orchestrator.md` lines 176-178
**References**: `@market-research`, `@business-modeling`, `@strategy`
**Resolution**: UPDATE -- replace with actual strategy skills
**Target**: `@doc-strategy` (exists), `@strategy-ref` (exists), `@standards` (global)
**Batch**: 20

#### Issue 14: @analytics, @experimentation, @research in intelligence/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/intelligence/agents/orchestrator.md` lines 186-188
**References**: `@analytics`, `@experimentation`, `@research`
**Resolution**: UPDATE -- replace with actual intelligence skills
**Target**: `@doc-intelligence` (exists), `@intelligence-ref` (exists), `@standards` (global)
**Batch**: 20

#### Issue 15: @development, @testing, @architecture in 10x-dev/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/10x-dev/agents/orchestrator.md` lines 205-207
**References**: `@development`, `@testing`, `@architecture`
**Resolution**: UPDATE -- replace with actual 10x-dev skills
**Target**: `@10x-workflow` (exists), `@10x-ref` (exists), `@standards` (global)
**Batch**: 20

#### Issue 16: @debt-management, @risk-assessment, @project-planning in debt-triage/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/debt-triage/agents/orchestrator.md` lines 169-171
**References**: `@debt-management`, `@risk-assessment`, `@project-planning`
**Resolution**: UPDATE -- replace with actual debt-triage skills
**Target**: `@debt-ref` (exists), `@shared-templates` (shared), `@standards` (global)
**Batch**: 20

#### Issue 17: @observability, @incident-response, @chaos-engineering in sre/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/sre/agents/orchestrator.md` lines 176-178
**References**: `@observability`, `@incident-response`, `@chaos-engineering`
**Resolution**: UPDATE -- replace with actual sre skills
**Target**: `@doc-sre` (exists), `@sre-ref` (exists), `@standards` (global)
**Batch**: 20

#### Issue 18: @code-quality, @refactoring, @testing in hygiene/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/hygiene/agents/orchestrator.md` lines 176-178
**References**: `@code-quality`, `@refactoring`, `@testing`
**Resolution**: UPDATE -- replace with actual hygiene skills
**Target**: `@hygiene-ref` (exists), `@smell-detection` (shared), `@standards` (global)
**Batch**: 20

#### Issue 19: @security-standards, @compliance in security/orchestrator

**Location**: `/Users/tomtenuta/Code/knossos/rites/security/agents/orchestrator.md` lines 186-187
**References**: `@security-standards`, `@compliance`
**Resolution**: UPDATE -- replace with actual security skills
**Target**: `@doc-security` (exists), `@security-ref` (exists), `@standards` (global)
**Batch**: 20

---

### Category C: Phantom Fragment Anchors

These reference specific template anchors that don't exist in the target skill.

#### Issue 20: @doc-ecosystem#audit-report-template does not exist

**Location**: `/Users/tomtenuta/Code/knossos/rites/hygiene/agents/audit-lead.md` line 93
**Reference**: `@doc-ecosystem#audit-report-template`
**Resolution**: CREATE -- add audit-report template to doc-ecosystem
**Target**: Create `/Users/tomtenuta/Code/knossos/rites/ecosystem/mena/doc-ecosystem/templates/audit-report.md` and add entry to INDEX
**Batch**: 19

**Rationale**: The smell-report and refactoring-plan templates already exist. The audit-report completes the hygiene workflow's template set (smell -> plan -> clean -> audit).

#### Issue 21: @doc-sre#infrastructure-change-template does not exist

**Location**: `/Users/tomtenuta/Code/knossos/rites/sre/agents/platform-engineer.md` lines 71, 75
**Reference**: `@doc-sre#infrastructure-change-template`
**Resolution**: CREATE -- add infrastructure-change template heading to doc-sre INDEX
**Target**: Add `## Infrastructure Change Template {#infrastructure-change-template}` section to `/Users/tomtenuta/Code/knossos/rites/sre/mena/doc-sre/INDEX.lego.md`
**Batch**: 19

#### Issue 22: @doc-sre#pipeline-design-template does not exist

**Location**: `/Users/tomtenuta/Code/knossos/rites/sre/agents/platform-engineer.md` lines 75, 77
**Reference**: `@doc-sre#pipeline-design-template`
**Resolution**: CREATE -- add pipeline-design template heading to doc-sre INDEX
**Target**: Add `## Pipeline Design Template {#pipeline-design-template}` section to `/Users/tomtenuta/Code/knossos/rites/sre/mena/doc-sre/INDEX.lego.md`
**Batch**: 19

#### Issue 23: @doc-sre#incident-communication-template does not exist

**Location**: `/Users/tomtenuta/Code/knossos/rites/sre/agents/incident-commander.md` lines 77, (referenced in AUDIT-sre-pack-agents.md:125)
**Reference**: `@doc-sre#incident-communication-template`
**Resolution**: CREATE -- add incident-communication template heading to doc-sre INDEX
**Target**: Add `## Incident Communication Template {#incident-communication-template}` section to `/Users/tomtenuta/Code/knossos/rites/sre/mena/doc-sre/INDEX.lego.md`
**Batch**: 19

---

### Category D: Cross-Rite Reference Concerns (Advisory)

These are references that resolve at the source level but may not materialize into the consuming rite's `.claude/skills/` directory. They work when the agent is told to "use this skill" (Claude loads it via Skill tool from any location) but represent architectural boundary leakage.

#### Issue 24: hygiene agents reference @doc-ecosystem (ecosystem-local skill)

**Location**: Multiple hygiene agents:
- `/Users/tomtenuta/Code/knossos/rites/hygiene/agents/architect-enforcer.md` line 87
- `/Users/tomtenuta/Code/knossos/rites/hygiene/agents/code-smeller.md` line 65
- `/Users/tomtenuta/Code/knossos/rites/hygiene/agents/audit-lead.md` line 93
**Reference**: `@doc-ecosystem#refactoring-plan-template`, `@doc-ecosystem#smell-report-template`, `@doc-ecosystem#audit-report-template`
**Resolution**: DEFER -- these work functionally because the Skill tool can load any skill by name regardless of rite boundaries. Moving the hygiene templates to a hygiene-local doc skill would be cleaner architecturally but is not a phantom reference bug.
**Batch**: N/A (advisory for future architectural cleanup)

#### Issue 25: intelligence/analytics-engineer references @doc-sre (sre-local skill)

**Location**: `/Users/tomtenuta/Code/knossos/rites/intelligence/agents/analytics-engineer.md` lines 118, 187
**Reference**: `@doc-sre#tracking-plan-template`
**Resolution**: DEFER -- intentional cross-rite reference. The analytics-engineer explicitly notes "tracking instrumentation lives in SRE domain". This is documented cross-domain collaboration, not a phantom.
**Batch**: N/A (by-design cross-rite reference)

---

### Category E: Template-Level Phantoms (Non-Agent Files)

#### Issue 26: @consult-ref phantom in forge/rite-development and forge/agent-curator

**Locations**:
- `/Users/tomtenuta/Code/knossos/rites/forge/mena/rite-development/INDEX.lego.md` line 133
- `/Users/tomtenuta/Code/knossos/rites/forge/agents/agent-curator.md` line 204
**Reference**: `@consult-ref`
**Resolution**: UPDATE -- reference should point to `@consult` (global mena at `mena/navigation/consult/INDEX.dro.md`)
**Target**: Replace `@consult-ref` with `consult` (the `/consult` command)
**Batch**: 22

**Note**: The `consult-ref` name implies a legomena reference guide for the Consultant, but the actual implementation is the `/consult` dromena command. The reference should be updated to reflect the actual name.

#### Issue 27: @shared/cross-rite-protocol phantom in forge/agent-template

**Locations**:
- `/Users/tomtenuta/Code/knossos/rites/forge/mena/rite-development/templates/agent-template.md` lines 301, 518
**Reference**: `@shared/cross-rite-protocol`
**Resolution**: UPDATE -- replace with `@cross-rite-handoff` (shared mena) or `@cross-rite` (global mena)
**Target**: Replace with `@cross-rite-handoff` since the template is for agent files that will use the shared skill
**Batch**: 22

---

## Resolution Not Required (Valid References)

The following `@references` were verified as resolving correctly:

| Reference | Resolves To | Used By |
|-----------|-------------|---------|
| @standards | `mena/guidance/standards/` (global) | 30+ agents |
| @file-verification | `mena/guidance/file-verification/` (global) | 20+ agents |
| @documentation | `mena/templates/documentation/` (global) | 15+ agents |
| @cross-rite | `mena/guidance/cross-rite/` (global) | 12+ agents |
| @prompting | `mena/guidance/prompting/` (global) | 5+ agents |
| @cross-rite-handoff | `rites/shared/mena/cross-rite-handoff/` (shared) | 5+ agents |
| @shared-templates | `rites/shared/mena/shared-templates/` (shared) | 5+ agents |
| @smell-detection | `rites/shared/mena/smell-detection/` (shared) | 3+ agents |
| @doc-ecosystem | `rites/ecosystem/mena/doc-ecosystem/` (ecosystem) | 5+ agents |
| @doc-sre | `rites/sre/mena/doc-sre/` (sre) | 6+ agents |
| @doc-strategy | `rites/strategy/mena/doc-strategy/` (strategy) | 4 agents |
| @doc-security | `rites/security/mena/doc-security/` (security) | 4 agents |
| @doc-rnd | `rites/rnd/mena/doc-rnd/` (rnd) | 4 agents |
| @doc-intelligence | `rites/intelligence/mena/doc-intelligence/` (intelligence) | 3 agents |
| @doc-reviews | `rites/docs/mena/doc-reviews/` (docs) | 3 agents |
| @rite-development | `rites/forge/mena/rite-development/` (forge) | 4 agents |
| @ecosystem-ref | `rites/ecosystem/mena/ecosystem-ref/` (ecosystem) | 1 agent |
| @10x-workflow | `rites/10x-dev/mena/10x-workflow/` (10x-dev) | 2+ references |

### False Positives (Not Skill References)

| Reference | Context | Reason |
|-----------|---------|--------|
| @product-lead | "Stakeholder contact: @product-lead" | Role/person reference |
| @api-team | "Owner: @api-team" | Team reference in example |
| @platform-team | "owner: @platform-team" | Team reference in example |
| @typescript-eslint/* | ESLint config examples | npm package scope |

---

## Batch Definitions (Revised)

### Phase A: Create Missing Skills and Templates (Batch 19)

**Scope**: 4 creation tasks
**Blast radius**: LOW -- adding new files, not modifying existing behavior
**Rollback**: Delete created files

| Task | Type | Target |
|------|------|--------|
| A1 | CREATE skill | `rites/shared/mena/orchestrator-templates/` (INDEX + 2 schema files) |
| A2 | CREATE template | `rites/ecosystem/mena/doc-ecosystem/templates/audit-report.md` |
| A3 | CREATE template section | `infrastructure-change-template` in doc-sre INDEX |
| A4 | CREATE template section | `pipeline-design-template` in doc-sre INDEX |
| A5 | CREATE template section | `incident-communication-template` in doc-sre INDEX |

**Verification**:
1. After A1: Confirm `rites/shared/mena/orchestrator-templates/INDEX.lego.md` exists with proper frontmatter
2. After A1: Confirm `rites/shared/mena/orchestrator-templates/schemas/consultation-request.md` exists
3. After A1: Confirm `rites/shared/mena/orchestrator-templates/schemas/consultation-response.md` exists
4. After A1: Confirm `rites/shared/manifest.yaml` lists `orchestrator-templates` in legomena
5. After A2: Confirm `rites/ecosystem/mena/doc-ecosystem/templates/audit-report.md` exists
6. After A2: Confirm doc-ecosystem INDEX references the new template
7. After A3-A5: Grep for each anchor ID in doc-sre INDEX, confirm match

**Commit boundary**: Single commit "feat(mena): create orchestrator-templates skill and missing doc templates"

### Phase B: Fix Phantom Skill References (Batches 20 + 22)

**Scope**: 10 update tasks across orchestrator agents + 2 template fixes
**Blast radius**: LOW -- changing reference names in markdown, not behavior
**Rollback**: Revert single commit

| Task | Type | Files |
|------|------|-------|
| B1-B8 | UPDATE | 8 orchestrator agents (Issues 12-19): replace phantom skill names with actual skill names |
| B9 | UPDATE | forge/rite-development INDEX: `@consult-ref` -> `consult` |
| B10 | UPDATE | forge/agent-curator: `@consult-ref` -> `consult` |
| B11 | UPDATE | forge/agent-template: `@shared/cross-rite-protocol` -> `@cross-rite-handoff` |

**Verification**:
1. Grep all agent files for the old phantom names -- zero matches expected
2. Grep for new reference names -- confirm correct count
3. Run `CGO_ENABLED=0 go build ./cmd/ari` -- confirm no build impact (markdown-only changes)

**Commit boundary**: Single commit "fix(agents): resolve phantom skill references in orchestrator and forge agents"

---

## Risk Assessment

| Phase | Risk Level | Blast Radius | Failure Detection | Recovery |
|-------|-----------|--------------|-------------------|----------|
| A (create) | LOW | New files only, no existing behavior affected | File existence check | Delete created files |
| B (update) | LOW | Markdown reference text in 11 files | Grep for old/new names | Revert commit |

---

## Scope Assessment

**Total distinct issues**: 27 (22 phantoms + 4 phantom anchors + 1 advisory item)
**Collapsible**: Issues 1-11 are one creation task. Issues 12-19 are one update pattern.
**Effective work items**: 12 (5 creations + 10 updates - 3 that are advisory/deferred)
**Recommendation**: This is a moderate scope. Batches 19-22 can be collapsed into 2 phases (A and B) as described above.

---

## Janitor Notes

1. **Commit convention**: Use `feat(mena):` prefix for Phase A (new content), `fix(agents):` prefix for Phase B (corrections)
2. **Test requirement**: No Go tests needed -- these are markdown-only changes. Verify with `CGO_ENABLED=0 go build ./cmd/ari` to confirm no impact on binary
3. **Critical ordering**: Phase A MUST complete before Phase B, because Phase B references depend on Phase A creations
4. **Materialization**: After both phases, run `ari materialize` to project new shared skills into `.claude/skills/`
5. **orchestrator-templates content**: Extract the CONSULTATION_REQUEST and CONSULTATION_RESPONSE schemas from the inline descriptions in any orchestrator agent (they are all identical). The hygiene orchestrator at `rites/hygiene/agents/orchestrator.md` lines 60-70 is a good source
6. **doc-sre new templates**: For the 3 new template sections (infrastructure-change, pipeline-design, incident-communication), follow the same heading + markdown code fence pattern used by the existing 6 templates in doc-sre INDEX. The SRE AUDIT at `rites/sre/AUDIT-sre-pack-agents.md` lines 44 and 125 confirms these were intentionally planned

---

## Attestation

| File | Read | Verified |
|------|------|----------|
| All 56 agent files in rites/*/agents/*.md | Yes (via Grep) | @ references extracted and cross-referenced |
| All INDEX.lego.md and INDEX.dro.md in mena/ | Yes (via find) | Available skills registry built |
| All INDEX.lego.md and INDEX.dro.md in rites/*/mena/ | Yes (via find) | Rite-local skills registry built |
| rites/ecosystem/mena/doc-ecosystem/INDEX.lego.md | Yes (Read) | Confirmed no audit-report-template anchor |
| rites/sre/mena/doc-sre/INDEX.lego.md | Yes (Read) | Confirmed 6 existing + 3 missing templates |
| rites/hygiene/manifest.yaml | Yes (Read) | Confirmed legomena: [hygiene-ref] only |
| rites/shared/manifest.yaml | Yes (implied via find) | Confirmed shared skill set |
| .claude/skills/ | Yes (ls) | Confirmed materialized output |
| rites/forge/mena/rite-development/INDEX.lego.md | Yes (Read) | Confirmed @consult-ref phantom |
| rites/forge/agents/agent-curator.md | Yes (Read) | Confirmed @consult-ref phantom |
| rites/forge/mena/rite-development/templates/agent-template.md | Yes (via Grep) | Confirmed @shared/cross-rite-protocol phantom |

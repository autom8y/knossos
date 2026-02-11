# CE Audit: Rite Structural Consistency

**Auditor**: Context Engineer
**Date**: 2026-02-09
**Scope**: `rites/*/` -- all 12 rites (11 workflow rites + shared infrastructure)

---

## Summary

- **Total rites audited**: 12 (11 workflow + 1 infrastructure)
- **Manifest schema consistency**: 83% (10/12 share identical field set; 2 have extra fields)
- **Orchestrator pattern compliance**: 73% (8/11 use minimal-comment style; 3 use verbose-comment style)
- **Workflow DAG validity**: 10/11 (rnd has 1 unreachable agent)
- **README consistency**: 75% (9/12 follow the standard structure; naming split between "Rite" and "Pack")
- **TODO quality**: 100% present for workflow rites; 0% for forge (missing) and shared (missing)
- **Legacy terminology residue**: 4 files still reference "teams" in orchestrator/workflow YAML

---

## Rite Inventory

| Rite | Agents (manifest) | Agents (files) | Legomena | Manifest | Orchestrator | Workflow | README | TODO |
|------|-------------------|----------------|----------|----------|--------------|----------|--------|------|
| 10x-dev | 5 | 5 | 5 | Y | Y | Y | Y | Y |
| debt-triage | 4 | 4 | 1 | Y | Y | Y | Y | Y |
| docs | 5 | 5 | 3 | Y | Y | Y | Y | Y |
| ecosystem | 6 | 6 | 3 | Y | Y | Y | Y | Y |
| forge | 7 | 7 | 3 | Y | Y | Y | N | N |
| hygiene | 5 | 5 | 1 | Y | Y | Y | Y | Y |
| intelligence | 5 | 5 | 2 | Y | Y | Y | Y | Y |
| rnd | 6 | 6 | 2 | Y | Y | Y | Y | Y |
| security | 5 | 5 | 2 | Y | Y | Y | Y | Y |
| shared | 0 | 0 | 4 | Y | N (correct) | N (correct) | Y | N |
| sre | 5 | 5 | 2 | Y | Y | Y | Y | Y |
| strategy | 5 | 5 | 2 | Y | Y | Y | Y | Y |

**Agent count by composition model** (per rite-composition.md canonical pattern):

| Model | Expected Agents | Rites | Actual Match |
|-------|----------------|-------|--------------|
| 3-Role (focused) | 3 + orchestrator = 4 | debt-triage | YES (4 agents, 3 phases) |
| 4-Role (standard) | 4 + orchestrator = 5 | docs, hygiene, intelligence, security, sre, strategy | YES (all have 5 agents, 4 phases) |
| 5-Role (full lifecycle) | 4 + orchestrator = 5 | 10x-dev | YES (5 agents, 4 phases) |
| 5-Phase (extended) | 5 + orchestrator = 6 | ecosystem, rnd | ecosystem: YES (6 agents, 5 phases); rnd: DRIFT (6 agents but only 4 phases -- tech-transfer orphaned) |
| 6-Phase (extended) | 6 + orchestrator = 7 | forge | YES (7 agents, 6 phases) |

---

## Critical Findings

### CRIT-1: rnd rite has unreachable agent `tech-transfer`

**Location**: `rites/rnd/manifest.yaml` line 48-49, `rites/rnd/agents/tech-transfer.md`

The `tech-transfer` agent is declared in the manifest and has an agent prompt file, but appears in NEITHER:
- `rites/rnd/workflow.yaml` (no phase routes to it)
- `rites/rnd/orchestrator.yaml` (no routing entry for it)

The orchestrator agent file (`rites/rnd/agents/orchestrator.md`) DOES reference tech-transfer in its routing table and phase handoff criteria. But the structured YAML files that drive materialization and tooling do not.

**Impact**: The orchestrator prompt knows about tech-transfer and will try to delegate to it, but the workflow DAG has no phase for it. This creates a semantic mismatch where the orchestrator believes it can route to an agent that the workflow engine does not recognize.

**Fix**: Add a `transfer` phase to `rites/rnd/workflow.yaml` after `future-architecture`, and add a `tech-transfer` routing entry to `rites/rnd/orchestrator.yaml`.

### CRIT-2: `docs/README.md` references non-existent `api-rite`

**Location**: `rites/docs/README.md` line 44

```markdown
- **api-rite**: When API documentation needs specialized treatment
```

No rite named `api-rite` exists in the system. This is a dangling reference that would confuse any agent or user consulting the README for routing guidance.

**Fix**: Remove this line or replace with a valid rite reference.

---

## High Findings

### HIGH-1: Legacy "teams" terminology persists in 4 orchestrator/workflow YAML files

**Locations**:
- `rites/rnd/orchestrator.yaml` line 53: `"@cross-rite for handoff patterns to other teams"`
- `rites/ecosystem/orchestrator.yaml` line 65: `"Designing without considering all 10 teams"`
- `rites/ecosystem/orchestrator.yaml` line 69: `"When changes affect other teams, escalate to user for coordination."`
- `rites/security/workflow.yaml` line 22: `"# Allows other teams (primarily 10x-dev) to invoke threat-modeler"`

The SL-008 terminology deep cleanse (completed 2026-02-09) resolved canaries across 65+ files and deleted the `teams/` directory. These 4 references survived. Per the terminology audit decision framework, refs to "teams" should be "rites" when the subject is Knossos workflow domains.

**Fix**: Mechanical replacement: "other teams" -> "other rites", "all 10 teams" -> "all rites", "other teams (primarily 10x-dev)" -> "other rites (primarily 10x-dev)".

### HIGH-2: Forge rite missing README.md and TODO.md

**Location**: `rites/forge/`

Forge is the meta-rite (creates other rites). It has 7 agents and 34 mena files -- the most complex rite in the system. Yet it is one of only two rites without a README (shared is the other) and the ONLY workflow rite without a TODO.

Every other workflow rite received a structured audit on 2026-01-02 that produced a TODO.md with validated improvements, deferred decisions, and cross-rite notes. Forge was not audited.

**Impact**: Forge is the rite that produces other rites. Its mena contains the canonical composition patterns, naming conventions, and glossaries. Without a TODO tracking its own structural gaps, the rite that enforces quality on others has no quality self-assessment.

**Fix**: Create `rites/forge/README.md` following the standard structure (When to Use, Quick Start, Agents table, Workflow, Related Rites). Create or schedule a TODO.md audit.

### HIGH-3: Inconsistent naming: "Pack" vs "Rite" in README headers

**Locations**: 6 READMEs say "Pack", 5 say "Rite"

| Header Style | Rites |
|-------------|-------|
| `# {Name} Pack` | 10x-dev, debt-triage, hygiene, rnd, security, sre |
| `# {Name} Rite` | docs, ecosystem, intelligence, strategy |
| Missing README | forge |

The SL-008 deep cleanse standardized "rite" throughout the codebase. "Pack" was the pre-SL-008 legacy term. The README headers were not caught because the cleanse focused on Go source, agent prompts, mena, and YAML config -- not markdown headers.

**Fix**: Rename all README headers to `# {Name} Rite` for consistency. This is a mechanical find-and-replace.

---

## Medium Findings

### MED-1: Orchestrator YAML comment style inconsistency

Three orchestrator files (rnd, security, ecosystem) use verbose inline comments with section headers (`# Rite metadata`, `# Frontmatter customization`, `# Specialist routing conditions`). The other eight use minimal comments (single header line only).

This is cosmetic but creates maintenance drift. When the orchestrator template is updated, eight files match the expected pattern and three do not.

**Fix**: Standardize all orchestrators to the minimal comment style. The verbose comments duplicate the field names and add no information.

### MED-2: `version` field present in only 2/11 workflow.yaml files

Only `ecosystem` and `forge` workflow.yaml files include a `version: "1.0.0"` field. The other 9 omit it. All 12 manifests consistently include the version field.

**Impact**: If tooling ever consumes the version field from workflow files, 9 rites would fail or return null. The inconsistency suggests the field was added to ecosystem and forge during a different development session and not backported.

**Fix**: Either add `version: "1.0.0"` to all workflow.yaml files, or remove it from the two that have it. Recommend the former for forward compatibility.

### MED-3: `mcp_servers` field present in only 2/12 manifest files

Only `10x-dev` (github MCP) and `ecosystem` (go-semantic, terraform) define MCP servers. This is not inherently wrong -- not all rites need MCP servers. However, the intelligence rite's TODO explicitly plans MCP integration (P5: autom8_data), and several rites would benefit from declared MCP server access (security could use OWASP tools, strategy could use market data APIs).

**Impact**: Low currently, but as MCP adoption grows, this field needs consistent treatment.

**Recommendation**: No action now, but when MCP servers are added, ensure the manifest schema documents this as an optional field with clear semantics.

### MED-4: `commands` field in workflow.yaml used by only 2 rites

Only `ecosystem` and `forge` workflow.yaml files include a `commands:` section mapping rite-specific dromena. Other rites define their commands via the `mena/` directory and `dromena: []` in the manifest.

This creates two patterns for command declaration: inline in workflow.yaml vs. via mena directory projection. The dual pattern is confusing for anyone creating a new rite.

**Recommendation**: Document which pattern is canonical. If `dromena: []` in the manifest is the standard, remove the inline `commands:` from workflow files or explain why these two rites are exceptions.

### MED-5: `external_consultation` in security manifest but not in workflow schema

The `security` manifest includes an `external_consultation` block (lines 72-79) that is unique to this rite. The same data appears in expanded form in `security/workflow.yaml` (lines 21-59). This is the only manifest that uses this field.

**Impact**: If the manifest schema is ever validated programmatically, this field would be flagged as unknown for all other rites or unexpected for security.

**Recommendation**: Either formalize `external_consultation` as an optional manifest field (it is a useful concept for cross-rite protocols) or move it entirely to workflow.yaml where the expanded version already lives.

---

## Low Findings

### LOW-1: `doc-rite` used as informal name in forge mena files

The forge rite's reference material (`rites/forge/mena/rite-development/`) uses `doc-rite` to refer to the `docs` rite in 8+ locations (patterns, glossary, artifacts). The actual rite name is `docs`, not `doc-rite`. While this is in descriptive text (not routing config), it could cause confusion when creating new rites.

**Locations**: `rite-composition.md`, `complexity-gating.md`, `command-mapping.md`, `agents.md`, `artifacts.md`

**Fix**: Replace `doc-rite` with `docs` in forge mena reference material.

### LOW-2: Shared rite missing TODO.md

The shared rite is infrastructure-only (no agents, no workflow). A TODO would track planned additions to shared legomena. The sre TODO P1 already identifies templates that should move to shared; the debt-triage TODO P1 identifies shared smell detection. Both are actionable items that lack a tracking home.

**Fix**: Create `rites/shared/TODO.md` capturing the backlog of items identified across other rite TODOs.

### LOW-3: Strategy workflow.yaml has the most elaborate back-routes

The strategy workflow.yaml is 113 lines with 6 back-routes (including `context_to_pass` and `expected_outcome` fields). Other rites use 3 back-routes with simpler schemas. While strategy's back-routes are well-designed, the schema divergence means tooling must handle both simple and extended back-route formats.

**Impact**: Very low -- back-routes appear to be documentation/guidance rather than machine-interpreted config currently. But if back-routes ever become programmatic, the strategy rite would require special handling.

### LOW-4: 10x-dev workflow.yaml includes security_consultation block

The `10x-dev/workflow.yaml` is 198 lines -- the longest workflow file -- partly because it includes a 53-line `security_consultation` section (lines 73-128). This defines trigger patterns, consultation policies, and invocation patterns for cross-rite security reviews.

The same information exists in `security/workflow.yaml` under `external_consultation`. The duplication is pragmatic (architect needs the triggers at hand) but creates dual maintenance burden.

**Recommendation**: Consider whether security_consultation should be a shared legomenon that both rites reference rather than inline config.

---

## Structural Recommendations

### R1: Standardize the Rite Scaffold

Based on this audit, the canonical rite scaffold should be:

```
rites/{name}/
  manifest.yaml     # REQUIRED - name, version, description, entry_agent, phases, agents, legomena, dependencies, complexity_levels, metadata
  orchestrator.yaml  # REQUIRED for workflow rites - rite, frontmatter, routing, workflow_position, handoff_criteria, skills, antipatterns, cross_rite_protocol
  workflow.yaml      # REQUIRED for workflow rites - name, workflow_type, description, entry_point, phases, complexity_levels, back_routes
  README.md          # REQUIRED - header "# {Name} Rite", sections: When to Use, Quick Start, Agents, Workflow, Related Rites
  TODO.md            # RECOMMENDED - audit date, status, validated improvements, deferred decisions, dependencies, cross-rite notes
  agents/            # REQUIRED for workflow rites
  mena/              # OPTIONAL - rite-specific legomena
  hooks/             # OPTIONAL - rite-specific hooks
```

### R2: Enforce Manifest-Workflow-Orchestrator Agent Alignment

Every agent listed in `manifest.yaml` MUST appear in EITHER:
- A `phases[].agent` entry in `workflow.yaml`, OR
- Be documented as an "out-of-workflow" agent with explicit justification

Current violation: `tech-transfer` in rnd.

Every agent in `orchestrator.yaml` routing MUST appear in `workflow.yaml` phases.

### R3: Resolve the Pack/Rite Naming Split

The SL-008 cleanse established "rite" as the canonical term. Apply the same standard to README headers. The fix is mechanical:

```
# 10x Dev Pack     -> # 10x Dev Rite
# Debt Triage Pack -> # Debt Triage Rite
# Hygiene Pack     -> # Hygiene Rite
# R&D Pack         -> # R&D Rite
# Security Pack    -> # Security Rite
# SRE Pack         -> # SRE Rite
```

### R4: Complete the Terminology Cleanse in YAML Config

4 YAML files still reference "teams" where "rites" is correct. This is a direct extension of SL-008 scope.

### R5: Audit the Forge Rite

Forge is the only workflow rite without a TODO.md. Given that forge is the meta-rite responsible for creating other rites, its structural patterns directly influence every other rite. An audit would validate its 7-agent, 6-phase design and identify whether it follows its own composition guidelines.

### R6: Formalize the Workflow.yaml Schema

The current workflow.yaml files share a core schema but diverge in extensions:
- `commands:` (ecosystem, forge only)
- `security_consultation:` (10x-dev only)
- `external_consultation:` (security only)
- `impact_routing:` (10x-dev only)
- `version:` (ecosystem, forge only)

Document the schema formally with required fields, optional fields, and rite-specific extension points. This prevents further schema drift.

### R7: Shared Mena Coverage Assessment

Current shared mena (4 legomena, 30 files):
- `cross-rite-handoff` -- schema + examples for inter-rite transitions
- `orchestrator-templates` -- consultation request/response schemas
- `shared-templates` -- debt ledger, risk matrix, sprint package templates
- `smell-detection` -- taxonomy, severity, integration patterns

**Items identified across rite TODOs as candidates for shared**:
- Generalized HANDOFF artifact pattern (debt-triage P2, referenced by 6+ rites)
- Shared behavior-preservation definition (hygiene P1)
- Intelligence-vs-Strategy boundary doc (intelligence P3, strategy P1)
- RND-vs-Spike distinction (rnd P2)

The cross-rite-handoff and shared-templates legomena were created in response to the rite audit. The TODO items from sre P1 (shared templates) and debt-triage P1 (smell detection) appear to have been implemented (both exist in shared mena). However, the generalized HANDOFF artifact from debt-triage P2 may overlap with the existing cross-rite-handoff legomenon -- verification needed.

---

## Appendix: Per-Rite Sizing Assessment

| Rite | Agents | Phases | Mena Files | Back-Routes | Assessment |
|------|--------|--------|------------|-------------|------------|
| 10x-dev | 5 | 4 | 13 | 3 | RIGHT-SIZED. Flagship rite, well-documented, mature. |
| debt-triage | 4 | 3 | 1 | 3 | RIGHT-SIZED. Minimal 3-role model, plans but does not execute. |
| docs | 5 | 4 | 17 | 3 | RIGHT-SIZED. Standard 4-role model, clear separation. |
| ecosystem | 6 | 5 | 21 | 3 | SLIGHTLY OVER-SCOPED. 5 phases with documentation-engineer as conditional. TODO acknowledges over-engineering (hub claims, satellite matrix). |
| forge | 7 | 6 | 34 | 3 | AT UPPER LIMIT. 7 agents is the maximum in the system. Justified as the meta-rite, but needs audit to confirm all 6 phases earn their existence. |
| hygiene | 5 | 4 | 3 | 3 | RIGHT-SIZED. Clean 4-role model. |
| intelligence | 5 | 4 | 2 | 3 | RIGHT-SIZED but UNDER-DOCUMENTED. Only 2 mena files for a data-focused rite. |
| rnd | 6 | 4 | 2 | 3 | MISALIGNED. 6 agents but only 4 workflow phases. tech-transfer is orphaned. |
| security | 5 | 4 | 2 | 3 | RIGHT-SIZED. Clean 4-role model with well-designed cross-rite consultation. |
| shared | 0 | 0 | 30 | 0 | RIGHT-SIZED as infrastructure. 30 mena files is substantial but they are shared across all rites. |
| sre | 5 | 4 | 2 | 3 | RIGHT-SIZED. Clean 4-role model, recently optimized (44% token reduction). |
| strategy | 5 | 4 | 2 | 6 | RIGHT-SIZED. Clean 4-role model with most elaborate back-routes. |

---

## Appendix: Cross-Rite Handoff Map

Based on README and TODO analysis, the declared cross-rite handoffs are:

```
debt-triage  --> hygiene     (execution: debt items for cleanup)
10x-dev      --> docs        (documentation: user-facing changes need docs)
10x-dev      --> sre         (validation: production readiness)
10x-dev      --> security    (assessment: SYSTEM complexity security review)
security     --> 10x-dev     (remediation: fix findings)
rnd          --> 10x-dev     (productionization: prototype to production)
rnd          --> strategy    (strategic input: moonshot findings)
intelligence --> 10x-dev     (implementation: experiment results driving features)
intelligence --> strategy    (strategic input: user insights informing roadmap)
strategy     --> 10x-dev     (implementation: strategic roadmap to build)
```

10x-dev is the primary hub with 5 inbound and 3 outbound handoff types. The cross-rite-handoff shared legomenon provides the schema for these transitions.

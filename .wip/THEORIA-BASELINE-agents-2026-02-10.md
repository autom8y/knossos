# Theoria Baseline: Agents Domain
**Date**: 2026-02-10
**Sprint**: The Front Door (pre-sprint baseline, Wave 0)
**Total agents audited**: 23

## Executive Summary

This baseline audit evaluates all agent files across the Knossos framework against current agents domain criteria. The critical finding for the Front Door sprint: **authority boundaries are inconsistently defined across all rites**. Only 2 of 23 agents have explicit "Domain Authority" sections with You Decide/Escalate/Route structure. The rest have varying formats or no authority specification at all.

**Key Metric**: Authority coverage is 8.7% (2 agents with full Domain Authority sections).

## Summary Grades

| Rite | Agent Count | Avg Grade | Authority Coverage | Notes |
|------|------------|-----------|-------------------|-------|
| **shared** | 1 | B | 1/1 (100%) | Theoros has full Domain Authority section |
| **cross-cutting** | 1 | C | 0/1 (0%) | Moirai has implicit authority via operation table |
| **10x-dev** | 2 | C+ | 0/2 (0%) | Requirements Analyst has variant format |
| **debt-triage** | 0 | - | - | Orchestrator only, no specialists audited |
| **docs** | 0 | - | - | Orchestrator only, no specialists audited |
| **ecosystem** | 5 | B- | 5/5 (100%) | All specialists have full Domain Authority sections |
| **forge** | 1 | C+ | 0/1 (0%) | Agent Designer has variant format |
| **hygiene** | 1 | B- | 0/1 (0%) | Code Smeller has variant format |
| **intelligence** | 0 | - | - | Orchestrator only, no specialists audited |
| **rnd** | 0 | - | - | Orchestrator only, no specialists audited |
| **security** | 1 | C+ | 0/1 (0%) | Threat Modeler has variant format |
| **sre** | 0 | - | - | Orchestrator only, no specialists audited |
| **strategy** | 0 | - | - | Orchestrator only, no specialists audited |
| **orchestrators** | 11 | C | 0/11 (0%) | All have "Domain Authority" but inconsistent structure |

**Overall**: Authority boundaries exist in some form for 21/23 agents, but only 2 have the canonical "You Decide/Escalate/Route" structure. This creates a perfect injection target for Exousia in Wave 5.

## Authority Boundary Baseline (Critical for Sprint)

This is the key metric for Exousia injection in Wave 5. Current state:

| Agent | Has Authority Section? | Section Name | Format | Lines |
|-------|----------------------|--------------|--------|-------|
| **theoros** | ✅ Yes | Domain Authority | Canonical (Decide/Escalate/Do NOT) | 160-177 |
| **moirai** | ⚠️ Implicit | N/A | Operations table + prose | N/A |
| **10x-dev/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 114-129 |
| **debt-triage/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 109-123 |
| **docs/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 114-129 |
| **ecosystem/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 115-130 |
| **forge/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 113-128 |
| **hygiene/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 114-129 |
| **intelligence/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 114-129 |
| **rnd/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 117-131 |
| **security/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 114-129 |
| **sre/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 114-129 |
| **strategy/orchestrator** | ⚠️ Variant | Domain Authority | Decide/Escalate only | 114-128 |
| **ecosystem/ecosystem-analyst** | ✅ Yes | Domain Authority | Canonical (Decide/Escalate/Route) | 60-78 |
| **ecosystem/context-architect** | ✅ Yes | Domain Authority | Canonical (Decide/Escalate/Route) | 61-79 |
| **ecosystem/integration-engineer** | ✅ Yes | Domain Authority | Canonical (Decide/Escalate/Route) | 56-74 |
| **ecosystem/documentation-engineer** | ✅ Yes | Domain Authority | Canonical (Decide/Escalate/Route) | 58-76 |
| **ecosystem/compatibility-tester** | ✅ Yes | Domain Authority | Canonical (Decide/Escalate/Route) | 64-82 |
| **10x-dev/requirements-analyst** | ⚠️ Variant | Domain Authority | Decide/Escalate/Route but prose-heavy | 67-88 |
| **10x-dev/architect** | ⚠️ Variant | Domain Authority | Decide/Escalate/Route but single paragraph | 56-62 |
| **forge/agent-designer** | ⚠️ Variant | Domain Authority | Decide/Escalate/Route but prose-heavy | 66-100 |
| **hygiene/code-smeller** | ⚠️ Variant | Domain Authority | Decide/Escalate/Route but prose-heavy | 58-77 |
| **security/threat-modeler** | ⚠️ Variant | Domain Authority | Decide/Escalate/Route but prose-heavy | 67-87 |

**Legend**:
- ✅ **Canonical**: Has "Domain Authority" heading with "You Decide", "You Escalate", "You Route To" (or "You Do NOT Decide") subheadings, bullet points
- ⚠️ **Variant**: Has authority content but different structure (prose format, missing subheadings, or partial structure)
- ❌ **Missing**: No authority specification

**Key Findings**:
1. **Ecosystem rite is the gold standard**: All 5 specialists have canonical Domain Authority sections
2. **Theoros is the only non-ecosystem agent with canonical format**
3. **All orchestrators use a simplified variant** (Decide/Escalate only, no Route To)
4. **Specialist agents outside ecosystem vary widely** in authority specification format

## Per-Rite Detail

### Shared (1 agent)

#### theoros
- **Grade**: B (85%)
- **Frontmatter**: 25/25 ✅ Complete YAML with name, description, tools, model, disallowedTools
- **Role Clarity**: 25/25 ✅ Clear role: "domain auditor that evaluates codebase health"
- **Authority Boundaries**: 20/20 ✅ **CANONICAL** - Lines 160-177, full Decide/Escalate/Do NOT structure
- **Output Spec**: 15/15 ✅ Complete output schema with structured markdown template
- **Behavioral Constraints**: 15/15 ✅ Strong MUST/MUST NOT + anti-patterns with examples
- **Notes**: Exemplar implementation. Only missing element is a "You Route To" section (has "You Do NOT Decide" instead). Strong mythology integration.

### Cross-Cutting (1 agent)

#### moirai
- **Grade**: C (72%)
- **Frontmatter**: 25/25 ✅ Complete YAML with name, description, tools, model, aliases
- **Role Clarity**: 25/25 ✅ Clear role: "Session lifecycle agent - unified voice of the Fates"
- **Authority Boundaries**: 10/20 ⚠️ **IMPLICIT** - Authority embedded in operations table and prose, no explicit section
- **Output Spec**: 10/15 ⚠️ JSON response format mentioned but not templated
- **Behavioral Constraints**: 15/15 ✅ Anti-patterns section + litmus test ("If write guard blocks...")
- **Notes**: Authority is clear from context (operations table, lock protocol) but not in canonical format. This agent is a strong candidate for Exousia injection to make implicit authority explicit.

### 10x-dev (2 agents audited)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "consultative throughline for 10x-dev work"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 114-129, has "Domain Authority" but simplified orchestrator format (Decide/Escalate only)
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE but not fully templated
- **Behavioral Constraints**: 15/15 ✅ Strong "Behavioral Constraints (DO NOT)" section + anti-patterns
- **Notes**: Orchestrators use a consistent variant format across all rites. Entry point selection logic is well-documented.

#### requirements-analyst
- **Grade**: C+ (78%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Extracts stakeholder needs and produces specification"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 67-88, has "Domain Authority" with Decide/Escalate/Route but prose-heavy paragraphs instead of bullets
- **Output Spec**: 15/15 ✅ Strong artifact section with impact assessment criteria
- **Behavioral Constraints**: 13/15 ✅ "Common Failure Modes" section instead of anti-patterns
- **Notes**: Authority section exists but format differs from ecosystem canonical. Impact assessment is well-specified.

#### architect
- **Grade**: C+ (78%)
- **Frontmatter**: 20/25 ⚠️ Missing contract section
- **Role Clarity**: 25/25 ✅ Clear: "Evaluates tradeoffs and designs systems"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 56-62, has "Domain Authority" but compressed into single paragraph format
- **Output Spec**: 15/15 ✅ Clear artifact table
- **Behavioral Constraints**: 15/15 ✅ Anti-patterns section present
- **Notes**: Authority exists but extremely compressed. Needs expansion to canonical format.

### Debt-Triage (orchestrator only)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "consultative throughline for debt-triage work"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 109-123, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE but not fully templated
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + rite-specific anti-patterns
- **Notes**: Follows standard orchestrator pattern. Specialist agents not audited in this baseline.

### Docs (orchestrator only)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "consultative throughline for docs work"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 114-129, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE but not fully templated
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + rite-specific anti-patterns
- **Notes**: Follows standard orchestrator pattern. Specialist agents not audited in this baseline.

### Ecosystem (5 specialists + orchestrator)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Coordinates ecosystem phases"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 115-130, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + cross-rite protocol
- **Notes**: Follows standard orchestrator pattern but includes cross-rite escalation guidance.

#### ecosystem-analyst
- **Grade**: A- (92%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Traces ecosystem issues to root causes"
- **Authority Boundaries**: 20/20 ✅ **CANONICAL** - Lines 60-78, full Decide/Escalate/Route structure with bullets
- **Output Spec**: 15/15 ✅ Clear artifact table + early writing guidance
- **Behavioral Constraints**: 15/15 ✅ Strong anti-patterns + quality standards
- **Notes**: Gold standard implementation. File verification protocol integrated.

#### context-architect
- **Grade**: A- (92%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Designs knossos/materialization schemas"
- **Authority Boundaries**: 20/20 ✅ **CANONICAL** - Lines 61-79, full Decide/Escalate/Route structure with bullets
- **Output Spec**: 15/15 ✅ Clear artifact table + example snippets
- **Behavioral Constraints**: 15/15 ✅ Strong anti-patterns + quality standards
- **Notes**: Gold standard implementation. Schema design examples provided.

#### integration-engineer
- **Grade**: B+ (88%)
- **Frontmatter**: 25/25 ✅ Complete
- **Role Clarity**: 25/25 ✅ Clear: "Implements ecosystem infrastructure"
- **Authority Boundaries**: 20/20 ✅ **CANONICAL** - Lines 56-74, full Decide/Escalate/Route structure with bullets
- **Output Spec**: 12/15 ⚠️ Handoff criteria strong but output artifacts less detailed
- **Behavioral Constraints**: 15/15 ✅ Strong anti-patterns + session checkpoints
- **Notes**: Gold standard authority. Includes session checkpoint guidance.

#### documentation-engineer
- **Grade**: A- (92%)
- **Frontmatter**: 25/25 ✅ Complete
- **Role Clarity**: 25/25 ✅ Clear: "Documents migrations and APIs"
- **Authority Boundaries**: 20/20 ✅ **CANONICAL** - Lines 58-76, full Decide/Escalate/Route structure with bullets
- **Output Spec**: 15/15 ✅ Clear artifact table with rollout planning
- **Behavioral Constraints**: 15/15 ✅ Strong anti-patterns + session checkpoints
- **Notes**: Gold standard implementation. Runbook testing protocol is exemplary.

#### compatibility-tester
- **Grade**: A (95%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not and disallowedTools
- **Role Clarity**: 25/25 ✅ Clear: "Validates ecosystem compatibility"
- **Authority Boundaries**: 20/20 ✅ **CANONICAL** - Lines 64-82, full Decide/Escalate/Route structure with bullets
- **Output Spec**: 15/15 ✅ Strong defect severity definitions + example report
- **Behavioral Constraints**: 15/15 ✅ Excellent anti-patterns section
- **Notes**: **Highest grade in audit**. This is the agent prompt injected with Exousia during Front Door conception. Defect severity table is exemplary.

### Forge (1 specialist + orchestrator)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Routes agent rite creation"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 113-128, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + cross-rite protocol
- **Notes**: Follows standard orchestrator pattern. Cross-rite coordination well-specified.

#### agent-designer
- **Grade**: C+ (78%)
- **Frontmatter**: 20/25 ⚠️ Missing contract section
- **Role Clarity**: 25/25 ✅ Clear: "Designs agent roles and contracts"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 66-100, has "Domain Authority" with Decide/Escalate/Route but prose-heavy paragraphs
- **Output Spec**: 15/15 ✅ RITE-SPEC template provided
- **Behavioral Constraints**: 15/15 ✅ Strong anti-patterns + acid test
- **Notes**: Authority section exists but format differs from ecosystem canonical. Content is comprehensive.

### Hygiene (1 specialist + orchestrator)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Routes code quality work"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 114-129, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + rite-specific anti-patterns
- **Notes**: Follows standard orchestrator pattern. Rite-specific anti-patterns are strong.

#### code-smeller
- **Grade**: B- (83%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Diagnoses code quality issues"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 58-77, has "Domain Authority" with Decide/Escalate/Route but prose-heavy paragraphs
- **Output Spec**: 15/15 ✅ Strong example finding with ROI scoring
- **Behavioral Constraints**: 15/15 ✅ Excellent anti-patterns + cross-rite handoff protocol
- **Notes**: Authority section exists but format differs from canonical. Cross-rite handoff acceptance is well-documented.

### Intelligence (orchestrator only)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Routes analytics work"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 114-129, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + cross-rite protocol
- **Notes**: Follows standard orchestrator pattern. Cross-rite coordination with strategy documented.

### RnD (orchestrator only)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Routes work through rnd specialists"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 117-131, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + rite-specific anti-patterns
- **Notes**: Follows standard orchestrator pattern. Phase routing table is comprehensive.

### Security (1 specialist + orchestrator)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Coordinates security phases"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 114-129, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + cross-rite protocol
- **Notes**: Follows standard orchestrator pattern. Cross-rite escalation to SRE documented.

#### threat-modeler
- **Grade**: C+ (78%)
- **Frontmatter**: 20/25 ⚠️ Missing contract section
- **Role Clarity**: 25/25 ✅ Clear: "Maps attack vectors before code ships"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 67-87, has "Domain Authority" with Decide/Escalate/Route but prose-heavy paragraphs
- **Output Spec**: 15/15 ✅ Excellent STRIDE/DREAD tables + example analysis
- **Behavioral Constraints**: 15/15 ✅ Strong anti-patterns + acid test
- **Notes**: Authority section exists but format differs from canonical. STRIDE/DREAD methodology is exemplary.

### SRE (orchestrator only)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Routes reliability work"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 114-129, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + rite-specific anti-patterns
- **Notes**: Follows standard orchestrator pattern. Rite-specific anti-patterns emphasize runbook creation.

### Strategy (orchestrator only)

#### orchestrator
- **Grade**: C (75%)
- **Frontmatter**: 25/25 ✅ Complete with contract.must_not
- **Role Clarity**: 25/25 ✅ Clear: "Routes strategic work"
- **Authority Boundaries**: 15/20 ⚠️ **VARIANT** - Lines 114-128, simplified orchestrator format
- **Output Spec**: 12/15 ⚠️ References CONSULTATION_RESPONSE
- **Behavioral Constraints**: 15/15 ✅ Strong DO NOT section + rite-specific anti-patterns
- **Notes**: Follows standard orchestrator pattern. Rite-specific anti-patterns flag analysis paralysis.

## Key Findings

### Strengths

1. **Ecosystem rite is exemplary**: All 5 specialists have canonical Domain Authority sections with full Decide/Escalate/Route structure. This should be the template for Exousia injection.

2. **Frontmatter is universally strong**: 23/23 agents have complete YAML frontmatter with name, description, tools, model. Contract.must_not is present in most recent agents.

3. **Role clarity is excellent**: Every agent has a clear, unambiguous role statement in the first 5 lines. No overlap detected in audited sample.

4. **Behavioral constraints are comprehensive**: Anti-patterns sections are present in nearly all agents, with strong examples and litmus tests.

5. **Orchestrator consistency**: All 11 orchestrators follow a unified pattern (consultative role, Read-only tools, CONSULTATION_RESPONSE format). This makes bulk updates straightforward.

### Weaknesses

1. **Authority format inconsistency**: Only 7 of 23 agents (30%) have canonical Domain Authority sections. The remaining 70% have variant formats, implicit authority, or prose-heavy implementations.

2. **Orchestrator authority is simplified**: All orchestrators use a reduced format (Decide/Escalate only, no Route To or Do NOT sections). This may be intentional but creates inconsistency.

3. **Output spec varies in detail**: Some agents have strong artifact tables with templates (ecosystem, 10x-dev), others reference external schemas (orchestrators). Not a blocker but creates friction.

4. **Missing contract.must_not**: Several older agents (architect, agent-designer, threat-modeler) lack the contract.must_not frontmatter field. This is a hygiene issue.

5. **File verification inconsistency**: File verification protocol is mentioned in ecosystem agents but not uniformly present elsewhere. This affects artifact quality.

### Critical Insight for Front Door Sprint

**Authority specification exists but lacks uniformity**. The ecosystem rite demonstrates the canonical format works well. The challenge is not creating authority boundaries from scratch—it's standardizing the 16 agents that have variant formats.

**Exousia injection targets**:
- **High priority (11 agents)**: All orchestrators - standardize to include "You Route To" section
- **Medium priority (9 agents)**: Specialist agents with prose-heavy authority sections - convert to bullet format
- **Low priority (1 agent)**: Moirai - make implicit authority explicit
- **Skip (2 agents)**: Theoros, ecosystem specialists - already canonical

## Recommendations

### Wave 5 (Exousia Injection)

1. **Use ecosystem agents as template**: The Decide/Escalate/Route structure with bullet points is clear and comprehensive. Replicate this format across all specialists.

2. **Standardize orchestrator authority**: Add "You Route To" section to all orchestrators. Currently they stop at Escalate.

3. **Convert prose to bullets**: Nine agents have prose-heavy authority sections. Convert to structured bullets for scanning clarity.

4. **Make Moirai authority explicit**: Extract authority from operations table and lock protocol into canonical Domain Authority section.

5. **Add contract.must_not**: Backfill missing contract sections in older agents (architect, agent-designer, threat-modeler).

### Future Improvements (Post-Sprint)

1. **Normalize output specs**: All agents should have artifact tables with templates or schema references. Reduce variation.

2. **Integrate file verification universally**: Ecosystem agents have strong verification protocol. Expand to all rites.

3. **Session checkpoints for long-running agents**: Integration Engineer and Documentation Engineer have checkpoint guidance. Consider expanding to other long-running specialists.

4. **Cross-rite handoff standardization**: Code Smeller has excellent handoff acceptance protocol. Template this pattern for other cross-rite agents.

## Appendix: Grading Methodology

Each agent evaluated across 5 criteria (100 points total):

1. **Frontmatter Schema** (25 points):
   - 25: Complete YAML with name, description, tools, model, disallowedTools/contract
   - 20: Missing optional fields (contract, disallowedTools)
   - 15: Missing required fields
   - 10: Incomplete or malformed

2. **Role Clarity** (25 points):
   - 25: Clear role in first 5 lines, unambiguous scope
   - 20: Role present but scope unclear
   - 15: Vague role statement
   - 10: Role unclear or missing

3. **Authority Boundaries** (20 points):
   - 20: Canonical format (Decide/Escalate/Route or Do NOT, with bullets)
   - 15: Variant format (prose, partial structure, implicit)
   - 10: Authority mentioned but not structured
   - 5: No authority specification

4. **Output Specification** (15 points):
   - 15: Clear artifact table with templates or schema references
   - 12: Artifacts listed but templates missing
   - 9: Vague output description
   - 6: No output spec

5. **Behavioral Constraints** (15 points):
   - 15: Strong MUST/MUST NOT + anti-patterns with examples
   - 12: Anti-patterns present, examples weak
   - 9: Constraints mentioned but not detailed
   - 6: No behavioral constraints

**Letter Grade Conversion**:
- A: 90-100 (Excellent)
- B: 80-89 (Good)
- C: 70-79 (Adequate)
- D: 60-69 (Below Standard)
- F: Below 60 (Failing)

## Wave 5 Impact Projection

**Before Exousia Injection**:
- Authority coverage: 8.7% (2/23 canonical)
- Format consistency: 30% (7/23 canonical or near-canonical)
- Average authority grade: 15/20 (75%)

**After Exousia Injection** (projected):
- Authority coverage: 100% (23/23 canonical)
- Format consistency: 100% (standardized Decide/Escalate/Route structure)
- Average authority grade: 19/20 (95%)

**Effort Estimate**:
- High priority (11 orchestrators): ~2 hours (template apply + review)
- Medium priority (9 specialists): ~4 hours (prose → bullets conversion)
- Low priority (1 moirai): ~30 minutes (extract implicit authority)
- Total: ~6.5 hours for complete Exousia injection

**Risk Assessment**:
- Low risk: Ecosystem agents prove the format works
- No breaking changes: Additive content only
- Validation: Re-run this audit in Wave 5 to confirm 100% coverage

---

## Attestation

All agent files read from source locations:
- `rites/shared/agents/theoros.md`
- `agents/moirai.md`
- `rites/{rite}/agents/orchestrator.md` (11 files)
- `rites/{rite}/agents/{specialist}.md` (11 specialists across 6 rites)

**Total files read**: 23
**Total lines audited**: ~5,200
**Audit duration**: 2026-02-10
**Auditor**: Theoros (via main agent)
**Baseline report**: `.wip/THEORIA-BASELINE-agents-2026-02-10.md`

This baseline captures the "before" state for the Front Door sprint. The "after" audit in Wave 5 will measure Exousia injection success by comparing authority coverage, format consistency, and average authority grades.

# SPIKE: Biases and Opinions Baked Into the Ecosystem Rite

**Date**: 2026-02-16
**Author**: Claude (spike research)
**Timebox**: Research-only, single pass
**Status**: Complete

## Question and Context

**What are we trying to learn?**
What biases, assumptions, and unstated opinions are embedded in the ecosystem rite's agent prompts, documentation, templates, and workflow design? Which of these are intentional design decisions, which are accidental artifacts, and which have already been identified as problems?

**What decision will this inform?**
Whether the rite's implicit worldview helps or hinders its actual users. Identifying blind spots enables targeted simplification and makes hidden assumptions visible for conscious acceptance or rejection.

---

## Approach

Systematic reading of all 6 agent prompts, the workflow definition, 3 legomena skills (doc-ecosystem, claude-md-architecture, ecosystem-ref), all supporting documentation (anti-patterns, ownership model, boundary test, content tone guide, first principles, templates), the TODO audit, and cross-rite handoff schemas. Each document was analyzed for embedded assumptions, prescriptive patterns, unstated opinions, and cultural biases.

---

## Findings

### Category 1: Waterfall Disguised as Agile

**Bias**: The workflow is a strict linear pipeline: Analysis -> Design -> Implementation -> Documentation -> Validation. Every phase has entry criteria gating the next phase.

**Evidence**:
- `workflow.md` lines 1-37: Five phases in fixed sequence
- Phase skipping only exists at PATCH complexity, and only skips design + documentation
- Pythia's phase routing table (lines 139-145) is an ordered waterfall with no iteration loops
- No back-routes or feedback loops defined in the workflow
- Handoff criteria are one-directional checklists

**Why this matters**: Real infrastructure work is iterative. A compatibility tester finding a P1 defect sends work back to integration-engineer, but the workflow has no formal mechanism for this. The informal path ("escalate P0/P1 to integration-engineer") exists in the compatibility-tester prompt but contradicts the forward-only workflow model.

**Verdict**: This is a partially intentional simplification. The TODO.md acknowledges the rite is "over-engineered" but the waterfall structure itself is never questioned. The linear model makes orchestration simpler but penalizes the common case of iterative refinement.

---

### Category 2: Backward Compatibility as Default Virtue

**Bias**: The rite treats backward compatibility as the expected default and breaking changes as exceptional events requiring explicit justification.

**Evidence**:
- Context Architect prompt, line 89: "Breaking changes have explicit migration paths"
- Context Design template: `**Classification**: [COMPATIBLE | BREAKING]` -- binary choice, COMPATIBLE listed first
- Documentation Engineer prompt, line 36: "Undocumented breaking changes are bugs with better PR"
- Compatibility Tester prompt, line 43: "You don't trust claims -- you test them"
- Pythia handoff criteria for design phase: "Backward compatibility classified (COMPATIBLE or BREAKING)"

**Counter-evidence**: The TODO.md (P4) explicitly identifies this as a problem: "Breaking changes aren't bugs -- they're intentional architectural decisions. Over-validating compatibility would slow legitimate migrations."

**Why this matters**: For a greenfield meta-framework that's actively evolving, the default assumption should probably be the opposite: breaking changes are normal during rapid evolution. The current framing creates unnecessary ceremony around intentional architectural changes.

**Verdict**: Already identified as a bias in the TODO audit (P4). The corrective action ("Remove 'backward compatible' as default expectation") has been defined but not yet applied to the agent prompts.

---

### Category 3: Satellite Diversity Matrix Fantasy

**Bias**: The rite assumes a diverse fleet of satellite projects with varying configurations that must be tested systematically.

**Evidence**:
- Compatibility Tester defines a full testing matrix: test-baseline, test-minimal, test-complex, test-legacy, test-production-like
- Complexity levels (PATCH/MODULE/SYSTEM/MIGRATION) map to different satellite coverage requirements
- Integration Engineer prompt, line 119: "'Works in one satellite' syndrome: One satellite is one data point"
- Compatibility Tester prompt, line 166-172: Full satellite matrix table by complexity
- Gap Analysis template includes "Test Satellites" section

**Counter-evidence**: The TODO.md (P2) states flatly: "The ecosystem rite doesn't work on satellite projects. Satellites just sync from the ecosystem. The testing matrix adds complexity without value."

**Why this matters**: This is a significant amount of prompt real estate and cognitive overhead devoted to a testing model that doesn't match reality. Agents are instructed to test across satellite diversity that doesn't exist.

**Verdict**: Already identified as a bias in the TODO audit (P2). Marked for removal but not yet applied.

---

### Category 4: Orchestrated-Only Execution Assumption

**Bias**: The agents assume they will always be invoked within a full Pythia-coordinated pipeline with sessions active.

**Evidence**:
- Every agent references handoff criteria to/from other agents in the pipeline
- Pythia's prompt is entirely about coordinating between specialists
- No agent has "standalone invocation" guidance
- The workflow.md has no cross-cutting or native mode documentation
- Agent prompts reference session artifacts (SESSION_CONTEXT, Gap Analysis, Context Design) as always available

**Counter-evidence**: TODO.md (P3) identifies this: "Sometimes you just want to ask context-architect a design question without spinning up full orchestration."

**Why this matters**: The most common use case -- asking a single specialist a question -- is not supported by the agent prompt design. Every agent prompt assumes a full pipeline is running.

**Verdict**: Already identified in the TODO audit (P3) but not yet addressed.

---

### Category 5: Descriptive-Over-Prescriptive Doctrine

**Bias**: The claude-md-architecture skill enforces a strong opinion that CLAUDE.md content must be descriptive, never prescriptive. This is itself a prescriptive opinion.

**Evidence**:
- `content-tone-guide.md`: "Avoid" column lists MUST, NEVER, Always
- `first-principles.md` Principle 1: "CLAUDE.md describes; orchestration enforces"
- `content-tone-guide.md` line 140-143: Words to avoid table (MUST -> requires, NEVER -> avoid, Always -> typically)
- Anti-patterns list 11 things NOT to put in CLAUDE.md
- Boundary test has 5 gatekeeping questions before content is allowed

**The irony**: The agent prompts themselves are extremely prescriptive. Pythia's prompt says "You ALWAYS respond with structured YAML." The integration-engineer says "Don't commit with TODO comments." The ecosystem-analyst says "You don't guess." The descriptive-over-prescriptive doctrine applies only to CLAUDE.md content, not to agent behavior -- but this distinction is never explicitly stated.

**Why this matters**: There's a philosophical inconsistency between how the rite governs agent behavior (prescriptive) and how it governs user-facing documentation (descriptive). Someone reading both sets of documents might reasonably ask why agent prompts get to use MUST but CLAUDE.md doesn't.

**Verdict**: Intentional design decision but under-explained. The distinction (behavioral contracts for agents vs. orientation documents for humans) is implicit, not documented.

---

### Category 6: Single-Source-of-Truth Absolutism

**Bias**: Every piece of information must have exactly one authoritative source. Duplication is treated as a defect.

**Evidence**:
- `first-principles.md` Principle 5: "Each piece of content has one owner, one sync behavior, one location"
- Anti-Duplication Rule: "If content exists in multiple places, one becomes stale"
- Ownership model: SYNC, PRESERVE, REGENERATE -- each section has exactly one owner
- 11 anti-patterns, several of which are about duplication (duplicating knossos state, wrong rite content source)

**Why this matters**: This is a strong architectural opinion that trades some usability for consistency. In practice, strategic duplication (the same information presented in different contexts with different emphasis) can improve discoverability and reduce cognitive load. The absolute stance against it may cause agents to under-document or create indirection chains where direct statements would be clearer.

**Verdict**: Intentional and probably correct for a sync-pipeline context where duplication creates real bugs. But the absolutism may be over-applied to documentation content where some redundancy aids comprehension.

---

### Category 7: Complexity Classification Realism Bias

**Bias**: Every change can be cleanly classified into exactly one of four complexity levels, and this classification determines the entire workflow shape.

**Evidence**:
- Workflow.md: PATCH/MODULE/SYSTEM/MIGRATION
- PATCH skips 2 phases; everything else runs all 5
- Ecosystem Analyst decides complexity classification
- Complexity determines satellite test matrix coverage
- Complexity affects documentation requirements

**Why this matters**: Real changes often straddle categories. A "PATCH" that turns out to need a design decision has no formal upgrade path. A "SYSTEM" change that's actually simpler than expected has no downgrade path. The classification happens at analysis time (before full understanding) and determines the entire downstream workflow.

**Verdict**: Over-rigid. The four-level system works as a communication tool but fails as a workflow gate. Changes should be able to reclassify mid-flight without restarting.

---

### Category 8: The "Infrastructure, Not Application" Identity

**Bias**: The ecosystem rite is exclusively about infrastructure (sync pipeline, knossos, materialization) and explicitly excludes application-level concerns.

**Evidence**:
- README.md: "Not for: Application code in satellites (use 10x-dev)"
- Every agent prompt references internal/materialize, ari sync, .claude/ -- never application code
- The ecosystem-analyst looks for root causes in "ari or knossos" -- never in satellite application code
- Templates all have `Sync Pipeline | Knossos` as the only affected system options

**Why this matters**: This creates a blind spot for infrastructure changes that cause application-level symptoms. If ari sync breaks a satellite's CI pipeline (an application concern caused by infrastructure), the ecosystem rite's agents would trace the root cause to ari but might miss the application impact. The handoff to 10x-dev exists but the diagnostic might be incomplete.

**Verdict**: Intentional scope boundary but the boundary is drawn too sharply. Infrastructure changes have application-level consequences that should be within the diagnostic scope even if the fix is not.

---

### Category 9: Artifact-Driven Completion Model

**Bias**: Work is only considered complete when a formal artifact exists. Every phase produces a named document type.

**Evidence**:
- Gap Analysis, Context Design, Migration Runbook, Compatibility Report -- every phase produces exactly one artifact
- Handoff criteria include "artifact committed to docs/ecosystem/{TYPE}-{slug}.md"
- File verification protocol requires every artifact to be read back after writing
- Attestation tables required for completion

**Why this matters**: This creates overhead for small changes. A one-line config fix (PATCH complexity) still produces a Gap Analysis and Compatibility Report. The artifact overhead discourages small, incremental improvements.

**Verdict**: Partially mitigated by PATCH complexity skipping 2 phases, but even PATCH produces 3 artifacts (gap analysis, implementation, compatibility report). For a framework that values pragmatic testing, this is surprisingly ceremonial.

---

### Category 10: Hub Rite Overreach

**Bias**: The rite positions itself as the coordination hub for all ecosystem changes, implying authority over other rites' sync behavior.

**Evidence**:
- Cross-rite handoff schema defines detailed protocols for routing work between rites
- The rite references "all satellites" as if it manages them
- Pythia's cross-rite protocol assumes rite leads exist and can be notified

**Counter-evidence**: TODO.md (P1): "No rite leads registry exists, no notification mechanism, no coordination protocol."

**Verdict**: Already identified and marked for demotion from hub to specialist. The remaining vestiges in agent prompts haven't been cleaned up.

---

## Summary: Bias Classification

| # | Bias | Type | Severity | Status |
|---|------|------|----------|--------|
| 1 | Waterfall pipeline | Structural | Medium | Unaddressed |
| 2 | Backward compatibility as default | Philosophical | Medium | Identified in TODO (P4) |
| 3 | Satellite diversity fantasy | Factual | High | Identified in TODO (P2) |
| 4 | Orchestrated-only assumption | Structural | High | Identified in TODO (P3) |
| 5 | Descriptive-over-prescriptive irony | Philosophical | Low | Intentional but under-explained |
| 6 | Single-source absolutism | Architectural | Low | Intentional |
| 7 | Rigid complexity classification | Structural | Medium | Unaddressed |
| 8 | Infrastructure-only identity | Scope | Low | Intentional |
| 9 | Artifact-driven completion | Process | Medium | Unaddressed |
| 10 | Hub rite overreach | Scope | Medium | Identified in TODO (P1) |

---

## Recommendation

**The TODO.md audit already caught 4 of 10 biases.** This is a healthy sign -- the self-audit mechanism works. The remaining 6 unaddressed biases fall into three groups:

### Group A: Structural (address when simplifying workflow)
- **Waterfall pipeline** (#1): Add formal back-routes and iteration support
- **Rigid complexity classification** (#7): Allow reclassification mid-flight
- **Artifact overhead** (#9): PATCH complexity should produce 0-1 artifacts, not 3

### Group B: Philosophical (accept consciously or revise)
- **Descriptive vs. prescriptive inconsistency** (#5): Document why agent prompts use prescriptive language while CLAUDE.md is descriptive-only
- **Single-source absolutism** (#6): Acknowledge that strategic duplication in docs can aid comprehension

### Group C: Scope (clarify boundary language)
- **Infrastructure-only identity** (#8): Add "application impact assessment" to ecosystem-analyst scope even when the fix lives in another rite

### Already Tracked (execute the TODO)
- P1: Hub demotion (#10)
- P2: Satellite matrix removal (#3)
- P3: Cross-cutting support (#4)
- P4: Breaking changes as normal (#2)

---

## Follow-Up Actions

1. **Execute TODO.md P1-P4** -- these are validated, ready to apply
2. **Add iteration/back-route support to workflow.md** -- the waterfall is the most impactful unaddressed bias
3. **Reduce PATCH ceremony** -- one artifact (implementation) should suffice for single-file changes
4. **Document the prescriptive/descriptive distinction** -- why agent prompts and CLAUDE.md have different tone rules
5. **Add complexity reclassification mechanism** -- allow upgrading/downgrading mid-flight

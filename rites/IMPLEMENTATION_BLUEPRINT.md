# Roster Ecosystem Implementation Blueprint

> Synthesized from 3 deep-dive analyses | 2026-01-02
> Pattern Extraction | Dependency Sequencing | Context Engineering

---

## The Eigenvalue Insight

After analyzing 23 improvements across 10 teams, one change provides **80% of the value**:

> **Create the Generalized HANDOFF Artifact as an Ecosystem-Level Primitive**

This single change:
- Eliminates 8+ different handoff formats
- Prevents N² format explosion with new teams
- Enables cross-rite workflow automation
- Provides consistent routing language
- Creates foundation for all other coordination improvements

Everything else builds on or benefits from this primitive.

---

## Five Architectural Primitives

The audit revealed five missing primitives that, once added, transform 23 point-fixes into a coherent system:

| Primitive | Purpose | Teams Affected | Compression |
|-----------|---------|----------------|-------------|
| **HANDOFF Artifact** | Standardize all cross-rite transitions | 10+ team pairs | 8→1 formats |
| **Boundary Matrix** | Explicit routing for overlapping concerns | 3 team pairs | 6→1 docs |
| **Proactive Gate Registry** | Downstream teams inject into upstream workflows | security, docs | Pattern not ad-hoc |
| **Shared Detection Skill** | Consolidated smell/debt patterns | debt, hygiene | 2→1 definitions |
| **Shared Templates Skill** | Multi-team templates at ecosystem level | debt, hygiene, sre | 3→1 ownership |

---

## Core Principles for Implementation

### 1. Specification Through Abstraction
Define schemas, not scripts. Examples trump rules. Let teams self-organize within shared primitives.

### 2. Preserve Autonomy While Enabling Coordination
Shared infrastructure doesn't reduce team autonomy—it enables it. HANDOFF is a contract, not control.

### 3. Centralize Metadata, Decentralize Execution
The ecosystem needs shared knowledge of what teams do (routing metadata), not a hub controlling what they do.

### 4. Make Implicit Boundaries Explicit
Document *why* teams are separate, *what* makes them different, and *when* you need both.

### 5. Treat Agent Quality as Ecosystem Risk
Weak agents are rite-level performance ceilings. Quality investment precedes scope expansion.

### 6. Inverse Conway Law
Handoff complexity reflects coordination complexity. Shared primitives grow *with* the ecosystem, not *against* it.

---

## Optimized Sprint Structure

### Sprint 0: Foundation (Week 1)
**Goal:** Clear quick wins, fix bugs, improve agent quality
**Parallelization:** 7 items across 5 teams simultaneously

| Item | Team | Effort | Dependencies |
|------|------|--------|--------------|
| P0: Fix model assignment bug | strategy-pack | 5 min | None |
| P1: Fix user-researcher agent | intelligence-pack | 3 hrs | None |
| P2: Fix insights-analyst agent | intelligence-pack | 3 hrs | None |
| P1: Fix ship-pack references | rnd-pack | 1 hr | None |
| P1: Demote from hub | ecosystem-pack | 2 hrs | None |
| P2: Add staleness detection | doc-rite-pack | 2 hrs | None |
| P1: Define behavior preservation | hygiene-pack | 2 hrs | None |

**Value Delivered:** +10% ecosystem quality, bugs fixed, foundation definitions in place
**Token Impact:** -50 tokens (ecosystem demotion removes bloat)

---

### Sprint 1: Shared Infrastructure (Weeks 2-4)
**Goal:** Create the three shared primitives that all other improvements depend on
**Critical Path:** This is serialized. Must complete before Sprint 2.

```
Week 2: Design Phase
├── HANDOFF artifact schema design (review with 3+ teams)
├── Smell detection patterns audit (what debt + hygiene share)
└── Template ownership analysis (which templates are truly shared)

Week 3: Creation Phase
├── Create skills/shared/smell-detection/ (ecosystem level)
├── Create skills/shared/cross-rite-handoff/ (ecosystem level)
└── Create skills/shared/templates/ (ecosystem level)

Week 4: Integration Phase
├── debt-collector → smell-detection reference
├── code-smeller → smell-detection reference
├── sprint-planner → HANDOFF artifact output
└── Move templates from doc-sre to shared-templates
```

**Value Delivered:** Shared infrastructure enables 7 downstream integrations
**Token Impact:** +600 tokens initially, amortizes to -200 after 2 uses each

---

### Sprint 2: 10x-Dev-Pack Coordination (Weeks 5-6)
**Goal:** Make 10x the flexible hub for downstream teams
**Dependency:** Sprint 1 complete (HANDOFF schema available)

```
Week 5: 10x Core Changes
├── P1: Impact assessment for requirements-analyst
├── P2: Flexible entry points in orchestrator
└── P3: Cross-rite handoff protocols

Week 6: Dependent Team Gates (parallel)
├── security-pack P1: Proactive threat-modeler gate
├── doc-rite-pack P1: Documentation gate in QA checklist
└── rnd-pack P2: Spike overlap clarification
```

**Value Delivered:** 10-15% faster velocity on small changes, 0 security/doc surprises
**Token Impact:** +200 tokens when gates triggered (conditional)

**MVI Threshold Crossed:** After Sprint 2, ecosystem delivers tangible improvement.

---

### Sprint 3: Cross-Team Formalization (Weeks 7-8)
**Goal:** Deploy HANDOFF pattern across all team pairs, clarify boundaries
**Parallelization:** Two workstreams, 4-5 threads

**Workstream A: Boundary Clarification**
```
├── intelligence-pack P3 + strategy-pack P1: Mirror boundary table
├── strategy-pack P2: Define R&D integration pathway
└── strategy-pack P3: Add missing back-routes
```

**Workstream B: HANDOFF Rollout**
```
├── hygiene-pack P2: Accept HANDOFF from debt-triage
├── sre-pack P2: Accept HANDOFF from 10x
├── security-pack P3: Accept HANDOFF from 10x
├── intelligence-pack P4: Produce HANDOFF for 10x/strategy
├── rnd-pack P3: Produce HANDOFF for 10x/strategy
└── strategy-pack P4: Produce HANDOFF for 10x
```

**Value Delivered:** All cross-rite handoffs formalized, routing confusion eliminated
**Token Impact:** -200 tokens per handoff (structured vs. ambiguous)

---

### Sprint 4: Stabilization (Week 9)
**Goal:** Integration testing, operational playbooks, edge case documentation

```
├── End-to-end workflow tests (3 representative scenarios)
├── Cross-rite coordination playbooks
├── Edge case documentation
└── Smoke tests for all handoff paths
```

**Value Delivered:** System-level confidence, operational documentation
**Token Impact:** Neutral (testing overhead balanced by reduced debugging)

---

## Beautiful Architecture: The Target State

### Skill Hierarchy (Progressive Disclosure)

```
.claude/skills/
├── core/                          # Always loaded (~200 tokens)
│   ├── orchestration/
│   ├── consult-ref/
│   └── cross-rite-handoff/        # NEW
│
├── shared/                        # Loaded on demand (~300 tokens each)
│   ├── smell-detection/           # NEW
│   ├── shared-templates/          # NEW
│   └── doc-artifacts/
│
└── rite-specific/                 # Loaded when team active
    ├── 10x-workflow/
    ├── doc-sre/
    ├── strategy-ref/
    └── ...
```

**Token Economics:**
- Cold start: Load `core/` only (200 tokens)
- Team activation: Add rite-specific (1500 tokens)
- Cross-rite transition: Add `shared/` once, cache (300 tokens, amortized)

### HANDOFF Artifact Design

**Format:** YAML frontmatter + Markdown body (best of both worlds)

```yaml
---
source_team: debt-triage-pack
target_team: hygiene-pack
handoff_type: execution
created: 2026-01-02
initiative: "Q1 Technical Debt Remediation"
---

# Handoff Summary

Sprint planner scored these as high-impact, low-effort refactoring targets.
Risk assessor validated blast radius is contained to validator modules.

## Items

### PKG-001: Extract shared email validator
- **Priority:** High
- **Location:** `src/api/validators/`
- **Acceptance Criteria:**
  - Single `validateEmail` function in shared module
  - All 3 API files import shared validator
  - Existing tests pass unchanged

### PKG-002: Consolidate date formatters
- **Priority:** Medium
- **Location:** `src/utils/dates/`
- **Acceptance Criteria:**
  - Unified DateFormatter class
  - ISO 8601 default output

## Notes for Receiving Team

Recommend starting with PKG-001 as it unblocks PKG-002.
Total estimated effort: 4-6 hours.
```

**Why This Format:**
- YAML header enables schema validation
- Markdown body optimizes Claude's comprehension
- Structured items prevent "forgot to include X"
- Notes section preserves institutional knowledge

### Proactive Gate Pattern

```yaml
# gates/security-threat-modeling.yaml
gate_id: threat-modeling-pre-tdd
trigger:
  upstream_team: 10x-dev-pack
  upstream_phase: design
  condition: complexity == SYSTEM AND (auth OR crypto OR pii)

downstream_team: security-pack
downstream_agent: threat-modeler
gate_type: consultation
time_budget: 2-4 hours

output: THREAT-*.md
incorporation:
  target: TDD security section
  binding: advisory
```

**Why Gates as Primitives:**
- New teams can register gates without modifying 10x
- Gates are discoverable and documented
- Orchestrator can auto-route based on registry

### Team Boundary Matrix Pattern

```markdown
## When to Use [Team A] vs [Team B]

| Dimension | Team A | Team B |
|-----------|--------|--------|
| Core question | "..." | "..." |
| Work scope | ... | ... |
| Success metric | ... | ... |
| Example scenarios | ... | ... |

**Rule of thumb:**
- Team A = [domain axis]
- Team B = [orthogonal axis]

**When you need both:** [cross-rite scenario]
```

**Where Applied:**
- Intelligence vs Strategy (product analytics vs market analysis)
- RND vs 10x /spike (exploration vs evaluation)
- Debt-triage vs Hygiene (planning vs execution)

---

## Context Engineering: Session Strategy

### Implementation Session Map

| Session | Focus | Duration | Token Budget |
|---------|-------|----------|--------------|
| S1 | Sprint 0: Bug fixes + agent quality | 3-4 hrs | 2500 |
| S2 | Sprint 1: Shared skill creation | 4-5 hrs | 3000 |
| S3 | Sprint 1: Integration (debt + hygiene) | 2-3 hrs | 2800 |
| S4 | Sprint 2: 10x core changes | 3-4 hrs | 2500 |
| S5 | Sprint 2: Downstream gates (parallel) | 2-3 hrs | 2200 |
| S6 | Sprint 3A: Boundary clarifications | 2-3 hrs | 2000 |
| S7 | Sprint 3B: HANDOFF rollout | 3-4 hrs | 2500 |
| S8 | Sprint 4: Validation + playbooks | 3 hrs | 2000 |

**Total:** 8 focused sessions, ~25-30 hours of implementation

### Context Preservation Between Sessions

Use `SESSION_CONTEXT.md` checkpoints:

```yaml
checkpoint: sprint-1-complete
completed_artifacts:
  - skills/shared/smell-detection/SKILL.md
  - skills/shared/cross-rite-handoff/SKILL.md
  - skills/shared/templates/SKILL.md
next_phase: sprint-2-10x-changes
dependencies_satisfied:
  - HANDOFF schema defined
  - Smell detection patterns available
```

---

## Value Delivery Timeline

```
Week 1 (Sprint 0)
├── Agent quality: +10%
├── Bugs fixed: 2
└── Foundation definitions: 2

Week 4 (Sprint 1 complete)
├── Shared infrastructure: 3 new skills
├── Token savings potential: -50% on cross-rite
└── 7 teams ready for integration

Week 6 (Sprint 2 complete) ← MVI THRESHOLD
├── Small project velocity: +10-15%
├── Security surprises: 0 (proactive gates)
├── Documentation gaps: 0 (proactive gates)
└── Cross-rite routing: Clear

Week 8 (Sprint 3 complete)
├── All handoffs formalized: 10+ team pairs
├── Boundary confusion: Eliminated
├── New team onboarding: Pattern-based

Week 9 (Sprint 4 complete)
├── System confidence: High
├── Operational playbooks: Complete
└── Edge cases: Documented
```

---

## Compression Summary

| Before (23 items) | After (5 primitives) | Compression |
|-------------------|---------------------|-------------|
| 8 handoff formats | 1 HANDOFF schema | 8→1 |
| 6 boundary docs | 1 matrix template | 6→1 |
| 2 detection patterns | 1 shared skill | 2→1 |
| 3 template locations | 1 shared skill | 3→1 |
| 2 ad-hoc gates | 1 gate registry | Pattern |

**Total:** 23 point-fixes compressed to 5 primitives + 10 integrations

---

## Success Criteria

### Sprint 0 Complete When:
- [ ] All P0 bugs fixed (strategy model assignment)
- [ ] Intelligence agent scores 23+/30
- [ ] Ship-pack references eliminated
- [ ] Behavior preservation defined

### Sprint 1 Complete When:
- [ ] HANDOFF schema documented and reviewed by 3+ teams
- [ ] Smell detection skill working for both debt-collector and code-smeller
- [ ] Shared templates accessible from ecosystem level
- [ ] Template references updated in all consuming teams

### Sprint 2 Complete When:
- [ ] 10x orchestrator accepts flexible entry points
- [ ] Impact assessment integrated for PATCH complexity
- [ ] Security threat-modeling gate operational
- [ ] Documentation gate in QA checklist

### Sprint 3 Complete When:
- [ ] Intelligence/Strategy boundary table in both READMEs
- [ ] All 6 HANDOFF producers updated
- [ ] All 4 HANDOFF consumers updated
- [ ] R&D integration pathway documented

### Sprint 4 Complete When:
- [ ] 3 end-to-end workflows tested
- [ ] Cross-rite playbooks written
- [ ] Edge cases documented
- [ ] Smoke tests passing

---

## The Beautiful Architecture

When complete, the roster ecosystem will have:

1. **Uniform Coordination Language** - Every cross-rite transition speaks HANDOFF
2. **Progressive Disclosure** - Skills load in tiers: core → shared → rite-specific
3. **Explicit Boundaries** - Routing tables eliminate "which team?" confusion
4. **Proactive Integration** - Downstream teams engage upstream via registered gates
5. **Shared Foundation** - Detection patterns and templates owned collectively
6. **Quality as Priority** - Weak agents block scope expansion, not just flagged

The architecture doesn't just fix 23 problems—it prevents similar problems from emerging as the ecosystem grows.

---

## Next Actions

1. **Approve this blueprint** (user decision point)
2. **Begin Sprint 0** (no dependencies, immediate start)
3. **Schedule Sprint 1 design review** (HANDOFF schema with 3+ team representatives)
4. **Assign session ownership** (8 sessions mapped above)

The ecosystem will reach MVI by Week 6, full maturity by Week 9.

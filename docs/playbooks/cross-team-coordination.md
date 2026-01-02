# Cross-Team Coordination Playbook

> Operational guidance for producing and consuming cross-team HANDOFF artifacts.
> Version: 1.0.0

## Overview

This playbook provides practical guidance for cross-team coordination in the roster ecosystem. It covers when to use HANDOFF artifacts, how to produce and consume them, common scenarios, and escalation paths.

---

## When to Use Cross-Team HANDOFF

### Use HANDOFF When

| Situation | Why HANDOFF |
|-----------|-------------|
| Your team's work is complete and another team should continue | Formal context transfer, clear ownership |
| Specialist review is required before proceeding | Blocking gate, documented assessment |
| Cross-team coordination is needed for shared initiative | Single source of truth for work items |
| Work originated elsewhere and you're receiving it | Structured intake, clear expectations |

### Do NOT Use HANDOFF When

| Situation | What Instead |
|-----------|--------------|
| Quick question for another team | Direct consultation via user |
| Internal phase transition within same team | Session state transition |
| Information sharing without action items | Documentation or meeting notes |
| Escalation without work transfer | Escalation protocol |

---

## HANDOFF Types Quick Reference

| Type | Flow | Key Requirement | Example |
|------|------|-----------------|---------|
| `execution` | Planning -> Execution | Acceptance criteria | debt-triage -> hygiene |
| `validation` | Dev -> Ops | Validation scope | 10x -> sre |
| `assessment` | Dev -> Specialist | Assessment questions | 10x -> security |
| `implementation` | Research -> Dev | Design references | strategy -> 10x |
| `strategic_input` | Research -> Strategy | Data sources | intelligence -> strategy |
| `strategic_evaluation` | R&D -> Strategy | Evaluation criteria | rnd -> strategy |

---

## Producing a HANDOFF

### Step-by-Step Process

```
1. IDENTIFY   - Determine handoff type and target team
2. GATHER     - Collect source artifacts and context
3. STRUCTURE  - Format items with required fields per type
4. VALIDATE   - Check against schema requirements
5. STORE      - Save in appropriate location
6. NOTIFY     - Inform user for routing to target team
```

### Required Fields by Type

#### All Types (Required)
```yaml
source_team: [your-team-pack]
target_team: [receiving-team-pack]
handoff_type: [type]
created: [YYYY-MM-DD]
initiative: [initiative name]
```

#### Type-Specific Requirements

| Type | Required Per Item |
|------|-------------------|
| `execution` | `Acceptance Criteria` |
| `validation` | `Validation Scope` |
| `assessment` | `Assessment Questions` |
| `implementation` | `Design References` |
| `strategic_input` | `Data Sources`, `Confidence` |
| `strategic_evaluation` | `Evaluation Criteria` |

### Quality Checklist for Producers

- [ ] Frontmatter complete and valid
- [ ] Context section explains why handoff exists
- [ ] Source artifacts listed with paths
- [ ] Each item has unique ID (e.g., PKG-001, SEC-001)
- [ ] Each item has priority (Critical/High/Medium/Low)
- [ ] Each item has summary (1-2 sentences)
- [ ] Type-specific fields present per item
- [ ] Notes section provides actionable guidance
- [ ] Blocking flag set if downstream cannot proceed

### Common Producer Mistakes

| Mistake | Impact | Fix |
|---------|--------|-----|
| Missing acceptance criteria | Target team can't verify completion | Add testable criteria |
| Vague summaries | Scope misunderstanding | Be concrete and specific |
| No source artifacts | Context lost | Always link source work |
| Wrong handoff type | Validation fails | Match type to flow pattern |
| Self-handoff | Invalid | Handoffs are between different teams |

---

## Consuming a HANDOFF

### Step-by-Step Process

```
1. INTAKE     - Receive HANDOFF from user/routing
2. REVIEW     - Read frontmatter for type and priority
3. CONTEXT    - Read Context section for background
4. ARTIFACTS  - Check Source Artifacts for references
5. PLAN       - Process Items as work queue
6. EXECUTE    - Complete work per acceptance criteria
7. RESPOND    - Produce output or return handoff
```

### Consumer Checklist

- [ ] Understand source team and why they handed off
- [ ] Review priority and blocking status
- [ ] Access all source artifacts
- [ ] Understand scope of each item
- [ ] Clarify assessment questions if unclear
- [ ] Plan work based on dependencies in Notes
- [ ] Produce response artifact when complete

### Handling Incomplete HANDOFFs

| Issue | Action |
|-------|--------|
| Missing source artifacts | Request from source team via user |
| Unclear acceptance criteria | Seek clarification before starting |
| Wrong handoff type | Notify source team, request correction |
| Missing context | Ask for background before proceeding |

---

## Common Scenarios

### Scenario 1: Feature Development with Security Gate

**Flow**: 10x-dev-pack -> security-pack -> 10x-dev-pack

1. **10x-dev-pack** completes PRD for payment processing
2. **Trigger**: SYSTEM complexity + security considerations detected
3. **Produce**: HANDOFF (assessment) to security-pack
   ```yaml
   handoff_type: assessment
   priority: critical
   blocking: true
   ```
4. **security-pack** produces threat model
5. **Return**: Threat model with verdict (APPROVED/BLOCKED)
6. **10x-dev-pack** continues with design phase

**Key Points**:
- `blocking: true` prevents design from starting
- Threat model becomes source artifact for TDD
- Security mitigations must appear in TDD

### Scenario 2: Debt Remediation

**Flow**: debt-triage-pack -> hygiene-pack

1. **debt-triage-pack** completes debt collection, assessment, planning
2. **Produce**: HANDOFF (execution) to hygiene-pack
   ```yaml
   handoff_type: execution
   priority: high
   ```
3. **hygiene-pack** executes sprint packages
4. **Output**: Remediation report with behavior preservation confirmation

**Key Points**:
- Each package has acceptance criteria
- Behavior preservation checklists required
- Audit signoff before completion

### Scenario 3: Research to Production

**Flow**: rnd-pack -> strategy-pack -> 10x-dev-pack

1. **rnd-pack** completes spike/prototype
2. **Produce**: HANDOFF (strategic_evaluation) to strategy-pack
   ```yaml
   handoff_type: strategic_evaluation
   priority: medium
   ```
3. **strategy-pack** evaluates viability, makes go/no-go decision
4. **If GO**: Produce HANDOFF (implementation) to 10x-dev-pack
   ```yaml
   handoff_type: implementation
   priority: high
   ```
5. **10x-dev-pack** builds production system

**Key Points**:
- R&D never directly hands to 10x-dev-pack
- Strategy gate ensures business alignment
- Implementation handoff includes design references

### Scenario 4: Production Readiness

**Flow**: 10x-dev-pack -> sre-pack

1. **10x-dev-pack** completes feature + QA
2. **Produce**: HANDOFF (validation) to sre-pack
   ```yaml
   handoff_type: validation
   priority: high
   ```
3. **sre-pack** validates production readiness
4. **Output**: Validation report with deployment approval

**Key Points**:
- Validation scope must be clear
- SRE checks observability, reliability, scalability
- Approval required before production deployment

### Scenario 5: Documentation Handoff

**Flow**: 10x-dev-pack -> doc-team-pack

1. **10x-dev-pack** completes feature
2. **Produce**: HANDOFF (assessment) to doc-team-pack
   ```yaml
   handoff_type: assessment
   priority: medium
   ```
3. **doc-team-pack** assesses documentation needs
4. **Output**: Documentation or doc-review

**Key Points**:
- Assessment questions guide doc scope
- May be optional depending on feature visibility
- Not blocking for deployment

---

## Routing Decisions

### Decision Tree: Which Team?

```
Is this about security/compliance?
  YES -> security-pack
  NO  -> Continue

Is this about production operations?
  YES -> sre-pack
  NO  -> Continue

Is this about technical debt?
  YES -> Is it planning or execution?
    Planning -> debt-triage-pack
    Execution -> hygiene-pack
  NO  -> Continue

Is this about documentation?
  YES -> doc-team-pack
  NO  -> Continue

Is this about research/exploration?
  YES -> rnd-pack
  NO  -> Continue

Is this about strategic direction?
  YES -> strategy-pack
  NO  -> Continue

Is this about market/user intelligence?
  YES -> intelligence-pack
  NO  -> Continue

Default: 10x-dev-pack (feature development)
```

### Multi-Team Coordination

When work requires multiple teams:

1. **Sequential**: Use chained handoffs
   ```
   Team A -> HANDOFF -> Team B -> HANDOFF -> Team C
   ```

2. **Parallel**: User coordinates multiple handoffs
   ```
   Team A -> HANDOFF -> Team B (security)
          -> HANDOFF -> Team C (docs)
   ```

3. **Hub**: Use strategy-pack or 10x orchestrator as coordinator
   ```
   Orchestrator coordinates multiple team handoffs
   ```

---

## Escalation Paths

### When to Escalate

| Situation | Escalate To |
|-----------|-------------|
| Blocking handoff not responded to | User (for routing) |
| Target team rejects handoff | Source team + user |
| Priority disagreement | User (for arbitration) |
| Schema validation failure | Ecosystem-pack (schema issue) |
| Cross-cutting initiative needs | Orchestrator |

### Escalation Process

1. **Document** the issue clearly
2. **Include** original HANDOFF reference
3. **State** what action is needed
4. **Propose** resolution if possible

### Urgent Bypass

For urgent situations requiring bypass of normal handoff flow:

```yaml
priority: critical
blocking: true
# Add in Notes:
## URGENT
Timeline: [specific deadline]
Escalation: Bypass normal queue, requires immediate attention
Justification: [why this is urgent]
```

---

## Handoff File Locations

### Naming Convention
```
HANDOFF-[source]-to-[target]-[date].md
```

### Storage Locations

| Context | Location |
|---------|----------|
| Active session | `.claude/sessions/{session-id}/` |
| Initiative docs | `docs/handoffs/` |
| Sprint context | `.claude/sprints/{sprint-id}/` |

### Discovery

To find existing handoffs:
```bash
# All handoffs
find . -name "HANDOFF-*.md"

# Handoffs to specific team
find . -name "HANDOFF-*-to-security-*.md"

# Recent handoffs
find . -name "HANDOFF-*.md" -mtime -7
```

---

## Metrics and Monitoring

### Handoff Health Indicators

| Metric | Healthy | Warning | Critical |
|--------|---------|---------|----------|
| Response time (blocking) | <24h | 24-48h | >48h |
| Response time (normal) | <72h | 72-120h | >120h |
| Rejection rate | <5% | 5-15% | >15% |
| Incomplete handoffs | <10% | 10-25% | >25% |

### Tracking

Track handoff status in session context:
```yaml
handoffs:
  - id: HANDOFF-10x-to-security-2026-01-02
    status: pending_response
    created: 2026-01-02
    blocking: true
```

---

## Anti-Patterns

### Producer Anti-Patterns

| Anti-Pattern | Problem | Solution |
|--------------|---------|----------|
| Kitchen sink handoff | Too many unrelated items | Split into focused handoffs |
| Missing blocking flag | Downstream proceeds prematurely | Set `blocking: true` when needed |
| Orphaned handoff | No one picks it up | Ensure user routes to target team |
| Version mismatch | References outdated artifacts | Use latest artifact versions |

### Consumer Anti-Patterns

| Anti-Pattern | Problem | Solution |
|--------------|---------|----------|
| Cherry picking | Incomplete work | Process all items or reject formally |
| Silent rejection | Source team unaware | Communicate rejection with reason |
| Scope creep | Work expands beyond handoff | Stick to handoff scope, create new handoff for additions |
| No output | Work done but not documented | Always produce response artifact |

---

## Quick Reference Card

### Producing (5 steps)
1. Choose handoff type
2. Fill frontmatter
3. Write Context
4. Structure Items with type-specific fields
5. Add Notes for target team

### Consuming (5 steps)
1. Check priority and blocking status
2. Read Context and Source Artifacts
3. Plan work from Items
4. Execute per acceptance criteria
5. Produce response

### Common Handoff Pairs
```
10x -> security : assessment (threat modeling)
10x -> sre      : validation (production readiness)
10x -> doc-team : assessment (documentation)
debt-triage -> hygiene : execution (debt remediation)
rnd -> strategy : strategic_evaluation (viability)
strategy -> 10x : implementation (go-to-market)
intelligence -> strategy : strategic_input (research)
```

---

## Related Documents

- [Cross-Team Handoff Schema](../../.claude/skills/shared/cross-team-handoff/schema.md)
- [Cross-Team Handoff SKILL](../../.claude/skills/shared/cross-team-handoff/SKILL.md)
- [Edge Cases: Cross-Team Workflows](../edge-cases/cross-team-workflows.md)
- [E2E Test: Feature Development](../testing/e2e-feature-development.md)
- [E2E Test: Security Workflow](../testing/e2e-security-workflow.md)
- [E2E Test: Debt Remediation](../testing/e2e-debt-remediation.md)

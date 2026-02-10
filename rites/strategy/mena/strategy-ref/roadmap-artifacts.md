# Roadmap Artifacts

> HANDOFF examples, RICE scoring templates, and strategic roadmap artifact patterns.

## HANDOFF Example (to 10x-dev)

```yaml
---
source_rite: strategy
target_rite: 10x-dev
handoff_type: implementation
created: 2026-01-02
initiative: Q1 Enterprise Expansion
priority: high
---

## Context

Q1 strategic roadmap prioritized enterprise tier as highest-value initiative. Resource allocation confirmed (3 engineers, Jan-Feb). OKRs established with measurable key results.

## Source Artifacts
- docs/strategy/ROADMAP-Q1-2026.md
- docs/strategy/PRIORITIZATION-enterprise-analysis.md
- docs/strategy/OKR-Q1-2026.md

## Items

### IMP-001: Enterprise tier implementation
- **Priority**: High (Rank #1 in Q1 roadmap)
- **Summary**: Build enterprise-grade subscription tier with team features
- **Strategic Rationale**:
  - RICE score: 30 (highest among Q1 candidates)
  - Revenue impact: Path to $100K MRR by Q2
  - Market timing: Competitor gaps in enterprise segment
- **OKR Alignment**:
  - O1: Accelerate revenue growth
  - KR1: Launch enterprise tier by Feb 15
  - KR2: Sign 5 enterprise customers by Mar 31
- **Resource Allocation**: 3 engineers, 6 weeks
- **Dependencies**: Auth system (complete), billing integration (in progress)
- **Acceptance Criteria**:
  - Rite management (invite, roles, permissions)
  - SSO integration (SAML, OAuth)
  - Usage-based billing
  - Admin dashboard

### IMP-002: API v2 quick win
- **Priority**: Medium (Rank #2 in Q1 roadmap)
- **Summary**: Modernize API for developer experience
- **Strategic Rationale**:
  - RICE score: 150 (quick win, high reach)
  - Developer ecosystem enablement
- **OKR Alignment**:
  - O2: Improve developer adoption
  - KR1: API v2 launched by Jan 31
- **Resource Allocation**: 2 engineers, 3 weeks
- **Dependencies**: None
- **Acceptance Criteria**:
  - RESTful design following OpenAPI 3.0
  - Rate limiting and authentication
  - Interactive documentation

## Notes for Target Rite

Enterprise tier is the critical path to Q2 revenue targets. API v2 is a quick win that unblocks future mobile initiative (deferred to Q2). Recommend parallel execution with separate engineering tracks.

Stakeholder contact: @product-lead for enterprise requirements clarification.
```

### Content Guidelines for Strategic HANDOFFs

1. **Strategic Rationale**: Always include the "why" from roadmap prioritization
2. **OKR Alignment**: Connect implementation to measurable business outcomes
3. **Resource Allocation**: Specify committed resources and timeline
4. **Dependencies**: Map what must complete first
5. **Acceptance Criteria**: Translate strategic requirements to implementable specs

---

## RICE Scoring Example

**Scenario**: Create Q1 roadmap for product team with 5 engineers

**Input**: Financial model showing $50K/mo burn, 3 initiatives proposed, need to hit $100K MRR by Q2

**Output (excerpt from Strategic Roadmap)**:
```markdown
## Prioritization Matrix (RICE Framework)

| Initiative | Reach | Impact | Confidence | Effort | Score | Rank |
|------------|-------|--------|------------|--------|-------|------|
| Enterprise tier | 100 | 3 | 80% | 8 | 30 | 1 |
| API v2 | 500 | 2 | 60% | 4 | 150 | 2 |
| Mobile app | 1000 | 1 | 40% | 12 | 33 | 3 |

**Decision**: Ship Enterprise tier (highest revenue impact) and API v2 (quick win). Defer mobile.

## Resource Allocation

| Initiative | Engineers | Timeline | Dependencies |
|------------|-----------|----------|--------------|
| Enterprise tier | 3 | Jan-Feb | Auth system (complete) |
| API v2 | 2 | Jan | None |
| Mobile app | 0 | Deferred to Q2 | API v2 |

## Q1 OKRs

**O1: Accelerate revenue growth**
- KR1: Launch enterprise tier by Feb 15
- KR2: Sign 5 enterprise customers by Mar 31
- KR3: Reach $80K MRR by end of Q1 (path to $100K)
```

**Why**: Framework applied consistently. Trade-offs explicit. Resource allocation realistic. OKRs measurable and connected to strategy.

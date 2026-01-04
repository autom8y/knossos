# Team Boundary Gap Analysis

> Analysis of team boundary clarity, overlaps, gaps, and misrouting patterns
> Session: session-20260104-013028-45c9ccac | Task: task-001
> Generated: 2026-01-04

## Executive Summary

This analysis audits 11 team packs across the roster ecosystem to identify boundary issues affecting routing accuracy and developer discoverability. Key findings:

- **5 boundary overlaps** identified where multiple teams claim similar territory
- **4 boundary gaps** identified where work falls between teams
- **5 common misrouting scenarios** documented with mitigation recommendations
- **All 11 teams** require enhanced "NOT for" sections for clarity

---

## Team Boundary Inventory (Current State)

### 1. 10x-dev-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Software development |
| **Triggers** | "Build a new feature", "PRD and technical design", "requirements to tested code", "implement API/service/module" |
| **Not for** | Documentation work, infrastructure automation, one-off scripts without testing requirements |
| **Produces** | PRD, TDD, ADRs, code, tests |
| **Complexity** | SCRIPT, MODULE, SERVICE, PLATFORM |

**Boundary Clarity**: Medium - Overlaps with hygiene-pack on refactoring, rnd-pack on spikes

---

### 2. doc-team-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Documentation lifecycle |
| **Triggers** | "Documentation scattered/inconsistent", "document this feature/system/API", "audit documentation", "reorganize docs" |
| **Not for** | Code implementation, infrastructure work, writing code comments |
| **Produces** | Audit report, doc structure, documentation, review signoff |
| **Complexity** | PAGE, SECTION, SITE |

**Boundary Clarity**: High - Clear separation from code-focused teams

---

### 3. hygiene-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Code quality |
| **Triggers** | "Codebase feels messy", "dead code/unused imports", "technical debt inventory", "refactoring", "cleanup commits" |
| **Not for** | New features (behavior changes), ecosystem infrastructure, quick formatting fixes |
| **Produces** | Smell Report, Refactoring Plan, Commit Stream, Audit Report |
| **Complexity** | SPOT, MODULE, CODEBASE |

**Boundary Clarity**: Medium - Overlaps with debt-triage-pack on debt identification

---

### 4. debt-triage-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Technical debt management |
| **Triggers** | "What technical debt do we have?", "prioritize debt paydown", "biggest technical risk", "debt cleanup sprint", "inherited codebase debt" |
| **Not for** | Feature development, active incidents, ongoing reliability work |
| **Produces** | Debt Ledger, Risk Report, Sprint Plan |
| **Complexity** | QUICK, AUDIT |

**Boundary Clarity**: Medium - Overlaps with hygiene-pack (assessment vs execution)

---

### 5. rnd-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Technology exploration |
| **Triggers** | "Evaluate new technology", "integrate with current stack", "proof-of-concept", "architecture in 2 years" |
| **Not for** | Production feature development, immediate shipping |
| **Produces** | tech-assessment, integration-map, prototype, moonshot-plan, TRANSFER, HANDOFF |
| **Complexity** | SPIKE, EVALUATION, MOONSHOT |

**Boundary Clarity**: Medium - Overlaps with 10x `/spike` mode

---

### 6. sre-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Site reliability engineering |
| **Triggers** | "Better monitoring/alerting", "production is down", "improve reliability", "handle failure scenario", "outage prevention", "noisy alerts" |
| **Not for** | Feature development, application code, debt management |
| **Produces** | Observability Report, Reliability Plan, Infrastructure Changes, Resilience Report |
| **Complexity** | ALERT, SERVICE, SYSTEM, PLATFORM |

**Boundary Clarity**: High - Clear infrastructure focus, but overlaps with security-pack on infrastructure security

---

### 7. intelligence-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Product analytics |
| **Triggers** | "How do users use this feature?", "track conversion", "A/B test", "what do metrics tell us" |
| **Not for** | Implementation or feature development |
| **Produces** | tracking-plan, research-findings, experiment-design, insights-report |
| **Complexity** | METRIC, FEATURE, INITIATIVE |

**Boundary Clarity**: High - Well-defined inward focus (our users, our product)

---

### 8. strategy-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Business strategy |
| **Triggers** | "TAM for market opportunity", "enter enterprise segment", "usage-based pricing", "prioritize initiatives" |
| **Not for** | Tactical feature decisions, engineering implementation, day-to-day product management |
| **Produces** | market-analysis, competitive-intel, financial-model, strategic-roadmap |
| **Complexity** | TACTICAL, STRATEGIC, TRANSFORMATION |

**Boundary Clarity**: High - Well-defined outward focus (market, competitors)

---

### 9. security-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Security assessment |
| **Triggers** | "Security-review auth system", "SOC 2 requirements", "pentest API", "security perspective on PR" |
| **Not for** | General code review without security implications, performance optimization, feature development |
| **Produces** | threat-model, compliance-requirements, pentest-report, security-signoff |
| **Complexity** | PATCH, FEATURE, SYSTEM |

**Boundary Clarity**: High - Clear security focus

---

### 10. ecosystem-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Ecosystem infrastructure |
| **Triggers** | "Satellite sync failures", "hook/skill registration not working", "design infrastructure patterns", "CEM/roster bugs", "breaking changes migration", "cross-satellite compatibility" |
| **Not for** | Application code in satellites (use 10x-dev-pack), team-specific workflows (use team-pack) |
| **Produces** | Gap Analysis, Context Design, Implementation, Migration Runbook, Compatibility Report |
| **Complexity** | PATCH, MODULE, SYSTEM, MIGRATION |

**Boundary Clarity**: High - Clear infrastructure focus for CEM/roster ecosystem

---

### 11. forge-pack

| Aspect | Definition |
|--------|------------|
| **Domain** | Agent team creation |
| **Triggers** | "New agent team concept", "agent prompts", "workflow configuration", "roster integration", "catalog update", "evaluation" |
| **Not for** | (Not explicitly defined in README - MISSING) |
| **Produces** | Team specification, agent prompts, workflow config, catalog update, evaluation report |
| **Complexity** | (Not defined in README) |

**Boundary Clarity**: Low - Missing "Not for" section, limited routing guidance

---

## Identified Overlaps

### Overlap 1: Refactoring Territory (hygiene-pack vs 10x-dev-pack)

**Contested Area**: Code changes that improve structure but also add functionality

| Signal | hygiene-pack Claim | 10x-dev-pack Claim |
|--------|-------------------|-------------------|
| "Refactor to add feature" | Refactoring is core domain | Feature work is core domain |
| "Clean up while implementing" | Cleanup is hygiene territory | Part of implementation workflow |

**Root Cause**: hygiene-pack handles "refactoring" but excludes "behavior changes", while feature work often requires both.

**Recommendation for hygiene-pack "NOT for"**:
> Refactoring that requires new test cases for new behavior (route to 10x-dev-pack). If behavior is preserved and only structure changes, use hygiene-pack.

---

### Overlap 2: Debt Identification (debt-triage-pack vs hygiene-pack)

**Contested Area**: Initial assessment of code quality issues

| Signal | debt-triage-pack Claim | hygiene-pack Claim |
|--------|------------------------|-------------------|
| "What debt do we have?" | Debt collection is phase 1 | code-smeller identifies issues |
| "Technical debt inventory" | Core debt-collector role | Smell Report catalogs issues |

**Root Cause**: Both teams have assessment agents (debt-collector vs code-smeller) with overlapping detection capabilities.

**Differentiation**:
- debt-triage-pack: Strategic assessment with risk scoring and sprint planning
- hygiene-pack: Tactical cleanup with atomic commits and behavior preservation

**Recommendation for debt-triage-pack "NOT for"**:
> Immediate code cleanup execution (route to hygiene-pack). Debt-triage assesses and prioritizes; hygiene-pack executes cleanup.

**Recommendation for hygiene-pack "NOT for"**:
> Strategic debt prioritization across multiple sprints (route to debt-triage-pack). Hygiene-pack handles cleanup execution within a sprint.

---

### Overlap 3: Spike vs Research (10x-dev-pack /spike vs rnd-pack)

**Contested Area**: Technology evaluation and exploration

| Signal | 10x /spike Claim | rnd-pack Claim |
|--------|-----------------|----------------|
| "Should we use React or Vue?" | Time-boxed decision | Technology evaluation |
| "Is this library suitable?" | Single-session spike | Integration analysis |

**Root Cause**: Both teams handle technology exploration, differentiated only by session duration and outcome type.

**Current Mitigation**: Both READMEs include decision guides distinguishing single-session decisions (/spike) from multi-session research (rnd-pack).

**Recommendation for 10x-dev-pack "NOT for"**:
> Multi-session research with learning-focused outcomes (route to rnd-pack). If you need to learn, experiment, and iterate across multiple sessions, use rnd-pack.

**Recommendation for rnd-pack "NOT for"**:
> Single-session technology comparisons with clear decision criteria (use 10x `/spike`). If you can answer it in one focused session with a decision at the end, use `/spike`.

---

### Overlap 4: Infrastructure Security (security-pack vs sre-pack)

**Contested Area**: Infrastructure-related security concerns

| Signal | security-pack Claim | sre-pack Claim |
|--------|---------------------|----------------|
| "Secure our infrastructure" | Security assessment domain | Infrastructure is platform-engineer territory |
| "Vulnerability in deployment" | Penetration testing scope | Incident response scope |

**Root Cause**: security-pack handles "security" broadly, while sre-pack handles "infrastructure" broadly. Infrastructure security falls in both domains.

**Current Mitigation**: security-pack orchestrator.yaml includes cross_team_protocol: "Escalate infrastructure security to sre-pack."

**Recommendation for security-pack "NOT for"**:
> Infrastructure hardening without security vulnerabilities (route to sre-pack). Security-pack handles threat assessment; sre-pack handles infrastructure configuration.

**Recommendation for sre-pack "NOT for"**:
> Security vulnerability assessment and compliance mapping (route to security-pack). SRE handles infrastructure operations; security handles vulnerability analysis.

---

### Overlap 5: Performance Issues (sre-pack vs 10x-dev-pack)

**Contested Area**: Performance problems in application code vs infrastructure

| Signal | sre-pack Claim | 10x-dev-pack Claim |
|--------|---------------|-------------------|
| "API is slow" | Observability + latency = SRE | Application code optimization |
| "Performance optimization" | SLO/SLI measurement | Code-level implementation |

**Root Cause**: Performance spans both infrastructure (where it's measured) and application code (where it's fixed).

**Recommendation for sre-pack "NOT for"**:
> Application code optimization to improve performance (route to 10x-dev-pack). SRE identifies performance issues via observability; 10x implements code-level fixes.

**Recommendation for 10x-dev-pack "NOT for"**:
> Infrastructure-level performance tuning (route to sre-pack). 10x handles application code; SRE handles infrastructure optimization, scaling, and caching layers.

---

## Identified Gaps

### Gap 1: API Documentation Ownership

**Work That Falls Through**:
- OpenAPI/Swagger spec generation
- API reference documentation with code examples
- SDK documentation

**Currently Claimed By**:
- doc-team-pack: General documentation, but "Not for: writing code comments"
- 10x-dev-pack: Code implementation, but "Not for: documentation work"

**Gap Description**: API documentation requires both code knowledge (to generate accurate specs) and documentation skills (to write clear references). Neither team fully claims this territory.

**Recommendation**:
- Add to doc-team-pack triggers: "API reference documentation"
- Add to doc-team-pack "NOT for": "API spec generation from code (route to 10x, then handoff)"
- Establish handoff pattern: 10x produces OpenAPI spec, doc-team enriches with examples

---

### Gap 2: CI/CD Pipeline Development

**Work That Falls Through**:
- New CI/CD pipeline creation
- Build automation scripts
- Deployment workflow development

**Currently Claimed By**:
- sre-pack: "CI/CD pipelines, IaC" (platform-engineer) but "Not for: feature development"
- 10x-dev-pack: "Build a new feature" but "Not for: infrastructure automation"

**Gap Description**: New pipeline development is neither "feature development" nor purely "infrastructure automation" - it's infrastructure development.

**Recommendation**:
- Clarify sre-pack scope: CI/CD pipeline creation is sre-pack territory
- Add to sre-pack triggers: "Create CI/CD pipeline", "automate builds"
- Add to 10x-dev-pack "NOT for": "CI/CD pipeline creation (route to sre-pack)"

---

### Gap 3: Quick Scripting / Automation

**Work That Falls Through**:
- One-off scripts that need quality (but not full 10x lifecycle)
- Automation scripts that aren't CI/CD
- Developer tooling scripts

**Currently Claimed By**:
- 10x-dev-pack: Explicitly "Not for: one-off scripts without testing requirements"
- hygiene-pack: "Not for: quick formatting fixes"
- No team claims general scripting

**Gap Description**: Quick but quality scripts have no home. 10x is too heavyweight, hygiene is for cleanup.

**Recommendation**:
- Consider rnd-pack for exploratory scripts (prototype-engineer)
- Add to 10x-dev-pack complexity: SCRIPT handles "<200 LOC" but needs clearer entry
- Alternative: Define a "utils" or "tooling" team for developer automation

---

### Gap 4: Observability Implementation

**Work That Falls Through**:
- Adding logging to existing code
- Implementing tracing instrumentation
- Creating custom metrics in application code

**Currently Claimed By**:
- sre-pack: observability-engineer owns "Metrics, logs, traces"
- 10x-dev-pack: Implementation of code changes

**Gap Description**: observability-engineer defines what to measure; 10x implements code. But who instruments the code? Observability design vs implementation gap.

**Recommendation**:
- Establish handoff: sre-pack produces tracking plan, 10x implements instrumentation
- Add to sre-pack "NOT for": "Adding instrumentation code to application (route to 10x with tracking plan)"
- Add cross_team_protocol to sre-pack: "Observability implementation handoff to 10x-dev-pack"

---

## Common Misrouting Scenarios

### Scenario 1: "Refactor and add feature"

**User Intent**: "I need to refactor the payment module and add a new payment provider"

**Common Misroute**: hygiene-pack (because "refactor" triggers)

**Correct Routing**: 10x-dev-pack (behavior change = feature work)

**Distinguishing Signal**: "add a new" = behavior change, not pure refactoring

**Recommendation**:
- Add to hygiene-pack "NOT for": "Refactoring that adds new functionality (behavior changes go to 10x-dev-pack)"

---

### Scenario 2: "Quick security check"

**User Intent**: "Can you do a quick security check on this PR before I merge?"

**Common Misroute**: Direct execution (seems simple, skip team routing)

**Correct Routing**: security-pack at PATCH complexity (security-reviewer agent)

**Distinguishing Signal**: Any security-related review, regardless of size, benefits from security-pack

**Recommendation**:
- Add to security-pack triggers: "quick security check", "security review before merge"
- Document PATCH complexity as the fast-path for quick reviews

---

### Scenario 3: "Investigate slow API"

**User Intent**: "Our API is slow and users are complaining"

**Common Misroute**: 10x-dev-pack (to "fix the code")

**Correct Routing**: sre-pack first (observability-engineer diagnoses), then 10x if code fix needed

**Distinguishing Signal**: Investigation/diagnosis phase before implementation

**Recommendation**:
- Add to sre-pack triggers: "investigate performance", "diagnose slow"
- Add to 10x-dev-pack "NOT for": "Performance investigation without diagnosis (start with sre-pack)"

---

### Scenario 4: "Document this new feature"

**User Intent**: "I just finished implementing feature X, now I need to document it"

**Common Misroute**: Doing it inline during 10x session

**Correct Routing**: Handoff to doc-team-pack after implementation complete

**Distinguishing Signal**: Post-implementation documentation is doc-team territory

**Recommendation**:
- Formalize in 10x "Related Teams": "After implementation, handoff documentation to doc-team-pack"
- Add doc-team-pack triggers: "document completed feature", "write docs for new feature"

---

### Scenario 5: "Technical debt sprint"

**User Intent**: "We need to do a technical debt sprint"

**Common Misroute**: hygiene-pack (because "cleanup")

**Correct Routing**: debt-triage-pack (for sprint planning), then hygiene-pack (for execution)

**Distinguishing Signal**: "sprint" implies planning and prioritization, not just execution

**Recommendation**:
- Add to debt-triage-pack triggers: "plan debt sprint", "debt sprint planning"
- Add to hygiene-pack "NOT for": "Sprint planning for debt paydown (start with debt-triage-pack)"
- Establish handoff: debt-triage produces Sprint Plan, hygiene executes

---

## Recommendations Summary

### Per-Team "NOT for" Enhancements

| Team | Add to "NOT for" |
|------|------------------|
| **10x-dev-pack** | Multi-session research (rnd-pack), CI/CD pipeline creation (sre-pack), performance investigation before diagnosis (sre-pack), infrastructure optimization (sre-pack) |
| **doc-team-pack** | API spec generation from code (10x first, then handoff) |
| **hygiene-pack** | Strategic debt prioritization (debt-triage-pack), refactoring that adds new behavior (10x-dev-pack) |
| **debt-triage-pack** | Immediate code cleanup execution (hygiene-pack) |
| **rnd-pack** | Single-session technology comparisons (10x `/spike`) |
| **sre-pack** | Security vulnerability assessment (security-pack), adding instrumentation code to application (10x with handoff) |
| **intelligence-pack** | (Current definition adequate) |
| **strategy-pack** | (Current definition adequate) |
| **security-pack** | Infrastructure hardening without vulnerabilities (sre-pack) |
| **ecosystem-pack** | (Current definition adequate) |
| **forge-pack** | **(CRITICAL)** Add "NOT for" section: Production feature development, existing team modification (ecosystem-pack), one-off agent creation without team context |

### Cross-Team Handoff Formalization

| From | To | Trigger |
|------|----|---------|
| sre-pack (observability) | 10x-dev-pack | Instrumentation implementation needed |
| debt-triage-pack | hygiene-pack | Sprint plan ready for execution |
| 10x-dev-pack | doc-team-pack | Feature complete, documentation needed |
| 10x-dev-pack | sre-pack | Production deployment validation |
| rnd-pack | 10x-dev-pack | Prototype ready for production |

### Missing Documentation

| Gap | Recommended Action |
|-----|-------------------|
| forge-pack README missing "NOT for" | Add section before next sync |
| forge-pack README missing complexity levels | Define TEAM, AGENT, SKILL levels |
| API documentation ownership undefined | Clarify in both doc-team and 10x READMEs |
| CI/CD pipeline development ambiguous | Add explicit trigger to sre-pack |

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This gap analysis | `/Users/tomtenuta/Code/roster/docs/analysis/team-boundary-gaps.md` | Created |
| 10x-dev-pack README | `/Users/tomtenuta/Code/roster/teams/10x-dev-pack/README.md` | Read |
| doc-team-pack README | `/Users/tomtenuta/Code/roster/teams/doc-team-pack/README.md` | Read |
| hygiene-pack README | `/Users/tomtenuta/Code/roster/teams/hygiene-pack/README.md` | Read |
| debt-triage-pack README | `/Users/tomtenuta/Code/roster/teams/debt-triage-pack/README.md` | Read |
| rnd-pack README | `/Users/tomtenuta/Code/roster/teams/rnd-pack/README.md` | Read |
| sre-pack README | `/Users/tomtenuta/Code/roster/teams/sre-pack/README.md` | Read |
| intelligence-pack README | `/Users/tomtenuta/Code/roster/teams/intelligence-pack/README.md` | Read |
| strategy-pack README | `/Users/tomtenuta/Code/roster/teams/strategy-pack/README.md` | Read |
| security-pack README | `/Users/tomtenuta/Code/roster/teams/security-pack/README.md` | Read |
| ecosystem-pack README | `/Users/tomtenuta/Code/roster/teams/ecosystem-pack/README.md` | Read |
| DESIGN-intent-matching.md | `/Users/tomtenuta/Code/roster/docs/design/DESIGN-intent-matching.md` | Read |
| cross-team-handoff schema | `/Users/tomtenuta/Code/roster/teams/shared/skills/cross-team-handoff/schema.md` | Read |

---

## Handoff Criteria Checklist

- [x] Root cause traced to specific boundary definitions in team READMEs
- [x] At least 3 overlap scenarios documented (5 documented)
- [x] At least 3 gap scenarios documented (4 documented)
- [x] Common misrouting scenarios documented (5 documented)
- [x] Recommendations for each team's "NOT for" content provided
- [x] Affected teams enumerated with specific recommendations
- [x] File verification completed

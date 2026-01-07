# Team Boundary Gap Analysis

> Analysis of team boundary clarity, overlaps, gaps, and misrouting patterns
> Session: session-20260104-013028-45c9ccac | Task: task-001
> Generated: 2026-01-04

## Executive Summary

This analysis audits 11 rites across the roster ecosystem to identify boundary issues affecting routing accuracy and developer discoverability. Key findings:

- **5 boundary overlaps** identified where multiple teams claim similar territory
- **4 boundary gaps** identified where work falls between teams
- **5 common misrouting scenarios** documented with mitigation recommendations
- **All 11 teams** require enhanced "NOT for" sections for clarity

---

## Team Boundary Inventory (Current State)

### 1. 10x-dev

| Aspect | Definition |
|--------|------------|
| **Domain** | Software development |
| **Triggers** | "Build a new feature", "PRD and technical design", "requirements to tested code", "implement API/service/module" |
| **Not for** | Documentation work, infrastructure automation, one-off scripts without testing requirements |
| **Produces** | PRD, TDD, ADRs, code, tests |
| **Complexity** | SCRIPT, MODULE, SERVICE, PLATFORM |

**Boundary Clarity**: Medium - Overlaps with hygiene on refactoring, rnd on spikes

---

### 2. docs

| Aspect | Definition |
|--------|------------|
| **Domain** | Documentation lifecycle |
| **Triggers** | "Documentation scattered/inconsistent", "document this feature/system/API", "audit documentation", "reorganize docs" |
| **Not for** | Code implementation, infrastructure work, writing code comments |
| **Produces** | Audit report, doc structure, documentation, review signoff |
| **Complexity** | PAGE, SECTION, SITE |

**Boundary Clarity**: High - Clear separation from code-focused teams

---

### 3. hygiene

| Aspect | Definition |
|--------|------------|
| **Domain** | Code quality |
| **Triggers** | "Codebase feels messy", "dead code/unused imports", "technical debt inventory", "refactoring", "cleanup commits" |
| **Not for** | New features (behavior changes), ecosystem infrastructure, quick formatting fixes |
| **Produces** | Smell Report, Refactoring Plan, Commit Stream, Audit Report |
| **Complexity** | SPOT, MODULE, CODEBASE |

**Boundary Clarity**: Medium - Overlaps with debt-triage on debt identification

---

### 4. debt-triage

| Aspect | Definition |
|--------|------------|
| **Domain** | Technical debt management |
| **Triggers** | "What technical debt do we have?", "prioritize debt paydown", "biggest technical risk", "debt cleanup sprint", "inherited codebase debt" |
| **Not for** | Feature development, active incidents, ongoing reliability work |
| **Produces** | Debt Ledger, Risk Report, Sprint Plan |
| **Complexity** | QUICK, AUDIT |

**Boundary Clarity**: Medium - Overlaps with hygiene (assessment vs execution)

---

### 5. rnd

| Aspect | Definition |
|--------|------------|
| **Domain** | Technology exploration |
| **Triggers** | "Evaluate new technology", "integrate with current stack", "proof-of-concept", "architecture in 2 years" |
| **Not for** | Production feature development, immediate shipping |
| **Produces** | tech-assessment, integration-map, prototype, moonshot-plan, TRANSFER, HANDOFF |
| **Complexity** | SPIKE, EVALUATION, MOONSHOT |

**Boundary Clarity**: Medium - Overlaps with 10x `/spike` mode

---

### 6. sre

| Aspect | Definition |
|--------|------------|
| **Domain** | Site reliability engineering |
| **Triggers** | "Better monitoring/alerting", "production is down", "improve reliability", "handle failure scenario", "outage prevention", "noisy alerts" |
| **Not for** | Feature development, application code, debt management |
| **Produces** | Observability Report, Reliability Plan, Infrastructure Changes, Resilience Report |
| **Complexity** | ALERT, SERVICE, SYSTEM, PLATFORM |

**Boundary Clarity**: High - Clear infrastructure focus, but overlaps with security on infrastructure security

---

### 7. intelligence

| Aspect | Definition |
|--------|------------|
| **Domain** | Product analytics |
| **Triggers** | "How do users use this feature?", "track conversion", "A/B test", "what do metrics tell us" |
| **Not for** | Implementation or feature development |
| **Produces** | tracking-plan, research-findings, experiment-design, insights-report |
| **Complexity** | METRIC, FEATURE, INITIATIVE |

**Boundary Clarity**: High - Well-defined inward focus (our users, our product)

---

### 8. strategy

| Aspect | Definition |
|--------|------------|
| **Domain** | Business strategy |
| **Triggers** | "TAM for market opportunity", "enter enterprise segment", "usage-based pricing", "prioritize initiatives" |
| **Not for** | Tactical feature decisions, engineering implementation, day-to-day product management |
| **Produces** | market-analysis, competitive-intel, financial-model, strategic-roadmap |
| **Complexity** | TACTICAL, STRATEGIC, TRANSFORMATION |

**Boundary Clarity**: High - Well-defined outward focus (market, competitors)

---

### 9. security

| Aspect | Definition |
|--------|------------|
| **Domain** | Security assessment |
| **Triggers** | "Security-review auth system", "SOC 2 requirements", "pentest API", "security perspective on PR" |
| **Not for** | General code review without security implications, performance optimization, feature development |
| **Produces** | threat-model, compliance-requirements, pentest-report, security-signoff |
| **Complexity** | PATCH, FEATURE, SYSTEM |

**Boundary Clarity**: High - Clear security focus

---

### 10. ecosystem

| Aspect | Definition |
|--------|------------|
| **Domain** | Ecosystem infrastructure |
| **Triggers** | "Satellite sync failures", "hook/skill registration not working", "design infrastructure patterns", "CEM/roster bugs", "breaking changes migration", "cross-satellite compatibility" |
| **Not for** | Application code in satellites (use 10x-dev), team-specific workflows (use team-pack) |
| **Produces** | Gap Analysis, Context Design, Implementation, Migration Runbook, Compatibility Report |
| **Complexity** | PATCH, MODULE, SYSTEM, MIGRATION |

**Boundary Clarity**: High - Clear infrastructure focus for CEM/roster ecosystem

---

### 11. forge

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

### Overlap 1: Refactoring Territory (hygiene vs 10x-dev)

**Contested Area**: Code changes that improve structure but also add functionality

| Signal | hygiene Claim | 10x-dev Claim |
|--------|-------------------|-------------------|
| "Refactor to add feature" | Refactoring is core domain | Feature work is core domain |
| "Clean up while implementing" | Cleanup is hygiene territory | Part of implementation workflow |

**Root Cause**: hygiene handles "refactoring" but excludes "behavior changes", while feature work often requires both.

**Recommendation for hygiene "NOT for"**:
> Refactoring that requires new test cases for new behavior (route to 10x-dev). If behavior is preserved and only structure changes, use hygiene.

---

### Overlap 2: Debt Identification (debt-triage vs hygiene)

**Contested Area**: Initial assessment of code quality issues

| Signal | debt-triage Claim | hygiene Claim |
|--------|------------------------|-------------------|
| "What debt do we have?" | Debt collection is phase 1 | code-smeller identifies issues |
| "Technical debt inventory" | Core debt-collector role | Smell Report catalogs issues |

**Root Cause**: Both teams have assessment agents (debt-collector vs code-smeller) with overlapping detection capabilities.

**Differentiation**:
- debt-triage: Strategic assessment with risk scoring and sprint planning
- hygiene: Tactical cleanup with atomic commits and behavior preservation

**Recommendation for debt-triage "NOT for"**:
> Immediate code cleanup execution (route to hygiene). Debt-triage assesses and prioritizes; hygiene executes cleanup.

**Recommendation for hygiene "NOT for"**:
> Strategic debt prioritization across multiple sprints (route to debt-triage). Hygiene-pack handles cleanup execution within a sprint.

---

### Overlap 3: Spike vs Research (10x-dev /spike vs rnd)

**Contested Area**: Technology evaluation and exploration

| Signal | 10x /spike Claim | rnd Claim |
|--------|-----------------|----------------|
| "Should we use React or Vue?" | Time-boxed decision | Technology evaluation |
| "Is this library suitable?" | Single-session spike | Integration analysis |

**Root Cause**: Both teams handle technology exploration, differentiated only by session duration and outcome type.

**Current Mitigation**: Both READMEs include decision guides distinguishing single-session decisions (/spike) from multi-session research (rnd).

**Recommendation for 10x-dev "NOT for"**:
> Multi-session research with learning-focused outcomes (route to rnd). If you need to learn, experiment, and iterate across multiple sessions, use rnd.

**Recommendation for rnd "NOT for"**:
> Single-session technology comparisons with clear decision criteria (use 10x `/spike`). If you can answer it in one focused session with a decision at the end, use `/spike`.

---

### Overlap 4: Infrastructure Security (security vs sre)

**Contested Area**: Infrastructure-related security concerns

| Signal | security Claim | sre Claim |
|--------|---------------------|----------------|
| "Secure our infrastructure" | Security assessment domain | Infrastructure is platform-engineer territory |
| "Vulnerability in deployment" | Penetration testing scope | Incident response scope |

**Root Cause**: security handles "security" broadly, while sre handles "infrastructure" broadly. Infrastructure security falls in both domains.

**Current Mitigation**: security orchestrator.yaml includes cross_team_protocol: "Escalate infrastructure security to sre."

**Recommendation for security "NOT for"**:
> Infrastructure hardening without security vulnerabilities (route to sre). Security-pack handles threat assessment; sre handles infrastructure configuration.

**Recommendation for sre "NOT for"**:
> Security vulnerability assessment and compliance mapping (route to security). SRE handles infrastructure operations; security handles vulnerability analysis.

---

### Overlap 5: Performance Issues (sre vs 10x-dev)

**Contested Area**: Performance problems in application code vs infrastructure

| Signal | sre Claim | 10x-dev Claim |
|--------|---------------|-------------------|
| "API is slow" | Observability + latency = SRE | Application code optimization |
| "Performance optimization" | SLO/SLI measurement | Code-level implementation |

**Root Cause**: Performance spans both infrastructure (where it's measured) and application code (where it's fixed).

**Recommendation for sre "NOT for"**:
> Application code optimization to improve performance (route to 10x-dev). SRE identifies performance issues via observability; 10x implements code-level fixes.

**Recommendation for 10x-dev "NOT for"**:
> Infrastructure-level performance tuning (route to sre). 10x handles application code; SRE handles infrastructure optimization, scaling, and caching layers.

---

## Identified Gaps

### Gap 1: API Documentation Ownership

**Work That Falls Through**:
- OpenAPI/Swagger spec generation
- API reference documentation with code examples
- SDK documentation

**Currently Claimed By**:
- docs: General documentation, but "Not for: writing code comments"
- 10x-dev: Code implementation, but "Not for: documentation work"

**Gap Description**: API documentation requires both code knowledge (to generate accurate specs) and documentation skills (to write clear references). Neither team fully claims this territory.

**Recommendation**:
- Add to docs triggers: "API reference documentation"
- Add to docs "NOT for": "API spec generation from code (route to 10x, then handoff)"
- Establish handoff pattern: 10x produces OpenAPI spec, doc-team enriches with examples

---

### Gap 2: CI/CD Pipeline Development

**Work That Falls Through**:
- New CI/CD pipeline creation
- Build automation scripts
- Deployment workflow development

**Currently Claimed By**:
- sre: "CI/CD pipelines, IaC" (platform-engineer) but "Not for: feature development"
- 10x-dev: "Build a new feature" but "Not for: infrastructure automation"

**Gap Description**: New pipeline development is neither "feature development" nor purely "infrastructure automation" - it's infrastructure development.

**Recommendation**:
- Clarify sre scope: CI/CD pipeline creation is sre territory
- Add to sre triggers: "Create CI/CD pipeline", "automate builds"
- Add to 10x-dev "NOT for": "CI/CD pipeline creation (route to sre)"

---

### Gap 3: Quick Scripting / Automation

**Work That Falls Through**:
- One-off scripts that need quality (but not full 10x lifecycle)
- Automation scripts that aren't CI/CD
- Developer tooling scripts

**Currently Claimed By**:
- 10x-dev: Explicitly "Not for: one-off scripts without testing requirements"
- hygiene: "Not for: quick formatting fixes"
- No team claims general scripting

**Gap Description**: Quick but quality scripts have no home. 10x is too heavyweight, hygiene is for cleanup.

**Recommendation**:
- Consider rnd for exploratory scripts (prototype-engineer)
- Add to 10x-dev complexity: SCRIPT handles "<200 LOC" but needs clearer entry
- Alternative: Define a "utils" or "tooling" team for developer automation

---

### Gap 4: Observability Implementation

**Work That Falls Through**:
- Adding logging to existing code
- Implementing tracing instrumentation
- Creating custom metrics in application code

**Currently Claimed By**:
- sre: observability-engineer owns "Metrics, logs, traces"
- 10x-dev: Implementation of code changes

**Gap Description**: observability-engineer defines what to measure; 10x implements code. But who instruments the code? Observability design vs implementation gap.

**Recommendation**:
- Establish handoff: sre produces tracking plan, 10x implements instrumentation
- Add to sre "NOT for": "Adding instrumentation code to application (route to 10x with tracking plan)"
- Add cross_team_protocol to sre: "Observability implementation handoff to 10x-dev"

---

## Common Misrouting Scenarios

### Scenario 1: "Refactor and add feature"

**User Intent**: "I need to refactor the payment module and add a new payment provider"

**Common Misroute**: hygiene (because "refactor" triggers)

**Correct Routing**: 10x-dev (behavior change = feature work)

**Distinguishing Signal**: "add a new" = behavior change, not pure refactoring

**Recommendation**:
- Add to hygiene "NOT for": "Refactoring that adds new functionality (behavior changes go to 10x-dev)"

---

### Scenario 2: "Quick security check"

**User Intent**: "Can you do a quick security check on this PR before I merge?"

**Common Misroute**: Direct execution (seems simple, skip team routing)

**Correct Routing**: security at PATCH complexity (security-reviewer agent)

**Distinguishing Signal**: Any security-related review, regardless of size, benefits from security

**Recommendation**:
- Add to security triggers: "quick security check", "security review before merge"
- Document PATCH complexity as the fast-path for quick reviews

---

### Scenario 3: "Investigate slow API"

**User Intent**: "Our API is slow and users are complaining"

**Common Misroute**: 10x-dev (to "fix the code")

**Correct Routing**: sre first (observability-engineer diagnoses), then 10x if code fix needed

**Distinguishing Signal**: Investigation/diagnosis phase before implementation

**Recommendation**:
- Add to sre triggers: "investigate performance", "diagnose slow"
- Add to 10x-dev "NOT for": "Performance investigation without diagnosis (start with sre)"

---

### Scenario 4: "Document this new feature"

**User Intent**: "I just finished implementing feature X, now I need to document it"

**Common Misroute**: Doing it inline during 10x session

**Correct Routing**: Handoff to docs after implementation complete

**Distinguishing Signal**: Post-implementation documentation is doc-team territory

**Recommendation**:
- Formalize in 10x "Related Teams": "After implementation, handoff documentation to docs"
- Add docs triggers: "document completed feature", "write docs for new feature"

---

### Scenario 5: "Technical debt sprint"

**User Intent**: "We need to do a technical debt sprint"

**Common Misroute**: hygiene (because "cleanup")

**Correct Routing**: debt-triage (for sprint planning), then hygiene (for execution)

**Distinguishing Signal**: "sprint" implies planning and prioritization, not just execution

**Recommendation**:
- Add to debt-triage triggers: "plan debt sprint", "debt sprint planning"
- Add to hygiene "NOT for": "Sprint planning for debt paydown (start with debt-triage)"
- Establish handoff: debt-triage produces Sprint Plan, hygiene executes

---

## Recommendations Summary

### Per-Team "NOT for" Enhancements

| Team | Add to "NOT for" |
|------|------------------|
| **10x-dev** | Multi-session research (rnd), CI/CD pipeline creation (sre), performance investigation before diagnosis (sre), infrastructure optimization (sre) |
| **docs** | API spec generation from code (10x first, then handoff) |
| **hygiene** | Strategic debt prioritization (debt-triage), refactoring that adds new behavior (10x-dev) |
| **debt-triage** | Immediate code cleanup execution (hygiene) |
| **rnd** | Single-session technology comparisons (10x `/spike`) |
| **sre** | Security vulnerability assessment (security), adding instrumentation code to application (10x with handoff) |
| **intelligence** | (Current definition adequate) |
| **strategy** | (Current definition adequate) |
| **security** | Infrastructure hardening without vulnerabilities (sre) |
| **ecosystem** | (Current definition adequate) |
| **forge** | **(CRITICAL)** Add "NOT for" section: Production feature development, existing team modification (ecosystem), one-off agent creation without rite context |

### Cross-Team Handoff Formalization

| From | To | Trigger |
|------|----|---------|
| sre (observability) | 10x-dev | Instrumentation implementation needed |
| debt-triage | hygiene | Sprint plan ready for execution |
| 10x-dev | docs | Feature complete, documentation needed |
| 10x-dev | sre | Production deployment validation |
| rnd | 10x-dev | Prototype ready for production |

### Missing Documentation

| Gap | Recommended Action |
|-----|-------------------|
| forge README missing "NOT for" | Add section before next sync |
| forge README missing complexity levels | Define TEAM, AGENT, SKILL levels |
| API documentation ownership undefined | Clarify in both doc-team and 10x READMEs |
| CI/CD pipeline development ambiguous | Add explicit trigger to sre |

---

## File Verification

| Artifact | Absolute Path | Status |
|----------|---------------|--------|
| This gap analysis | `/Users/tomtenuta/Code/roster/docs/analysis/team-boundary-gaps.md` | Created |
| 10x-dev README | `/Users/tomtenuta/Code/roster/rites/10x-dev/README.md` | Read |
| docs README | `/Users/tomtenuta/Code/roster/rites/docs/README.md` | Read |
| hygiene README | `/Users/tomtenuta/Code/roster/rites/hygiene/README.md` | Read |
| debt-triage README | `/Users/tomtenuta/Code/roster/rites/debt-triage/README.md` | Read |
| rnd README | `/Users/tomtenuta/Code/roster/rites/rnd/README.md` | Read |
| sre README | `/Users/tomtenuta/Code/roster/rites/sre/README.md` | Read |
| intelligence README | `/Users/tomtenuta/Code/roster/rites/intelligence/README.md` | Read |
| strategy README | `/Users/tomtenuta/Code/roster/rites/strategy/README.md` | Read |
| security README | `/Users/tomtenuta/Code/roster/rites/security/README.md` | Read |
| ecosystem README | `/Users/tomtenuta/Code/roster/rites/ecosystem/README.md` | Read |
| DESIGN-intent-matching.md | `/Users/tomtenuta/Code/roster/docs/design/DESIGN-intent-matching.md` | Read |
| cross-rite-handoff schema | `/Users/tomtenuta/Code/roster/rites/shared/skills/cross-rite-handoff/schema.md` | Read |

---

## Handoff Criteria Checklist

- [x] Root cause traced to specific boundary definitions in team READMEs
- [x] At least 3 overlap scenarios documented (5 documented)
- [x] At least 3 gap scenarios documented (4 documented)
- [x] Common misrouting scenarios documented (5 documented)
- [x] Recommendations for each team's "NOT for" content provided
- [x] Affected teams enumerated with specific recommendations
- [x] File verification completed

# Agent Quality Audit: Context Engineering Assessment

**Date**: 2026-02-09
**Auditor**: Context Engineer (Opus 4.6)
**Scope**: All agents across `rites/*/agents/*.md`
**References**: Canonical template (`rites/forge/mena/rite-development/templates/agent-template.md`), Archetype generator (`internal/agent/archetype.go`)

---

## Summary

| Metric | Value |
|--------|-------|
| Total agents audited | 58 |
| Rites covered | 11 |
| Orchestrators | 11 |
| Specialists/Designers/Engineers/Analysts | 47 |
| Reviewers (with contract.must_not) | 7 |
| Agents >200 lines | 19 |
| Critical findings | 5 |
| High findings | 8 |
| Medium findings | 7 |
| Low findings | 6 |

### Archetype Distribution

| Archetype (programmatic) | Mapped `type` values | Count |
|--------------------------|---------------------|-------|
| orchestrator | orchestrator | 11 |
| specialist | specialist, engineer, analyst, designer | 40 |
| reviewer | reviewer | 7 |

The `type` field uses fine-grained values (analyst, designer, engineer, specialist) that map to one of the 3 programmatic archetypes in `archetype.go`. This is intentional semantic richness -- not a defect -- but the mapping is implicit and undocumented.

---

## Critical Findings

### C-01: Forge agents follow canonical template; 41/47 non-orchestrator agents outside forge do not

**Severity**: CRITICAL
**Scope**: 41 agents across 10 rites (all non-forge specialists)

The canonical template (`agent-template.md`) specifies multi-line YAML `|` descriptions with three required components: (1) "When to use this agent" section with 3+ use cases, (2) at least one `<example>` block, (3) role summary + trigger phrases. Only **7 forge agents** and **11 orchestrators** use multi-line descriptions. The remaining **41 agents use single-line descriptions** that omit use cases and examples entirely.

**Impact**: CC uses the `description` field as the primary discovery mechanism for Task tool routing. Single-line descriptions like `"Diagnoses code quality issues"` provide far less signal for CC to match user intent than the forge-style multi-line format with explicit triggers and examples. This means CC routing precision degrades for 70% of agents.

**Agents affected**: All non-orchestrator agents in 10x-dev, debt-triage, docs, ecosystem, hygiene, intelligence, rnd, security, sre, strategy.

**Example (current, 10x-dev/architect)**:
```yaml
description: "Evaluates tradeoffs and designs systems. Use when: system design, technical analysis, or architecture decisions needed. Triggers: design, architect, tradeoffs, system design, technical analysis."
```

**Example (forge canonical, agent-designer)**:
```yaml
description: |
  The rite architecture specialist who designs agent roles, boundaries, and contracts.
  Invoke when creating a new rite, adding agents to existing rites, or restructuring...

  When to use this agent:
  - Designing a new rite from scratch
  - Adding or modifying agents in an existing rite
  ...

  <example>
  Context: User wants to create a new rite for API development
  user: "I need a rite..."
  assistant: "Invoking Agent Designer..."
  </example>
```

**Recommendation**: Migrate all 41 agents to multi-line descriptions. This is the single highest-impact improvement for CC routing precision.

---

### C-02: maxTurns values wildly diverge from archetype defaults with no documented rationale

**Severity**: CRITICAL
**Scope**: All 58 agents

The archetype defaults in `archetype.go` specify: orchestrator=3, specialist=25, reviewer=15. Actual values in agent files:

| Archetype | Default | Actual Range | Most Common |
|-----------|---------|-------------|-------------|
| orchestrator | 3 | 40 (all 11) | 40 |
| specialist | 25 | 150-250 | 150 |
| reviewer | 15 | 100 (all 7) | 100 |

**Every single agent overrides the archetype default.** Orchestrators use 13x the default. Specialists use 6-10x. Reviewers use 6.7x.

This means the archetype defaults in `archetype.go` are functionally dead code -- they describe a system that does not exist. Either the defaults are wrong (should be updated to match reality) or the agent files are over-provisioned (wasting token budget on turns that never execute).

**Impact**: The archetype contract becomes unreliable. Code that relies on `ArchetypeDefaults.MaxTurns` for validation or scaffolding will produce values the platform immediately overrides. New agents created from archetypes will have the "wrong" maxTurns until manually corrected.

**Recommendation**: Update archetype defaults to reflect actual practice: orchestrator=40, specialist=150, reviewer=100. Or document the rationale for the divergence.

---

### C-03: 10 agents >200 lines embed reference knowledge that should be in skills

**Severity**: CRITICAL
**Scope**: 10 agents with embedded reference content

Agent prompts are context-loaded into every subagent invocation. Embedded reference knowledge pays full token cost on every turn. The following agents embed substantial reference material that would be cheaper as on-demand skills:

| Agent | Lines | Embedded Content | Est. Token Waste/Turn |
|-------|-------|-----------------|----------------------|
| `rnd/tech-transfer` | 301 | 2 complete HANDOFF format examples (~100 lines) | ~400 |
| `forge/eval-specialist` | 297 | Validation checklists, Rite Maturity Model (~100 lines) | ~400 |
| `forge/agent-curator` | 295 | Consultant Sync Checklist, Versioning Scheme (~90 lines) | ~360 |
| `forge/workflow-engineer` | 294 | Workflow Patterns Library, 6 pattern definitions (~80 lines) | ~320 |
| `intelligence/insights-analyst` | 281 | Complete HANDOFF example with full item detail (~80 lines) | ~320 |
| `forge/platform-engineer` | 257 | ari sync CLI reference documentation (~60 lines) | ~240 |
| `strategy/roadmap-strategist` | 246 | HANDOFF format template + complete example (~70 lines) | ~280 |
| `docs/doc-auditor` | 233 | Staleness Detection Mode with full CLI spec (~80 lines) | ~320 |
| `forge/agent-designer` | 224 | Complexity level definitions, RITE-SPEC template (~40 lines) | ~160 |
| `debt-triage/sprint-planner` | 216 | HANDOFF format template (~50 lines) | ~200 |

**Token cost**: These 10 agents carry ~3,000 tokens of reference content that loads on every invocation regardless of whether the specific reference is needed that turn.

**Recommendation**: Extract HANDOFF templates, validation checklists, CLI references, and pattern libraries into skills. Replace inline content with `@skill-name` references. Target: all agents under 200 lines.

---

### C-04: `role` field exists in 47 agents but is absent from canonical template frontmatter schema

**Severity**: CRITICAL
**Scope**: 47 non-orchestrator agents

Every non-orchestrator agent has a `role` field (e.g., `role: "Designs agent roles and contracts"`). Zero orchestrators have it. The canonical template frontmatter schema lists only: `name`, `description`, `tools`, `model`, `color`. The `role` field is not documented.

Similarly, `type` (orchestrator/specialist/analyst/designer/engineer/reviewer), `maxTurns`, `disallowedTools`, and `contract` are present in agents but absent from the canonical template schema.

**Impact**: The canonical template is incomplete as a specification. Authors following it will produce agents missing fields that the platform expects. The `role` field appears to serve as a one-line summary for CLAUDE.md agent tables (distinct from the longer `description`), but this purpose is undocumented.

**Recommendation**: Update canonical template frontmatter schema to document all 8 fields actually in use: `name`, `role`, `description`, `type`, `tools`, `model`, `color`, `maxTurns` (plus conditional `disallowedTools` and `contract`).

---

### C-05: `type` field uses 6 values but maps to only 3 programmatic archetypes

**Severity**: CRITICAL
**Scope**: All 58 agents

The `type` field uses: `orchestrator` (11), `specialist` (9), `analyst` (8), `designer` (7), `engineer` (10), `reviewer` (7). The archetype generator only knows 3: orchestrator, specialist, reviewer.

The mapping is: analyst/designer/engineer/specialist all resolve to the `specialist` archetype. This is implicit -- there is no documented mapping, no validation that catches `type: analyst` and routes to `specialist` archetype defaults.

**Impact**: If `ari agent scaffold --type=analyst` is ever built, it would fail because `analyst` is not a recognized archetype. The fine-grained types provide useful semantic information for humans but create a gap between the type taxonomy in agent files and the archetype taxonomy in Go code.

**Recommendation**: Either (a) add analyst/designer/engineer as archetype aliases in `archetype.go`, or (b) add a documented `type -> archetype` mapping, or (c) move to a two-field model: `archetype: specialist` + `type: analyst`.

---

## High Findings

### H-01: Orchestrator boilerplate is ~160 lines of identical content across 11 agents

**Severity**: HIGH
**Scope**: 11 orchestrators (194-230 lines each)

All 11 orchestrators share identical content for: Consultation Role (~25 lines), Tool Access (~8 lines), Consultation Protocol (~30 lines), Behavioral Constraints (~20 lines), Handling Failures (~15 lines), The Acid Test (~8 lines), Anti-Patterns (~15 lines). That is ~120-160 lines of platform-owned content repeated 11 times.

The archetype generator (`archetype.go`) correctly marks these sections as `OwnerPlatform`, indicating they should be generated not authored. But currently each orchestrator file contains manually-maintained copies.

**Token impact**: 11 x ~160 lines = ~1,760 lines of duplicated content across source files. When materialized, each orchestrator pays ~640 tokens for boilerplate.

**Recommendation**: Implement platform section injection during materialization. Author-owned sections (Domain Authority, Phase Routing, Cross-Rite Protocol) remain in agent files; platform sections inject from archetype definitions.

---

### H-02: Section heading inconsistencies across rites

**Severity**: HIGH
**Scope**: 47 non-orchestrator agents

The canonical template and archetype generator define precise section headings. Agents deviate:

| Canonical Heading | Variant Used | Count | Rites |
|-------------------|-------------|-------|-------|
| "Core Responsibilities" | "Responsibilities" | 9 | ecosystem (5), forge (partial) |
| "Core Responsibilities" | "Core Purpose" | 10 | reviewers use this per archetype |
| "Approach" | "When Invoked" | 13 | ecosystem, intelligence, strategy, rnd |
| "Approach" | "When Invoked (First Actions)" | 5 | ecosystem |
| "Approach" | "How You Work" | 6 | forge |
| "Skills Reference" | "Related Skills" | 6 | 10x-dev |
| "Anti-Patterns to Avoid" | "Anti-Patterns" | 33 | mixed across rites |
| "Cross-Rite Routing" | "Cross-Rite Notes" | 5 | forge |
| "Cross-Rite Routing" | "Cross-Rite Protocol" | 4 | some orchestrators |

**Note on "Core Purpose"**: The reviewer archetype in `archetype.go` defines `core-purpose` as its first section (not `core-responsibilities`). So the 10 agents using "Core Purpose" are actually following their archetype correctly. This is an archetype-level divergence from the canonical template, not an agent-level error.

**Impact**: Automated validation tools that check for canonical section names will report false positives. Inconsistency makes cross-agent navigation harder for humans reviewing agent files.

**Recommendation**: Align on archetype-specific heading sets. The canonical template should acknowledge that reviewer archetype uses "Core Purpose" and specialist uses "Core Responsibilities". Normalize the other variants: "When Invoked" -> "Approach", "Related Skills" -> "Skills Reference".

---

### H-03: 7 reviewers have `disallowedTools: [Task]` but archetype also specifies WebFetch/WebSearch in default tools

**Severity**: HIGH
**Scope**: 7 reviewers (qa-adversary, compatibility-tester, doc-auditor, doc-reviewer, eval-specialist, security-reviewer, audit-lead)

The reviewer archetype default tools include `WebFetch` and `WebSearch`. But only 1 of 7 reviewers (qa-adversary) actually lists these tools. The other 6 reviewers use `Bash, Glob, Grep, Read, Edit, Write, TodoWrite, Skill` -- the specialist tool set minus WebFetch/WebSearch.

**Impact**: Either the reviewer archetype default is wrong (most reviewers do not need web access), or 6 reviewers are under-provisioned. The archetype contract says reviewers get web tools; reality says they do not.

**Recommendation**: Remove WebFetch/WebSearch from reviewer archetype defaults. Add them explicitly only to agents that need web access (qa-adversary gets them from its tools list already).

---

### H-04: `code-smeller` has Write but not Edit tool -- analysis agent with write access

**Severity**: HIGH
**Scope**: `rites/hygiene/agents/code-smeller.md`

The code-smeller is typed as `analyst` and its role is purely diagnostic ("Diagnose only -- fixes are the Architect Enforcer's domain"). The canonical template recommends analysis agents use `Bash, Glob, Grep, Read, Task, TodoWrite` (no Edit/Write). But code-smeller has `Write` (to produce its smell report) and lacks `Edit`.

Having `Write` without `Edit` is an unusual combination. An analyst that can create new files but not modify existing ones is a specific constraint -- but it is undocumented in the agent whether this is intentional.

**Recommendation**: Document the rationale. If the agent needs Write only for report generation, this is defensible. But consider whether the smell report should instead be produced via skill template routing, removing the need for Write entirely.

---

### H-05: 21 agents have "File Verification" section not defined in any archetype

**Severity**: HIGH
**Scope**: 21 agents across docs, ecosystem, hygiene, intelligence rites

Twenty-one agents include a `## File Verification` section (or reference to it) that is not part of any archetype section definition. This section typically says: "See `file-verification` skill for artifact verification protocol."

This is a case of organic section growth -- a useful pattern that emerged but was never codified in the archetype. It sits outside the template validation rules.

**Impact**: Minor structural drift. The content is useful (artifact verification is important). But uncodified sections accumulate over time and bloat agent prompts.

**Recommendation**: Either add `file-verification` as a derived section to specialist/reviewer archetypes, or fold the reference into the existing "Handoff Criteria" section (which already covers verification).

---

### H-06: 4 strategy agents have `<example>` in body but not in description frontmatter

**Severity**: HIGH
**Scope**: strategy/business-model-analyst, strategy/competitive-analyst, strategy/market-researcher, strategy/roadmap-strategist

These agents have `## Example` sections in their markdown body (examples of their output) but their frontmatter descriptions are single-line strings without the `<example>` blocks the canonical template requires.

**Impact**: The `<example>` blocks in description frontmatter serve CC discovery. Body examples serve human readers. These agents have the human version but not the CC version. CC routing will not benefit from the examples.

**Recommendation**: Add `<example>` blocks to the description frontmatter of these agents.

---

### H-07: Orchestrator color assignments do not consistently use purple

**Severity**: HIGH
**Scope**: 11 orchestrators

The canonical template and archetype defaults specify `purple` for orchestrators. Actual colors:

| Orchestrator | Color |
|-------------|-------|
| ecosystem | purple |
| rnd | purple |
| forge | cyan |
| 10x-dev | blue |
| docs | green |
| debt-triage | orange |
| hygiene | green |
| intelligence | cyan |
| security | red |
| sre | orange |
| strategy | yellow |

Only **2 of 11** orchestrators use the archetype default color `purple`. This defeats the visual convention where purple = orchestrator.

**Impact**: Users scanning agent lists cannot quickly identify orchestrators by color. The archetype contract is violated.

**Recommendation**: Standardize all orchestrators to purple. Reassign their current colors to other agents in the same rite if needed.

---

### H-08: `doc-auditor` typed as reviewer but has Edit and Write tools

**Severity**: HIGH
**Scope**: `rites/docs/agents/doc-auditor.md`

The doc-auditor has `type: reviewer` and a `contract.must_not` that includes "Write or rewrite documentation content" and "Delete documentation files." Yet it has `Edit` and `Write` in its tool list. The contract tries to constrain what the tools allow, creating a tension: the tools permit actions the contract forbids.

The reviewer archetype defaults include Edit and Write (for producing review reports), so the tool list is technically archetype-compliant. But a reviewer with `must_not: Write documentation` having Write access is a weak enforcement model.

**Recommendation**: If the reviewer genuinely must not write docs, remove Write/Edit and route report production through a skill template or have it output to stdout. If Write is needed for the audit report (not documentation), clarify the contract to distinguish "documentation content" from "audit artifacts."

---

## Medium Findings

### M-01: 13 agents use "When Invoked" instead of "Approach" for their methodology section

**Severity**: MEDIUM
**Scope**: 13 agents across ecosystem, intelligence, rnd, strategy

The canonical template uses "## Approach" for the agent's working methodology. 13 agents use "## When Invoked" or "## When Invoked (First Actions)" instead. The content is functionally equivalent but the heading name differs.

**Recommendation**: Rename to "## Approach" for consistency. This is a mechanical change.

---

### M-02: No agents outside forge have multi-line `<example>` descriptions, but 41 agents have effective trigger phrases

**Severity**: MEDIUM
**Scope**: 41 non-forge specialists

While these agents lack the canonical multi-line format, they do have inline trigger phrases: `"Use when: X, Y, Z. Triggers: a, b, c."` This compressed format provides some CC routing signal, just less than the canonical format.

**Recommendation**: Prioritize C-01 migration for the 15 most-invoked agents first. The remaining agents can migrate incrementally.

---

### M-03: `type` field absent from all orchestrators but present in all non-orchestrators

**Severity**: MEDIUM
**Scope**: 11 orchestrators

Wait -- correction. Grep results show all orchestrators DO have `type: orchestrator`. This was a false signal from the previous session. Verified: all 58 agents have the `type` field.

**Status**: Not a finding. All agents are consistent.

---

### M-04: `role` field absent from all 11 orchestrators

**Severity**: MEDIUM
**Scope**: 11 orchestrators

All 47 non-orchestrator agents have a `role` field. All 11 orchestrators omit it. The CLAUDE.md agent table uses a "Role" column populated from this field. Orchestrators get their table entry from the description instead.

**Impact**: Minor inconsistency. The orchestrator description is sufficient for the table, but the absence of `role` means the frontmatter schema has an implicit conditional: "required for specialists, absent for orchestrators."

**Recommendation**: Either add `role` to orchestrators or document it as specialist-only in the schema.

---

### M-05: 6 agents in rnd rite omit Bash tool -- may limit diagnostic capability

**Severity**: MEDIUM
**Scope**: rnd/integration-researcher, rnd/moonshot-architect, rnd/tech-transfer, rnd/technology-scout (4 agents)

Four rnd agents lack the `Bash` tool. For research-oriented agents this may be intentional (they analyze and write, not execute), but `Bash` is needed for `git log`, `wc -l`, `go test`, and other diagnostic commands that even non-implementation agents may need.

**Recommendation**: Review whether these agents actually need shell access for their diagnostic workflows. If they only read and write docs, the omission is fine.

---

### M-06: Duplicate section "Cross-Rite" appears under 3 different names

**Severity**: MEDIUM
**Scope**: 19 agents

Cross-rite routing appears as: "Cross-Rite Routing" (10), "Cross-Rite Notes" (5), "Cross-Rite Protocol" (4). The archetype defines it as `cross-rite-protocol` for orchestrators. The canonical template calls it "Cross-Rite Routing."

**Recommendation**: Normalize to "Cross-Rite Routing" for specialists, "Cross-Rite Protocol" for orchestrators (matching respective archetypes).

---

### M-07: forge agents have frontmatter at lines 1-35; all others at lines 1-17

**Severity**: MEDIUM
**Scope**: 7 forge agents

Forge agents have ~30-line frontmatter blocks (due to multi-line descriptions with examples). All other rites have ~15-line frontmatter. This is not a defect -- forge follows the canonical template more closely -- but it means forge agents are structurally different from the rest.

**Impact**: When C-01 is resolved (migrating all agents to multi-line descriptions), all agents will have forge-style frontmatter. This finding will self-resolve.

---

## Low Findings

### L-01: 3 rites have color collisions within the rite

**Severity**: LOW
**Scope**: docs (2 blue agents), sre (2 orange-ish), security (2 red agents)

The canonical template says "Avoid duplicates within the same pantheon." Some rites have color overlaps:
- docs: doc-auditor (blue) and tech-writer (blue)
- security: orchestrator (red) and security-reviewer (red)

**Recommendation**: Reassign one agent in each collision pair.

---

### L-02: 48 agents use opus, only 6 use sonnet

**Severity**: LOW
**Scope**: All 58 agents

The canonical template recommends sonnet for engineers and documentation agents. Actual usage:
- **opus (48)**: All orchestrators, all analysts, all architects, all researchers, most engineers
- **sonnet (6)**: ecosystem/integration-engineer, forge/agent-curator, forge/platform-engineer, rnd/prototype-engineer, hygiene/janitor, all 4 docs agents

Most engineers use opus despite the template recommending sonnet. This is a cost decision: opus provides better judgment but costs more per turn.

**Recommendation**: Accept this as intentional. The model recommendation in the template should note "opus preferred for agents requiring deep reasoning; sonnet acceptable for high-turn-count implementation agents."

---

### L-03: "Anti-Patterns to Avoid" vs "Anti-Patterns" heading split

**Severity**: LOW
**Scope**: 57 agents (33 use "Anti-Patterns", 24 use "Anti-Patterns to Avoid")

Canonical template says "Anti-Patterns to Avoid." The archetype says `anti-patterns`. Minor inconsistency.

**Recommendation**: Pick one. "Anti-Patterns" is shorter and the archetype uses it. Update the canonical template.

---

### L-04: `contract.must_not` enforcement is behavioral, not mechanical

**Severity**: LOW
**Scope**: 7 reviewer agents

The `contract.must_not` field in reviewer frontmatter contains natural-language constraints like "Write or rewrite documentation content." CC does not enforce these mechanically -- they are instructions in the system prompt, not tool restrictions.

**Impact**: If a reviewer agent "decides" to violate its must_not, there is no guardrail. The `disallowedTools: [Task]` field IS mechanically enforced by CC. The `must_not` field is not.

**Recommendation**: Document that `must_not` is advisory (prompt-level constraint) while `disallowedTools` is enforced (CC strips these tools). Consider whether critical must_not items should be mapped to disallowedTools where possible.

---

### L-05: Handoff Criteria present in all 58 agents but quality varies

**Severity**: LOW
**Scope**: All 58 agents

All agents have "## Handoff Criteria" (good -- 100% compliance). But quality ranges from 3-item checklists to 10-item checklists. The canonical template recommends 5-10 items.

Agents with fewer than 5 criteria: ~12 agents across various rites.

**Recommendation**: Add criteria to thin checklists during the next content pass.

---

### L-06: `strategy/roadmap-strategist` embeds a full HANDOFF example that duplicates sprint-planner's

**Severity**: LOW
**Scope**: strategy/roadmap-strategist, debt-triage/sprint-planner

Both agents embed similar HANDOFF format templates. The HANDOFF schema is already available via the `cross-rite-handoff` skill. Embedding it in agents is redundant.

**Impact**: ~70 lines of duplicated template content across 2 agents.

**Recommendation**: Replace inline HANDOFF templates with `@cross-rite-handoff` skill references.

---

## Per-Agent Assessment

### Legend
- **TF**: Template Fidelity (section compliance)
- **FM**: Frontmatter Completeness
- **TS**: Tool Scoping (appropriate for role)
- **DA**: Domain Authority (decide/escalate/route)
- **HC**: Handoff Criteria quality
- **CC**: Context Cost (line count)
- **OP**: Orchestrator Pattern adherence
- **XR**: Cross-Rite consistency

Scale: PASS / WARN / FAIL

### 10x-dev

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 230 | PASS | WARN^1 | PASS | PASS | PASS | WARN | PASS | PASS |
| requirements-analyst | 189 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | WARN^4 |
| architect | 164 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | WARN^4 |
| principal-engineer | 157 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | WARN^4 |
| qa-adversary | 151 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |

### debt-triage

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 194 | PASS | WARN^1 | PASS | PASS | PASS | PASS | PASS | PASS |
| debt-collector | 168 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| risk-assessor | 163 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| sprint-planner | 216 | WARN^2 | WARN^3 | PASS | PASS | PASS | WARN | n/a | PASS |

### docs

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 200 | PASS | WARN^1 | PASS | PASS | PASS | PASS | PASS | PASS |
| doc-auditor | 233 | WARN^5 | WARN^3 | WARN^6 | PASS | PASS | FAIL | n/a | PASS |
| information-architect | 157 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| tech-writer | 157 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| doc-reviewer | 150 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |

### ecosystem

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 215 | PASS | WARN^1 | PASS | PASS | PASS | WARN | PASS | PASS |
| ecosystem-analyst | 161 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | WARN^7 |
| context-architect | 159 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | WARN^7 |
| integration-engineer | 159 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | WARN^7 |
| documentation-engineer | 153 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | WARN^7 |
| compatibility-tester | 145 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | WARN^7 |

### forge

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 214 | PASS | PASS | PASS | PASS | PASS | WARN | PASS | PASS |
| agent-designer | 224 | PASS | PASS | PASS | PASS | PASS | WARN | n/a | PASS |
| prompt-architect | 153 | PASS | PASS | PASS | PASS | PASS | PASS | n/a | PASS |
| workflow-engineer | 294 | WARN^8 | PASS | PASS | PASS | PASS | FAIL | n/a | PASS |
| platform-engineer | 257 | WARN^8 | PASS | PASS | PASS | PASS | FAIL | n/a | PASS |
| eval-specialist | 297 | WARN^8 | PASS | PASS | PASS | PASS | FAIL | n/a | PASS |
| agent-curator | 295 | WARN^8 | PASS | PASS | PASS | PASS | FAIL | n/a | PASS |

### hygiene

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 201 | PASS | WARN^1 | PASS | PASS | PASS | PASS | PASS | PASS |
| code-smeller | 176 | WARN^2 | WARN^3 | WARN^9 | PASS | PASS | PASS | n/a | PASS |
| architect-enforcer | 162 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| janitor | 157 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| audit-lead | 177 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |

### intelligence

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 211 | PASS | WARN^1 | PASS | PASS | PASS | WARN | PASS | PASS |
| analytics-engineer | 202 | WARN^2 | WARN^3 | PASS | PASS | PASS | WARN | n/a | PASS |
| user-researcher | 157 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| experimentation-lead | 203 | WARN^2 | WARN^3 | PASS | PASS | PASS | WARN | n/a | PASS |
| insights-analyst | 281 | WARN^2 | WARN^3 | PASS | PASS | PASS | FAIL | n/a | PASS |

### rnd

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 206 | PASS | WARN^1 | PASS | PASS | PASS | WARN | PASS | PASS |
| technology-scout | 152 | WARN^2 | WARN^3 | WARN^10 | PASS | PASS | PASS | n/a | PASS |
| integration-researcher | 148 | WARN^2 | WARN^3 | WARN^10 | PASS | PASS | PASS | n/a | PASS |
| prototype-engineer | 149 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| moonshot-architect | 143 | WARN^2 | WARN^3 | WARN^10 | PASS | PASS | PASS | n/a | PASS |
| tech-transfer | 301 | WARN^2 | WARN^3 | WARN^10 | PASS | PASS | FAIL | n/a | PASS |

### security

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 211 | PASS | WARN^1 | PASS | PASS | PASS | WARN | PASS | PASS |
| threat-modeler | 174 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| compliance-architect | 161 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| penetration-tester | 177 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| security-reviewer | 179 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |

### sre

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 201 | PASS | WARN^1 | PASS | PASS | PASS | PASS | PASS | PASS |
| observability-engineer | 159 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| incident-commander | 158 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| platform-engineer | 153 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| chaos-engineer | 152 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |

### strategy

| Agent | Lines | TF | FM | TS | DA | HC | CC | OP | XR |
|-------|-------|----|----|----|----|----|----|----|----|
| orchestrator | 201 | PASS | WARN^1 | PASS | PASS | PASS | PASS | PASS | PASS |
| market-researcher | 149 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| competitive-analyst | 153 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| business-model-analyst | 155 | WARN^2 | WARN^3 | PASS | PASS | PASS | PASS | n/a | PASS |
| roadmap-strategist | 246 | WARN^2 | WARN^3 | PASS | PASS | PASS | FAIL | n/a | PASS |

### Footnotes

1. ^1: Orchestrator missing `role` field (all 11); color not purple (9/11)
2. ^2: Section heading variants (see H-02); missing `<example>` in description
3. ^3: Single-line description; missing "When to use" and `<example>` blocks (see C-01)
4. ^4: Uses "Related Skills" instead of "Skills Reference"
5. ^5: Reviewer typed but has Edit/Write tools with contract tension (see H-08)
6. ^6: Write tool present despite must_not contract constraint
7. ^7: Uses "When Invoked" / "When Invoked (First Actions)" instead of "Approach"; uses "Responsibilities" instead of "Core Responsibilities"
8. ^8: Embeds reference content that should be in skills (see C-03)
9. ^9: Has Write but not Edit; analysis agent with file creation capability (see H-04)
10. ^10: Missing Bash tool; may limit diagnostic capability (see M-05)

---

## Recommendations (ordered by impact)

### Tier 1: Structural fixes (high impact, systematic)

1. **Update archetype defaults to match reality** (C-02)
   - Change maxTurns: orchestrator 3->40, specialist 25->150, reviewer 15->100
   - Or add documentation explaining why overrides are expected
   - Effort: S (1-2 hours)

2. **Update canonical template frontmatter schema** (C-04)
   - Add: `role`, `type`, `maxTurns`, `disallowedTools`, `contract`
   - Document which fields are per-archetype
   - Effort: S (1-2 hours)

3. **Add type->archetype mapping** (C-05)
   - Add `analyst`, `designer`, `engineer` as archetype aliases in `archetype.go`
   - Or document the mapping explicitly in the template
   - Effort: S (2-4 hours)

4. **Standardize orchestrator colors to purple** (H-07)
   - Mechanical change across 9 orchestrator files
   - Effort: XS (30 minutes)

### Tier 2: Description migration (highest ROI for CC routing)

5. **Migrate 41 agents to multi-line descriptions** (C-01)
   - Prioritize by invocation frequency
   - Add "When to use" and `<example>` blocks
   - Effort: L (8-16 hours across multiple sessions)

### Tier 3: Content extraction (context cost reduction)

6. **Extract embedded reference content to skills** (C-03)
   - HANDOFF templates -> `@cross-rite-handoff` skill
   - Validation checklists -> per-rite skills
   - CLI references -> platform skills
   - Pattern libraries -> domain skills
   - Target: 10 agents below 200 lines
   - Effort: M (4-8 hours)

7. **Implement orchestrator boilerplate injection** (H-01)
   - Platform sections generated during materialization
   - Author only maintains Domain Authority, Phase Routing, Cross-Rite Protocol
   - Effort: L (8-16 hours; requires materialization pipeline changes)

### Tier 4: Normalization (consistency polish)

8. **Normalize section headings** (H-02, M-01, M-06, L-03)
   - "When Invoked" -> "Approach" (13 agents)
   - "Related Skills" -> "Skills Reference" (6 agents)
   - "Anti-Patterns to Avoid" <-> "Anti-Patterns" (pick one)
   - "Cross-Rite Notes" -> "Cross-Rite Routing" (5 agents)
   - Effort: M (4-8 hours; mechanical but touches many files)

9. **Resolve reviewer tool/contract tension** (H-08, H-03)
   - Review WebFetch/WebSearch in reviewer archetype
   - Clarify doc-auditor Write access vs must_not
   - Effort: S (2-4 hours)

10. **Add `role` to orchestrators or document as optional** (M-04)
    - Effort: XS (30 minutes)

---

## Appendix: Stale Terminology Check

The SL-008 terminology deep cleanse (completed 2026-02-09, 9 commits) resolved all "team" -> "rite" references. Grep of all agent files for "team" returns **zero matches**. The stale references noted in the previous session summary have been cleaned.

---

## Appendix: Token Budget Analysis

Estimated per-invocation token cost by archetype (prompt only, excluding conversation):

| Category | Avg Lines | Est. Tokens | Agents | Total Budget |
|----------|-----------|-------------|--------|-------------|
| Orchestrator | 208 | ~830 | 11 | ~9,130 |
| Specialist (<200 lines) | 158 | ~630 | 27 | ~17,010 |
| Specialist (>200 lines) | 252 | ~1,010 | 13 | ~13,130 |
| Reviewer | 163 | ~650 | 7 | ~4,550 |

The 13 oversized specialists carry ~4,940 extra tokens vs. the <200-line average. At CC rates, this is a meaningful overhead per session, especially for agents invoked multiple times.

---

*Generated by Context Engineer audit, 2026-02-09. All findings reference source files in `rites/*/agents/*.md` against `rites/forge/mena/rite-development/templates/agent-template.md` and `internal/agent/archetype.go`.*

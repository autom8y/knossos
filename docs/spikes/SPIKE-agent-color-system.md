# SPIKE: Agent Color System Design

**Status**: COMPLETE
**Date**: 2026-02-19
**Duration**: 1 hour
**Scope**: Agent color validation, Pythia identity colors, role-archetype palette, non-standard color remediation

## Executive Summary

Validated the hybrid color system proposal (Pythia = rite identity color, specialists = role-archetype color) against the actual codebase. Found **13 agents using invalid/non-standard colors** that Claude Code cannot render, a **missing color (yellow) from Knossos documentation**, and confirmed the proposal is sound with minor adjustments needed.

### Key Findings

1. **CC Valid Palette**: 8 named colors: red, blue, green, yellow, purple, orange, pink, cyan
2. **Knossos Template Gap**: Forge template documents only 7 colors (missing yellow)
3. **Invalid Colors in Production**: 13 agents use colors CC does not recognize (crimson, magenta, gold, indigo, rust, bone, amber, steel)
4. **Current Pythia Colors**: Already mostly aligned with the proposal -- 10 of 13 match
5. **Hybrid System**: Viable and already partially in effect organically

### Decision

Adopt Approach C (Hybrid) with the remediation plan below. Update forge templates to include yellow. Fix all 13 invalid-color agents.

---

## 1. Claude Code Color Contract

### 1.1 Valid Colors (Verified)

Source: CC `/agents` interactive wizard, confirmed via [CC subagent documentation](https://code.claude.com/docs/en/sub-agents) and third-party documentation.

| Color | Available | In Knossos Template |
|-------|-----------|-------------------|
| red | Yes | Yes |
| blue | Yes | Yes |
| green | Yes | Yes |
| yellow | Yes | **NO** (missing) |
| purple | Yes | Yes |
| orange | Yes | Yes |
| pink | Yes | Yes |
| cyan | Yes | Yes |

The CC agent schema does not enforce an enum -- the `color` field accepts any string matching `^[a-z]+$`. However, CC only renders the 8 colors above. Non-standard values may be silently ignored or produce undefined UI behavior.

### 1.2 Knossos Schema

The agent schema (`internal/validation/schemas/agent.schema.json`) uses pattern `^[a-z]+$` with no enum constraint. There is no Go-level validation of agent colors. The only documented palette is in the forge template, which lists 7 of 8 valid colors.

---

## 2. Current State Inventory

### 2.1 Color Distribution (78 agent assignments across 13 rites + shared)

| Color | Count | Valid? |
|-------|-------|--------|
| cyan | 16 | Yes |
| orange | 14 | Yes |
| green | 11 | Yes |
| red | 9 | Yes |
| purple | 8 | Yes |
| blue | 5 | Yes |
| **magenta** | **4** | **NO** |
| yellow | 2 | Yes |
| **gold** | **2** | **NO** |
| **crimson** | **2** | **NO** |
| **steel** | **1** | **NO** |
| **rust** | **1** | **NO** |
| pink | 1 | Yes |
| **indigo** | **1** | **NO** |
| **bone** | **1** | **NO** |
| **amber** | **1** | **NO** |

**13 agents (17%) use invalid colors.**

### 2.2 Current Pythia Colors by Rite

| Rite | Current Pythia Color | Proposed | Match? |
|------|---------------------|----------|--------|
| 10x-dev | blue | blue | Yes |
| arch | cyan | cyan | Yes |
| debt-triage | orange | orange | Yes |
| docs | green | green | Yes |
| ecosystem | purple | purple | Yes |
| forge | cyan | cyan (or blue) | Yes |
| hygiene | green | green | Yes |
| intelligence | cyan | cyan (adjusted) | Yes |
| rnd | purple | purple | Yes |
| security | red | red | Yes |
| slop-chop | **crimson** (invalid) | red | Needs fix |
| sre | orange | orange | Yes |
| strategy | yellow | yellow | Yes |

**12 of 13 Pythia colors already match the proposal.** Only slop-chop needs remediation (crimson -> red).

### 2.3 Invalid Color Agents (Remediation Required)

| Agent | Rite | Current (Invalid) | Proposed (Valid) | Rationale |
|-------|------|--------------------|-----------------|-----------|
| requirements-analyst | 10x-dev | magenta | pink | User-facing requirements gathering = researcher/empathy archetype |
| sprint-planner | debt-triage | magenta | yellow | Planner/coordinator archetype |
| documentation-engineer | ecosystem | magenta | blue | Knowledge/documentation/curator archetype |
| user-researcher | intelligence | magenta | pink | Human-facing research = researcher/empathy archetype |
| moirai | shared (.claude) | indigo | purple | Fates mythology = visionary/strategic archetype |
| theoros | shared | gold | yellow | Sacred observer = illumination/insight archetype |
| pythia | slop-chop | crimson | red | Rite identity = aggressive cutting |
| cruft-cutter | slop-chop | rust | orange | Degradation finder = analyst/observer archetype |
| gate-keeper | slop-chop | bone | cyan | Structural gatekeeper = architect archetype |
| hallucination-hunter | slop-chop | crimson | red | Adversary archetype |
| logic-surgeon | slop-chop | amber | yellow | Careful analytical work = planner/assessor archetype |
| remedy-smith | slop-chop | steel | green | Builder/fixer = builder/implementer archetype |

---

## 3. The Hybrid Color System

### 3.1 Two Rules

**Rule 1 -- Pythia carries the rite's identity color.** This color evokes what the rite is about. It is the "brand" of the rite.

**Rule 2 -- Specialist agents use a role-archetype palette.** The color signals what the agent *does*, not which rite it belongs to.

### 3.2 Role-Archetype Palette

| Color | Archetype | Example Agents |
|-------|-----------|---------------|
| red | Adversary / Auditor / Breaker | qa-adversary, audit-lead, security-reviewer, eval-specialist |
| cyan | Architect / Structural Thinker | architect, architect-enforcer, compliance-architect, structure-evaluator |
| green | Builder / Engineer / Implementer | principal-engineer, integration-engineer, workflow-engineer, janitor |
| orange | Analyst / Observer / Scout | ecosystem-analyst, analytics-engineer, technology-scout, threat-modeler |
| purple | Strategist / Visionary | moonshot-architect, roadmap-strategist, insights-analyst |
| blue | Knowledge / Documentation / Curator | doc-auditor, tech-writer, agent-curator, tech-transfer |
| yellow | Planner / Coordinator / Assessor | risk-assessor, sprint-planner |
| pink | Researcher / Human-Facing / Empathy | user-researcher, requirements-analyst |

### 3.3 Rite Identity Colors (Pythia)

| Rite | Pythia Color | Vibe |
|------|-------------|------|
| 10x-dev | blue | Craftsmanship, depth, trust |
| arch | cyan | Structural clarity, blueprints |
| debt-triage | orange | Warning/urgency, triage |
| docs | green | Growth, freshness, living docs |
| ecosystem | purple | Holistic, strategic, big picture |
| forge | cyan | Creative structure, precise creation |
| hygiene | green | Clean, healthy, fresh |
| intelligence | cyan | Precision, data, clarity |
| rnd | purple | Exploration, imagination, moonshot |
| security | red | Danger-awareness, alertness |
| slop-chop | red | Aggressive cutting, no mercy |
| sre | orange | Operational vigilance |
| strategy | yellow | Illumination, planning |

**Shared identity colors are acceptable** -- forge/arch/intelligence share cyan; hygiene/docs share green; rnd/ecosystem share purple; sre/debt-triage share orange; security/slop-chop share red. Rites with similar energy naturally cluster.

### 3.4 Harmonic Validation

When Pythia's rite-identity color matches the dominant archetype of the rite, the palette feels harmonious:
- Security (red Pythia) surrounded by adversarial agents (red) -- harmonic
- Strategy (yellow Pythia) surrounded by planners (yellow) -- harmonic
- Docs (green Pythia) surrounded by knowledge workers (blue) -- Pythia stands apart as oracle

Both patterns work thematically: harmony reinforces the rite's identity; contrast signals "I'm the oracle overseeing this domain."

---

## 4. Conflict Analysis

### 4.1 Uniqueness Not Required

The proposal does not require every rite to have a unique Pythia color. With 13 rites and 8 colors, uniqueness is mathematically impractical. The system works because:
- Color carries *vibe*, not *identity* -- the rite name is the identifier
- Within a rite, Pythia + specialists create a palette, not a single color
- Cross-rite scanning uses agent *names* not colors

### 4.2 Two Reds in slop-chop

Pythia (red) and hallucination-hunter (red) share a color. This is acceptable because both are adversarial in nature. The rite is aggressively deletion-oriented, and double-red reinforces that identity.

### 4.3 No Pink Over-Use

Pink is used sparingly (2 agents: requirements-analyst, user-researcher). This is correct -- the researcher/empathy archetype is genuinely rare.

---

## 5. Documentation Updates Required

### 5.1 Forge Template (3 files)

Add `yellow` to the valid color list:

| File | Current | Proposed |
|------|---------|----------|
| `rites/forge/mena/agent-prompt-engineering/template.md` | `{purple\|pink\|cyan\|green\|red\|orange\|blue}` | `{purple\|pink\|cyan\|green\|red\|orange\|blue\|yellow}` |
| `rites/forge/mena/agent-prompt-engineering/validation/checklist.md` | `One of: purple, pink, cyan, green, red, orange, blue` | `One of: purple, pink, cyan, green, red, orange, blue, yellow` |
| `rites/forge/mena/rite-development/templates/agent-template.md` | `purple \| pink \| cyan \| green \| red \| orange \| blue` | `purple \| pink \| cyan \| green \| red \| orange \| blue \| yellow` |

### 5.2 Agent Schema (Optional Enhancement)

Could add an enum constraint to `agent.schema.json`:

```json
"color": {
  "type": "string",
  "enum": ["red", "blue", "green", "yellow", "purple", "orange", "pink", "cyan"],
  "description": "Display color for agent identification (CC valid palette)"
}
```

This would catch invalid colors at validation time rather than allowing silent rendering failures.

---

## 6. Remediation Plan

### Phase 1: Fix Invalid Colors (13 agents, mechanical)

All changes are to frontmatter `color:` field only. No behavioral changes.

```
# slop-chop (6 agents -- all colors invalid)
rites/slop-chop/agents/pythia.md:           crimson -> red
rites/slop-chop/agents/cruft-cutter.md:     rust    -> orange
rites/slop-chop/agents/gate-keeper.md:      bone    -> cyan
rites/slop-chop/agents/hallucination-hunter.md: crimson -> red
rites/slop-chop/agents/logic-surgeon.md:    amber   -> yellow
rites/slop-chop/agents/remedy-smith.md:     steel   -> green

# magenta agents (4 agents -- magenta not in CC palette)
rites/10x-dev/agents/requirements-analyst.md:    magenta -> pink
rites/debt-triage/agents/sprint-planner.md:      magenta -> yellow
rites/ecosystem/agents/documentation-engineer.md: magenta -> blue
rites/intelligence/agents/user-researcher.md:    magenta -> pink

# shared/platform agents (3 agents)
rites/shared/agents/theoros.md:     gold   -> yellow
.claude/agents/moirai.md:          indigo -> purple
```

Note: theoros appears in both rites/shared/agents/ and may be materialized to .claude/agents/ by the sync pipeline. Fix the source.

### Phase 2: Update Documentation (3 files)

Add yellow to all forge template color lists.

### Phase 3: Optional Schema Enforcement

Add enum to agent.schema.json to prevent future drift.

---

## 7. Follow-Up Actions

- [ ] **P0**: Fix 13 invalid-color agents (Phase 1 above)
- [ ] **P0**: Add yellow to forge template color lists (Phase 2)
- [ ] **P1**: Consider adding enum constraint to agent.schema.json (Phase 3)
- [ ] **P2**: Add `ari lint` check for agent colors against CC palette
- [ ] **P2**: Document the hybrid color system in doctrine (role-archetype palette + Pythia identity)
- [ ] **P3**: Validate color rendering behavior when CC encounters unknown color values (does it fall back to no-color, or error?)

---

## Appendix: Full Rite Palette Visualization

```
10x-dev:       Pythia(blue)  architect(cyan) principal-eng(green) qa-adversary(red) req-analyst(pink*)
arch:          Pythia(cyan)  dep-analyst(purple) remediation-planner(pink) structure-eval(cyan) topo-cart(orange)
debt-triage:   Pythia(orange) debt-collector(orange) risk-assessor(yellow) sprint-planner(yellow*)
docs:          Pythia(green)  doc-auditor(blue) doc-reviewer(red) info-architect(cyan) tech-writer(blue)
ecosystem:     Pythia(purple) compat-tester(red) context-arch(cyan) doc-engineer(blue*) eco-analyst(orange) integ-eng(green)
forge:         Pythia(cyan)  agent-designer(purple) prompt-arch(cyan) workflow-eng(green) platform-eng(orange) eval-spec(red) agent-curator(blue)
hygiene:       Pythia(green)  arch-enforcer(cyan) audit-lead(red) code-smeller(orange) janitor(green)
intelligence:  Pythia(cyan)  analytics-eng(orange) experimentation-lead(cyan) insights-analyst(purple) user-researcher(pink*)
rnd:           Pythia(purple) integ-researcher(cyan) moonshot-arch(purple) prototype-eng(green) tech-transfer(blue) tech-scout(orange)
security:      Pythia(red)   compliance-arch(cyan) pen-tester(green) security-reviewer(red) threat-modeler(orange)
slop-chop:     Pythia(red*)  cruft-cutter(orange*) gate-keeper(cyan*) hallucination-hunter(red*) logic-surgeon(yellow*) remedy-smith(green*)
sre:           Pythia(orange) chaos-eng(red) incident-cmd(purple) observability-eng(orange) platform-eng(cyan)
strategy:      Pythia(yellow) business-model(green) competitive-analyst(cyan) market-researcher(orange) roadmap-strategist(purple)

* = color changed from invalid to valid in this proposal
```

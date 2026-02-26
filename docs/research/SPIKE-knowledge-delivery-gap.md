# SPIKE: Knowledge-Delivery Gap Analysis

> **Spike ID**: spike-20260226-165553
> **Created**: 2026-02-26
> **Question**: Does knossos have a systemic knowledge-delivery gap where operational knowledge exists in skills but lacks discovery paths to agents?
> **Timebox**: 1 hour
> **Deliverable**: decision
> **Status**: complete

## Executive Summary

**Verdict: CONFIRMED — gaps are larger than initially estimated.**

Knossos encodes operational knowledge into skills (legomena) but fails to wire those skills to the agents that need them. The platform has a robust *knowledge layer* and a robust *enforcement layer* but a weak *delivery layer* connecting them. This creates a systemic pattern where agents bypass conventions not out of defiance but out of ignorance — they literally cannot discover the relevant skill.

## The Knowledge-Delivery Gap Model

Every operational convention needs a **discovery path** from agent context to the relevant skill. Four reliability tiers exist:

| Path | Reliability | Token Cost | Mechanism |
|------|-------------|------------|-----------|
| **Direct** (skill preload via frontmatter) | Highest | ~500-800 tok | `skills: [conventions]` in agent frontmatter |
| **Referenced** (prompt mention) | Medium | ~30 tok | "Load commit:behavior for git ops" in prompt body |
| **Autonomous** (CC keyword matching) | Lowest | Zero | CC's Skill tool description matching |
| **Enforcement-driven** (hook denial + skill ref) | Reactive but reliable | Zero | "Load skill X" in denial reason string |

Currently, most operational knowledge relies entirely on the **Autonomous** path — which is probabilistic and unreliable.

## Validated Findings

### Gap 1: guidance/standards — Implementation Agents Unprotected

**Claim**: Loaded by 2 non-implementation agents. Zero implementation agents have it.
**Actual**: **CONFIRMED exactly.**

| Metric | Count |
|--------|-------|
| Agents WITH guidance/standards | 2 |
| — compatibility-tester (ecosystem, reviewer role) | 1 |
| — risk-assessor (debt-triage, analyst role) | 1 |
| Implementation agents WITHOUT | 59 |

**Key agents missing it**: principal-engineer, janitor, prototype-engineer, integration-engineer, architect — the five agents most likely to produce code that should follow standards.

**Skill location**: `mena/guidance/standards/INDEX.lego.md` with 8 sub-files covering code conventions, repository map, tech stack (Go, Python, API, core, infrastructure), and tool selection.

**Impact**: HIGH — implementation agents write code without access to naming conventions, tech stack decisions, or repository organization guidelines.

### Gap 2: guidance/file-verification — Write-Capable Agents Unprotected

**Claim**: Loaded by 3 ecosystem agents. ~50 agents with Write access unprotected.
**Actual**: **CONFIRMED and understated.** 65 unprotected, not ~50.

| Metric | Count |
|--------|-------|
| Total agents with Write access | 68 |
| Agents WITH file-verification | 3 (4.4%) |
| Agents WITHOUT file-verification | 65 (95.6%) |

**Protected agents** (all ecosystem rite): compatibility-tester, context-architect, ecosystem-analyst.

**Unprotected across 14 rites**: Every rite has Write-capable agents without this skill. Notable: integration-engineer and documentation-engineer in the ecosystem rite itself lack it.

**Skill purpose**: Prevent hallucinated file operations by mandating Write → Read → Confirm verification protocol.

**Impact**: MEDIUM — 95.6% of Write-capable agents can claim file creation without verification, breaking artifact handoff chains.

### Gap 3: validate.go Denial Messages — No Skill References

**Claim**: Blocks force-push, --no-verify, reset --hard but doesn't reference skills.
**Actual**: **CONFIRMED exactly.** Proven by counter-example.

**validate.go** (`internal/cmd/hook/validate.go`) has 5 denial messages, **zero skill references**:

| Line | Blocked Operation | Denial Message | Skill Reference |
|------|-------------------|----------------|-----------------|
| 162 | rm -rf protected | "Cannot rm -rf protected path: {path}" | None |
| 169 | Force push main | "Use --force-with-lease or push to feature branch" | None |
| 174 | --no-verify | "Pre-commit hooks exist for a reason" | None |
| 179 | git reset --hard | "Use git stash or git checkout" | None |
| 184 | git clean -fd | "Use git stash or manual cleanup" | None |

**Counter-example proving the pattern works**: `gitconventions.go` (line 34-36):
```go
const conventionDenyReason = "Commit message does not follow conventional format. " +
    "Load skill commit:behavior for full specification. " +
    "Expected: type(scope): subject..."
```

This denial message explicitly names the skill. The test file (`gitconventions_test.go`) even verifies the skill reference is present. This is the **carrot-stick-recovery pattern** working correctly — and validate.go simply needs the same treatment.

**Hook comparison**:

| Hook | File | Skill Ref in Denial? |
|------|------|---------------------|
| git-conventions | gitconventions.go | YES — "Load skill commit:behavior" |
| writeguard | writeguard.go | PARTIAL — Task(moirai) recovery, no skill ref |
| validate | validate.go | NO — generic guidance only |
| agent-guard | agentguard.go | NO |

**Impact**: LOW effort, completes existing pattern. ~10 line change.

### Gap 4: Tier 2 Agent Discovery — Prompt-Level Skill References

**Claim**: ~9 agents with Bash+Skill tools have no prompt-level discovery path.
**Actual**: **MASSIVELY UNDERSTATED.** 46 agents affected, not ~9.

| Metric | Count |
|--------|-------|
| Total agents with Bash+Skill tools | 48 |
| With explicit skill discovery in prompt | 2 (4.2%) |
| Without explicit skill discovery | 46 (95.8%) |

**Only 2 agents have explicit skill discovery guidance**:
1. `clinic/pathologist` — "Playbook Loading" section with skill-to-system mapping table
2. `shared/theoros` — "Load skills as needed for domain context"

**46 agents across all rites** have Bash+Skill tools but zero prompt-level instruction on when/how to load operational skills. Distribution: 10x-dev (4), clinic (3), debt-triage (3), docs (4), ecosystem (5), forge (6), hygiene (4), intelligence (4), rnd (1), security (4), sre (4), strategy (4).

**Critical nuance**: Some agents have `skills: [conventions]` in frontmatter (preloaded) but zero mention in prompt body. The frontmatter preloads the skill at spawn, but the agent's prompt doesn't reinforce when to use it — a weaker form of the gap.

**Impact**: MEDIUM — fixable with ~30 tokens per agent (one-line reference).

## The Carrot-Stick-Recovery Pattern

The three-layer model generalizes beyond git conventions:

```
1. CARROT: Skill preloaded via frontmatter → agent has knowledge at spawn
2. STICK:  Hook enforces convention → malformed action denied
3. RECOVERY: Denial names the skill → agent self-corrects on next attempt
```

**Current state by domain**:

| Domain | Carrot | Stick | Recovery | Status |
|--------|--------|-------|----------|--------|
| Commit format | conventions preload | gitconventions hook | "Load skill commit:behavior" | COMPLETE |
| Git safety | conventions preload | validate hook | Generic guidance only | MISSING Recovery |
| File verification | 3 agents only | (no hook) | N/A | MISSING Carrot + Stick |
| Code standards | 2 agents only | (no hook) | N/A | MISSING Carrot |
| Session writes | (not applicable) | writeguard hook | Task(moirai) guidance | PARTIAL |

**Where it applies**: Objective, enforceable conventions (commit format, git safety, file verification).
**Where it does NOT apply**: Subjective behaviors (architecture decisions, code quality, naming aesthetics).

## Revised Wave Recommendations

Based on validated findings, the original wave plan is adjusted:

### Wave 1: Skill Wiring (HIGH impact, LOW effort)

| Action | Agents Affected | Effort | Impact |
|--------|----------------|--------|--------|
| Wire guidance/standards to implementation agents | ~10 highest-priority | 15 min | HIGH |
| Wire guidance/file-verification to artifact-producing agents | ~10 highest-priority | 15 min | MEDIUM |

**Priority targets for guidance/standards**: principal-engineer, architect, integration-engineer (all rites), janitor, prototype-engineer, platform-engineer (forge, sre).

**Priority targets for file-verification**: documentation-engineer, integration-engineer (ecosystem), tech-writer (docs), platform-engineer (forge, sre), principal-engineer (10x-dev).

**Not recommended**: Wiring ALL 65 agents. Apply to the ~10-15 highest-impact agents first, measure effect, then expand.

### Wave 2: Hook Denial Skill References (LOW effort, completes pattern)

| Action | Files | Effort | Impact |
|--------|-------|--------|--------|
| Add skill references to validate.go denial messages | 1 file, ~10 lines | 10 min | LOW |
| Add skill references to agentguard.go denial messages | 1 file, ~5 lines | 5 min | LOW |

**Pattern to follow**: gitconventions.go line 34-36. Each denial message adds "Load skill {name} for {guidance}."

### Wave 3: Prompt-Level Skill Discovery (MEDIUM effort, MEDIUM impact)

| Action | Agents Affected | Effort | Impact |
|--------|----------------|--------|--------|
| Add 1-line skill references to agent prompts | 46 agents (prioritize ~15) | 30 min | MEDIUM |

**Recommended pattern** (minimal, ~30 tokens):
```markdown
Load `conventions` skill for git operations and code standards.
```

**Or tabular** (pathologist pattern, ~80 tokens):
```markdown
## Skill Loading
| Skill | Load When |
|-------|-----------|
| `conventions` | Git commits, code standards |
| `guidance/standards` | Tech stack, naming conventions |
```

**Priority**: Agents in 10x-dev, ecosystem, forge, hygiene first (highest code-production volume).

### Wave 4: Systematic Audit (META, deferred)

Full agent-capability-to-skill mapping audit. Produces a matrix of every agent's tools vs. available skills, identifying remaining gaps. Recommended after Waves 1-3 are measured.

## Decision

**Recommendation: Proceed with Waves 1-3 as implementation tasks.**

- Wave 1 is highest ROI: 30 minutes of frontmatter edits protecting the most impactful agents.
- Wave 2 completes an already-proven pattern (gitconventions.go).
- Wave 3 addresses the largest gap (46 agents) with minimal per-agent cost.
- Wave 4 deferred until Waves 1-3 effects are observable.

**Conversion path**: Each wave becomes a `/task` at PATCH complexity.

## Artifacts

- This report: `docs/research/SPIKE-knowledge-delivery-gap.md`
- No POC code (analysis spike, not implementation)
- No benchmarks (qualitative analysis)

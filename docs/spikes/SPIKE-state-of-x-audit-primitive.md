# SPIKE: "State of the {X}" Audit Primitive

> **Session**: session-20260209-230451-d4df362a
> **Date**: 2026-02-09
> **Rite**: forge
> **Status**: Complete — ready for Phase 2 (Interview)

---

## Question

Which CC primitive type should the "State of the {X}" audit pattern become, how should it compose with existing primitives, and what's the schema for the audit spec?

## Decision This Informs

Whether to build a Dromena, Legomena, Workflow, Agent, Forge template, or composite primitive — and how to parameterize the N-agent swarm for parallel domain auditing.

---

## Approach

Two parallel research agents grounded in doctrine:
1. **context-engineer**: Evaluated each primitive type against the mena model, existing precedent, and mortal limits doctrine
2. **claude-code-guide**: Researched CC features (Task parallelism, background agents, progressive disclosure, token economics)

Additionally: manual review of existing dromena patterns (`/new-rite`, `/validate-rite`, `/eval-agent`, `/spike`), the rite-pack schema, workflow types, and the shared mena structure.

---

## Findings

### The Pattern Decomposed

| Operation | Nature | CC Primitive |
|-----------|--------|-------------|
| **Invocation** | User decides to audit | Dromena |
| **Parameterization** | Target, domains, depth | Dromena args |
| **Audit schemas/rubrics** | Reference knowledge (persistent) | Legomena |
| **N-agent dispatch** | Parallel Task calls from main thread | Agent via Task tool |
| **Per-domain evaluation** | Independent subagent work | Agent (isolated context) |
| **Synthesis** | Consolidate N reports | Main thread (or synthesis agent) |
| **Output artifact** | "State of the {X}" document | Side effect (dromena territory) |

### Comparison Matrix

| Criterion | A: Dro | B: Lego | C: Agent | D: Dro+Agent | **E: Dro+Lego+Agent** | F: Workflow | G: Forge Tpl |
|-----------|--------|---------|----------|-------------|----------------------|------------|-------------|
| **Fit score** | 2/5 | 1/5 | 2/5 | 3/5 | **5/5** | 3/5 | 2/5 |
| User invocation | Yes | No | No | Yes | **Yes** | Yes | Yes |
| Knowledge persistence | No | Yes | No | No | **Yes** | Yes | Yes |
| N-agent dispatch | Yes* | No | No | Yes | **Yes** | Awkward | Overkill |
| Progressive disclosure | No | Yes | No | No | **Yes** | Yes | Yes |
| Token efficiency | Poor | N/A | Poor | Medium | **Good** | Medium | Poor |
| Mena model fit | Violates | Partial | Partial | Partial | **Clean** | Partial | Overkill |

*With inlined knowledge, defeating token efficiency.

### Why Option E Wins

**Each primitive does exactly what it's designed for:**

- **Dromena** (`/state-of <target>`): Thin entry point (~80 lines). User invocation, argument parsing, dispatch orchestration. TRANSIENT lifecycle.
- **Legomena** (`state-of-ref/`): Domain registry, grading rubric, per-domain evaluation criteria, report/synthesis schemas. PERSISTENT lifecycle. Progressive disclosure via subdirectories.
- **Agent** (generic `domain-auditor`): Executes evaluation. Receives domain-specific criteria from dispatch prompt (constructed by main thread reading legomena). TRANSIENT lifecycle.

**The dromena is the trigger. The legomena is the brain. The agents are the hands.**

### Why Other Options Fail

| Option | Fatal Flaw |
|--------|-----------|
| **A: Dromena only** | Stuffing persistent knowledge (rubrics, schemas) into a transient primitive violates mena lifecycle model |
| **B: Legomena only** | Cannot initiate, cannot dispatch, no side effects — reference only |
| **C: Agent only** | Agents cannot spawn agents. One agent doing 7 domains = context window monolith |
| **D: Dromena + Agent** | Knowledge layer (rubrics, schemas) has no home — defaults to bloated dromena or inlined prompts |
| **F: Workflow/Rite** | Creates entire rite for a cross-cutting operation. N-agent parallel dispatch doesn't map to sequential workflow phases. Chicken-and-egg: audit rite would lose context of rite being audited |
| **G: Forge Template** | Over-engineering. Audit structure is constant; what varies is data (domains), not structure |

### CC Feature Validation

| Feature | Supports Pattern? | Notes |
|---------|-------------------|-------|
| Task tool parallelism | **YES** | Main thread can launch multiple Task calls concurrently |
| `run_in_background` | **YES** | Launch N agents, collect results after completion |
| Agent isolation | **YES** | Each agent gets own context (~200K tokens), work doesn't bloat main session |
| No agent nesting | **Constraint** | Main thread must be the dispatcher — agents cannot spawn agents |
| Skills progressive disclosure | **YES** | INDEX.lego.md description in context; full content loads on-demand |
| Token budget awareness | **NO** | No API for agents to check remaining budget; auto-compaction handles limits |

### Composition with Existing Primitives

| Primitive | Relationship |
|-----------|-------------|
| `/spike` | **Sequential predecessor** — spike researches what to audit, then `/state-of` executes it |
| `/qa` | **Sequential successor** — audit findings generate work items, `/qa` validates fixes |
| `/wrap` | **Session conclusion** — captures audit session state for handoff |
| `/validate-rite` | **Structural precedent** — same pattern: dromena triggers, agent evaluates, legomena provides schemas |
| `/new-rite` | **Structural precedent** — dromena + forge-ref legomena + forge agents |

### N-Agent Swarm Parameterization

**Recommendation: Registry in legomena with auto-discovery fallback.**

The legomena should contain a domain registry mapping targets to domain lists:

| Target Pattern | Domains | N |
|---------------|---------|---|
| `framework` | dromena, legomena, agents, inscription, rules, hooks, pipeline | 7 |
| `rite:{name}` | agents, workflow, mena, manifest | 4 |
| `pipeline` | sync, materialize, hooks, validate | 4 |
| `mena` | skill descriptions, progressive disclosure, index quality, schemas | 4 |
| `satellite:{path}` | inscription, sync status, rite alignment, skill coverage | 4 |

User overrides via `--domains=dromena,legomena,agents`.

### Rite Ownership

| Component | Owner | Rationale |
|-----------|-------|-----------|
| Dromena (`/state-of`) | **forge** | Forge owns rite validation (`/validate-rite`, `/eval-agent`). Audit extends validation from component to holistic level |
| Legomena (`state-of-ref/`) | **shared** | Cross-cutting knowledge, available to any rite running an audit |
| Agent (`domain-auditor`) | **shared** | Generic evaluator, not forge-specific |

### Token Budget Analysis (7-domain audit)

| Phase | Input Tokens | Output Tokens |
|-------|-------------|--------------|
| Dromena invocation | ~200 | ~100 |
| Legomena load (INDEX) | ~300 | 0 |
| Per-domain criteria (x7) | ~700 | 0 |
| Agent prompt (x7) | ~2,800 | 0 |
| Per-domain evaluation (x7) | ~14,000 (codebase reads) | ~3,500 |
| Synthesis | ~3,900 (7 reports + prompt) | ~1,500 |
| **Total overhead** | **~21,900** | **~5,100** |

Sustainable. Each agent invocation stays under 5K overhead tokens. The synthesis agent receives ~3,500 tokens of reports.

### Mena File Structure

```
rites/forge/mena/
  state-of/INDEX.dro.md              # Dromena: thin dispatcher (~80 lines)

rites/shared/mena/
  state-of-ref/
    INDEX.lego.md                    # Domain registry, grading rubric overview
    domains/
      dromena.md                     # Evaluation criteria for slash commands
      legomena.md                    # Evaluation criteria for skills
      agents.md                      # Evaluation criteria for agent prompts
      inscription.md                 # Evaluation criteria for CLAUDE.md
      rules.md                       # Evaluation criteria for rules files
      hooks.md                       # Evaluation criteria for hook configs
      pipeline.md                    # Evaluation criteria for sync/materialize
    schemas/
      report-schema.md              # Per-domain report format
      synthesis-schema.md           # Consolidated "State of" format
      scorecard-schema.md           # Current vs. target scorecard
    templates/
      domain-report.md              # Template for individual reports
      synthesis-report.md           # Template for final document
```

**Total framework footprint**: ~1,100 lines across ~15 files. Each file small, focused, progressively disclosed.

### Doctrine Alignment

| Doctrine Principle | Alignment |
|-------------------|-----------|
| **Mortal Limits** | Each domain agent gets focused context, not everything. Progressive disclosure prevents token bloat |
| **Athena's Wisdom** | Knowing what to bring: each agent loads only its domain criteria. What to leave: other domains stay out of context |
| **The Clew Is Sacred** | Each agent dispatch and synthesis step can be recorded as events |
| **Rites Over Teams** | The audit uses shared primitives, not a dedicated rite. Invoke what you need |
| **Honest Signals** | Grading rubric produces health scores, not self-reported confidence |
| **Mena Lifecycle** | TRANSIENT dromena, PERSISTENT legomena, TRANSIENT agents — matches exactly |

---

## Recommendation

**Build as Option E: Dromena + Legomena + Agent composite.**

- **Dromena**: `/state-of <target>` in `rites/forge/mena/state-of/INDEX.dro.md`
- **Legomena**: `state-of-ref/` in `rites/shared/mena/state-of-ref/INDEX.lego.md`
- **Agent**: Generic `domain-auditor` in `rites/shared/agents/domain-auditor.md`

This is a novel primitive composition for Knossos — the first to use `workflow_type: parallel` semantics at the dromena level (all existing rites are sequential). It establishes a reusable pattern for any future N-agent parallel operations.

---

## Open Questions for Phase 2 Interview

1. **Scope**: Framework-level only, or should `/state-of` work on any target (rite, pipeline, satellite, arbitrary directory)?
2. **Parameterization**: How much should the user specify vs. auto-detect? Is the domain registry enough, or should there be a discovery agent?
3. **Output**: `.wip/STATE-OF-{X}-{date}.md` vs. `docs/audits/` vs. both?
4. **Integration**: Should findings auto-generate sprint backlogs? Connect to `/task` or `/sprint`?
5. **Agent model**: Single generic `domain-auditor` agent with per-domain criteria injection, or specialized per-domain agents (more tokens, better quality)?
6. **Synthesis approach**: Main thread synthesizes, or dedicated synthesis agent? Main thread saves a Task call but requires all reports in main context.
7. **Grading scale**: Letter grades (A-F)? Numeric (1-10)? Traffic light (green/yellow/red)? Must be compatible with scorecard format.
8. **Cross-domain patterns**: Should synthesis explicitly look for patterns across domains, or just aggregate?

---

## Follow-Up Actions

1. **Phase 2**: Interview on open questions above
2. **Phase 3**: Build the primitive composition
   - Dromena in forge mena
   - Legomena in shared mena (domain registry + evaluation criteria)
   - Agent in shared agents
   - Test with 2-3 domain subset before full 7-domain audit
3. **Phase 4**: Validate by running `/state-of framework` and comparing output to manual "State of the Framework" audit
4. **Future**: Consider whether this pattern warrants a forge template for generating domain-specific audit configurations

# Sprint Decisions: The Front Door

**Date**: 2026-02-10
**Sprint**: Deep overhaul of cold-start UX, semantic identity, and protocol hygiene
**Source**: 6-round structured interview informed by 5 context-engineer audits
**Status**: Decisions Complete. Ready for orchestrator re-plan.

---

## Governing Principles

### P1: Three-Tier Model (ACCEPTED)

All Knossos terminology is governed by three tiers with different change-impact rules:

| Tier | Name | Audience | Change Impact |
|------|------|----------|---------------|
| **Tier 1: Operational** | Protocol terms in prompts/inscriptions | Claude (via prompts) | Breaking change. Requires migration. |
| **Tier 2: CLI/Infrastructure** | Go package names, CLI commands | Go runtime / user CLI | Code change. Standard software refactor. |
| **Tier 3: Doctrinal** | Mythology in the Coda, glossary, concordance | Human developers | Documentation change. No behavioral impact. |

### P2: Naming as Reinforcement (ACCEPTED)

Naming is a **reinforcement layer** — it shapes behavior and creates coherence but is not the primary constraint mechanism. Tool restrictions and structured output do the heavy lifting; naming makes the constraints feel natural.

### P3: Progressive Disclosure with /consult On-Ramp (ACCEPTED)

Users encounter plain English first. Mythology surfaces gradually via `/consult` as the designated mythology on-ramp. The entry experience is mythology-free; the depth experience is mythology-rich.

### P4: Comprehensive Inscription over Token Budget (ACCEPTED)

The inscription (CLAUDE.md) should be **comprehensive and correct** rather than minimized for tokens. Quality of always-on context matters more than token savings. The redesign is holistic, not compressive.

---

## Decision 1: Operationalize Pythia

**Choice**: Rename `orchestrator.md` to `pythia.md` across all rites.

**Rationale**: Claude will access artifacts to get better context about the framework's specific orchestration principles, minimizing inference. The sacrifice of innate understanding of the no-op pattern is offset by proper context engineering and contractual enforcement in the Pythia prompt.

**Scope**: Every rite that has an `orchestrator.md` (10+ rites). Manifests, agent prompts, cross-references, inscription templates.

**Execution**: Combined pass — rename + Exousia injection + contextual read of each prompt to distinguish "orchestrator" as a name reference vs behavioral description. NOT a mechanical find-and-replace.

**Tier**: Tier 1 (Operational). This is a protocol change.

---

## Decision 2: Moirai Stays One Agent; Fates Are Progressive Disclosure Primitives

**Choice**: The Moirai agent is NOT decomposed. Clotho, Lachesis, and Atropos become dromena and legomena that the Moirai agent loads for context.

**Architecture**:

```
User invokes /start
  -> Moirai agent (the unified executor)
    -> Moirai loads Clotho (progressive disclosure of session creation SOPs)
      -> Moirai executes ari commands informed by Clotho's context
```

**Fate Mapping**:

| Fate | Primitive Type | Function |
|------|---------------|----------|
| **Clotho** (The Spinner) | Dromenon | Session creation procedures and SOPs |
| **Lachesis** (The Measurer) | Legomenon | Session state query and tracking reference |
| **Atropos** (The Cutter) | Dromenon | Session end/wrap procedures and SOPs |

**User-Facing Commands**: Stay plain English. `/start`, `/park`, `/continue`, `/wrap` remain unchanged. The Fate names live in the mena layer (source files) — Tier 1 protocol names, not Tier 0 UX names.

**Relationship**: The Fates are what Moirai reads to know how to act. They are reference knowledge and behavioral specs loaded via progressive disclosure, not wrappers or replacements for session commands.

**Tier**: Fates are Tier 1 (Operational, in Moirai's context). Session command names are Tier 0 (user-facing UX, never changed).

---

## Decision 3: Exousia (Authority Contract)

**Choice**: Name the "You decide / You escalate / You do NOT decide" pattern **Exousia** (Greek: authority/power/jurisdiction).

**Format**: Strict template. Every agent MUST have:

```markdown
## Exousia

### You Decide
- {list of decisions within this agent's authority}

### You Escalate
- {list of situations requiring escalation to user or Pythia}

### You Do NOT Decide
- {list of decisions explicitly outside this agent's scope}
```

**No exceptions.** This becomes a hard requirement auditable by `/theoria` against the agents domain criteria.

**Tier**: Tier 1 (Operational). Appears in every agent prompt as a section header.

---

## Decision 4: /go Location and Behavior

**Location**: `rites/shared/mena/go.dro.md` — shared across all rites, always materialized.

**Pre-Rite Behavior**: When no rite is active, `/go` dispatches to `/consult` for routing guidance. `/consult` is the universal pre-rite entry point. Pythia is rite-scoped; `/consult` is universal.

**Autopark Trigger**: Always park if session is active on exit. Simple rule. Near-zero cost. `/go` handles stale sessions on resume.

**Scenarios (Phase 1)**:
1. ALREADY_ACTIVE -> Show status + next step (~3s)
2. RESUME_PARKED -> Auto-resume (~8s)
3. NEW_WORK -> /consult for routing, create session (~12s)
4. ORIENTATION -> Dashboard + options (~5s)

Scenarios 5-6 (RESUME_ORPHANED, CROSS_WORKTREE) deferred to Phase 2.

---

## Decision 5: Inscription Holistic Redesign

**Choice**: Full redesign of inscription templates, not just compression.

**Approach**: Context-engineer designs the new structure; integration-engineer implements in templates. Orchestrator (Pythia) coordinates.

**Principles**:
- Comprehensive coverage over token budget concerns
- Must reflect Pythia (not orchestrator) as the coordination identity
- Must reference Exousia as the authority contract standard
- Must be correct about Fates-as-primitives architecture
- Progressive disclosure: inscription teaches the current state, skills provide depth

**Timing**: Designed AFTER Pythia + Exousia work is complete (inscription is the SUMMARY of the new protocol).

---

## Decision 6: Theoria as Validation

**Choice**: Run `/theoria agents` as baseline BEFORE the sprint and as validation AFTER. Use as signal, not gate (since theoria criteria may need updating for Exousia).

**Bonus**: Update agents domain criteria to include Exousia compliance BEFORE the post-sprint run. This makes the audit meaningful.

**Dogfooding**: The orchestrator (soon Pythia) coordinates the sprint, which includes its own rename. Natural dogfood of the coordination layer.

---

## Decision 7: Three-Tier Document Location

**Choice**: Doctrine addition. New section in `knossos-doctrine.md` (likely Section II.5 or integrated into the existing naming section).

---

## Decision 8: Sprint Session Cadence

**Choice**: Re-invoke orchestrator with all decisions to determine optimal cadence. The expanded scope (holistic inscription, Pythia+Exousia combined pass, Fates-as-primitives) requires fresh planning.

---

## Dependency Graph (Updated)

```
Decision Dependencies (all resolved above):
  D1 (Pythia) + D3 (Exousia) -> Combined agent pass
  D2 (Fates) -> Moirai primitive creation
  D4 (/go) -> Depends on /consult being stable, autopark existing
  D5 (Inscription) -> Depends on D1, D2, D3 being implemented
  D6 (Theoria) -> Baseline before Wave 1, validation after final wave
  D7 (Tier doc) -> After all implementations, captures the final state

Implementation Order (must be serial where noted):
  1. Theoria baseline (can run immediately)
  2. Fates primitives (independent of Pythia work)
  3. Pythia rename + Exousia injection (the big combined pass)
  4. Autopark hook
  5. /go dromena
  6. Inscription redesign (depends on 2, 3)
  7. Three-tier doctrine addition (depends on all)
  8. Theoria validation + agents criteria update
```

---

## Open Items for Orchestrator

- Exact session count and wave structure given expanded scope
- Whether Fates primitives and Pythia+Exousia can be parallelized across sessions
- Review gate placement
- Model tier assignments per work item
- Context flow: which items share context beneficially

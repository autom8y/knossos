# Design Principles

> Revelations encoded as architecture. Principles that give Knossos its shape.

These eight principles (from Section XI of the Knossos Doctrine) are the architectural DNA of the platform—not rules to follow but truths to embody.

---

## 1. The Clew Is Sacred

**Statement:** Every action must knot the clew. If an event goes unrecorded, the path becomes uncertain. When the path becomes uncertain, return becomes unlikely.

**Implementation:**
- Event recording via `/roster/internal/hook/clewcontract/` (Go package defining clew contract types)
- Session event persistence through `ari session` commands (`/roster/cmd/ari/`)
- Hook-based event capture (Go implementation: `/roster/internal/hook/`, rite-specific scripts: `/roster/rites/[rite]/hooks/`, materialized: `.claude/hooks/`)
- Events append to `events.jsonl` in session directories

**Application Guidance:**
- All state mutations flow through Moirai (see Principle 3)
- Tool calls, decisions, artifacts trigger event recording
- Session lifecycle transitions are auditable through event log
- Recovery and debugging rely on complete event provenance

---

## 2. Honest Signals Over Comfortable Lies

**Statement:** White Sails exist because the easy answer is often wrong. Gray is not failure—gray is honesty about uncertainty. Ship gray with eyes open rather than white with false confidence.

**Implementation:**
- White Sails confidence computation (`/roster/internal/sails/`)
- Three-color system (WHITE/GRAY/BLACK) with anti-gaming rules
- Quality gate execution at session wrap
- QA upgrade mechanism (independent review can elevate GRAY → WHITE)

**Application Guidance:**
- Spikes and hotfixes default to GRAY (inherent uncertainty)
- Open questions trigger gray ceiling automatically
- Humans can downgrade confidence, never self-upgrade
- Missing proofs prevent WHITE status
- Confidence signals guide deployment decisions

---

## 3. Mutation Through the Fates

**Statement:** Only the Moirai may modify session state. Clotho creates. Lachesis tracks. Atropos terminates. When mutations flow through divine authority, validation is guaranteed.

**Implementation:**
- Moirai agent definition (`/roster/user-agents/moirai.md`)
- Session state management (`/roster/internal/session/`)
- Write guard hooks prevent direct `SESSION_CONTEXT.md` modification
- State mutations validated against lifecycle schemas

**Application Guidance:**
- Never directly edit `SESSION_CONTEXT.md` or `SPRINT_CONTEXT.md`
- Use `ari session` commands or invoke Moirai agent
- Clotho: session initialization (`session_start` events)
- Lachesis: progress tracking, phase transitions, budget monitoring
- Atropos: session termination, archival, White Sails generation
- Pre-tool-use hooks intercept invalid mutations

---

## 4. Rites Over Teams

**Statement:** Teams imply fixed membership. Rites are flexible practices—invoke what you need, swap when you must. A rite with only skills is not incomplete; it is a simple rite.

**Implementation:**
- Rite source definitions (`/roster/rites/`)
- Manifest-based rite composition (`manifest.yaml` per rite)
- Materialization system (`/roster/internal/materialize/`, invoked via `ari sync materialize`)
- Rite loading and context injection (`/roster/internal/rite/`)

**Application Guidance:**
- Rites can contain: agents, skills, hooks, workflows (any combination)
- Simple rites (skills-only) are valid and useful
- Invoke rites mid-journey without full context switch
- Swap rites for major practice changes
- Current rite tracked in `ACTIVE_RITE` file
- Cost model: `invoke-rite` (cheap) vs `swap-rite` (expensive)

---

## 5. Heroes Are Mortal

**Statement:** Context is finite. Heroes tire, forget, can only carry so much. Design for summoning with rich context, not for heroes who know everything.

**Implementation:**
- Cognitive budget tracking (configurable via `ARIADNE_MSG_WARN`, `ARIADNE_MSG_PARK`)
- Auto-park suggestions via hooks when budget depletes
- Task tool delegation with context curation
- Session context (`SESSION_CONTEXT.md`) as compact summoning context

**Application Guidance:**
- Summon specialist agents (heroes) with session context
- Don't load every skill/rite simultaneously
- Park when cognitive budget warns (not failure, wisdom)
- Better clew → better summoning → better heroes
- Context window constraints are design inputs, not bugs

---

## 6. The Labyrinth Grows

**Statement:** Complexity is not static. The labyrinth extends as you explore it. New passages open. Old paths shift. The clew must accommodate growth without breaking.

**Implementation:**
- Extensible rite system (add rites to `/roster/rites/`)
- Dynamic skill loading
- Hook registration system (new hooks integrate without platform changes)
- Manifest-driven materialization (rites define their own structure)

**Application Guidance:**
- Rites are additive—new practices don't break existing ones
- Skills can be added/removed from rites via manifest
- Platform grows through rite composition, not core modification
- Backward compatibility via schema versioning
- The labyrinth evolves; the clew mechanism remains constant

---

## 7. Return Is the Victory

**Statement:** Slaying the Minotaur matters less than returning to Athens. A merged PR with honest confidence is more valuable than heroic effort lost to context collapse.

**Implementation:**
- Session wrap workflow (`ari session wrap`)
- Quality gates at termination
- White Sails confidence required before close
- Session archival preserves journey record
- PR creation includes test plan and confidence signal

**Application Guidance:**
- Incomplete session with honest GRAY signal > abandoned session
- Parking preserves state for resumption
- Wrap workflow validates completion criteria
- Merged PR (Athens) is the goal, not in-progress heroics
- Context collapse is the enemy; clew is the defense

---

## 8. The Inscription Prepares

**Statement:** CLAUDE.md is not documentation—it is the labyrinth speaking. Keep the Inscription current, and travelers arrive prepared.

**Implementation:**
- Inscription templates (`/roster/knossos/templates/CLAUDE.md.tpl`)
- Dynamic generation via `ari sync inscription`
- SessionStart hook reads Inscription (`.claude/CLAUDE.md`)
- Knossos-managed sections (delimited by `<!-- KNOSSOS:START -->` ... `<!-- KNOSSOS:END -->`)

**Application Guidance:**
- Never manually edit Knossos-managed sections
- Use `ari sync inscription` to regenerate from source
- Inscription declares: available rites, agents, execution mode, hooks
- Custom sections (outside Knossos delimiters) are preserved
- The labyrinth speaks to Theseus at entry—prepare the entrance

---

## Principle Relationships

These principles interlock:

```
Principle 1 (Clew Sacred) ──enables──▶ Principle 7 (Return)
                  ▲
                  │
Principle 3 (Mutation via Fates) ──validates──▶ Principle 1
                  │
                  ▼
Principle 2 (Honest Signals) ──informs──▶ Principle 7

Principle 5 (Heroes Mortal) ──constrains──▶ Principle 4 (Rites > Teams)
                  │
                  ▼
Principle 6 (Labyrinth Grows) ──requires──▶ Principle 4

Principle 8 (Inscription) ──prepares──▶ Principle 5
```

The principles form a system, not a checklist. Violating one weakens the others.

---

## Anti-Patterns (Principle Violations)

| Anti-Pattern | Violated Principle | Consequence |
|--------------|-------------------|-------------|
| Direct `SESSION_CONTEXT.md` edits | 3 (Mutation via Fates) | Validation bypassed, inconsistent state |
| Ignoring GRAY signals, forcing WHITE | 2 (Honest Signals) | Aegeus problem—false confidence |
| Loading all rites simultaneously | 5 (Heroes Mortal) | Context overflow, diluted summoning |
| Abandoning sessions instead of wrapping | 7 (Return) | Orphaned sessions on Naxos |
| Manual Inscription edits in Knossos sections | 8 (Inscription) | Materialization conflicts |
| Unrecorded decisions or actions | 1 (Clew Sacred) | Lost provenance, no audit trail |

---

## Evolution

These principles emerged from production experience navigating complex codebases. They are **descriptive** (what works) as much as **prescriptive** (what to do).

As the platform evolves, new principles may emerge. The current eight represent foundational truths discovered through practice.

---

**See Also:**
- `philosophy/knossos-doctrine.md` (Section XI: source of these principles)
- `philosophy/mythology-concordance.md` (mapping myth to implementation)
- `reference/INDEX.md` (navigation hub)

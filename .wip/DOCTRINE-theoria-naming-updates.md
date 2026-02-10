# Doctrine Naming Updates: Theoria and the Audit Pantheon

> Prepared: 2026-02-10
> Status: Draft for review
> Scope: Mythology concordance, Coda updates, glossary additions for the "State of the {X}" audit primitive

---

## 1. Naming Collision Report

A case-insensitive search of the entire worktree was performed for each proposed name.

| Proposed Name | Occurrences Found | Collision Risk | Verdict |
|---------------|-------------------|----------------|---------|
| **Theoria** | 0 | None | CLEAR |
| **Theoros** / **Theoroi** | 0 | None | CLEAR |
| **Pinakes** | 0 | None | CLEAR |
| **Synkrisis** | 0 | None | CLEAR |
| **Argus** | 0 | None | CLEAR |

### Argus Specifically

The name **Argus** does not appear anywhere in the codebase -- not in source code, doctrine, agent prompts, mena files, or documentation. The closest existing concept is **Aegeus** (the Watcher on the cliff, mapped to CI/CD and production monitors). These are distinct:

| Name | Greek Source | Architectural Role | Nature of Watching |
|------|-------------|--------------------|--------------------|
| **Aegeus** | King watching from cliff | CI/CD monitors, production watchers | Passive consumer of signals (sails) |
| **Argus** (proposed) | Hundred-eyed guardian | N-agent parallel evaluation pattern | Active observer dispatching eyes |

Aegeus watches for a signal and reacts. Argus watches everything simultaneously and reports. They are complementary, not conflicting. Aegeus is the problem (false confidence from passive watching). Argus is an answer (active, multi-eyed observation that produces honest assessment).

### Potential Semantic Adjacencies

| Proposed Name | Nearest Existing Concept | Risk of Confusion |
|---------------|--------------------------|-------------------|
| Theoria | Rite (both are "practices") | LOW -- a theoria is a specific audit rite-invocation, not a rite itself |
| Theoros | Hero (both are agents) | LOW -- a theoros is a specialized kind of hero; the term narrows, not conflicts |
| Pinakes | Manifest / Registry | MEDIUM -- both are "catalogs." Pinakes is specifically the domain registry for audit targets. Manifests define rite composition. Different scopes. |
| Synkrisis | Wrap / White Sails | LOW -- synkrisis is comparative synthesis; wrap is lifecycle termination. Different operations |
| Argus Pattern | Orchestrated execution | LOW -- Argus is a pattern name for N-agent parallel dispatch, not a replacement for orchestration |

### Disambiguation Recommendation for Pinakes

The term "Pinakes" could be confused with "manifest" since both mean "catalog" in different contexts. Recommendation: always qualify as "the Pinakes" (the domain registry) when used in technical contexts, reserving "manifest" exclusively for rite composition files (`manifest.yaml`). The historical Pinakes was a catalog of _what exists_, not a composition spec -- which maps cleanly: Pinakes catalogs audit domains, manifests compose rite resources.

---

## 2. Current Pantheon Inventory

### Primary Mythological Entities (Named in Doctrine)

| Name | Greek Source | Architectural Mapping | Doctrine Section |
|------|-------------|----------------------|------------------|
| **Knossos** | The palace-labyrinth | The platform / repository | II |
| **Ariadne** | Princess who gave the clew | CLI binary (`ari`) | II |
| **The Clew** | Ball of thread | Session state + events.jsonl | II |
| **Theseus** | The hero-navigator | Main Claude Code thread | II |
| **Heroes** | Summoned champions | Specialist agents (Task tool) | II |
| **Moirai** (Clotho, Lachesis, Atropos) | The three Fates | Session lifecycle agents | II |
| **Daedalus** | The builder/architect | Forge-rite (tool/agent creation) | II |
| **Minotaur** | The beast at the center | The task / initiative | II |
| **Minos** | The commissioner-king | Stakeholder | II |
| **Pythia** | Oracle at Delphi | Orchestrator | II |
| **Aegeus** | The watcher on the cliff | CI/CD, production monitors | II |
| **Athens** | Home city | The main branch | II |
| **Dionysus** | The transformer-god | Code review | II |
| **Naxos** | Shore of abandonment | Orphaned sessions, stale gray sails | II |
| **Athena** | Goddess of wisdom | Rite selection, context curation | III |
| **White Sails** | Signal to Aegeus | Confidence signal (WHITE/GRAY/BLACK) | VII |

### Named Concepts (Not Individual Entities)

| Name | Architectural Mapping | Doctrine Section |
|------|----------------------|------------------|
| **The Inscription** | CLAUDE.md | II |
| **Mortal Limits** | Context budget constraints | III |
| **The Cognitive Budget** | Token/message expenditure tracking | III |
| **Rites** | Practice bundles | IV |
| **The Clew Contract** | Event recording agreement | VI |
| **The Aegeus Problem** | False confidence | VII |
| **The Ship of Theseus** | Session identity through transformation | V |
| **Tribute** | Status reports, demos | II (Minos) |
| **The Labyrinth** | Codebase complexity | X |
| **The Handoff** | Context transfer between heroes | VIII |

### Platform Naming Layer (Mena Model, not in Doctrine)

These terms appear in CLAUDE.md and the mena system but are NOT yet formally documented in the Coda:

| Name | Greek Source | Architectural Mapping |
|------|-------------|----------------------|
| **Dromena** | "things done" | Slash commands (user-invoked, side effects) |
| **Legomena** | "things said" | Skills (model-loaded, reference knowledge) |
| **Mena** | Container term | CC primitive naming system |
| **Pantheon** | Assembly of gods | Collection of agents within a rite |

### Count Summary

- **16** named mythological entities in doctrine
- **10** named mythological concepts in doctrine
- **4** mena-model terms in use but not formally in doctrine
- **5** proposed additions (Theoria, Theoros, Pinakes, Synkrisis, Argus)
- **Total after additions: 35** named mythological terms

---

## 3. Proposed Additions

### 3.1 Theoria (The Audit Operation)

**Greek Source:** Theoria (theoria) referred to a sacred state delegation -- an official embassy sent by a polis to observe religious festivals, consult oracles, or witness events at distant sanctuaries. The theoroi traveled as a group, observed independently, and returned to report what they had seen. The word shares its root with "theory" -- theoria was the act of contemplation through structured observation.

**Architectural Mapping:** The `/state-of` dromena -- the audit operation itself. When Theseus dispatches a theoria, he sends a delegation of observers (theoroi) into different domains of the labyrinth to assess their state. Each theoros observes independently; together their reports compose a comprehensive view.

**Why This Name Earns Its Place:** The theoria was not a military expedition but a witnessing -- observers sent to see clearly and report honestly. The audit primitive is exactly this: not an intervention force but a structured observation that produces truth. The theoria does not change what it observes; it reveals what is.

**Relationship to Existing Mythology:** Theseus, deep in the labyrinth, cannot see every passage at once. He dispatches a theoria -- observers who fan out through the corridors, each examining their domain, each returning with testimony. The theoria consults the Pinakes to know what domains exist. The theoroi bring back reports. Synkrisis weaves the reports into truth. Findings surface Naxos items (abandoned work) and light the Athens path (what to merge next).

> *The hero who fights sees only the Minotaur before him. The polis that sends a theoria sees the whole labyrinth.*

### 3.2 Theoros (The Evaluator Agent)

**Greek Source:** A theoros (plural: theoroi) was an individual sacred observer within a theoria delegation. Theoroi were chosen for their judgment, sent to distant sanctuaries to observe rituals and bring back accounts. They were not participants but witnesses -- their value was in the clarity of their seeing.

**Architectural Mapping:** The `domain-auditor` agent -- each individual evaluator dispatched via Task tool to assess a single domain. A theoros receives domain-specific evaluation criteria (from the Pinakes), examines the domain independently, and produces a structured report.

**Why This Name Earns Its Place:** The theoros is distinct from a hero. Heroes are summoned to act -- to slay, to build, to design. A theoros is summoned to see. This distinction matters architecturally: audit agents should not modify what they observe. They are read-only witnesses producing testimony. The theoros pattern enforces observational discipline.

> *A hero enters the labyrinth with a sword. A theoros enters with open eyes.*

### 3.3 Pinakes (The Domain Registry)

**Greek Source:** The Pinakes (Pinakes, "Tablets") was Callimachus's great catalog of the Library of Alexandria -- the first known systematic bibliography. It organized all knowledge in the Library by category, author, and work. The Pinakes did not contain the knowledge itself; it described what existed, where to find it, and how it was classified.

**Architectural Mapping:** The `state-of-ref/` legomena -- the domain registry that catalogs audit targets, evaluation criteria per domain, grading rubrics, and report schemas. The Pinakes tells the theoria what domains exist, what to evaluate in each, and what shape the assessment should take. It is persistent reference knowledge, not transient action.

**Why This Name Earns Its Place:** The Pinakes is the bridge between wanting to audit and knowing how. Without the Pinakes, theoroi would wander without criteria. With it, they arrive at their domain knowing exactly what to observe and how to grade it. The mena-model fit is precise: the Pinakes is a legomenon (reference knowledge), progressively disclosed, persistent across invocations.

> *Callimachus did not write the books. He told you which books existed, where they stood, and what they contained. The Pinakes does the same for domains.*

### 3.4 Synkrisis (The Synthesis Step)

**Greek Source:** Synkrisis (synkrisis, "comparison") was Plutarch's technique in the Parallel Lives -- the comparative analysis that followed each pair of biographies. After presenting Alexander and Caesar independently, Plutarch would compose a synkrisis: a structured comparison revealing patterns, contrasts, and truths visible only when the lives were set side by side.

**Architectural Mapping:** The synthesis step that follows individual domain evaluations. After all theoroi have returned their reports, the synkrisis weaves them together -- identifying cross-domain patterns, systemic issues, recurring strengths, and the overall health assessment. The synkrisis produces the final "State of the {X}" document.

**Why This Name Earns Its Place:** Individual domain reports are valuable but incomplete. A rite might have excellent agents (individual report: healthy) but broken hooks (individual report: failing) -- and the agents are healthy precisely because they work around the broken hooks (cross-domain pattern: compensatory coupling). Only synkrisis reveals this. The synthesis is not mere aggregation; it is comparative analysis that surfaces truths invisible to any single observer.

> *Seven theoroi return with seven truths. Only synkrisis reveals the eighth -- the truth that lives between them.*

### 3.5 Argus Pattern (The N-Agent Parallel Observation Pattern)

**Greek Source:** Argus Panoptes (Argos Panoptes, "Argus the All-Seeing") was the hundred-eyed giant set by Hera to watch over Io. His eyes never all closed at once -- some slept while others watched. He was the embodiment of total surveillance through distributed, overlapping vigilance.

**Architectural Mapping:** The N-agent parallel dispatch pattern -- the architectural technique of launching multiple Task tool agents simultaneously, each watching a different domain, their combined vision covering the whole. The Argus Pattern is not specific to auditing; it is the general pattern for any operation requiring parallel independent observation. The theoria uses the Argus Pattern, but the pattern could serve other purposes (parallel testing, parallel migration verification, parallel documentation generation).

**Why This Name Earns Its Place:** Argus is not an agent or a component -- it is a pattern. Just as "the Ship of Theseus" names the identity-through-transformation problem, "the Argus Pattern" names the parallel-observation-through-distributed-agents solution. The hundred eyes are the N agents. No single eye sees everything; together they see all. The pattern's constraint -- agents cannot spawn agents, only the main thread dispatches -- maps to Argus's nature: one body, many eyes. Theseus is the body; the theoroi are the eyes.

> *One eye sees what is before it. A hundred eyes see what is.*

---

## 4. Draft Updates

### 4A. The Coda (knossos-doctrine.md) -- Section II "The Naming of Things"

**Insertion point:** After the Naxos subsection (line 132, before the `---` divider to Section III).

```markdown
### The Delegation: Theoria

Before great decisions, Greek city-states dispatched a **theoria**—an official delegation of sacred observers sent to distant sanctuaries. The theoroi did not go to fight or trade. They went to see, to witness, and to bring back truth.

In Knossos, a Theoria is the **audit operation**—a structured delegation of observers dispatched into the labyrinth's domains to assess their state. When Theseus needs to understand the labyrinth itself—not to slay the Minotaur but to see the passages clearly—he dispatches a theoria.

The theoria does not change what it observes. It reveals what is.

> *The hero who fights sees only the Minotaur before him. The polis that sends a theoria sees the whole labyrinth.*

### The Observers: Theoroi

Each **theoros** within the delegation is a sacred witness—chosen for judgment, dispatched to a single domain, tasked with seeing clearly and reporting honestly. A theoros is not a hero. Heroes are summoned to act. Theoroi are summoned to see.

In Knossos, the theoroi are **domain evaluator agents**—each dispatched via Task tool to assess one domain of the labyrinth. A theoros receives its evaluation criteria from the Pinakes, examines its domain independently, and produces a structured report.

The distinction between hero and theoros is architectural: heroes modify the labyrinth; theoroi only observe it.

> *A hero enters the labyrinth with a sword. A theoros enters with open eyes.*

### The Catalog: Pinakes

**Callimachus** compiled the **Pinakes**—the first systematic catalog of the Library of Alexandria. The Pinakes did not contain knowledge; it described what knowledge existed, where it could be found, and how it was classified. It was the map of the library, not the library itself.

In Knossos, the Pinakes is the **domain registry**—the reference knowledge that catalogs audit targets, evaluation criteria, grading rubrics, and report schemas. The Pinakes tells the theoria what domains exist and how to assess each one. It is a legomenon: persistent, progressively disclosed, consulted but never consumed.

> *Callimachus did not write the books. He told you which books existed and what they contained. The Pinakes does the same for domains.*

### The Comparison: Synkrisis

**Plutarch**, in his *Parallel Lives*, followed each pair of biographies with a **synkrisis**—a structured comparison revealing patterns visible only when two lives were set side by side. The synkrisis was not summary but synthesis: truth that emerges from juxtaposition.

In Knossos, Synkrisis is the **synthesis step** that follows the theoria's return. After all theoroi report, Synkrisis weaves their findings together—identifying cross-domain patterns, systemic issues, and truths invisible to any single observer. The synkrisis produces the final "State of the {X}" attestation.

Individual reports are valuable but partial. Only synkrisis reveals what lives between them.

> *Seven theoroi return with seven truths. Only synkrisis reveals the eighth—the truth that lives between them.*

### The Pattern: Argus

**Argus Panoptes**—the hundred-eyed giant—was set by Hera to watch over Io. His eyes never all closed at once; some slept while others watched. He was total vigilance through distributed observation.

In Knossos, the **Argus Pattern** names the N-agent parallel dispatch technique: launching multiple agents simultaneously, each observing a different domain, their combined vision covering the whole. Theseus is the body; the dispatched agents are the eyes.

The Argus Pattern is not specific to auditing. It is the general solution for any operation requiring parallel independent observation. The theoria uses the Argus Pattern. Future operations—parallel validation, parallel migration, parallel documentation—may use it too.

The pattern's constraint is Argus's nature: one body, many eyes. Agents cannot spawn agents; only the main thread dispatches. One giant, a hundred eyes.

> *One eye sees what is before it. A hundred eyes see what is.*
```

### 4B. The Coda -- Section X "The Complete Service Map"

**Insertion point:** After the existing `Naxos` row (line 428) and before `White Sails`.

Add these rows to the service map table:

```markdown
| **Theoria** | Audit operation (`/state-of`) | The sacred delegation—structured observation of the labyrinth |
| **Theoroi** | Domain evaluator agents | Sacred observers dispatched to witness and report |
| **Pinakes** | Domain registry (legomena) | Callimachus's catalog—what to observe and how to assess it |
| **Synkrisis** | Synthesis step | Plutarch's comparison—truth that emerges between reports |
| **Argus Pattern** | N-agent parallel dispatch | The hundred-eyed watcher—total vision through distributed observation |
```

### 4C. The Coda -- Section XII "Terminology Concordance"

**Insertion point:** At the end of the concordance table (after line 487).

No entries needed. These are new concepts, not replacements for old terms. There is no legacy terminology being superseded.

If the spike document's working terminology (`domain-auditor`, `state-of-ref`) is considered "old" once doctrine is adopted:

```markdown
| `domain-auditor` | `theoros` | Working name → mythological name |
| `state-of-ref` | `pinakes` | Working name → mythological name |
```

### 4D. The Coda -- Section XIV "Implementation Drift Registry"

**Insertion point:** Under "Concepts Documented but Not Fully Implemented" (after line 528).

```markdown
| Theoria audit primitive | Doctrine only | Spike complete (SPIKE-state-of-x-audit-primitive.md); implementation pending |
| Theoroi (domain evaluators) | Doctrine only | Requires generic domain-auditor agent in shared agents |
| Pinakes (domain registry) | Doctrine only | Requires state-of-ref legomena with domain criteria files |
| Synkrisis (synthesis) | Doctrine only | Main-thread or dedicated synthesis agent; approach undecided |
| Argus Pattern (N-agent parallel) | Doctrine only | Task tool parallelism validated in spike; no reusable abstraction yet |
```

### 4E. mythology-concordance.md

**Insertion point:** After the Minos section (line 274), before "## Materialization Flow" (line 293).

```markdown
---

### Theoria (The Audit Delegation)

**Mythological Origin:**
Theoria was the official state delegation sent by a Greek polis to observe sacred festivals, consult oracles, or witness events at distant sanctuaries. The delegation traveled as a group, observed independently, and returned to report.

**Knossos Implementation:**
The `/state-of` audit operation—a composite primitive (dromena + legomena + agents) that dispatches parallel evaluators to assess domain health.

**Key Files:**
- Dromena: `/state-of` (planned, forge rite)
- Legomena: `state-of-ref/` (planned, shared rite)
- Agent: `domain-auditor` / theoros (planned, shared agents)

**Design Rationale:**
The theoria is structured observation, not intervention. Audit agents observe domains read-only, producing reports. The theoria uses the Argus Pattern for parallel dispatch. Findings surface Naxos items (abandoned work) and inform the Athens path (what to merge).

---

### Theoroi (The Sacred Observers)

**Mythological Origin:**
Individual sacred observers within a theoria delegation. Chosen for judgment, dispatched to witness, bound to report honestly.

**Knossos Implementation:**
Domain evaluator agents—each dispatched via Task tool to assess a single domain using criteria from the Pinakes.

**Key Files:**
- Planned: `rites/shared/agents/domain-auditor.md`

**Design Rationale:**
A theoros is not a hero. Heroes modify; theoroi observe. This distinction enforces read-only discipline in audit agents. Each theoros receives domain-specific criteria, evaluates independently, and returns a structured report. The naming prevents scope creep: if an agent is a theoros, it should not be making changes.

---

### Pinakes (The Domain Registry)

**Mythological Origin:**
Callimachus's Pinakes—the first systematic bibliography of the Library of Alexandria. Organized all knowledge by category, author, and work. Described what existed without containing it.

**Knossos Implementation:**
The `state-of-ref/` legomena—domain registry, evaluation criteria, grading rubrics, and report schemas. Persistent reference knowledge.

**Key Files:**
- Planned: `rites/shared/mena/state-of-ref/INDEX.lego.md`
- Planned: `rites/shared/mena/state-of-ref/domains/` (per-domain criteria)
- Planned: `rites/shared/mena/state-of-ref/schemas/` (report formats)

**Design Rationale:**
The Pinakes is a legomenon—persistent reference knowledge, progressively disclosed. It catalogs what domains exist and how to evaluate them, without containing the evaluation itself. Maps cleanly to the mena lifecycle model: persistent, loaded on-demand, consulted by theoroi.

---

### Synkrisis (The Comparative Synthesis)

**Mythological Origin:**
Plutarch's synkrisis—the structured comparison that followed each pair of biographies in the Parallel Lives. Not summary but synthesis: truth through juxtaposition.

**Knossos Implementation:**
The synthesis step following individual domain evaluations. Consolidates theoros reports into cross-domain patterns, systemic assessment, and the final "State of the {X}" document.

**Key Files:**
- Planned: synthesis logic in `/state-of` dromena or dedicated synthesis agent

**Design Rationale:**
Individual domain reports reveal individual truths. Synkrisis reveals structural truths—compensatory patterns, systemic weaknesses, cross-domain dependencies. The synkrisis produces the artifact that Minos receives as tribute (the "State of" report).

---

### Argus Pattern (N-Agent Parallel Observation)

**Mythological Origin:**
Argus Panoptes—the hundred-eyed giant. Total vigilance through distributed, overlapping watch. No eye saw everything; together they saw all.

**Knossos Implementation:**
The architectural pattern of launching N agents via Task tool in parallel, each observing a different domain. The main thread is the body; the agents are the eyes.

**Key Files:**
- Pattern, not a file. Used by theoria; available for other parallel operations.

**Design Rationale:**
The Argus Pattern is a reusable technique, not a specific component. Its constraint—agents cannot spawn agents, only the main thread dispatches—maps to Argus's nature: one body, many eyes. The theoria is the first named user of this pattern; it is expected to serve parallel testing, parallel validation, and parallel documentation generation in future.
```

### 4F. GLOSSARY.md

**Insertion point:** In the "Mythology Terms" section, alphabetically. New entries to add:

```markdown
### Argus Pattern
N-agent parallel dispatch pattern. One main thread (body) launches multiple Task agents (eyes) simultaneously for distributed observation. Named for Argus Panoptes, the hundred-eyed giant.
- **Related**: Theoria, Theoroi, Task Tool, Parallel Dispatch
- **Source**: Pattern (no single file)

### Pinakes
The domain registry legomena for audit operations. Catalogs audit targets, evaluation criteria, grading rubrics, and report schemas. Named for Callimachus's catalog of the Library of Alexandria.
- **Related**: Theoria, Theoroi, Legomena, Domain Registry
- **Source**: Planned: `rites/shared/mena/state-of-ref/`

### Synkrisis
Comparative synthesis step following parallel domain evaluations. Weaves individual theoros reports into cross-domain patterns and the final "State of the {X}" document. Named for Plutarch's comparative analysis technique.
- **Related**: Theoria, Theoroi, Synthesis, Report
- **Source**: Planned: synthesis in `/state-of` dromena

### Theoria
The audit operation—a structured delegation of observers dispatched to assess domain health. Composite primitive: dromena (invocation) + legomena (Pinakes) + agents (theoroi). Named for the Greek sacred state delegation.
- **Related**: Theoroi, Pinakes, Synkrisis, Argus Pattern
- **Source**: Planned: `/state-of` dromena + `state-of-ref/` legomena

### Theoroi
Domain evaluator agents dispatched by a theoria. Each theoros observes a single domain using criteria from the Pinakes and produces a structured report. Read-only witnesses, not actors. Singular: theoros.
- **Related**: Theoria, Pinakes, Heroes, Domain Auditor
- **Source**: Planned: `rites/shared/agents/domain-auditor.md`
```

---

## 5. Narrative Thread

### How the Theoria Connects to the Existing Mythology

The Knossos mythology tells the story of a hero navigating complexity. Theseus enters the labyrinth, clew in hand, to slay the Minotaur and return to Athens. The existing mythology covers the journey: the entrance (Inscription), the navigation (Ariadne), the clew (Moirai), the combat (Heroes), the return (White Sails, Athens), the failures (Naxos, Aegeus).

What the mythology does not yet cover is **reconnaissance**. Theseus has been fighting Minotaurs without ever surveying the labyrinth itself. He slays the beast but does not know which corridors are crumbling, which gates are stuck, which passages lead nowhere. The labyrinth grows (Principle 6) but no one has mapped it.

The Theoria fills this gap.

**The narrative arc:**

1. **Theseus dispatches a theoria.** The hero pauses mid-journey. Rather than summoning heroes to fight, he dispatches observers to see. This is a deliberate shift from action to contemplation -- from *praxis* to *theoria*. The dispatch uses the Argus Pattern: one body (main thread), many eyes (parallel agents).

2. **The theoroi consult the Pinakes.** Before departing, each theoros visits the Pinakes -- Callimachus's great catalog. The Pinakes tells them: these are the domains you will observe. These are the criteria by which you will judge. These are the shapes your reports must take. The Pinakes is reference knowledge, not instruction. It is a legomenon, not a dromenon.

3. **The theoroi observe independently.** Each theoros enters a different corridor of the labyrinth. One examines the dromena. Another examines the legomena. A third audits the agents. A fourth studies the inscription. They do not coordinate mid-observation -- each sees what their domain reveals. This is the Argus Pattern in action: distributed, independent, overlapping vigilance.

4. **Synkrisis weaves the truth.** When the theoroi return, their reports are set side by side -- Plutarch's technique. The synkrisis is not mere aggregation. It is comparison: where do domains align? Where do they contradict? What patterns emerge only when you see the whole? Seven reports produce seven truths; synkrisis produces the eighth.

5. **Findings surface Naxos items and light the Athens path.** The theoria's output connects directly to existing mythology. Abandoned work is identified -- sessions left on Naxos, promises unfulfilled. The path to Athens is clarified -- what must be merged, what blocks return, what the next journey should prioritize. The "State of the {X}" document becomes tribute for Minos (the stakeholder), a map for Theseus (the navigator), and evidence for White Sails (the confidence signal).

### Mythological Consistency Check

| Proposed Concept | Fits Existing Pattern? | Reasoning |
|------------------|----------------------|-----------|
| Theoria as dromena | Yes | User-invoked, produces side effects (the report). Like other dromena (`/spike`, `/validate-rite`), it is a transient operation with lasting output |
| Theoros as agent | Yes | Summoned mid-journey via Task tool, like heroes. But distinguished by being read-only observers, not actors -- a refinement of the hero concept, not a contradiction |
| Pinakes as legomena | Yes | Persistent reference knowledge, progressively disclosed. Follows the exact mena lifecycle model |
| Synkrisis as synthesis | Yes | Operates after observation, before reporting. Fits the handoff pattern: theoroi produce artifacts, synkrisis consumes them, "State of" document crystallizes |
| Argus as pattern name | Yes | Pattern names already exist in doctrine: "Ship of Theseus" (identity problem), "Aegeus Problem" (false confidence). "Argus Pattern" is the parallel observation solution |

### The Deeper Resonance

The theoria concept introduces something the mythology has been missing: **self-awareness of the labyrinth**. Until now, the myth has been about navigating complexity. The theoria is about understanding complexity -- stepping back from the fight to see the battlefield.

This maps to a real architectural need. Knossos has grown (68 CLI commands, 11 rites, hundreds of mena files, extensive doctrine). No single session can hold the whole. The theoria is how the system examines itself -- not through one exhausted hero trying to see everything, but through a disciplined delegation of observers, each with clear criteria, producing reports that are synthesized into truth.

The myth continues to be the architecture. Theseus fights. The theoria observes. Both are necessary. Without the theoria, the labyrinth grows unseen. Without Theseus, nothing is slain. The hero and the observer are complementary faces of the same system.

---

## Appendix A: Source File Reference

All doctrine files audited for this document:

| File | Path | Lines |
|------|------|-------|
| The Coda | `/Users/tomtenuta/Code/knossos/.worktrees/wt-20260209-225839-bfdd/docs/doctrine/philosophy/knossos-doctrine.md` | 553 |
| Design Principles | `/Users/tomtenuta/Code/knossos/.worktrees/wt-20260209-225839-bfdd/docs/doctrine/philosophy/design-principles.md` | 213 |
| Mythology Concordance | `/Users/tomtenuta/Code/knossos/.worktrees/wt-20260209-225839-bfdd/docs/doctrine/philosophy/mythology-concordance.md` | 331 |
| Glossary | `/Users/tomtenuta/Code/knossos/.worktrees/wt-20260209-225839-bfdd/docs/doctrine/reference/GLOSSARY.md` | 253 |
| Spike: State of X | `/Users/tomtenuta/Code/knossos/.worktrees/wt-20260209-225839-bfdd/docs/spikes/SPIKE-state-of-x-audit-primitive.md` | 214 |

## Appendix B: Search Results Summary

| Search Term | Files Found | Collision? |
|-------------|------------|------------|
| `argus` (case-insensitive) | 0 | No |
| `theoria` (case-insensitive) | 0 | No |
| `theoros` (case-insensitive) | 0 | No |
| `pinakes` (case-insensitive) | 0 | No |
| `synkrisis` (case-insensitive) | 0 | No |
| `callimachus` (case-insensitive) | 0 | No |
| `plutarch` (case-insensitive) | 0 | No |
| `aegeus` (case-insensitive) | 9 files | No collision (different concept) |
| `athena` (case-insensitive) | 4 files | No collision (different concept) |
| `pantheon` (case-insensitive) | 20+ files | No collision (different concept) |

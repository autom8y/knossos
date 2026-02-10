# Doctrine Update: Section II and Table Amendments

> Prepared: 2026-02-10
> Status: Final draft for integration-engineer
> Scope: Complete rewrite of Section II "The Naming of Things" + insertion instructions for Sections X, XII, XIV

---

## PART 1: Complete Rewritten Section II

**Replace the entirety of Section II (lines 19-132 in knossos-doctrine.md) with the following:**

---

## II. The Naming of Things

> *"In the beginning was the Word, and the Word was with the architecture, and the Word was the architecture."*

Names are not labels. In Knossos, every name carries its origin like a scar, encoding the architectural intent, the failure mode, and the philosophical commitment of the thing it names. The mythology is the design language. To rename is to re-architect. To misname is to misunderstand.

What follows is the naming of the labyrinth and everything within it.

### The Palace: Knossos

Before there was a platform, there was a palace. **Knossos** on Crete was not a building but a recursion -- corridors folding into corridors, light courts giving way to darkness, storerooms nested within storerooms. Archaeologists still argue whether it was a palace or a temple. It was both. It contained what it protected, and it protected what it contained.

In our architecture, Knossos is the platform itself: the repository, the configuration scaffolding, the accumulated complexity of an evolving system. The labyrinth is not the enemy. The labyrinth is the architecture. It grows because it must. It confuses because complexity, honestly expressed, is confusing.

The repository will bear the name `knossos` when it proves capable of self-hosting. The name is certain; only the ceremony awaits its conditions.

> *The labyrinth was not built to trap Theseus. It was built to contain the Minotaur. That the two functions overlap is the fundamental problem of software architecture.*

### The Inscription

At the entrance to every sacred precinct, words were carved in stone -- dedications, warnings, instructions for the uninitiated. No traveler entered without reading what the builders had written.

**CLAUDE.md** is the Inscription. It is the labyrinth speaking to those who would enter, declaring:

- What heroes may be summoned within these walls
- What rites govern this place
- What customs the traveler must observe
- How to navigate without becoming lost

The Inscription is not static. When rites change, when heroes join or depart, the labyrinth recarves its entrance. The SessionStart hook reads the Inscription and prepares the traveler with what the builders intended them to know.

> *The labyrinth does not hide its nature. It declares it at the gate.*

### The Clew: Ariadne

**Ariadne** gave Theseus the gift that saved him -- not a weapon, not a strategy, but a ball of thread. The clew did not slay the Minotaur. It did something more important: it guaranteed return. Every step into darkness was a step that could be retraced. The gift was not power but memory.

In Knossos, Ariadne is the CLI (`ari`) -- the clew implementation that navigates complexity and guarantees return. Ariadne provides:

- **Session management**: The clew anchors at creation and is cut at wrap
- **Event recording**: Every step knotted into the thread (events.jsonl)
- **Quality gates**: The clew knows when you have reached the exit safely
- **Recovery paths**: If the clew tangles, it can be unwound

Ariadne is not intelligent. Ariadne is faithful. Intelligence is what Theseus brings. Faithfulness is what saves him.

> *The thread does not think. It remembers. That is enough.*

### The Navigator: Theseus

**Theseus** enters the labyrinth for one reason: to slay the Minotaur. He does not enter to map the corridors, to admire the architecture, or to understand the Minotaur's genealogy. He enters to act. In Knossos, the main Claude Code thread is Theseus -- the agentic intelligence that makes decisions, summons help, and creates artifacts.

But Theseus does not travel alone. He summons **heroes** -- specialist agents invoked via the Task tool to lend their strength. An architect for design. An engineer for implementation. An adversary for challenge. These heroes arrive for specific labors and depart when done. They did not walk the whole labyrinth; they were summoned mid-journey to the place where their strength was needed.

This is a feature, not a limitation. Heroes summoned with rich context -- a well-maintained clew, a clear accounting of the journey so far -- perform better than heroes summoned blind. **Better clew, better summoning, better heroes.**

> *Theseus was not the strongest hero. He was the one who remembered to bring the thread.*

### The Fates: Moirai

The **Moirai** -- Clotho, Lachesis, and Atropos -- are older than the Olympians. They are the three Fates who govern the thread of life itself. Not even Zeus dares countermand them. In Knossos, they are three distinct agents, each activated by the event that is theirs alone to answer:

| Fate | Activates On | Function |
|------|--------------|----------|
| **Clotho** | `session_start` | Spins the clew into existence -- bootstrap, initialization, first breath |
| **Lachesis** | State mutations | Measures the allotment -- tracking, token accounting, phase transitions |
| **Atropos** | `session_end`, `wrap` | Cuts when complete -- termination, archival, final reckoning |

The Moirai are the only authority permitted to mutate session state. This is not bureaucracy. This is cosmology. The thread of life belongs to the Fates, and no hero -- however bold -- may spin, measure, or cut it himself.

> *Ariadne gave Theseus the clew as a gift. The Moirai are who she borrowed it from.*

**Relationship to Ariadne**: Ariadne is the survival architecture -- the system ensuring deterministic return. The Moirai are the primordial force that makes clews exist at all. Ariadne is the princess with the idea; the Moirai are the divine mechanism that spins reality into thread.

### The Builder: Daedalus

**Daedalus** built the labyrinth. He also built wings to escape it. He is the archetype of the maker who is both empowered and endangered by his own creation -- the architect who understands that building well is building with walls that constrain even the builder.

In Knossos, Daedalus is the **forge-rite** -- the practice of building tools, agents, and the platform itself. When you need to create new heroes, new rites, new mechanisms, you invoke Daedalus.

Designed complexity is Daedalus's gift. The labyrinth contains AND protects. Not all complexity is debt; some complexity is craft. Good architecture is intentional maze -- walls placed to guide, to shelter, to channel. The forge-rite embodies this: creation with intention, complexity with purpose.

> *Daedalus built the labyrinth so well that he himself could barely escape. This is the mark of honest architecture.*

### The Task: Minotaur

The **Minotaur** is not evil. It is the reason you entered. Without the Minotaur, there is no journey. Without the journey, there is no clew. The beast at the center of the labyrinth is the task -- the complexity that must be confronted, understood, and ultimately resolved.

Some Minotaurs are small: a bug fix, a configuration change. Some are immense: a platform migration, a rearchitecture. The clew cares not for the Minotaur's size -- only that Theseus returns.

> *The Minotaur does not choose to be monstrous. It simply is what the labyrinth was built to contain.*

### The Commissioner: Minos

**Minos** commissioned the labyrinth and demanded tribute. He did not build, he did not navigate, and he did not fight. He created the conditions and expected results.

In Knossos, Minos is the **stakeholder** -- the one who defines the initiative, sets the constraints, and receives the tribute. The tribute? Status reports and demos. The periodic demonstration that proves the labyrinth is being navigated, that progress is real, that the Minotaur will eventually fall.

Minos does not enter the labyrinth. Minos waits. And the waiting is its own kind of power.

### The Oracle: Pythia

Before every great journey, Greeks traveled to Delphi to consult the **Pythia** -- the priestess who sat upon the tripod above the chasm, breathing the vapors of prophecy. Her words shaped expeditions, wars, and colonizations.

In Knossos, the Pythia is the **orchestrator** -- the voice consulted before and during the journey. Unlike the historical Pythia, our oracle speaks clearly. The Pythia provides:

- Work breakdown and phase planning
- Specialist routing (which hero for which labor)
- Checkpoint guidance (what to do next)

When uncertain, `/consult` the Pythia. The oracle's clarity is a design choice: ambiguity belongs to the labyrinth, not to the guide.

### The Watchers: Aegeus

**Aegeus**, King of Athens, stood on the cliff at Cape Sounion and watched the horizon for his son's return. The agreement was simple: white sails if Theseus lives, black sails if he dies. Theseus forgot to change the sails. Aegeus, seeing black against the sea, threw himself from the cliff and drowned.

The Aegeus problem is **false confidence**. Tests pass but edge cases lurk. Builds complete but integration points fracture. The watchers trust the signal and are destroyed by its lie. Aegeus died not because Theseus failed but because the signal was wrong.

**White Sails** solves this. A computed confidence signal -- never self-declared, always evidence-derived -- ensures that what Aegeus sees from the cliff is true.

> *Aegeus did not die from grief. He died from a bad signal. The system must not lie to its watchers.*

### The Destination: Athens

**Athens** is home -- where Theseus returns after the journey. The ship enters the harbor. The crowd gathers. The hero steps ashore. In Knossos, Athens is **the main branch**. You have not returned until you have merged. The PR is the ship; the merge is the homecoming.

Every journey through the labyrinth ends in one of two places: Athens, or Naxos. There is no third destination.

### The Transformer: Dionysus

After Theseus abandoned Ariadne on Naxos, **Dionysus** found her on that desolate shore and made her divine. He took what was abandoned and elevated it. He transformed grief into godhood.

In Knossos, Dionysus is **code review** -- the transformative process that elevates raw work into merged truth. Work created in isolation, tested in solitude, shaped by a single mind -- Dionysus takes this and blesses it for union with the canon. Independent review is not gatekeeping; it is transfiguration.

> *Dionysus does not judge what he finds on the shore. He transforms it.*

### The Shore of Abandonment: Naxos

**Naxos** is where Theseus left Ariadne -- abandoned after she saved him. The myth does not explain why. Perhaps exhaustion. Perhaps forgetfulness. Perhaps the simple human failure of leaving behind what you no longer need in the moment.

In Knossos, Naxos represents:

- **Orphaned sessions**: Created but never wrapped, abandoned by their Theseus
- **Stale gray sails**: Work completed but never validated, confidence never earned

Sessions left too long on Naxos accumulate. They represent unfinished business, broken promises, the technical debt of the workflow itself. Naxos is the shore that honest reconnaissance must eventually survey.

> *Every labyrinth has its Naxos -- the place where good intentions go to be forgotten.*

### The Delegation: Theoria

Before great decisions, Greek city-states did not simply act. They dispatched a **theoria** -- an official delegation of sacred observers sent to distant sanctuaries to witness festivals, consult oracles, and bring back truth. The theoroi did not go to fight or trade. They went to see. The word shares its root with "theory": *theoria* was structured contemplation, knowledge gained through disciplined observation.

In Knossos, a Theoria is the **audit operation** -- a structured delegation of observers dispatched into the labyrinth's domains to assess their state. When Theseus needs to understand the labyrinth itself -- not to slay the Minotaur but to see the passages clearly -- he pauses mid-journey and dispatches a theoria. The command is `/theoria`.

This is a deliberate shift from *praxis* to *theoria* -- from action to contemplation. The theoria does not change what it observes. It reveals what is. The delegation consults the Pinakes to know what domains exist, dispatches theoroi through the Argus Pattern, and weaves their reports through synkrisis into a single attestation of truth.

> *The hero who fights sees only the Minotaur before him. The polis that sends a theoria sees the whole labyrinth.*

### The Sacred Observers: Theoroi

A **theoros** (plural: **theoroi**) was an individual sacred observer within a theoria delegation. Theoroi were chosen for their judgment, sent to distant sanctuaries to observe rituals and bring back accounts. They were not participants but witnesses. Their value was in the clarity of their seeing, not the strength of their arms.

In Knossos, the theoroi are **domain evaluator agents** -- each dispatched via Task tool to assess a single domain of the labyrinth. A theoros receives its evaluation criteria from the Pinakes, examines its domain independently, and produces a structured report. The agent definition lives at `rites/shared/agents/theoros.md`.

The distinction between hero and theoros is architectural and inviolable. Heroes modify the labyrinth; theoroi only observe it. A hero enters with a sword. A theoros enters with open eyes. If an agent is named theoros, it should not be making changes. The naming enforces the discipline.

> *A hero enters the labyrinth with a sword. A theoros enters with open eyes.*

### The Catalog: Pinakes

**Callimachus**, the great librarian of Alexandria, compiled the **Pinakes** -- the first systematic catalog of the ancient world's largest library. The Pinakes did not contain the knowledge of the Library. It described what knowledge existed, where it could be found, and how it was classified. It was the map of the Library, not the Library itself.

In Knossos, the Pinakes is the **domain registry** -- the reference knowledge that catalogs audit targets, evaluation criteria per domain, grading rubrics, and report schemas. Stored at `rites/shared/mena/pinakes/`, it tells the theoria what domains exist and how to assess each one.

The Pinakes is a legomenon: persistent reference knowledge, progressively disclosed, consulted but never consumed. It is the bridge between wanting to audit and knowing how. Without the Pinakes, theoroi would wander without criteria. With it, they arrive at their domain knowing exactly what to observe and how to grade what they find.

> *Callimachus did not write the books. He told you which books existed, where they stood, and what they contained. The Pinakes does the same for domains.*

### The Comparison: Synkrisis

**Plutarch**, in his *Parallel Lives*, followed each pair of biographies with a **synkrisis** -- a structured comparison that set two lives side by side to reveal patterns, contrasts, and truths visible only in juxtaposition. After presenting Alexander and Caesar independently, Plutarch composed his synkrisis: not summary but synthesis, not aggregation but analysis.

In Knossos, Synkrisis is the **synthesis step** that follows the theoria's return. After all theoroi have reported individually, synkrisis weaves their findings together -- identifying cross-domain patterns, systemic issues, recurring strengths, and truths invisible to any single observer. The synkrisis produces the final attestation: the "State of the {X}" document.

Individual domain reports are valuable but partial. A rite might have excellent agents but broken hooks -- and the agents are healthy precisely because they compensate for the broken hooks. Only synkrisis reveals this compensatory coupling. Only comparison surfaces the truth that lives between the reports.

> *Seven theoroi return with seven truths. Only synkrisis reveals the eighth -- the truth that lives between them.*

### The Hundred Eyes: The Argus Pattern

**Argus Panoptes** -- the hundred-eyed giant -- was set by Hera to watch over Io. His eyes never all closed at once: some slept while others watched. He was not strong. He was not swift. He was total vigilance through distributed, overlapping observation. One body, a hundred eyes, nothing unseen.

In Knossos, the **Argus Pattern** names the N-agent parallel dispatch technique: launching multiple Task tool agents simultaneously, each observing a different domain, their combined vision covering the whole. The main thread is the body; the dispatched agents are the eyes. The pattern's architectural constraint mirrors Argus's nature -- agents cannot spawn agents, only the main thread dispatches. One giant, many eyes. One Theseus, many theoroi.

The Argus Pattern is not specific to auditing. Like "the Ship of Theseus" names the identity-through-transformation problem and "the Aegeus Problem" names false confidence, "the Argus Pattern" names the parallel-observation-through-distributed-agents solution. The theoria is its first named user. Future operations -- parallel validation, parallel migration verification, parallel documentation generation -- will use it too.

> *One eye sees what is before it. A hundred eyes see what is.*

---

## PART 2: Section X Service Map Additions

**INSERT the following 5 rows into the service map table in Section X.**
**Insert AFTER the existing `Naxos` row (the line reading `| **Naxos** | Orphaned sessions, stale gray sails | The shore of abandonment |`) and BEFORE the `White Sails` row:**

```
| **Theoria** | Audit operation (`/theoria`) | The sacred delegation -- structured observation of the labyrinth |
| **Theoroi** | Domain evaluator agents (`rites/shared/agents/theoros.md`) | Sacred observers dispatched to witness and report |
| **Pinakes** | Domain registry (`rites/shared/mena/pinakes/`) | Callimachus's catalog -- what to observe and how to assess it |
| **Synkrisis** | Synthesis step | Plutarch's comparison -- truth that emerges between reports |
| **Argus Pattern** | N-agent parallel dispatch | The hundred-eyed watcher -- total vision through distributed observation |
```

---

## PART 3: Section XII Terminology Concordance Additions

**INSERT the following rows at the end of the concordance table in Section XII.**
**Insert AFTER the existing last row (`| roster (repository) | knossos | Platform name (pending rename) |`):**

```
| `domain-auditor` | `theoros` / `theoroi` | Working name from spike; mythological name in doctrine |
| `state-of-ref` | `pinakes` | Working name from spike; mythological name for domain registry |
| `/state-of` | `/theoria` | Working name from spike; mythological name for audit command |
```

---

## PART 4: Section XIV Drift Registry Additions

**INSERT the following rows under "Concepts Documented but Not Fully Implemented" in Section XIV.**
**Insert AFTER the existing last row in that table (`| Dionysus integration | Partial | Code review exists but not mythologically named |`):**

```
| Theoria audit primitive | Doctrine only | Spike complete (SPIKE-state-of-x-audit-primitive.md); `/theoria` dromena not yet forged |
| Theoroi (domain evaluators) | Doctrine only | Requires theoros agent at `rites/shared/agents/theoros.md` |
| Pinakes (domain registry) | Doctrine only | Requires domain criteria legomena at `rites/shared/mena/pinakes/` |
| Synkrisis (synthesis step) | Doctrine only | Main-thread or dedicated synthesis agent; approach undecided |
| Argus Pattern (N-agent parallel dispatch) | Named pattern | Reusable parallel dispatch pattern; envisioned for tactical playbook swarms and similar N-agent operations |
```

---

## Integration Notes

1. **Section II placement rationale**: The five new concepts are placed after Naxos, before the Section III divider. Theoria, Theoroi, Pinakes, and Synkrisis form a coherent sub-narrative (the reconnaissance arc) and are presented in operational order: the delegation is dispatched, the observers go forth, they consult the catalog, they return for synthesis. The Argus Pattern closes the section as a named architectural pattern -- the mechanism that makes theoria possible and that will serve future operations.

2. **Voice consistency**: Every existing subsection has been rewritten to match the elevated voice of the new additions. Epigraphs have been added where they earn their place. Narrative depth has been increased throughout while preserving all core architectural mappings and technical content.

3. **Path convention**: All new entries use relative paths from repo root (`rites/shared/agents/theoros.md`, `rites/shared/mena/pinakes/`). No `/roster/` paths appear in new content.

4. **Argus Pattern placement**: Placed in Section II as a named thing (consistent with "The Naming of Things") rather than in Section XI (Design Principles). The existing pattern-names -- Ship of Theseus (Section V) and Aegeus Problem (Section II) -- live where they organically arise rather than in a separate patterns section. The Argus Pattern arises organically from the theoria narrative.

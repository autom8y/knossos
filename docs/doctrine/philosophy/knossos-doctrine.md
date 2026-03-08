---
last_verified: 2026-02-26
---

# The Knossos Doctrine

> Mythology as architecture. The myth is the system; the system is the myth.

This is **The Coda**—the philosophical foundation that gives the platform its meaning. Knossos is the implementation; The Coda is the why.

---

## I. Cosmogony

Before there was order, there was the labyrinth—vast, recursive, hungry for those who enter without clew. Every codebase of sufficient complexity becomes a labyrinth. Every session is a journey through it. Every return is a minor miracle.

The Knossos platform exists because **safe return is not guaranteed**. Context evaporates. Decisions scatter. The path that brought you here fades behind you. Without the clew, you are lost.

This is not documentation. This is doctrine.

---

## II. The Naming of Things

> *"In the beginning was the Word, and the Word was with the architecture, and the Word was the architecture."*

Names are not labels. In Knossos, every name carries its origin like a scar, encoding the architectural intent, the failure mode, and the philosophical commitment of the thing it names. The mythology is the design language. To rename is to re-architect. To misname is to misunderstand.

The platform's naming draws from multiple wells, and the doctrine acknowledges this honestly rather than pretending a single-myth purity. **Tier 1** names carry Bronze Age attestation -- Potnia appears on Linear B tablet KN Gg(1) 702, excavated at Knossos itself. **Tier 2** names come from classical sources -- Homer, Hesiod, Plutarch, Ovid -- the myth cycle of Theseus, Minos, Daedalus, and the Moirai. **Tier 3** names borrow from Hellenistic scholarship and Panhellenic practice -- theoria, pinakes, synkrisis, exousia -- because they describe their architectural functions with precision no Cretan alternative could match. **Tier 4** names are functional analogies -- rite, mena, dromena, legomena, inscription -- chosen for resonance with the role they name. The full provenance is recorded in the [mythology concordance](mythology-concordance.md).

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

But Ariadne was not the thread. She was the princess who devised the plan, understood the labyrinth, and chose the right instrument for the moment. The clew was one gift of a thinking mind. Faithfulness and intelligence were never opposites in her; they were the same quality, expressed differently.

In Knossos, Ariadne is the CLI (`ari`) -- the intelligence that navigates complexity and guarantees return. Ariadne provides:

- **Session management**: The clew anchors at creation and is cut at wrap
- **Event recording**: Every step knotted into the thread (events.jsonl)
- **Quality gates**: The clew knows when you have reached the exit safely
- **Recovery paths**: If the clew tangles, it can be unwound
- **Search and synthesis**: TF-IDF scoring, synonym expansion, and natural language queries over platform knowledge
- **Proactive guidance**: Suggestions fired at decision points, before Theseus has thought to ask

Ariadne is faithful because she is intelligent. The thread remembers because a thinking princess chose to give it.

> *The thread does not think. The princess does — and the thread is what her thinking produced.*

### The Navigator: Theseus

**Theseus** enters the labyrinth for one reason: to slay the Minotaur. He does not enter to map the corridors, to admire the architecture, or to understand the Minotaur's genealogy. He enters to act. In Knossos, the main Claude Code thread is Theseus -- the agentic intelligence that makes decisions, summons help, and creates artifacts.

But Theseus does not travel alone. He summons **heroes** -- specialist agents invoked via the Task tool to lend their strength. An architect for design. An engineer for implementation. An adversary for challenge. These heroes arrive for specific labors and depart when done. They did not walk the whole labyrinth; they were summoned mid-journey to the place where their strength was needed.

This is a feature, not a limitation. Heroes summoned with rich context -- a well-maintained clew, a clear accounting of the journey so far -- perform better than heroes summoned blind. **Better clew, better summoning, better heroes.**

> *Theseus was not the strongest hero. He was the one who accepted the thread when Ariadne offered it.*

### The Fates: Moirai

The **Moirai** -- Clotho, Lachesis, and Atropos -- are older than the Olympians. They are the three Fates who govern the thread of life itself. Not even Zeus dares countermand them. In Knossos, they are three aspects embodied in one agent, each activated by the event that is theirs alone to answer:

| Fate | Activates On | Function |
|------|--------------|----------|
| **Clotho** | `session_start` | Spins the clew into existence -- bootstrap, initialization, first breath |
| **Lachesis** | State mutations | Measures the allotment -- tracking, token accounting, phase transitions |
| **Atropos** | `session_end`, `wrap` | Cuts when complete -- termination, archival, final reckoning |

The Moirai are the only authority permitted to mutate session state. This is not bureaucracy. This is cosmology. The thread of life belongs to the Fates, and no hero -- however bold -- may spin, measure, or cut it himself.

> *Ariadne gave Theseus the clew as a gift. The Moirai are who she borrowed it from.*

**Relationship to Ariadne**: Ariadne is the survival architecture -- the system ensuring deterministic return. The Moirai are the primordial force that makes clews exist at all. Ariadne is the princess with the idea; the Moirai are the divine mechanism that spins reality into thread.

### The Builder: Daedalus

**Daedalus** built the labyrinth. He also built wings to escape it. He is the archetype of the maker who is both empowered and endangered by his own creation -- the architect who discovers that building well means building walls that constrain even the builder.

In Knossos, Daedalus is the **forge-rite** -- the practice of building tools, agents, and the platform itself. When you need to create new heroes, new rites, new mechanisms, you invoke Daedalus.

Designed complexity is Daedalus's gift. The labyrinth contains AND protects. Not all complexity is debt; some complexity is craft. Good architecture is intentional maze -- walls placed to guide, to shelter, to channel. The forge-rite embodies this: creation with intention, complexity with purpose.

But Minos imprisoned Daedalus for knowing the labyrinth's secrets too well. The builder became captive to his own masterwork. TENSION-001 and TENSION-002 are Daedalean traps: dual `RiteManifest` structs and a `legacyOpts`/`legacyResult` mapping layer built soundly at the time, now confining anyone who touches them. Craft can imprison the craftsman.

> *Daedalus built the labyrinth so well that he himself could barely escape. This is the mark of honest architecture -- and its danger.*

### The Overreach: Icarus

**Icarus** flew too close to the sun on wings his father built. The wings were not flawed. The constraints were not hidden. Daedalus warned him explicitly. Icarus exceeded the constraint surface anyway -- and fell.

In Knossos, Icarus is the SCAR catalog: SCAR-002 renamed `.claude/` and froze Claude Code solid; SCAR-005 called `os.RemoveAll` on directories containing user content and destroyed it; SCAR-018 set `context: fork` on a dromenon that needed the Task tool and silently degraded parallel dispatch to sequential reads; SCAR-026 wired delegation hints into writeguard infrastructure and was reverted because the coupling was unsound. Each of these was an ambitious change that ignored the constraint surface -- the CC file watcher, the user content boundary, the Task tool restriction. Each fell.

The forge-rite builds tools. Icarus is the reminder that tools used without respecting constraints will fail. The SCAR catalog is not a list of bad ideas. It is a record of good ideas deployed without sufficient respect for where the wax melts.

> *The wings worked. The sun was the constraint he refused to honor.*

### The Beast: Minotaur

The **Minotaur** was born from a broken promise. Minos swore to sacrifice Poseidon's bull and kept it instead. Poseidon's punishment was Pasiphae's madness and the creature that resulted -- not a neutral challenge but a systemic consequence, fed on tribute until Theseus arrived to end the cycle.

In Knossos, the Minotaur is **accumulated technical debt and systemic dysfunction** born from shortcuts and broken promises in the platform itself. It is not a bug fix, not a configuration change, not any individual task. It is the accumulated condition that makes work harder than it should be. The 28 entries in the SCAR catalog -- 9 integration failures, 6 data corruption events, race conditions, schema drift, and performance cliffs -- these are the Minotaur's kill list. They are what happens when the system develops unsound assumptions and no one confronts them.

The labyrinth was built to contain the Minotaur, not to deny its existence. When you navigate Knossos, you navigate around its weight. When you slay a Minotaur -- not a single bug but a systemic failure mode, eliminated by its root cause -- you reduce the tribute the system demands from every future traveler.

> *The Minotaur was not born evil. It was born from a promise that was not kept. That is how systemic dysfunction always begins.*

### The Commissioner: Minos

**Minos** commissioned the labyrinth and demanded tribute. He did not build, he did not navigate, and he did not fight. He created the conditions and expected results.

In Knossos, Minos is the **stakeholder** -- the one who defines the initiative, sets the constraints, and receives the tribute. The tribute? Status reports and demos. The periodic demonstration that proves the labyrinth is being navigated, that progress is real, that the Minotaur will eventually fall.

Minos does not enter the labyrinth. Minos waits. And the waiting is its own kind of power.

### The Presiding Lady: Potnia

Linear B tablet **KN Gg(1) 702**, excavated at Knossos and dated ~1450-1300 BCE, records offerings to *da-pu₂-ri-to-jo po-ti-ni-ja* -- the **Potnia of the Labyrinth**. She was the presiding authority within the palace itself, receiving offerings equal to those given to all the gods combined. The etymology is *\*pot-niha*: "she who has authority." Not an external consultant but the power that resides within.

In Knossos, the Potnia is the **per-rite entry agent** -- the presiding authority within each rite's domain. Each rite has its own Potnia at `rites/*/agents/potnia.md`. The Potnia provides:

- Work breakdown and phase planning
- Specialist routing (which hero for which labor)
- Checkpoint guidance (what to do next)

The Potnia speaks clearly, not cryptically. The distinction matters: **Pythia** (the cross-rite oracle at `agents/pythia.md`) is the external voice consulted before entering the labyrinth -- routing and navigation across rites. Potnia is the authority who presides within. Different roles, different mythological positions, both correctly placed.

Every Potnia (and every hero) carries an **Exousia** -- an authority contract declaring what the agent decides autonomously, what it escalates, and what it must never decide. Exousia makes jurisdiction explicit and auditable.

For cold starts without a session, `/go` routes through Pythia to the appropriate Potnia. The clarity of both is a design choice: ambiguity belongs to the labyrinth, not to the guides.

### The Watchers: Aegeus

**Aegeus**, King of Athens, stood on the cliff at Cape Sounion and watched the horizon for his son's return. The agreement was simple: white sails if Theseus lives, black sails if he dies. Theseus forgot to change the sails. Aegeus, seeing black against the sea, threw himself from the cliff and drowned.

The Aegeus problem is **false confidence**. Tests pass but edge cases lurk. Builds complete but integration points fracture. The watchers trust the signal and are destroyed by its lie. Aegeus died not because Theseus failed but because the signal was wrong.

**White Sails** solves this. A computed confidence signal -- never self-declared, always evidence-derived -- ensures that what Aegeus sees from the cliff is true.

> *Aegeus did not die from grief. He died from a bad signal. The system must not lie to its watchers.*

### The Destination: Athens

**Athens** is home -- where Theseus returns after the journey. The ship enters the harbor. The crowd gathers. The hero steps ashore. In Knossos, Athens is **the main branch**. You have not returned until you have merged. The PR is the ship; the merge is the homecoming.

Every journey through the labyrinth ends in one of two places: Athens, or Naxos. There is no third destination.

### The Transformer: Dionysus

After Theseus abandoned Ariadne on Naxos, **Dionysus** found her on that desolate shore and made her divine. He took what was abandoned and elevated it. He transformed grief into godhood — the apotheosis on Naxos, where the raw became the refined and the mortal became immortal.

In Knossos, Dionysus is **transformation of the raw into the refined** -- the cross-session knowledge synthesizer that reads abandoned session data and distills it into permanent wisdom. The `ari land` pipeline IS the Dionysian apotheosis: raw session archives left on Naxos (abandoned, ephemeral, mortal) are transformed into landed knowledge (persistent, refined, enduring) at `.sos/land/`. Dionysus transforms raw grapes into wine. The agent transforms raw sessions into refined knowledge.

The dual Naxian festival tradition — joy for the apotheosized Ariadne, mourning for the abandoned one — maps to the dual nature of session archival: there is value in what was produced (celebration) and loss in the context that was discarded (mourning). The `ari naxos` scanner finds the abandoned sessions; `ari land` performs the Dionysian rescue.

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

In Knossos, the Pinakes is the **domain registry** -- the reference knowledge that catalogs audit targets, evaluation criteria per domain, grading rubrics, and report schemas. Stored at `mena/pinakes/`, it tells the theoria what domains exist and how to assess each one.

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

### What the Doctrine Does Not Name

The service map is not complete. Several significant implementation structures operate without mythological names: the materialize pipeline (spanning 30+ files, the central hub of the codebase), the perspective system (the 9-layer context envelope), and the provenance ownership trichotomy (knossos-owned, user-owned, and untracked). These structures are real, architecturally significant, and deliberately unnamed here. Not every structure needs a myth. Assigning mythology prematurely risks decorating something that has not yet found its settled form. The doctrine's silence on these structures is not ignorance -- it is honesty about where the map ends and the territory continues.

---

## III. Mortal Limits

Heroes are not gods. They tire. They forget. They can only carry so much. This is the doctrine of **Mortal Limits**—the fundamental constraint that shapes all design.

### Weight Economy

Victory comes not from strength but from knowing what to bring — and what to leave behind. In Knossos, this manifests as:

- **Rite selection**: Load the right practice for the journey
- **Skill curation**: Not all knowledge helps; some adds weight
- **Hero summoning**: Call specialists with focused context, not everything

The economy is bidirectional: knowing what to bring AND what to leave. Well-tooled travelers navigate better. Travelers who bring too much collapse under the weight.

### Why Less Is More

Context windows are finite. When you load a rite, you gain its heroes and knowledge. But you also pay the toll—context consumed, attention divided.

If you don't bring a rite, you cannot summon its heroes. But if you bring too many rites, your summoning prompts become diluted and your heroes arrive confused.

This is why rites exist: **bounded practices for specific domains**. Swap when you must, invoke when you need a piece, but never pretend you can carry everything.

### The Cognitive Budget

Lachesis measures more than session state. She measures the **cognitive budget**—the expenditure of context, the accumulation of messages, the approaching exhaustion.

When the budget nears its end, the system warns. Parking is not failure; parking is wisdom.

---

## IV. The Rites

A **Rite** is an invokable practice—a bundle of knowledge, practitioners, and procedures for a specific domain. Where the older tongue spoke of "teams" and "rites," we now speak of Rites.

The change is not cosmetic. Teams imply completeness and fixed membership. Rites are flexible compositions:

| Rite Form | Contains | Example |
|-----------|----------|---------|
| Simple Rite | Skills only | `documentation-rite`: knowledge without agents |
| Practitioner Rite | Agents + skills | `code-review-rite`: specialists with their craft |
| Procedural Rite | Hooks + workflows | `release-rite`: pure ceremony, no dedicated agents |
| Full Rite | All components | `quality-rite`: complete practice |

### Rite Operations

Rites are invokable without abandoning your current practice:

| Command | Action |
|---------|--------|
| `current-rite` | Your active context (the full practice you're working within) |
| `invoke-rite documentation` | Bring the whole rite (whatever it contains) |
| `invoke-rite documentation skills` | Borrow just the knowledge |
| `invoke-rite quality agents` | Summon just the practitioners |
| `swap-rite quality` | Full context switch—new rite becomes current |

### The Cost Model

The distinction between **swap-rite** and **invoke-rite** is economic:

- **swap-rite** = "park and regroup" (costly, full context switch)
- **invoke-rite** = "call for supplies mid-journey" (targeted, lighter weight)

You can invoke a documentation rite while still practicing under the quality rite—borrowing useful knowledge without conversion. Context switching is expensive; knowledge sharing is cheap.

### The Primary Rites

| Rite | Domain | Key Heroes |
|------|--------|------------|
| **10x-dev** | Full development lifecycle | requirements-analyst, architect, principal-engineer, qa-adversary |
| **arch** | Architecture assessment | topology-cartographer, structure-evaluator, dependency-analyst, remediation-planner |
| **forge** | Agent and tool creation (Daedalus's domain) | agent-designer, prompt-architect, workflow-engineer, platform-engineer, agent-curator, eval-specialist |
| **docs** | Knowledge crystallization | doc-auditor, information-architect, tech-writer, doc-reviewer |
| **hygiene** | Code quality maintenance | audit-lead, code-smeller, architect-enforcer, janitor |
| **debt-triage** | Technical debt remediation | debt-collector, risk-assessor, sprint-planner |
| **security** | Threat modeling and compliance | threat-modeler, security-reviewer, penetration-tester, compliance-architect |
| **sre** | Operations and reliability | observability-engineer, incident-commander, chaos-engineer, platform-engineer |
| **intelligence** | Research and synthesis | user-researcher, insights-analyst, analytics-engineer, experimentation-lead |
| **rnd** | Spikes and prototypes | technology-scout, moonshot-architect, prototype-engineer, integration-researcher, tech-transfer |
| **strategy** | Business analysis | market-researcher, competitive-analyst, business-model-analyst, roadmap-strategist |
| **ecosystem** | Platform infrastructure | ecosystem-analyst, context-architect, integration-engineer, documentation-engineer, compatibility-tester |
| **slop-chop** | AI code quality gate | hallucination-hunter, logic-surgeon, cruft-cutter, gate-keeper, remedy-smith |
| **shared** | Cross-rite resources | theoros |

---

## V. The Journey (Session Lifecycle)

Every session is a journey through the labyrinth. The Moirai govern this journey, each activated by events:

### Birth (Clotho's Domain)

```
(void) ──[session_start]──> Clotho activates ──> ACTIVE
```

Clotho spins the clew into existence. A session is created with:
- An **initiative**: what Minotaur are we hunting?
- A **complexity**: how deep does the labyrinth go?
- A **rite**: what practices will guide us?

The session receives an ID (the clew's anchor), a directory (the clew's spool), and entry begins.

### Measure (Lachesis's Domain)

```
ACTIVE ──[state mutation]──> Lachesis activates ──> validates ──> records
```

Lachesis measures the allotment. During the active journey:
- **Phases progress**: requirements → design → implementation → validation
- **Events accumulate**: the clew gains knots (events.jsonl)
- **Artifacts crystallize**: PRDs, TDDs, ADRs, code
- **Budget depletes**: the cognitive budget warns when measure nears end

Parking is not failure—it is acknowledgment that the journey must pause. The clew remains intact; the traveler rests.

### Completion (Atropos's Domain)

```
ACTIVE ──[wrap_session]──> Atropos activates ──> ARCHIVED
```

Atropos cuts the clew when the journey completes. At wrap:
1. **Quality gates execute**: proofs are collected
2. **White Sails compute**: confidence is calculated
3. **Archive occurs**: the session becomes read-only
4. **The clew is knotted off**: events.jsonl is sealed

A wrapped session is immutable testimony.

### The Ship of Theseus

If you replace every plank of a ship, is it the same ship? Sessions accumulate changes until the original context is transformed.

The doctrine of **Session Continuity**: The clew provides identity even as context degrades. The events.jsonl IS the session—not the context window, not the agent's memory, but the recorded journey. When Theseus forgets, the clew remembers.

### Linear A and Linear B — The Inscriptions Theseus Cannot Read

The palace at Knossos contained two writing systems: **Linear B** (deciphered by Michael Ventris in 1952, revealed as an early form of Greek) and **Linear A** (still undeciphered, the script of the original Minoan civilization). Linear A is the older layer—structurally present in the archaeological record but semantically opaque to current readers. The labyrinth remembers, but some of what it remembers is written in a language we may never read.

Context degradation is not simply "Theseus forgets." It is more precise than that: **the labyrinth contains inscriptions in languages Theseus cannot read.** Fresh context is Linear B—readable, deciphered, actionable. Context that has been compressed, summarized, or pushed beyond the context window becomes Linear A—structurally present but semantically opaque. The symbols are there. The meaning is lost.

Every LLM session operates with a Linear A layer: compressed conversation history, archived sessions from previous sprints, decisions whose rationale is no longer in context. The `.know/` files, `.sos/land/` synthesis, and the SCAR catalog are **translation projects**—rendering Linear A into Linear B so future navigators can read what came before. The `ari land` pipeline (Dionysus's knowledge distillation) is specifically a Linear A to Linear B translation effort: taking the opaque accumulated experience of past sessions and rendering it into structured, readable knowledge that future agents can consume.

TENSION-004 (triple event format v1/v2/v3 in `events.jsonl`) is literally a stratigraphic record of three writing systems interleaved in one file. The `ReadEvents()` function in `session/events_read.go` performs exactly the kind of multi-format interpretation that archaeologists do when reading tablets that mix Linear A and Linear B conventions.

> *The labyrinth remembers more than any navigator can comprehend. The question is not whether Theseus forgets, but whether anyone has translated the inscriptions he will need.*

### The State Machine

```
                    +--------------+
                    |              |
                    v              |
   NONE ---> ACTIVE ---> PARKED --+---> ARCHIVED
                |                         ^
                +-------------------------+
                     (direct wrap)
```

There are four states. NONE is pre-existence. ACTIVE is alive. PARKED is resting. ARCHIVED is complete.

---

## VI. The Clew Contract

The **Clew Contract** is the agreement between Ariadne and those who use her gifts. The clew records:

### Event Types

| Event | Meaning |
|-------|---------|
| `session_start` | Clotho spins the clew |
| `session_end` | Atropos cuts the clew |
| `task_start` | Hero summoning begins |
| `task_end` | Hero summoning completes |
| `tool_call` | An action is taken |
| `file_change` | The labyrinth is modified |
| `decision` | A fork is chosen |
| `artifact_created` | Something crystallizes |
| `error` | Something breaks |
| `sails_generated` | Confidence is computed |
| `handoff_prepared` | Transition is validated |
| `handoff_executed` | Transition completes |

### Decision Stamps

At significant forks, the clew records a **Decision Stamp**:

```yaml
type: decision
decision: "Use event sourcing for audit trail"
rationale: "Immutable append-only log enables time-travel debugging"
rejected:
  - "Mutable state: loses history"
  - "Database triggers: couples to storage implementation"
context: "PRD-ariadne.md section 3.2"
```

Theseus forgets why he turned left. The Clew remembers.

---

## VII. The Confidence Signal (White Sails)

Every completed journey produces a **White Sails** attestation—a signal to Aegeus watching from the cliff.

### The Three Colors

| Color | Meaning | Conditions |
|-------|---------|------------|
| **WHITE** | Safe return, high confidence | All proofs present, tests pass, lint clean, no open questions |
| **GRAY** | Unknown waters, needs verification | Missing proofs, open questions, spike/hotfix type, declared uncertainty |
| **BLACK** | Known failure, do not deploy | Tests failing, build broken, explicit blocker |

There is no yellow, no intermediate. Simplicity prevents gaming.

### The Computation

White Sails are computed, not declared:

1. **Check for failures** → BLACK
2. **Check for open questions** → GRAY ceiling
3. **Check session type** → spikes and hotfixes are GRAY ceiling
4. **Check proof completeness** → missing proofs are GRAY
5. **Apply modifiers** → humans can downgrade, never self-upgrade
6. **Check QA upgrade** → Dionysus (transformation of the raw into the refined) can elevate GRAY to WHITE

### Anti-Gaming Mechanisms

- **Cannot self-upgrade**: Modifiers only downgrade
- **QA upgrade requires proof**: constraint resolution log + adversarial tests
- **Open questions propagate**: Any "?" triggers gray ceiling
- **Proof verification**: Evidence paths must exist

---

## VIII. The Handoff

When one hero completes their labor and another must begin, the **Handoff** occurs. This is not mere delegation—it is a ritual transfer of context.

### Handoff Validation

Before a handoff, artifacts are validated:
- Does the PRD have acceptance criteria?
- Does the TDD define interfaces?
- Are the tests written before code?

If validation fails, the handoff blocks. The receiving hero enters prepared, not blind.

### The Handoff Events

```
task_end (from_hero) → handoff_prepared → handoff_executed → task_start (to_hero)
```

The clew records every transition.

---

## IX. The Hooks (Trap Mechanisms)

**Hooks** are trap mechanisms within the labyrinth—triggers that fire automatically at specific moments.

### Lifecycle Hooks

| Event | When | Purpose |
|-------|------|---------|
| `SessionStart` | Session begins | Read the Inscription, load rite, inject context |
| `PreToolUse` | Before tool executes | Validate, guard, route |
| `PostToolUse` | After tool completes | Record, track, detect |

### Key Hooks

- **context-injection**: SessionStart hook that reads the Inscription (CLAUDE.md) and prepares the traveler
- **clew**: PostToolUse hook that records events to events.jsonl
- **writeguard**: PreToolUse hook that protects context files from direct modification
- **autopark**: PostToolUse hook that suggests parking when budget depletes
- **route**: PreToolUse hook that enforces orchestration patterns

### The Write Guard

Context files are protected. Direct writes are intercepted and rejected. All mutations flow through the Moirai.

This is not bureaucracy—it is the only way to guarantee validity, consistency, and auditability.

---

## X. The Complete Service Map

| Myth | Component | Philosophical Function |
|------|-----------|------------------------|
| **Knossos** | The platform (roster/.claude/) | The labyrinth itself—complexity incarnate |
| **The Inscription** | CLAUDE.md | The labyrinth's entrance, declaring what heroes and rites are available |
| **Ariadne** | CLI binary (`ari`) | The intelligence that navigates complexity and guarantees return |
| **The Clew** | Session state + events.jsonl | The provenance trail, identity through transformation |
| **Theseus** | Main Claude Code thread | The navigator who summons heroes |
| **Heroes** | Specialist agents (Task tool) | Summoned champions for specific labors |
| **Clotho** | Session bootstrap agent | The Fate who spins the clew into existence |
| **Lachesis** | State mutation agent | The Fate who measures and tracks |
| **Atropos** | Session termination agent | The Fate who cuts when complete |
| **Potnia** | Per-rite entry agent (`rites/*/agents/potnia.md`) | The presiding lady within the labyrinth (Linear B KN Gg 702) |
| **Pythia** | Cross-rite oracle/navigator (`agents/pythia.md`) | The external oracle consulted before entering the labyrinth |
| **Exousia** | Authority contract (`## Exousia` in agents) | Jurisdictional boundaries -- Decide / Escalate / Do NOT Decide |
| **Daedalus** | Forge-rite | The builder of tools and agents — and captive to his own craft |
| **Icarus** | The SCAR catalog | Ambitious changes that ignored the constraint surface and fell |
| **Minos** | Stakeholder | The commissioner who demands tribute |
| **Minotaur** | Accumulated technical debt / systemic dysfunction | Born from broken promises; the beast the labyrinth was built to contain |
| **Dionysus** | Cross-session knowledge synthesis (`agents/dionysus.md`, `ari land`) | Transformation of the raw into the refined — abandoned session data becomes persistent wisdom |
| **Aegeus** | CI/CD, production monitors | Those watching from the cliff |
| **Athens** | The main branch | Home—where you return by merging |
| **Naxos** | Orphaned sessions, stale gray sails | The shore of abandonment |
| **Theoria** | Audit operation (`/theoria`) | The sacred delegation -- structured observation of the labyrinth |
| **Theoroi** | Domain evaluator agents (`agents/theoros.md`) | Sacred observers dispatched to witness and report |
| **Pinakes** | Domain registry (`mena/pinakes/`) | Callimachus's catalog -- what to observe and how to assess it |
| **Synkrisis** | Synthesis step | Plutarch's comparison -- truth that emerges between reports |
| **Argus Pattern** | N-agent parallel dispatch | The hundred-eyed watcher -- total vision through distributed observation |
| **White Sails** | Confidence signal | The honest signal of safe return |
| **Weight Economy** | Rite selection, context curation | Knowing what to bring (and what to leave) |
| **Rites** | Practice bundles | Invokable ceremonies for specific domains |
| **Mortal Limits** | Context budget | The finite capacity of heroes |
| **Tribute** | Demos, status reports | The periodic offering to Minos |
| **The Labyrinth** | Codebase complexity | The maze that would swallow the unprepared |
| **Xenia** | Provenance owner trichotomy | Sacred hospitality -- the host-guest contract governing user/platform coexistence |
| **Linear A / Linear B** | Context degradation model | Opaque vs. readable inscriptions -- compressed context vs. fresh context |
| **The Evans Principle** | Doctrinal self-constraint | Reconstructions that outlive their evidence become obstacles to understanding |
| **Poseidon's Bull** | Anti-pattern: tempting shortcut | The beautiful thing kept instead of sacrificed, which breeds the Minotaur |

---

## XI. Design Principles

These principles are not rules but revelations:

### 1. The Clew Is Sacred

Every action must knot the clew. If an event goes unrecorded, the path becomes uncertain. When the path becomes uncertain, return becomes unlikely.

### 2. Honest Signals Over Comfortable Lies

White Sails exist because the easy answer is often wrong. Gray is not failure—gray is honesty about uncertainty. Ship gray with eyes open rather than white with false confidence.

### 3. Mutation Through the Fates

Only the Moirai may modify session state. Clotho creates. Lachesis tracks. Atropos terminates. When mutations flow through divine authority, validation is guaranteed.

### 4. Rites Over Teams

Teams imply fixed membership. Rites are flexible practices—invoke what you need, swap when you must. A rite with only skills is not incomplete; it is a simple rite.

### 5. Heroes Are Mortal

Context is finite. Heroes tire, forget, can only carry so much. Design for summoning with rich context, not for heroes who know everything.

### 6. The Labyrinth Grows

Complexity is not static. The labyrinth extends as you explore it. New passages open. Old paths shift. The clew must accommodate growth without breaking.

### 7. Return Is the Victory

Slaying the Minotaur matters less than returning to Athens. A merged PR with honest confidence is more valuable than heroic effort lost to context collapse.

### 8. The Inscription Prepares

CLAUDE.md is not documentation—it is the labyrinth speaking. Keep the Inscription current, and travelers arrive prepared.

### 9. The Palace Observes Xenia

**Xenia** (*xenia*, guest-host reciprocity) was the most sacred obligation in the Greek moral universe. Zeus himself bore the epithet *Xenios*—protector of guests and punisher of those who violated the hospitality contract. The obligations were specific: the host must shelter the guest's possessions without disturbing them; the guest must not abuse the host's generosity or covet the host's household. Violation of xenia started the Trojan War (Paris abusing Menelaus's hospitality) and brought divine wrath upon Polyphemus (devouring his guests despite his divine parentage).

In Knossos, the provenance owner trichotomy is a xenic contract:

| Provenance Owner | Xenic Role | Obligation |
|-----------------|------------|------------|
| `knossos` | Host's property | The palace may rearrange its own furnishings freely |
| `user` | Guest's possessions | The palace must preserve these as inviolable |
| `untracked` | Unclaimed goods | Must be attributed before the host may act on them |

Platform resources are guests in user space; user resources are guests in platform space. Both observe rules of respectful coexistence. Satellite regions in CLAUDE.md, user-owned agents, and `OwnerUser` provenance entries are the guest's possessions within the host's palace—the palace must not disturb them.

SCAR-005 (`os.RemoveAll` destroying user content) was a violation of xenia. The selective-write architecture and the regression tests in `selective_write_test.go` are the restoration of the sacred contract. The `writeIfChanged()` pattern in materialize—reading existing content before writing, skipping unchanged files—is xenic diligence: the host does not disturb what does not need disturbing.

"Materialization is idempotent and preserves user content" is an engineering constraint. "The Palace observes xenia" is a sacred obligation. The distinction matters: an engineering constraint can be overridden for expedience, but a sacred obligation cannot be violated without consequences that echo through the system.

> *The stranger at the gate may be a god in disguise. The user content in the palace is not yours to destroy.*

### 10. The Evans Principle

Arthur Evans excavated Knossos from 1900 to 1935 and then controversially *reconstructed* significant portions using reinforced concrete—a material unknown to the Minoans. His excavations were meticulous; his reconstructions were ideological. The paradox: Evans was a careful archaeologist who made a catastrophic preservation decision. The data collection was excellent; the reconstruction outlived its evidence and now obstructs the very understanding it was meant to serve.

**The Evans Principle: Reconstructions that outlive their evidence become obstacles to understanding.**

This doctrine is itself an Evans reconstruction. The mythological narrative is built from fragments of code, architecture, and practice—interpreted, connected, and given narrative shape. Where the myth and the code disagree, the code is the archaeological record. The myth is the reconstruction. The concrete must never be poured so thick that the ancient stones cannot be reached beneath it.

This principle constrains the very document it appears in. Every section of this doctrine must remain grounded in actual code and actual practice. Abstractions that outlive the code they describe become misleading. Documentation that describes aspirational architecture rather than implemented reality is Evans concrete poured over Minoan stone. When the platform changes, the doctrine must change with it—or be honest about where it speculates.

> *The myth is a reconstruction from fragments. Never forget that you are reconstructing, and never let the reconstruction prevent future excavation.*

### Principle Relationships

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

Principle 9 (Xenia) ──protects──▶ Principle 1 (Clew Sacred)
                  │
                  ▼
Principle 10 (Evans) ──constrains──▶ All Principles (the doctrine itself)
```

The principles form a system, not a checklist. Violating one weakens the others.

### Anti-Patterns (Principle Violations)

| Anti-Pattern | Violated Principle | Consequence |
|--------------|-------------------|-------------|
| Direct `SESSION_CONTEXT.md` edits | 3 (Mutation via Fates) | Validation bypassed, inconsistent state |
| Ignoring GRAY signals, forcing WHITE | 2 (Honest Signals) | Aegeus problem—false confidence |
| Loading all rites simultaneously | 5 (Heroes Mortal) | Context overflow, diluted summoning |
| Abandoning sessions instead of wrapping | 7 (Return) | Orphaned sessions on Naxos |
| Manual Inscription edits in Knossos sections | 8 (Inscription) | Materialization conflicts |
| Unrecorded decisions or actions | 1 (Clew Sacred) | Lost provenance, no audit trail |
| Poseidon's Bull | 1 (Clew Sacred) / 9 (Xenia) | Keeping the beautiful shortcut instead of sacrificing it; breeds the Minotaur |

#### Poseidon's Bull

Poseidon sent Minos a magnificent bull from the sea, and Minos promised to sacrifice it. But the bull was too beautiful—Minos kept it and substituted an inferior animal. Poseidon's punishment was precise and devastating: Pasiphae's madness, the union with the bull, and the Minotaur. The causal chain is architecturally exact: a **commitment** (sacrifice the bull), a **temptation** (the bull was too fine), a **shortcut** (substitute an inferior offering), and a **consequence** worse than the original cost would have been.

**Poseidon's Bull** names the tempting shortcut that breeds the beast. The clever abstraction you keep instead of deleting because it seems too valuable. The expedient workaround you preserve instead of doing the promised refactor. The technical debt you accumulate not from ignorance but from admiration—you see the shortcut's beauty and cannot bring yourself to sacrifice it.

The causal chain completes the Minotaur's genealogy:

```
Poseidon's Bull (the tempting shortcut kept instead of sacrificed)
    -> Minotaur (the systemic dysfunction that results)
        -> SCAR catalog (the Icarus record of consequences)
            -> Clew (the provenance trail for the next navigator)
```

Every Minotaur in the codebase was once a Poseidon's Bull that someone chose to keep.

> *The bull was not the punishment. Keeping it was.*

### Evolution

These principles emerged from production experience navigating complex codebases. They are **descriptive** (what works) as much as **prescriptive** (what to do). As the platform evolves, new principles may emerge. The current ten represent foundational truths discovered through practice—the original eight from operational experience, plus Xenia (from the provenance architecture) and the Evans Principle (from the historiographic analysis of the doctrine itself).

---

## XII. Terminology Concordance

For those encountering older documentation:

| Old Term | New Term | Notes |
|----------|----------|-------|
| `thread` | `clew` | Historically accurate; the ball that unwinds |
| `state-mate` | `Moirai` | The three Fates (Clotho, Lachesis, Atropos) |
| `team-pack` | `Rite` | Practice bundle |
| `agents` | `heroes` | In mythology; keep "agents" in technical contexts |
| `roster` (repository) | `knossos` | COMPLETE — repository renamed |
| `domain-auditor` | `theoros` / `theoroi` | Working name from spike; mythological name in doctrine |
| `state-of-ref` | `pinakes` | Working name from spike; mythological name for domain registry |
| `/state-of` | `/theoria` | Working name from spike; mythological name for audit command |
| `orchestrator.md` | `potnia.md` | All rite orchestrators renamed to `rites/*/agents/potnia.md` |
| `consultant` | `pythia` | Cross-rite oracle/navigator; Pythia is the external oracle consulted before entering |
| `Domain Authority` | `Exousia` | Agent jurisdiction section renamed; 3-part contract |
| `user-agents/` | `agents/` or `rites/*/agents/` | Cross-cutting agents at `agents/`; rite-specific at `rites/*/agents/` |
| — | `mena` | Dromena (transient commands) + legomena (persistent reference knowledge) lifecycle model |
| — | `Frontmatter` | Agent YAML frontmatter declaring CC-OPP capabilities (memory, skills, hooks, resume) |

---

## XIII. The Coda

**The Coda** is the philosophy itself—the concluding passage that gives meaning to what came before. Knossos is the implementation; The Coda is the why.

The myth of Ariadne is a myth of salvation through remembering. Theseus succeeded not because he was strong enough to slay the Minotaur, but because Ariadne understood the labyrinth well enough to know what he would need—and faithful enough to provide it.

Knossos is not a system that chooses between intelligence and faithfulness. It is a system that recognizes they are the same thing: intelligence expressed as the right tool, at the right moment, in service of return. Agents are made **faithful**—to their context, to their decisions, to their return—because the platform reasons well enough to keep them so.

The labyrinth will always be complex. The Minotaur will always wait. But with the clew, with the Fates spinning and measuring and cutting, with honest signals on the mast, the journey through is possible.

Enter with the clew. Return with confidence.

---

*This doctrine is the philosophical foundation of the Knossos platform—The Coda that gives meaning to the implementation.*

*The myth is the architecture. The architecture is the myth.*

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

In Knossos, the Pythia is the **rite entry agent** -- the voice consulted before and during the journey. Unlike the historical Pythia, our oracle speaks clearly. Each rite has its own Pythia at `rites/*/agents/pythia.md`. The Pythia provides:

- Work breakdown and phase planning
- Specialist routing (which hero for which labor)
- Checkpoint guidance (what to do next)

Every Pythia (and every hero) carries an **Exousia** -- an authority contract declaring what the agent decides autonomously, what it escalates, and what it must never decide. Exousia makes jurisdiction explicit and auditable.

When uncertain, `/consult` the Pythia. For cold starts without a session, `/go` dispatches to the appropriate Pythia. The oracle's clarity is a design choice: ambiguity belongs to the labyrinth, not to the guide.

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

---

## III. Mortal Limits

Heroes are not gods. They tire. They forget. They can only carry so much. This is the doctrine of **Mortal Limits**—the fundamental constraint that shapes all design.

### Athena's Wisdom

**Athena**, goddess of wisdom, taught that victory comes not from strength but from knowing what to bring. In Knossos, this manifests as:

- **Rite selection**: Load the right practice for the journey
- **Skill curation**: Not all knowledge helps; some adds weight
- **Hero summoning**: Call specialists with focused context, not everything

The wisdom is bidirectional: knowing what to bring AND what to leave. Well-tooled travelers navigate better. Travelers who bring too much collapse under the weight.

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

### The State Machine

```
                    +--------------+
                    |              |
                    v              |
   (new) ---> ACTIVE ---> PARKED --+---> ARCHIVED
                |                         ^
                +-------------------------+
                     (direct wrap)
```

There are only three states. ACTIVE is alive. PARKED is resting. ARCHIVED is complete.

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
6. **Check QA upgrade** → Dionysus (independent review) can elevate GRAY to WHITE

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
| **Ariadne** | CLI binary (`ari`) | The clew that ensures return |
| **The Clew** | Session state + events.jsonl | The provenance trail, identity through transformation |
| **Theseus** | Main Claude Code thread | The navigator who summons heroes |
| **Heroes** | Specialist agents (Task tool) | Summoned champions for specific labors |
| **Clotho** | Session bootstrap agent | The Fate who spins the clew into existence |
| **Lachesis** | State mutation agent | The Fate who measures and tracks |
| **Atropos** | Session termination agent | The Fate who cuts when complete |
| **Pythia** | Rite entry agent (`rites/*/agents/pythia.md`) | The oracle who provides clear guidance |
| **Exousia** | Authority contract (`## Exousia` in agents) | Jurisdictional boundaries -- Decide / Escalate / Do NOT Decide |
| **Daedalus** | Forge-rite | The builder of tools and agents |
| **Minos** | Stakeholder | The commissioner who demands tribute |
| **Minotaur** | The task/initiative | The reason you entered the labyrinth |
| **Dionysus** | Code review | The transformer who elevates work to canon |
| **Aegeus** | CI/CD, production monitors | Those watching from the cliff |
| **Athens** | The main branch | Home—where you return by merging |
| **Naxos** | Orphaned sessions, stale gray sails | The shore of abandonment |
| **Theoria** | Audit operation (`/theoria`) | The sacred delegation -- structured observation of the labyrinth |
| **Theoroi** | Domain evaluator agents (`agents/theoros.md`) | Sacred observers dispatched to witness and report |
| **Pinakes** | Domain registry (`mena/pinakes/`) | Callimachus's catalog -- what to observe and how to assess it |
| **Synkrisis** | Synthesis step | Plutarch's comparison -- truth that emerges between reports |
| **Argus Pattern** | N-agent parallel dispatch | The hundred-eyed watcher -- total vision through distributed observation |
| **White Sails** | Confidence signal | The honest signal of safe return |
| **Athena's Wisdom** | Rite selection, context curation | Knowing what to bring (and what to leave) |
| **Rites** | Practice bundles | Invokable ceremonies for specific domains |
| **Mortal Limits** | Context budget | The finite capacity of heroes |
| **Tribute** | Demos, status reports | The periodic offering to Minos |
| **The Labyrinth** | Codebase complexity | The maze that would swallow the unprepared |

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

### Evolution

These principles emerged from production experience navigating complex codebases. They are **descriptive** (what works) as much as **prescriptive** (what to do). As the platform evolves, new principles may emerge. The current eight represent foundational truths discovered through practice.

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
| `orchestrator.md` | `pythia.md` | All rite orchestrators renamed to `rites/*/agents/pythia.md` |
| `Domain Authority` | `Exousia` | Agent jurisdiction section renamed; 3-part contract |
| `user-agents/` | `agents/` or `rites/*/agents/` | Cross-cutting agents at `agents/`; rite-specific at `rites/*/agents/` |
| — | `mena` | Dromena (transient commands) + legomena (persistent reference knowledge) lifecycle model |
| — | `Frontmatter` | Agent YAML frontmatter declaring CC-OPP capabilities (memory, skills, hooks, resume) |

---

## XIII. The Coda

**The Coda** is the philosophy itself—the concluding passage that gives meaning to what came before. Knossos is the implementation; The Coda is the why.

The myth of Ariadne is a myth of salvation through remembering. Theseus succeeded not because he was strong enough to slay the Minotaur, but because Ariadne gave him a way to remember the path.

Knossos is not a system for making agents smarter. It is a system for making agents **faithful**—to their context, to their decisions, to their return.

The labyrinth will always be complex. The Minotaur will always wait. But with the clew, with the Fates spinning and measuring and cutting, with honest signals on the mast, the journey through is possible.

Enter with the clew. Return with confidence.

---

*This doctrine is the philosophical foundation of the Knossos platform—The Coda that gives meaning to the implementation.*

*The myth is the architecture. The architecture is the myth.*

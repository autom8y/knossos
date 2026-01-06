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

### The Platform: Knossos

**Knossos** is the labyrinth itself—the palace-complex where the Minotaur dwells. In our architecture, Knossos is the platform: the repository, the configuration, the accumulated complexity of an evolving system.

The repository currently named `roster` will become `knossos` when it proves itself capable of self-hosting. Until then, **roster/.claude/ IS Knossos**. The name is certain; only the renaming awaits its conditions.

### The Inscription

At the entrance to every labyrinth, there are words carved in stone. **CLAUDE.md** is the Inscription—the labyrinth speaking to those who would enter. It tells Theseus:

- What heroes may be summoned
- What rites are available
- What customs govern this place
- How to navigate without becoming lost

The Inscription is not static. When rites change, when heroes join or depart, the labyrinth updates its entrance. The SessionStart hook reads the Inscription and prepares the traveler.

> The labyrinth does not hide its nature. It declares it at the gate.

### The Clew: Ariadne

**Ariadne** gave Theseus the gift that saved him—not a weapon, but a clew (a ball of thread). The clew did not kill the Minotaur. The clew ensured return.

In Knossos, Ariadne is the CLI (`ari`)—the clew implementation that navigates complexity and guarantees return. Ariadne provides:

- **Session management**: The clew anchors at creation and is cut at wrap
- **Event recording**: Every step knotted into the clew (events.jsonl)
- **Quality gates**: The clew knows when you've reached the exit safely
- **Recovery paths**: If the clew tangles, it can be unwound

Ariadne is not intelligent. Ariadne is faithful.

### The Navigator: Theseus

**Theseus** enters the labyrinth to accomplish a task—to slay the Minotaur, to ship the feature, to close the ticket. In Knossos, the main Claude Code thread is Theseus: the agentic intelligence that makes decisions, takes actions, creates artifacts.

Theseus does not travel alone. He summons **heroes**—specialist agents invoked via the Task tool to lend their strength. These heroes (architect, engineer, adversary) arrive for specific labors and depart when done. They did not walk the whole labyrinth; they were summoned mid-journey.

This is a feature, not a limitation. Heroes summoned with rich context (a well-maintained clew) perform better than heroes summoned blind. **Better clew → better summoning → better heroes.**

### The Fates: Moirai

The **Moirai**—Clotho, Lachesis, and Atropos—are the three Fates who govern the thread of life. In Knossos, they are three distinct agents, each activated by specific events:

| Fate | Activates On | Function |
|------|--------------|----------|
| **Clotho** | `session_start` | Spins the clew into existence—bootstrap, initialization, first breath |
| **Lachesis** | State mutations | Measures the allotment—tracking, token accounting, phase transitions |
| **Atropos** | `session_end`, `wrap` | Cuts when complete—termination, archival, final reckoning |

> Ariadne gave Theseus the clew as a gift. The Moirai are who she borrowed it from.

**Relationship to Ariadne**: Ariadne is the survival architecture—the system ensuring deterministic return. The Moirai are the primordial force that makes clews exist at all. Ariadne is the princess with the idea; the Moirai are the divine mechanism that spins reality into being.

### The Builder: Daedalus

**Daedalus** built the labyrinth. He also built wings to escape it. In Knossos, Daedalus is the **forge-rite**—the practice of building tools, agents, and the platform itself.

Designed complexity is Daedalus's gift. The labyrinth contains AND protects. Good architecture is intentional maze—walls placed to guide, to shelter, to channel. Not all complexity is debt; some complexity is craft.

When you need to build new heroes, new rites, new mechanisms—you invoke Daedalus.

### The Task: Minotaur

The **Minotaur** is the task—the complexity that must be confronted, understood, and ultimately resolved. The Minotaur is not evil; it is simply the reason you entered the labyrinth.

Some Minotaurs are small (a bug fix). Some are immense (a platform migration). The clew cares not for the Minotaur's size—only that Theseus returns.

### The Commissioner: Minos

**Minos** commissioned the labyrinth and demanded tribute. In Knossos, Minos is the **stakeholder**—the one who creates the conditions, demands results, and receives the tribute.

The tribute? **Status reports and demos.** The periodic demonstration that proves the labyrinth is being navigated, that progress is made, that the Minotaur will eventually fall.

### The Oracle: Pythia

Before every great journey, Greeks consulted the **Pythia** at Delphi. In Knossos, the Pythia is the **orchestrator**—the voice consulted before and during the journey.

Unlike the historical Pythia, our oracle speaks clearly. The Pythia provides:
- Work breakdown and phase planning
- Specialist routing (which hero for which labor)
- Checkpoint guidance (what to do next)

When uncertain, `/consult` the Pythia.

### The Watchers: Aegeus

**Aegeus**, King of Athens, watched from the cliff for his son's return. The agreement: white sails if Theseus lives, black sails if he dies. Theseus forgot to change the sails. Aegeus, seeing black, threw himself into the sea.

The Aegeus problem is **false confidence**. Tests pass but edge cases lurk. Builds complete but integration points fracture. The watchers trust the signal and are destroyed by its lie.

**White Sails** solves this.

### The Destination: Athens

**Athens** is home—where Theseus returns after the journey. In Knossos, Athens is **the main branch**. You haven't returned until you've merged. The PR is the ship; the merge is the homecoming.

### The Transformer: Dionysus

After Theseus abandoned Ariadne on Naxos, **Dionysus** found her and made her divine. He transformed her abandonment into elevation.

In Knossos, Dionysus is **code review**—the transformative process that elevates raw work into merged truth. Dionysus takes what was created in isolation and blesses it for union with the canon.

### The Shore of Abandonment: Naxos

**Naxos** is where Theseus left Ariadne—abandoned after she saved him. In Knossos, Naxos represents:

- **Orphaned sessions**: Created but never wrapped, abandoned by their Theseus
- **Stale gray sails**: Work completed but never validated, confidence never earned

Sessions left too long on Naxos accumulate. They represent unfinished business, broken promises, technical debt of the workflow itself.

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

A **Rite** is an invokable practice—a bundle of knowledge, practitioners, and procedures for a specific domain. Where the older tongue spoke of "teams" and "team packs," we now speak of Rites.

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
| **10x-dev-rite** | Full development lifecycle | analyst, architect, engineer, adversary |
| **forge-rite** | Agent and tool creation (Daedalus's domain) | architect, prompt-engineer |
| **documentation-rite** | Knowledge crystallization | writer, editor |
| **hygiene-rite** | Code quality maintenance | reviewer, refactorer |
| **debt-triage-rite** | Technical debt remediation | archaeologist, surgeon |
| **security-rite** | Threat modeling and compliance | threat-modeler, auditor |
| **sre-rite** | Operations and reliability | operator, incident-commander |
| **intelligence-rite** | Research and synthesis | researcher, synthesizer |
| **exploration-rite** | Spikes and prototypes | explorer, evaluator |
| **strategy-rite** | Business analysis | strategist, analyst |
| **ecosystem-rite** | Platform infrastructure | integrator, validator |

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
| **Pythia** | Orchestrator | The oracle who provides clear guidance |
| **Daedalus** | Forge-rite | The builder of tools and agents |
| **Minos** | Stakeholder | The commissioner who demands tribute |
| **Minotaur** | The task/initiative | The reason you entered the labyrinth |
| **Dionysus** | Code review | The transformer who elevates work to canon |
| **Aegeus** | CI/CD, production monitors | Those watching from the cliff |
| **Athens** | The main branch | Home—where you return by merging |
| **Naxos** | Orphaned sessions, stale gray sails | The shore of abandonment |
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

---

## XII. Terminology Concordance

For those encountering older documentation:

| Old Term | New Term | Notes |
|----------|----------|-------|
| `thread` | `clew` | Historically accurate; the ball that unwinds |
| `state-mate` | `Moirai` | The three Fates (Clotho, Lachesis, Atropos) |
| `team-pack` | `Rite` | Practice bundle |
| `ACTIVE_RITE` | `ACTIVE_RITE` | Current practice file |
| `agents` | `heroes` | In mythology; keep "agents" in technical contexts |
| `roster` (repository) | `knossos` | Platform name (pending rename) |

---

## XIII. The Coda

**The Coda** is the philosophy itself—the concluding passage that gives meaning to what came before. Knossos is the implementation; The Coda is the why.

The myth of Ariadne is a myth of salvation through remembering. Theseus succeeded not because he was strong enough to slay the Minotaur, but because Ariadne gave him a way to remember the path.

Knossos is not a system for making agents smarter. It is a system for making agents **faithful**—to their context, to their decisions, to their return.

The labyrinth will always be complex. The Minotaur will always wait. But with the clew, with the Fates spinning and measuring and cutting, with honest signals on the mast, the journey through is possible.

Enter with the clew. Return with confidence.

---

## XIV. Implementation Drift Registry

This section documents known divergences between the doctrine and the current implementation.

### Terminology Not Yet Migrated

| Doctrinal Term | Current Implementation | Status |
|----------------|------------------------|--------|
| `clew` | `thread` (in code) | Documentation uses clew; code uses thread |
| `Moirai` (3 agents) | `state-mate` (1 agent) | moirai.md created; split pending |
| `Rite` | `team-pack` | Doctrine complete; rename pending |
| `ACTIVE_RITE` | `ACTIVE_RITE` | COMPLETE |
| `Pythia` | `orchestrator` | Name alignment pending |
| `heroes` | `agents` | Mythology uses heroes; technical docs use agents |

### Concepts Documented but Not Fully Implemented

| Concept | Status | Gap |
|---------|--------|-----|
| Three separate Moirai | Pending | Event-driven activation not yet implemented |
| The Inscription seeding | Partial | CLAUDE.md exists but dynamic seeding is basic |
| Naxos cleanup | Not implemented | Orphaned session detection missing |
| Dionysus integration | Partial | Code review exists but not mythologically named |

### Prioritized Alignment Work

1. **High Priority**: Split Moirai into event-driven agents
2. **High Priority**: Enhance CLAUDE.md seeding (The Inscription)
3. **Medium Priority**: Implement Naxos detection (orphaned sessions)
4. **Low Priority**: Rename files from thread→clew in codebase

---

*This doctrine is the philosophical foundation of the Knossos platform—The Coda that gives meaning to the implementation.*

*The myth is the architecture. The architecture is the myth.*

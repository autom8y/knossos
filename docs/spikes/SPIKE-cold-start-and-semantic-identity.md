# Spike: Cold Start UX and Semantic Identity

> **Date**: 2026-02-10
> **Status**: Findings Complete
> **Decision**: Pending

---

## The Question

Two intertwined questions that emerged from the theoria build sprint:

1. How should a user cold-start into productive work without mythology knowledge?
2. Is the mythological naming system load-bearing architecture or decorative ceremony?

These look like separate problems. They are not. The cold-start is hostile *because* the system assumes mythological fluency. The mythology is questioned *because* it gates the cold-start. The spike treats them as one problem.

## Context

The theoria sprint required dispatching 5 context-engineer agents to audit the framework's own primitives. In the process of building that audit infrastructure, a harder question surfaced: **the framework is beautiful inside, but the front door is locked and the key is written in Ancient Greek.**

Three parallel agents investigated the cold-start problem (UX gap analysis, naming/identity design, behavioral specification). A fourth agent conducted a deep adversarial semantic audit of every mythological term. A fifth counter-audited the fourth, correcting a fundamental category error. Together they produced a unified picture.

---

## The Red Pill

The cold-start path is 11 steps, 2-5 minutes, and 8+ knowledge prerequisites before a user writes their first meaningful prompt:

1. Remember/find the correct worktree directory (timestamp-based paths)
2. `cd` + `claude`
3. Read CLAUDE.md inscription (jargon-heavy)
4. Discover session state (`/sessions` or `ari session list`)
5. Choose which session to resume
6. Resume session (`/continue`)
7. Verify rite alignment
8. Switch rite if mismatched
9. Read sprint context
10. Choose execution mode
11. Formulate first meaningful prompt

*-- Expert 1*

The mythological naming system operates on three tiers. Only one is load-bearing at the Claude execution layer:

| Tier | Count | Examples | Claude Visibility |
|------|-------|----------|-------------------|
| Operational | ~7 | theoros, dromena, legomena, rite, moirai, inscription, sails | **In prompts. IS the protocol.** |
| CLI/Infrastructure | ~5 | naxos, clewcontract, moirai (lock), sails (Go pkg) | Go code only. Never reaches Claude. |
| Doctrinal | 8+ | Theseus, Pythia, Daedalus, Minos, Minotaur, Athens, Dionysus, Aegeus | Doctrine docs only. Invisible to Claude. |

*-- Expert 4, validated by Expert 5*

Implementation drift is actively eroding protocol coherence:

- **Pythia/Orchestrator**: Doctrine names the orchestrator "Pythia." Code calls it `orchestrator`. Claude sees `orchestrator`.
- **Moirai 3-vs-1**: Doctrine describes three agents (Clotho, Lachesis, Atropos). Code has one `moirai.md`. The three-fate split is doctrinal fiction.
- **Clew/Thread**: Doctrine names it "clew." Code uses `clewcontract`. Users never type either word.

*-- Expert 4*

The CC Primitives table in CLAUDE.md costs ~70 tokens per session telling Claude that slash commands are called "dromena" and skills are called "legomena" -- information Claude needs to parse zero prompts correctly, because dispatch uses the CC-native names.

*-- Expert 4*

The most important coordination mechanism in the system -- "You decide / You escalate / You do NOT decide" -- has no name. The thing that actually governs agent behavior is unnamed. The things that decorate agent behavior have elaborate mythological identities.

*-- Expert 4*

The framework was designed inside-out: "How should agents coordinate?" The entry experience needs to be designed outside-in: "I just sat down. What do I do?"

*-- Expert 1*

---

## The Blue Pill

The comfortable narrative:

- "The mythology is internally consistent and beautiful." It is. The Coda reads like architecture poetry. Every term has a genealogy.
- "Users will learn the terms over time." They will -- if they stay long enough. The question is whether the cold-start filters out everyone who would have stayed.
- "The doctrine serves human understanding." It does. Developers reading the Coda understand *why* sessions exist, *why* agents are summoned, *why* the clew matters. This is real value.
- "The naming system is the design language." True. "To rename is to re-architect." The doctrine says so explicitly.

What this costs:

- Every new session starts with a user who must already know the vocabulary to use the vocabulary
- Terms like "theoros" appear in prompts that Claude executes, while terms like "Theseus" appear only in docs that Claude never reads -- but both are presented with equal weight in the doctrine
- Implementation drift accumulates silently because the mythology gives the *feeling* of coherence even as the code diverges
- The 11-step cold-start is the tax paid for ceremony over ergonomics

---

## What's Actually True (Between the Pills)

Expert 5 identified the category error that unlocks the whole problem:

> "Expert 4 committed a fundamental error by applying traditional software thinking to an LLM-native system. The system has TWO execution environments: Go runtime (where naming is decorative) and Claude runtime (where naming IS the program)."

This reframes everything. The question is not "is the mythology load-bearing?" The question is "load-bearing *where*?"

**Naming IS implementation in the Claude runtime.** When an agent prompt says `subagent_type="theoros"`, mythology is a literal function parameter. When CLAUDE.md says "Dromena have side effects and are user-controlled. Legomena are reference knowledge Claude loads autonomously," that sentence IS the dispatch protocol. Prompt text is executable code. Expert 5 is right: you cannot evaluate these terms atomically. They form an interlocking vocabulary -- a shared protocol that is emergent from the set, not reducible to individual terms.

**But only Tier 1 terms reach the Claude runtime.** Theseus, Pythia, Daedalus, Minos, Athens, Dionysus, Aegeus, Minotaur -- none of these appear in agent prompts or skill content that Claude executes. They live in doctrine, glossary, and human-facing documentation. They serve developers, not the model. And that is *fine* -- design communication is a legitimate function. The error is treating all tiers as equivalent.

**The problem is not the mythology. The problem is drift between tiers.** When doctrine says "Pythia" and code says "orchestrator," the mythology stops being a reliable map. When doctrine describes three Moirai agents and code ships one, the naming system actively misleads. When the most critical coordination pattern ("you decide / you escalate / you do NOT decide") has no name while decorative concepts have elaborate ones, the naming system has inverted its priorities.

**The cold-start problem is a *symptom* of the inside-out design, not a separate issue.** The inscription (CLAUDE.md) speaks in operational-tier vocabulary because it addresses Claude. But the user reads it too, and encounters terms they have no reason to know. The fix is not to strip the mythology -- it is to build a plain-English entry point that uses the operational tier invisibly.

---

## The `/go` Proposal

Expert 2 evaluated 7 candidates (`/go`, `/work`, `/in`, `/ready`, `/run`, `/up`, `/hey`) and recommended `/go`:

- **Two characters, one syllable.** Universally understood as "proceed."
- **Zero mythology.** No learning prerequisite.
- **Lifecycle symmetry**: `/go` (enter) -- work -- `/park` (pause) -- `/go` (resume) -- `/wrap` (complete).
- **Subsumes the decision** between `/start`, `/continue`, and "check status."

### The Six Scenarios

| # | State | Action | Target |
|---|-------|--------|--------|
| 1 | ALREADY_ACTIVE | Show status + next step | ~3s |
| 2 | RESUME_PARKED | Auto-resume, show sprint context | ~8s |
| 3 | NEW_WORK (user provides intent) | Route to rite, create session | ~12s |
| 4 | RESUME_ORPHANED (WIP exists, no session) | Ask what to do | 1 question max |
| 5 | CROSS_WORKTREE | Offer to switch | 1 question max |
| 6 | ORIENTATION (nothing at all) | Dashboard + options | ~5s |

*-- Expert 2*

### The Autopark-Go Loop

Expert 2's key insight: combine the Stop hook (autopark) with `/go` (auto-resume) to create a zero-thought cold start:

```
User closes Claude Code
  --> Stop hook fires
  --> Session auto-parked

User opens Claude Code
  --> Types /go
  --> Detects parked session
  --> Auto-resumes
  --> Shows sprint context
  --> Working in 8 seconds
```

### Design Principles

Expert 3 specified the behavioral contract:

- **One question maximum per invocation.** If the state is unambiguous, act. If ambiguous, ask exactly one question and act on the answer.
- **Dispatch, don't execute.** `/go` routes to existing infrastructure (`/start`, `/continue`, `/consult`, rite discovery). It does not replace them.
- **Read everything, ask almost nothing.** Collect all state (6 parallel `ari` CLI calls in <3s), then decide.
- **Does NOT add mythology.** Uses the operational-tier vocabulary that already exists, invisibly.

### Data Collection (Expert 3 spec)

On invocation, `/go` executes in parallel:

1. `ari session status` -- current session state
2. `ari session list --limit=5` -- recent sessions
3. `ari worktree status` -- current worktree state
4. `ari worktree list` -- all worktrees
5. `ari sync --status` -- rite alignment
6. Check for `.wip/` artifacts -- orphaned work

Total collection time: <3 seconds. Decision tree executes on the collected state.

---

## Semantic Identity Recommendations

### Strengthen

These terms ARE load-bearing protocol. Invest in their precision:

| Term | Function | Action |
|------|----------|--------|
| **theoros** / **theoroi** | Agent type distinction (observer vs. actor) | Ensure all agent prompts use consistently |
| **dromena** / **legomena** | File convention that drives materialization | Document the behavioral contract explicitly |
| **rite** | Workflow boundary that scopes everything | Already strong. Maintain. |
| **inscription** | CLAUDE.md as protocol declaration | Already strong. Maintain. |
| **moirai** | Session lifecycle authority | Resolve the 1-vs-3 drift (see below) |

### Resolve

Implementation drift that is actively eroding coherence:

| Drift | Doctrine Says | Code Does | Recommendation |
|-------|--------------|-----------|----------------|
| **Pythia/Orchestrator** | Orchestrator is "Pythia, the Oracle" | Agent file is `orchestrator.md`, prompts say "orchestrator" | Pick one. If "orchestrator" is what Claude sees, either rename in doctrine or rename the file. |
| **Moirai 1-vs-3** | Three agents (Clotho, Lachesis, Atropos) | One `moirai.md` agent | Either implement the three-fate split or update doctrine to reflect single-agent reality. The gap erodes trust in the naming system. |
| **Clew/Thread** | "The Clew is Sacred" | `clewcontract` package, but users never interact with it by name | Low priority. Internal naming is fine. But stop capitalizing "Clew" in user-facing text as if it is a command. |

### Clarify

Draw an explicit line between operational and doctrinal tiers:

- **Operational tier** (appears in prompts, drives behavior): theoros, dromena, legomena, rite, moirai, inscription, sails. These are protocol. Treat changes as breaking changes.
- **Doctrinal tier** (appears in Coda and glossary, serves human understanding): Theseus, Pythia, Daedalus, Minos, Minotaur, Athens, Dionysus, Aegeus, Athena. These are design communication. Valuable, but changes are documentation changes, not protocol changes.
- **Infrastructure tier** (appears in Go code, invisible to Claude): naxos, clewcontract, sails (Go package). Standard software naming. No mythology tax applies.

This three-tier distinction should be documented in the glossary or doctrine so contributors know which tier a term belongs to and what the change implications are.

### Add

Missing terminology for patterns that deserve names:

| Pattern | Current State | Recommendation |
|---------|--------------|----------------|
| "You decide / You escalate / You do NOT decide" | Unnamed. THE most important coordination mechanism. | Name it. This is Tier 1 -- it appears in every agent prompt. Expert 4 is right that the unnamed critical pattern is a worse problem than the named decorative ones. |
| Source vs. Projection | Described in glossary but not named as a principle | Consider elevating to a named doctrine principle. The SOURCE/PROJECTION distinction is fundamental to every sync and materialize operation. |
| The autopark-go loop | New concept from this spike | Name it if it ships. It is a lifecycle pattern, not a one-off feature. |

### Compress

Token waste in always-on context:

| Item | Current Cost | Recommendation |
|------|-------------|----------------|
| CC Primitives table in CLAUDE.md | ~70 tokens/session | Compress. Claude does not need to be told that slash commands exist. The mapping to Knossos names can live in the `lexicon` skill, loaded on demand. |
| Full agent list in CLAUDE.md | ~120 tokens/session | Already gated by rite. Confirm only active rite agents are materialized. |
| Execution Mode table | ~80 tokens/session | Evaluate whether `/go` can subsume mode selection, removing the need to present modes upfront. |

---

## The Spectrum

Expert 5's key contribution -- mythology operates on a spectrum, not a binary:

```
LOAD-BEARING                                                    DECORATIVE
(Claude executes)                                          (humans read)
     |                                                          |
     v                                                          v

  theoros --- dromena/legomena --- rite --- moirai --- inscription
                                                |
                                              sails --- clew --- naxos
                                                                  |
                                                        Pythia --- Theseus
                                                                      |
                                                          Daedalus --- Minos
                                                                        |
                                                             Dionysus --- Athens
```

Every term on the left side of this spectrum is a protocol element. Changes are breaking changes. Every term on the right side is design communication. Changes are documentation. The problems arise in the middle, where terms like "moirai" and "sails" straddle tiers with different implementations at each level.

---

## Follow-Up Actions

### If you take the Red Pill

1. **Build `/go`** as a dromena. Expert 3's behavioral spec is implementation-ready. Start with scenarios 1 (ALREADY_ACTIVE) and 2 (RESUME_PARKED) -- they cover 80% of cold starts.
2. **Implement autopark** in the Stop hook. This is the other half of the zero-thought loop.
3. **Resolve Pythia/Orchestrator drift.** Pick one name. Update whichever side is wrong.
4. **Resolve Moirai 1-vs-3.** Either build the three-fate split or update doctrine. The gap is actively misleading.
5. **Name the coordination pattern.** "You decide / You escalate / You do NOT decide" deserves a Tier 1 name.
6. **Compress the inscription.** Remove the CC Primitives table from CLAUDE.md. Move to `lexicon` skill.
7. **Document the three tiers** in the glossary with explicit change-impact guidance.

### If you take the Blue Pill

1. Keep the current cold-start path.
2. Write an onboarding guide that teaches the mythology.
3. Accept that the 11-step, 2-5 minute cold start is the price of ceremony.
4. Hope that implementation drift self-corrects.

### If you take what's actually true

1. Do items 1-7 from the Red Pill list.
2. Keep the doctrine. The Coda is good writing and good design communication. It serves a real purpose for human contributors.
3. But stop pretending all tiers are equal. Tier 1 is protocol. Tier 3 is literature. Both are valuable. They are not the same thing.

---

## Methodology

- 5 context-engineer agents dispatched: 3 parallel initial (cold-start UX, naming/identity, behavioral spec) + 1 sequential adversarial audit + 1 counter-audit
- Source files only -- never materialized `.claude/` projections
- Adversarial framing: each subsequent agent challenged the previous
- ~25 mythological terms traced from source definition through materialization to behavioral impact
- Cold-start path mapped empirically (11 steps, timed)
- 7 command name candidates evaluated against 6 scenarios
- Full behavioral specification with timing targets and token budgets
- Three-tier model validated by tracing every term to its actual appearance (or non-appearance) in Claude's execution context

---
last_verified: 2026-02-26
---

# Mythology Concordance

> The myth is the architecture. Greek thought is the language.

This concordance maps Greek mythology to the **SOURCE implementation** of Knossos—the canonical Go code, rite definitions, and platform source that **generates** the materialized `.claude/` projections.

**Critical Distinction:**
- **SOURCE** = Knossos repository (versioned, canonical, what Knossos IS)
- **PROJECTION** = `.claude/` directories (gitignored, materialized by `ari sync materialize`)

---

## Naming Provenance

The platform's naming draws from multiple wells. Acknowledging this honestly — rather than pretending a single-myth purity — is itself a design decision.

> The core architectural metaphor is the Knossos myth and its palace — Ariadne, Theseus, the Minotaur, the labyrinth itself, the Potnia who presided within, and the Pythia who was consulted before entering. The Potnia of the Labyrinth is attested at Knossos on Linear B tablet KN Gg(1) 702; Pythia is Delphic, but she occupies her correct mythological position as the external oracle consulted before a journey begins. Beyond the myth, architecture needs a language: the Moirai predate the Olympians; the mystery religions (dromena, legomena) may have Cretan roots; the theoria delegations, the Pinakes of Alexandria — these are borrowed from the wider Greek tradition because they describe their architectural functions with precision no Cretan alternative could match. The concordance marks this provenance honestly.

### Provenance Tiers

| Tier | Label | Criterion | Examples |
|------|-------|-----------|----------|
| **1** | Bronze Age Attestation | Name attested on Linear B tablets or Bronze Age archaeological evidence from Knossos itself | Potnia (KN Gg 702), Knossos, Ariadne (KN Gg 702 — as "Mistress of the Labyrinth"), the Labyrinth |
| **2** | Classical Source | Name drawn from Homer, Hesiod, Herodotus, Pausanias, Plutarch, or other classical literary/historical texts | Theseus, Minos, Daedalus, Icarus, Minotaur, Moirai, Pythia, Aegeus, Athens, Naxos, Dionysus, Argus |
| **3** | Hellenistic/Scholarly | Name drawn from post-classical scholarship, Hellenistic institutions, or Panhellenic practice documented in classical sources | Theoria, Theoroi, Pinakes, Synkrisis, Exousia |
| **4** | Functional Analogy | Name chosen for functional resonance with the architectural role, not mythological accuracy | Rite, Mena, Dromena, Legomena, Inscription, White Sails, Clew, Heroes |

---

## Core Components

| Myth | Implementation | Key Files | Provenance | Rationale |
|------|---------------|-----------|------------|-----------|
| **Knossos** (The Labyrinth) | The platform repository itself | Repository root, `internal/`, `rites/`, `cmd/ari/` | Tier 1 — Bronze Age site | The labyrinth is the palace, not the rooms within it. Per ADR-0009: the repository IS the platform; `.claude/` directories are projections. |
| **Ariadne** (The Clew) | CLI binary (`ari`) — survival architecture | `cmd/ari/`, `internal/session/`, `internal/hook/clewcontract/`, `internal/lock/`, `internal/search/`, `internal/suggest/`, `internal/perspective/` | Tier 1 — KN Gg 702 (as Mistress of the Labyrinth) | Faithful because intelligent. Ariadne was the strategist, not the thread. The clew was one instrument of her understanding; `ari` expresses both. |
| **Potnia** (Presiding Lady) | Rite entry agents | `rites/*/agents/potnia.md` | Tier 1 — Linear B tablet KN Gg 702: *da-pu₂-ri-to-jo po-ti-ni-ja* ("Potnia of the Labyrinth") | Each rite has its own Potnia. The presiding authority within the labyrinth, not the external oracle. Clear guidance, not cryptic prophecy. |
| **Pythia** (Cross-Rite Oracle) | Cross-rite oracle/navigator | `agents/pythia.md` | Tier 2 — Herodotus, Pausanias (Delphic oracle) | The external oracle consulted before entering the labyrinth. Cross-rite routing and navigation, not per-rite authority. Positionally correct: Pythia was consulted before the journey, not during it. |
| **The Clew/Thread** | Session state + event log + provenance trail | `internal/hook/clewcontract/`, `[session-dir]/events.jsonl` | Tier 4 — functional metaphor from the Theseus myth | The only source of truth when context degrades. Identity persists through transformation. |
| **Theseus** (Navigator) | Main Claude Code thread | N/A (the LLM agent itself) | Tier 2 — Homer, Plutarch, Bacchylides | Has agency but amnesia. The clew compensates. |
| **Heroes** | Specialist agents via Task tool | `rites/*/agents/`, `agents/` | Tier 4 — functional analogy | Summoned mid-journey with clew context. 75 agents across 14 rites. |
| **Moirai** (The Fates) | Session lifecycle agent (one agent, three aspects) | `agents/moirai.md`, `internal/session/` | Tier 2 — Hesiod *Theogony*, Plato *Republic* | Only authority to mutate session state. Clotho creates, Lachesis measures, Atropos cuts. Pre-Olympian primordial beings. |
| **The Inscription** | `CLAUDE.md` | `knossos/templates/CLAUDE.md.tpl`, `internal/inscription/` | Tier 4 — functional analogy (temple architecture) | The labyrinth speaks at entry. Knossos-managed sections regenerated by `ari sync inscription`. |
| **White Sails** | Confidence signal (WHITE/GRAY/BLACK) | `internal/sails/` | Tier 4 — functional metaphor from the Aegeus myth | Solves the Aegeus problem. Computed, never self-declared. |
| **Rites** | Manifest-driven practice bundles | `rites/`, `rites/*/manifest.yaml`, `internal/rite/` | Tier 4 — functional analogy (Greek religious ceremony) | Flexible compositions. 14 rites operational. Invoke (cheap) vs swap (expensive). |
| **Daedalus** (Builder) | The `forge-rite` | `rites/forge/` | Tier 2 — Homer *Iliad*, Ovid *Metamorphoses* | Designed complexity is intentional architecture — and the builder's own structures can imprison the builder. TENSION-001/TENSION-002 are Daedalean traps: soundly built, now confining. |
| **Icarus** (The Fall) | The SCAR catalog | `.know/scar-tissue.md` | Tier 2 — Ovid *Metamorphoses*, Diodorus Siculus | Ambitious changes that ignored the constraint surface. The wings were real; the constraints were ignored. |
| **Exousia** | Agent authority contract | `## Exousia` section in every agent `.md` | Tier 3 — Greek political vocabulary (Aristotle, NT usage) | Three-part: You Decide / You Escalate / You Do NOT Decide. |
| **Athens** | The `main` branch | Git branch | Tier 2 — Homer, Thucydides, Plutarch | Return = merged PR. |
| **Naxos** | Orphaned sessions | `internal/naxos/` | Tier 2 — Plutarch *Theseus*, Catullus | Sessions abandoned without wrapping. Detected by `ari naxos scan`. |
| **Dionysus** | Cross-session knowledge synthesis | `agents/dionysus.md`, `ari land` | Tier 2 — Homer, Hesiod, Euripides *Bacchae* | Transformation of the raw into the refined. On Naxos, Dionysus found abandoned Ariadne and made her divine; `ari land` finds abandoned session data and distills it into persistent wisdom at `.sos/land/`. |
| **Minos** | Stakeholders | `internal/tribute/` | Tier 2 — Homer *Odyssey*, Thucydides | Demands tribute (status reports and demos). |
| **Minotaur** | Accumulated technical debt / systemic dysfunction | `.know/scar-tissue.md`, `SESSION_CONTEXT.md` | Tier 2 — Plutarch, Ovid, Apollodorus | Born from shortcuts and broken promises. Not any individual task — the systemic condition that makes work harder than it should be. |
| **Aegeus** | CI/CD, production monitors | Conceptual | Tier 2 — Plutarch *Theseus* | Those watching from the cliff. The false-confidence problem. |
| **Theoria** | Audit operation (`/theoria`) | `/theoria` dromena, `mena/pinakes/`, `agents/theoros.md` | Tier 3 — Panhellenic practice (Herodotus, Thucydides) | Structured observation. Uses Argus Pattern for parallel dispatch. |
| **Theoroi** | Domain evaluator agents | `agents/theoros.md` | Tier 3 — Panhellenic practice (Herodotus, Thucydides) | Sacred observers — read-only witnesses. Singular: theoros. |
| **Pinakes** | Domain registry legomena | `mena/pinakes/` | Tier 3 — Callimachus of Cyrene, Hellenistic Alexandria | Callimachus's catalog — what to observe and how to assess. |
| **Synkrisis** | Synthesis step | Part of `/theoria` dromena | Tier 3 — Plutarch *Parallel Lives* | Plutarch's comparison — cross-domain patterns from individual reports. |
| **Argus Pattern** | N-agent parallel dispatch | Pattern (no single file) | Tier 2 — Apollodorus, Ovid (Argive myth) | One body (main thread), many eyes (agents). Reusable technique. |
| **Mena** | Lifecycle model (dromena + legomena) | `internal/mena/`, `rites/*/mena/` | Tier 4 — functional analogy (mystery religions, arguably Cretan origin) | Dromena (transient commands) and legomena (persistent reference). Context lifecycle distinguishes them. |
| **Dromena** | Transient commands (`.dro.md`) | `rites/*/mena/*.dro.md` | Tier 4 — mystery religions (Clement of Alexandria; Diodorus claims Cretan origin) | Execute and exit. User-invoked actions. |
| **Legomena** | Persistent reference (`.lego.md`) | `rites/*/mena/*.lego.md` | Tier 4 — mystery religions (Clement of Alexandria; Diodorus claims Cretan origin) | Stay in context. Consulted but never consumed. |

All paths relative to repository root.

---

## Materialization Flow

SOURCE (versioned, canonical) generates PROJECTION (gitignored, ephemeral):

| Source | Projection |
|--------|------------|
| `rites/` | `.claude/rites/` |
| `rites/*/agents/` + `agents/` | `.claude/agents/` |
| `rites/*/mena/` | `.claude/skills/` + `.claude/commands/` |
| `internal/hook/` (Go) + `rites/*/hooks/` | `.claude/hooks/` |
| `knossos/templates/` | `.claude/CLAUDE.md` (rendered) |
| `rites/*/manifest.yaml` | `.knossos/ACTIVE_RITE` |

**Command:** `ari sync materialize` reads SOURCE and writes PROJECTION. The labyrinth creates the rooms; the rooms are not the labyrinth.

---

**See Also:**
- [knossos-doctrine.md](knossos-doctrine.md) (complete mythological framework)
- [design-principles.md](design-principles.md) (architectural principles derived from mythology)
- [../reference/INDEX.md](../reference/INDEX.md) (navigation hub)

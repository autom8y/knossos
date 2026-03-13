---
last_verified: 2026-02-26
---

# Knossos Doctrine Compliance Status

> Current state of the Knossos platform as of February 2026.

---

## Platform Metrics

| Metric | Value |
|--------|-------|
| Go source lines | 105,609 |
| Go files | 340 (212 source + 128 test) |
| Go packages | 53 |
| CLI command families | 20 |
| CLI commands total | 84+ |
| Specialist agents | 75 across 14 rites |
| Active rites | 14 |
| ADRs documented | 27 |
| Skills (legomena) | 150 |
| Commands (dromena) | 84 |

---

## I. Doctrine Alignment

### Category-by-Category Status

| Category | Status | Notes |
|----------|--------|-------|
| **Clew Contract** | Complete | All 12 doctrine event types + 2 additions (command, context_switch) |
| **White Sails** | Complete | Three-color system, anti-gaming, complexity thresholds |
| **Session FSM** | Complete | ACTIVE/PARKED/ARCHIVED, formally verified (TLA+) |
| **Naxos Detection** | Complete | Full orphan scanning with configurable thresholds |
| **Inscription System** | Complete | Marker-based regeneration, backup/restore, section management |
| **Terminology** | Complete | team→rite, roster→knossos, orchestrator→potnia all migrated |
| **Rite Operations** | Complete | invoke/swap/release/pantheon operational across 14 rites |
| **Moirai Authority** | Complete | Unified agent with skill-based routing (beneficial simplification) |
| **Theoria/Pinakes** | Complete | Domain registry, theoros agent, `/theoria` dromena, synkrisis |
| **Hook Infrastructure** | Complete | SessionStart, PreToolUse, PostToolUse + agent-guard enforcement |
| **CC-OPP Uplift** | Complete | Memory (17 agents), Skills (68 agents), Hooks (10 agents), Resume (ecosystem) |
| **Handoff Protocol** | Partial | Events recorded, validation gates partial |
| **Cognitive Budget** | Partial | Infrastructure ready, active tracking incomplete |

### Capabilities Beyond Doctrine

| Capability | Status |
|------------|--------|
| Worktree system (parallel sessions) | Complete — 11 commands, full lifecycle |
| TLA+ formal verification | Complete — session FSM formally specified |
| Artifact registry with querying | Complete — full query API |
| Tribute generation | Complete — session summary system |
| Agent frontmatter (CC-OPP) | Complete — memory, skills, hooks, resume capabilities |
| Agent-guard hook enforcement | Complete — triple-layer enforcement (disallowedTools + hooks + tool restrictions) |
| Materialization pipeline | Complete — `ari sync materialize` with agent transforms |
| Provenance tracking | Complete — `internal/provenance/` |

---

## II. The Working System

### Ariadne CLI (`ari`) — 84+ commands across 20 families

```
ari
├── session (15)     # Full lifecycle management
├── rite (10)        # Practice operations
├── worktree (11)    # Parallel session isolation
├── sync (8)         # Artifact materialization
├── hook (11)        # Hook infrastructure + agent-guard
├── handoff (4)      # Agent transition protocol
├── inscription (5)  # context file management
├── artifact (4)     # Registry management
├── validate (3)     # Schema enforcement
├── manifest (4)     # Manifest operations
├── agent (3)        # Agent operations
├── initialize (2)   # Project initialization
├── migrate (2)      # Migration utilities
├── lint (2)         # Lint and validation
├── provenance (2)   # Provenance tracking
├── sails (1)        # Confidence computation
├── naxos (1)        # Orphan detection
├── tribute (1)      # Session reports
├── completion (4)   # Shell autocompletion
└── version (1)      # Version info
```

### Session Lifecycle — Formally Verified

Three states (ACTIVE, PARKED, ARCHIVED), five transitions enforced by FSM. TLA+ specification for concurrent access guarantees. Lock management with stale detection. Event sourcing via `events.jsonl`.

### White Sails Confidence System

Three colors (WHITE/GRAY/BLACK). Five proof types. Complexity-aware thresholds (70%/80%/90%). QA upgrade path. Cannot self-upgrade; modifiers only downgrade.

### Rite System — 14 Rites

10x-dev, arch, debt-triage, docs, ecosystem, forge, hygiene, intelligence, rnd, security, shared, slop-chop, sre, strategy.

### CC-OPP Agent Capability Uplift

- **Memory** (17 agents): 3-tier seeding, 150-line soft cap, self-curating
- **Skills** (68 agents): Frontmatter `skills:` field, ~3,500 token ceiling
- **Hooks** (10 agents): `ari hook agent-guard`, triple-layer enforcement
- **Resume** (ecosystem only): Throughline protocol for Potnia continuity

See [reference/agent-capabilities.md](../reference/agent-capabilities.md) for details.

### Theoria Audit Infrastructure

- **Pinakes**: Domain registry at `mena/pinakes/` with per-domain criteria
- **Theoros**: Domain evaluator agent at `agents/theoros.md`
- **Synkrisis**: Cross-domain synthesis producing "State of the {X}" reports
- **Argus Pattern**: N-agent parallel dispatch for distributed observation

---

## III. Known Gaps

### Honest Assessment

| Gap | Severity | Notes |
|-----|----------|-------|
| 5 CLI families underdocumented | Low | `agent`, `initialize`, `migrate`, `lint`, `provenance` — trust `ari <family> --help` |
| Cognitive budget active tracking | Medium | Infrastructure ready, active warning system incomplete |
| Handoff validation gates | Medium | Events recorded, pre-execution enforcement unclear |
| arch rite doc was missing | Fixed | Added `docs/doctrine/rites/arch.md` |
| slop-chop rite doc was missing | Fixed | Added `docs/doctrine/rites/slop-chop.md` |

### Beneficial Divergences

| Doctrine Says | Implementation Does | Why Better |
|---------------|-------------------|------------|
| Three separate Moirai agents | Unified agent with skill routing | Reduces context switching, maintains logical separation |
| — | Agent frontmatter capabilities | CC-OPP uplift enables memory, skills, hooks per agent |
| — | Materialization agent transforms | Automatic injection of capabilities during `ari sync materialize` |

---

## IV. Subsystem Architecture

See [reference/architecture-map.md](../reference/architecture-map.md) for the full subsystem table.

```
┌─────────────────────────────────────────────────────────────┐
│                        Ariadne CLI                          │
├─────────────────────────────────────────────────────────────┤
│  Session      Rite         Artifact      White Sails        │
│  Management   Operations   Registry      Confidence         │
│      │            │            │              │              │
│      └────────────┴────────────┴──────────────┘              │
│                         │                                    │
│              ┌──────────┴──────────┐                        │
│              │   Hook System       │                        │
│              │  (context, clew,    │                        │
│              │   writeguard,       │                        │
│              │   agent-guard)      │                        │
│              └──────────┬──────────┘                        │
│                         │                                    │
│              ┌──────────┴──────────┐                        │
│              │   Moirai Authority  │                        │
│              │  (state mutations)  │                        │
│              └──────────┬──────────┘                        │
│                         │                                    │
│              ┌──────────┴──────────┐                        │
│              │   Clew Contract     │                        │
│              │   (events.jsonl)    │                        │
│              └─────────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

---

*Enter with the clew. Return with confidence.*

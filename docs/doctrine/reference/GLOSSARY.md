---
last_verified: 2026-03-26
---

# Knossos Glossary

> Terminology reference for the Knossos platform.

## Critical Distinction

| Term | Definition |
|------|------------|
| **SOURCE** | The Knossos repository—versioned, canonical platform code |
| **PROJECTION** | channel directories—gitignored, materialized by `ari sync materialize` |

## Mythology Terms

### Apolysis
Agent dismissal — the operation that removes a previously summoned agent from your user-level harness configuration. Inverse of [Klesis](#klesis). Implemented as `ari agent dismiss <name>`.
- **Related**: Klesis, Katalogos, Summonable Heroes, Agent
- **Command**: `ari agent dismiss`

### Ariadne
The CLI binary (`ari`) that provides the clew—ensuring return through event recording and session management.
- **Related**: Clew, Theseus, Session
- **Source**: `cmd/ari/`

### Argus Pattern
N-agent parallel dispatch pattern. One main thread (body) launches multiple Agent tool agents (eyes) simultaneously for distributed observation. Named for Argus Panoptes, the hundred-eyed giant. A reusable technique, not a specific component; envisioned for tactical playbook swarms and similar N-agent operations.
- **Related**: Theoria, Theoroi, Agent Tool, Parallel Dispatch
- **Source**: Pattern (no single file)

### Athens
The `main` branch—destination of merged PRs. Return from the labyrinth is incomplete until work reaches Athens.
- **Related**: Pull Request, Merge, Ship
- **Source**: Git branch (not a file)

### CC-OPP
CC Operational Platform Properties — the capability uplift giving agents memory, skills, hooks, and resume. Declared via YAML frontmatter in agent `.md` files.
- **Related**: Agent Frontmatter, Memory, Skills, Hooks
- **Source**: `internal/agent/frontmatter.go`

### Channel
A projection target for materialization — either `.claude/` (Claude Code) or `.gemini/` (Gemini CLI). Each channel has its own `ChannelCompiler` that translates rite artifacts into the channel's native format. The `--channel` global flag selects which channel to target.
- **Related**: Materialization, Projection, Multi-Channel Architecture, ADR-0031
- **Source**: `internal/materialize/`, `internal/paths/channel.go`

### Clew
Session state + event log (`events.jsonl`) + provenance trail. The thread that unwinds through the labyrinth, providing a path back when context degrades.
- **Related**: Thread, Events, Session
- **Source**: `internal/hook/clewcontract/`

### Daedalus
The builder—represented by the `forge-rite` for creating agents, tools, and platform infrastructure. Designed complexity is intentional architecture.
- **Related**: Forge, Builder, Infrastructure
- **Source**: `rites/forge/`

### Dionysus
The transformer—cross-session knowledge synthesizer that transforms raw session data into refined persistent knowledge. On Naxos, Dionysus found the abandoned Ariadne and made her divine; in Knossos, the `ari land` pipeline finds abandoned session data and distills it into permanent cross-session wisdom in `.sos/land/`.
- **Related**: Land, Naxos, Knowledge Synthesis, Session Archives
- **Source**: `agents/dionysus.md`, `ari land`

### Dromena
Transient commands (`.dro.md` files materialized to `.channel/commands/`). Execute and exit. User-invoked actions. Part of the mena lifecycle model.
- **Related**: Legomena, Mena, Commands
- **Source**: `rites/*/mena/*.dro.md`

### Exousia
Authority contract defining an agent's jurisdictional boundaries. Every agent declares three subsections: You Decide (autonomous authority), You Escalate (requires Potnia or user input), You Do NOT Decide (hard boundaries). Named from the Greek exousia (authority, jurisdiction).
- **Related**: Potnia, Agent Contract, Jurisdiction
- **Source**: `## Exousia` section in every agent `.md` file

### Glint
A signal of an undocumented pattern found by [Myron](#myron) during feature discovery scans. Each glint identifies a capability or behavior present in the codebase but absent from documentation. Glints are collected into Myron's discovery report.
- **Related**: Myron, Scout, Feature Discovery
- **Source**: Myron agent output

### Heroes
Specialist agents invoked via Agent tool for specific labors, defined in rite manifests. Summoned mid-journey with clew context. Three tiers: Standing (always active), Rite (active with materialized rite), and Summonable Heroes (on-demand via `ari agent summon`).
- **Related**: Agents, Agent Tool, Rite, Summonable Heroes
- **Source**: `rites/[rite-name]/agents/`

### Inscription
The context file declaring available rites, agents, execution mode, and hooks. Words carved at the labyrinth entrance. Mentions Potnia for routing, Exousia for authority contracts, `/go` for cold-start entry, and the Fates for session lifecycle.
- **Related**: Context File, Knossos Sections, Potnia, Exousia, `/go`
- **Source**: `knossos/templates/CLAUDE.md.tpl`

### Legomena
Persistent reference knowledge (`.lego.md` files materialized to `.channel/skills/`). Stay in context. Consulted but never consumed. Part of the mena lifecycle model.
- **Related**: Dromena, Mena, Skills
- **Source**: `rites/*/mena/*.lego.md`

### Katalogos
The agent roster — the listing of standing, summoned, and available agents. Implemented as `ari agent roster`. Shows which Summonable Heroes are currently active and which are available.
- **Related**: Klesis, Apolysis, Summonable Heroes
- **Command**: `ari agent roster`

### Klesis
Agent summoning — the operation that materializes a Summonable Hero into your user-level harness configuration. Inverse of [Apolysis](#apolysis). Implemented as `ari agent summon <name>`.
- **Related**: Apolysis, Katalogos, Summonable Heroes
- **Command**: `ari agent summon`

### Knossos
The labyrinth—the Knossos repository itself (SOURCE), not the channel directories it generates (PROJECTION).
- **Related**: SOURCE, Platform, Repository
- **Source**: Repository root

### Metis
Standing context-engineering agent — the Titaness of strategic wisdom and craft intelligence. Metis provides strategic meta-level guidance on agent design, context engineering, and platform decisions. A standing agent (always active), not a summonable hero.
- **Related**: Pythia, Potnia, Standing Agents
- **Source**: Historical (no agent file — role absorbed by Myron)

### Minotaur
The initiative, feature, or task being pursued. The challenge at the labyrinth's heart—reason for the journey.
- **Related**: Initiative, Task, Goal
- **Source**: Defined in `SESSION_CONTEXT.md`

### Minos
Stakeholders who create initiatives and demand tribute (status reports and demos).
- **Related**: Tribute, Status Reports
- **Source**: `internal/tribute/`

### Myron
Feature discovery scout agent — performs wide-scan discovery of undocumented patterns and capabilities in the codebase. Produces [Glint](#glint) reports identifying features present in code but absent from documentation. Named for the sculptor who captured motion. Agent type: `scout`.
- **Related**: Glint, Scout, Feature Discovery, Summonable Heroes
- **Source**: `agents/myron.md` (or rite-scoped variant)

### Moirai
The Fates—centralized session lifecycle agent. Clotho spins (create), Lachesis measures (update), Atropos cuts (end).
- **Related**: Session Lifecycle, State Authority
- **Source**: `agents/moirai.md`

### Naxos
Shore of abandonment—orphaned sessions created but never wrapped. Detected for cleanup or resumption.
- **Related**: Orphaned Sessions, Cleanup
- **Source**: `internal/naxos/`

### Pinakes
The domain registry legomena for audit operations. Catalogs audit targets, evaluation criteria, grading rubrics, and report schemas. Named for Callimachus's catalog of the Library of Alexandria. Always qualified as "the Pinakes" to distinguish from manifest files.
- **Related**: Theoria, Theoroi, Legomena, Domain Registry
- **Source**: `mena/pinakes/`

### Procession
A template-defined, station-based workflow that coordinates work across multiple rites within a session. Each station maps to a rite-scoped work unit; progress advances via `ari procession proceed`. The `completed_stations` log is append-only — `ari procession recede` repositions without erasing history.
- **Related**: Session, Rite, Station
- **Command**: `ari procession *`
- **Source**: `internal/procession/`

### Potnia
The Presiding Lady — per-rite entry agents providing work breakdown, specialist routing, and checkpoint guidance. Attested on Linear B tablet KN Gg(1) 702 as *da-pu₂-ri-to-jo po-ti-ni-ja* ("Potnia of the Labyrinth"), the presiding authority within the palace at Knossos (~1450-1300 BCE). Each rite has its own Potnia: the authority who presides within, not the external oracle consulted before entering.
- **Provenance**: Tier 1 — Bronze Age Attestation (Linear B tablet KN Gg 702)
- **Related**: Pythia, Orchestrator, Routing, Exousia
- **Source**: `rites/*/agents/potnia.md` (each rite has its own)

### Pythia
The cross-rite oracle/navigator — consulted before entering the labyrinth for routing and navigation across rites. Named for the oracle at Delphi, where Greeks sought guidance before great undertakings. Pythia occupies the correct mythological position: the external oracle consulted before a journey, not the authority who presides within. Distinct from Potnia (per-rite presiding authority).
- **Provenance**: Tier 2 — Classical Source (Herodotus, Pausanias)
- **Related**: Potnia, Cross-Rite Navigation, Routing
- **Source**: `agents/pythia.md`

### Scout
Agent type for wide-scan discovery agents. Scout agents (`type: scout` in frontmatter) perform broad observation without modifying state — they look for signals across the codebase and report findings. [Myron](#myron) is the canonical scout agent.
- **Related**: Myron, Glint, Analyst, Agent Types
- **Source**: Agent frontmatter `type: scout`

### Synkrisis
Comparative synthesis step following parallel domain evaluations. Weaves individual theoros reports into cross-domain patterns and the final "State of the {X}" attestation. Named for Plutarch's comparative analysis technique in the *Parallel Lives*.
- **Related**: Theoria, Theoroi, Synthesis, Report
- **Source**: Synthesis step within the `/theoria` dromena

### Theoria
The audit operation—a structured delegation of observers dispatched to assess domain health. Composite primitive: dromena (`/theoria`) + legomena (Pinakes) + agents (theoroi). Uses the Argus Pattern for parallel dispatch. Named for the Greek sacred state delegation.
- **Related**: Theoroi, Pinakes, Synkrisis, Argus Pattern
- **Source**: `/theoria` dromena + `mena/pinakes/` legomena

### Theoroi
Domain evaluator agents dispatched by a theoria. Each theoros observes a single domain using criteria from the Pinakes and produces a structured report. Read-only witnesses, not actors. Singular: theoros.
- **Related**: Theoria, Pinakes, Heroes, Domain Evaluator
- **Source**: `rites/shared/agents/theoros.md`

### Theseus
The navigator—main AI harness agent with agency but amnesia. The agentic intelligence making decisions and summoning heroes.
- **Related**: Main Agent, Navigator, Context Degradation
- **Source**: N/A (the LLM agent itself)

### White Sails
Confidence signal (WHITE/GRAY/BLACK) generated at session wrap. Solves the Aegeus problem of false confidence.
- **Related**: Confidence, Quality Signal, Wrap
- **Source**: `internal/sails/`

## Technical Terms

### Active Rite
The currently materialized rite, indicated by `.knossos/ACTIVE_RITE` file.
- **Related**: Rite, Materialization
- **Source**: Generated by `ari sync materialize`

### Artifact
Work product produced during a session (PRD, TDD, code, tests, docs). Tracked in session context and event log.
- **Related**: Session, Deliverable
- **Source**: Registered in `SESSION_CONTEXT.md`

### Artifact Registry
System tracking created artifacts with verification. Ensures work products are recorded and accessible.
- **Related**: Artifact, Verification
- **Source**: `internal/artifact/`

### Clew Contract
Event schema and types defining what gets recorded to `events.jsonl`.
- **Related**: Events, Schema, Contract
- **Source**: `internal/hook/clewcontract/`

### Cognitive Budget
Tool usage tracking with configurable thresholds (warn, park). Monitors message count to prevent context degradation.
- **Related**: Budget, Context, Park
- **Environment**: `ARIADNE_MSG_WARN`, `ARIADNE_MSG_PARK`, `ARIADNE_BUDGET_DISABLE`

### Complexity
Session complexity level (TRIVIAL, LOW, MEDIUM, HIGH, EXTREME). Determines workflow rigor and artifact requirements.
- **Related**: Session, Workflow
- **Source**: Defined in session context

### Event Log
`events.jsonl` file recording session decisions, tool calls, artifacts, errors. Physical manifestation of the clew.
- **Related**: Clew, Provenance, Audit
- **Source**: `[session-dir]/events.jsonl`

### Execution Mode
Operating mode (Native, Cross-Cutting, Orchestrated) determining session tracking and agent delegation patterns.
- **Related**: Session, Orchestration, Workflow
- **Source**: Declared in context file

### Forge
The `forge-rite` for platform development—building agents, tools, infrastructure. Daedalus's domain.
- **Related**: Builder, Infrastructure, Platform
- **Source**: `rites/forge/`

### Frontmatter (Agent)
YAML frontmatter in agent `.md` files declaring CC-OPP capabilities: `skills:`, `hooks:`, `disallowedTools:`, `memory:`. Parsed by `internal/agent/frontmatter.go` during materialization.
- **Related**: CC-OPP, Agent Transform, Materialization
- **Source**: `internal/agent/frontmatter.go`

### Hook
Code executed at lifecycle events (SessionStart, SessionStop, PostToolUse). Injects context and enforces contracts. Includes `agent-guard` for runtime tool restriction enforcement via triple-layer enforcement (disallowedTools + hooks + tool restrictions).
- **Related**: Lifecycle, Contract, Context Injection, Agent-Guard
- **Source**: `internal/hook/`, `internal/cmd/hook/`, materialized to the channel hooks directory

### Knossos Sections
Delimited blocks in `CLAUDE.md` (`<!-- KNOSSOS:START -->` ... `<!-- KNOSSOS:END -->`) regenerated by `ari sync inscription`.
- **Related**: Inscription, CLAUDE.md, Materialization
- **Source**: Managed by Knossos, custom sections preserved

### Mena
The lifecycle model for rite knowledge: dromena (transient commands, `.dro.md`) and legomena (persistent reference, `.lego.md`). Context lifecycle distinguishes them, not just routing.
- **Related**: Dromena, Legomena, Skills, Commands
- **Source**: `rites/*/mena/`

### Manifest
YAML file defining rite composition (agents, skills, hooks, workflows).
- **Related**: Rite, Composition, Configuration
- **Source**: `rites/[rite-name]/manifest.yaml`

### Materialization
Process of generating the channel directory PROJECTION from SOURCE via `ari sync materialize`.
- **Related**: SOURCE, PROJECTION, Sync
- **Command**: `ari sync materialize`

### Orphan
Session created but never wrapped—abandoned work detected by Naxos scan.
- **Related**: Naxos, Session, Cleanup
- **Command**: `ari naxos scan`

### Park
Pause current work session, preserving state for later resumption. Triggered manually or by cognitive budget threshold.
- **Related**: Resume, Session, Budget
- **Command**: `ari session park`

### Projection
Channel directories generated from SOURCE. Gitignored, ephemeral, recreated via materialization.
- **Related**: SOURCE, Materialization, Ephemeral
- **Location**: channel directory (e.g., `.channel/`)

### Rite
Manifest-driven practice bundle containing agents, skills, hooks, and workflows. Invocable ceremonies for specific purposes.
- **Related**: Manifest, Practice, Workflow
- **Source**: `rites/`

### Resume
Restore parked session with full context from clew.
- **Related**: Park, Session, Context
- **Command**: `ari session resume` (via `/resume` skill)

### Session
Tracked work context with lifecycle (create, park, resume, wrap). Managed by Moirai, recorded via clew.
- **Related**: Moirai, Clew, Lifecycle
- **Source**: `.sos/sessions/[session-id]/`

### Session Context
`SESSION_CONTEXT.md` file containing session state (initiative, complexity, phase, tasks, artifacts).
- **Related**: Session, State, Moirai
- **Source**: `.sos/sessions/[session-id]/SESSION_CONTEXT.md`

### Skill
Reusable capability invoked via Skill tool. Defined in rite manifests. Legomena that persist in context.
- **Related**: Rite, Tool, Legomena, Mena
- **Source**: `rites/*/mena/*.lego.md`, materialized to `.channel/skills/`

### Summonable Heroes
Third tier of agent classification — on-demand agents materialized to `~/.claude/agents/` via `ari agent summon`. Unlike standing agents (always active) and rite agents (active with materialized rite), summonable heroes exist only when explicitly called. Declared with `tier: summonable` in source frontmatter. Operations: [Klesis](#klesis) (summon), [Apolysis](#apolysis) (dismiss), [Katalogos](#katalogos) (roster).
- **Related**: Heroes, Klesis, Apolysis, Katalogos, Agent Tiers
- **Command**: `ari agent summon`, `ari agent dismiss`, `ari agent roster`

### SOURCE
The Knossos repository—canonical, versioned platform code that generates PROJECTIONS.
- **Related**: PROJECTION, Knossos, Repository
- **Location**: Repository root

### Sprint
Multi-task coordinated workflow tracked in `SPRINT_CONTEXT.md`.
- **Related**: Tasks, Coordination, Planning
- **Source**: `.sos/sessions/[sprint-id]/SPRINT_CONTEXT.md`

### TLA+
Formal specification language used to verify session state machine correctness.
- **Related**: Formal Verification, FSM, Correctness
- **Source**: Session FSM specification

### Thread
See **Clew**. (Note: code sometimes uses "thread"; doctrine prefers "clew")
- **Related**: Clew, Session, Events

### Tribute
Status reports and demos demanded by Minos. Periodic demonstration of progress.
- **Related**: Minos, Status, Reporting
- **Source**: `internal/tribute/`

### Worktree
Git worktree for parallel AI coding sessions with filesystem isolation. Enables multiple simultaneous rites.
- **Related**: Parallel Sessions, Isolation, Git
- **Command**: `ari worktree create`, `ari worktree cleanup`

### Wrap
Complete and finalize work session with validation and White Sails confidence signal generation.
- **Related**: Session, Confidence, Completion
- **Command**: `/wrap` skill

## Environment Variables

### ARIADNE_BUDGET_DISABLE
Disable cognitive budget tracking when set to `1`.
- **Default**: Enabled
- **Related**: Cognitive Budget

### ARIADNE_MSG_PARK
Message count threshold for park suggestion.
- **Default**: None (warn only)
- **Related**: Cognitive Budget, Park

### ARIADNE_MSG_WARN
Message count warning threshold.
- **Default**: 250
- **Related**: Cognitive Budget

## See Also

- `philosophy/mythology-concordance.md` - Extended mythological explanations
- `philosophy/knossos-doctrine.md` - Complete philosophical framework
- `reference/INDEX.md` - Navigation hub
- `../decisions/ADR-0009-knossos-roster-identity.md` - SOURCE vs PROJECTION clarification

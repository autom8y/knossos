# Knossos Glossary

> Terminology reference for the Knossos platform.

## Critical Distinction

| Term | Definition |
|------|------------|
| **SOURCE** | The Knossos repository—versioned, canonical platform code |
| **PROJECTION** | `.claude/` directories—gitignored, materialized by `ari sync materialize` |

## Mythology Terms

### Ariadne
The CLI binary (`ari`) that provides the clew—ensuring return through event recording and session management.
- **Related**: Clew, Theseus, Session
- **Source**: `cmd/ari/`

### Argus Pattern
N-agent parallel dispatch pattern. One main thread (body) launches multiple Task agents (eyes) simultaneously for distributed observation. Named for Argus Panoptes, the hundred-eyed giant. A reusable technique, not a specific component; envisioned for tactical playbook swarms and similar N-agent operations.
- **Related**: Theoria, Theoroi, Task Tool, Parallel Dispatch
- **Source**: Pattern (no single file)

### Athens
The `main` branch—destination of merged PRs. Return from the labyrinth is incomplete until work reaches Athens.
- **Related**: Pull Request, Merge, Ship
- **Source**: Git branch (not a file)

### Clew
Session state + event log (`events.jsonl`) + provenance trail. The thread that unwinds through the labyrinth, providing a path back when context degrades.
- **Related**: Thread, Events, Session
- **Source**: `internal/hook/clewcontract/`

### Daedalus
The builder—represented by the `forge-rite` for creating agents, tools, and platform infrastructure. Designed complexity is intentional architecture.
- **Related**: Forge, Builder, Infrastructure
- **Source**: `rites/forge/`

### Dionysus
The transformer—code review process that elevates work from isolation to merged canon. Turns abandonment into elevation.
- **Related**: Code Review, QA, Elevation
- **Source**: Review workflows, White Sails QA upgrade

### Exousia
Authority contract defining an agent's jurisdictional boundaries. Every agent declares three subsections: You Decide (autonomous authority), You Escalate (requires Pythia or user input), You Do NOT Decide (hard boundaries). Named from the Greek exousia (authority, jurisdiction).
- **Related**: Pythia, Agent Contract, Jurisdiction
- **Source**: `## Exousia` section in every agent `.md` file

### Heroes
Specialist agents invoked via Task tool for specific labors, defined in rite manifests. Summoned mid-journey with clew context.
- **Related**: Agents, Task Tool, Rite
- **Source**: `rites/[rite-name]/agents/`

### Inscription
The `CLAUDE.md` file declaring available rites, agents, execution mode, and hooks. Words carved at the labyrinth entrance. Mentions Pythia for routing, Exousia for authority contracts, `/go` for cold-start entry, and the Fates for session lifecycle.
- **Related**: CLAUDE.md, Knossos Sections, Pythia, Exousia, `/go`
- **Source**: `knossos/templates/CLAUDE.md.tpl`

### Knossos
The labyrinth—the Knossos repository itself (SOURCE), not the `.claude/` directories it generates (PROJECTION).
- **Related**: SOURCE, Platform, Repository
- **Source**: Repository root

### Minotaur
The initiative, feature, or task being pursued. The challenge at the labyrinth's heart—reason for the journey.
- **Related**: Initiative, Task, Goal
- **Source**: Defined in `SESSION_CONTEXT.md`

### Minos
Stakeholders who create initiatives and demand tribute (status reports and demos).
- **Related**: Tribute, Status Reports
- **Source**: `internal/tribute/`

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
- **Source**: Planned: `rites/shared/mena/pinakes/`

### Pythia
The Oracle—rite entry agents providing work breakdown, specialist routing, and checkpoint guidance. Speaks clearly, not cryptically. Each rite has its own Pythia.
- **Related**: Orchestrator, Consultation, Routing, Exousia
- **Source**: `rites/*/agents/pythia.md` (each rite has its own)

### Synkrisis
Comparative synthesis step following parallel domain evaluations. Weaves individual theoros reports into cross-domain patterns and the final "State of the {X}" attestation. Named for Plutarch's comparative analysis technique in the *Parallel Lives*.
- **Related**: Theoria, Theoroi, Synthesis, Report
- **Source**: Planned: synthesis in `/theoria` dromena

### Theoria
The audit operation—a structured delegation of observers dispatched to assess domain health. Composite primitive: dromena (`/theoria`) + legomena (Pinakes) + agents (theoroi). Uses the Argus Pattern for parallel dispatch. Named for the Greek sacred state delegation.
- **Related**: Theoroi, Pinakes, Synkrisis, Argus Pattern
- **Source**: Planned: `/theoria` dromena + `rites/shared/mena/pinakes/` legomena

### Theoroi
Domain evaluator agents dispatched by a theoria. Each theoros observes a single domain using criteria from the Pinakes and produces a structured report. Read-only witnesses, not actors. Singular: theoros.
- **Related**: Theoria, Pinakes, Heroes, Domain Evaluator
- **Source**: Planned: `rites/shared/agents/theoros.md`

### Theseus
The navigator—main Claude Code thread with agency but amnesia. The agentic intelligence making decisions and summoning heroes.
- **Related**: Main Agent, Navigator, Context Degradation
- **Source**: N/A (the LLM agent itself)

### White Sails
Confidence signal (WHITE/GRAY/BLACK) generated at session wrap. Solves the Aegeus problem of false confidence.
- **Related**: Confidence, Quality Signal, Wrap
- **Source**: `internal/sails/`

## Technical Terms

### Active Rite
The currently materialized rite, indicated by `.claude/ACTIVE_RITE` file.
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
- **Source**: Declared in `CLAUDE.md`

### Forge
The `forge-rite` for platform development—building agents, tools, infrastructure. Daedalus's domain.
- **Related**: Builder, Infrastructure, Platform
- **Source**: `rites/forge/`

### Hook
Code executed at lifecycle events (SessionStart, SessionStop, PostToolUse). Injects context and enforces contracts.
- **Related**: Lifecycle, Contract, Context Injection
- **Source**: `hooks/`, materialized to `.claude/hooks/`

### Knossos Sections
Delimited blocks in `CLAUDE.md` (`<!-- KNOSSOS:START -->` ... `<!-- KNOSSOS:END -->`) regenerated by `ari sync inscription`.
- **Related**: Inscription, CLAUDE.md, Materialization
- **Source**: Managed by Knossos, custom sections preserved

### Manifest
YAML file defining rite composition (agents, skills, hooks, workflows).
- **Related**: Rite, Composition, Configuration
- **Source**: `rites/[rite-name]/manifest.yaml`

### Materialization
Process of generating `.claude/` PROJECTION from SOURCE via `ari sync materialize`.
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
`.claude/` directories generated from SOURCE. Gitignored, ephemeral, recreated via materialization.
- **Related**: SOURCE, Materialization, Ephemeral
- **Location**: `.claude/`

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
- **Source**: `.claude/sessions/[session-id]/`

### Session Context
`SESSION_CONTEXT.md` file containing session state (initiative, complexity, phase, tasks, artifacts).
- **Related**: Session, State, Moirai
- **Source**: `.claude/sessions/[session-id]/SESSION_CONTEXT.md`

### Skill
Reusable capability invoked via Skill tool. Defined in rite manifests or user skills directory.
- **Related**: Rite, Tool, Capability
- **Source**: `user-skills/`, `rites/[rite-name]/skills/`

### SOURCE
The Knossos repository—canonical, versioned platform code that generates PROJECTIONS.
- **Related**: PROJECTION, Knossos, Repository
- **Location**: Repository root

### Sprint
Multi-task coordinated workflow tracked in `SPRINT_CONTEXT.md`.
- **Related**: Tasks, Coordination, Planning
- **Source**: `.claude/sprints/[sprint-id]/SPRINT_CONTEXT.md`

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
Git worktree for parallel Claude sessions with filesystem isolation. Enables multiple simultaneous rites.
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

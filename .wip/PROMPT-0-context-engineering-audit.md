# Prompt 0: Knossos Context Engineering Deep Audit

> Paste this entire document as your first message in a fresh Claude Code session opened in `/Users/tomtenuta/Code/knossos`.

---

## What You're Looking At

**Knossos** is a context-engineering meta-framework — "Rails for Claude Code." It provides structured workflows, agent orchestration, and context lifecycle management for CC (Claude Code) projects. The Go binary is `ari` (Ariadne).

You are about to do a **deep context-engineering audit** of the entire framework. The goal: take Knossos from a 6/10 to a 9/10 from a context-engineering perspective, evaluated against how CC actually operates.

## CC's Context Engineering Model (Ground Truth)

CC has exactly **6 context primitives**. Everything Knossos does maps to these:

| CC Primitive | How CC Uses It | Knossos Name | Source Location |
|---|---|---|---|
| **CLAUDE.md** | Always loaded into system prompt. Hierarchical: `~/.claude/CLAUDE.md` < project `.claude/CLAUDE.md`. Regions owned by different sources. | **Inscription** | `knossos/templates/sections/*.md.tpl` → `.claude/CLAUDE.md` |
| **Skills** (`.claude/skills/`) | Model autonomously loads via `Skill("name")` tool. Loaded on-demand based on `description` field matching. Persistent in context once loaded. | **Legomena** (`.lego.md`) | `mena/**/*.lego.md` + `rites/*/mena/**/*.lego.md` → `.claude/skills/` |
| **Commands** (`.claude/commands/`) | User invokes via `/name`. Transient — injects prompt, executes, exits. Can use `context:fork` to isolate. | **Dromena** (`.dro.md`) | `mena/**/*.dro.md` + `rites/*/mena/**/*.dro.md` → `.claude/commands/` |
| **Agents** (`.claude/agents/`) | Model spawns via `Task(subagent_type)`. Runs in subprocess with own context. Cannot spawn sub-agents (CC strips Task tool). | **Agents** | `rites/*/agents/*.md` → `.claude/agents/` |
| **Hooks** (`.claude/settings.json`) | Auto-fire on lifecycle events (PreToolUse, PostToolUse, SessionStart, etc.). Receive JSON on stdin. Return JSON to stdout. | **Hooks** | `ari` binary is the hook handler |
| **Rules** (`.claude/rules/`) | Path-scoped instructions. CC loads matching rules when files in that path are touched. | **Rules** | `knossos/templates/rules/*.md` → `.claude/rules/` |

### Critical Context Engineering Properties

1. **CLAUDE.md is always-on cost** — every token in it is paid on every turn. Must be minimal behavioral contract, not knowledge dump.
2. **Skills are the progressive disclosure mechanism** — CC loads them autonomously when `description` triggers match. The `Use when:` and `Triggers:` fields in descriptions are the routing signal. Poor descriptions = skills never loaded or loaded wrong.
3. **Commands are transient** — they inject and exit. `context:fork` isolates their context cost from the main conversation. Without fork, heavy commands pollute context.
4. **Agents are context-isolated** — they get their own context window. They cannot see the parent conversation. They cannot spawn sub-agents. They return a single result message.
5. **Hooks are the only way to inject ephemeral context** — session state, git state, environment info. They run outside CC's context window and inject via stdout.
6. **Rules are path-conditional** — they're zero-cost until you touch a matching file. They're the right place for "when editing X, remember Y" instructions.

### The Context Budget Equation

```
Per-turn cost = CLAUDE.md + loaded skills + conversation history + tool results
```

Anything that bloats CLAUDE.md or causes unnecessary skill loads is a direct tax on every turn. The audit should evaluate: **is each piece of content in the right primitive for its lifecycle and access pattern?**

## Knossos Architecture (What You Need to Know)

### Directory Structure
```
knossos/
├── mena/                    # Top-level mena (shared across all rites)
│   ├── guidance/            # Standards, prompting patterns, cross-rite routing
│   ├── navigation/          # Rite switching, worktrees, sessions, consult
│   ├── operations/          # commit, pr, qa, code-review, spike, architect
│   ├── session/             # start, park, handoff, common schemas
│   ├── meta/                # minus-1, zero, one (meta-session patterns)
│   ├── workflow/            # hotfix
│   ├── templates/           # Doc artifact templates (PRD, ADR, etc.)
│   ├── cem/                 # CEM debug skill
│   └── rite-switching/      # Rite switch commands per rite
├── rites/                   # Rite definitions (11 rites + shared)
│   ├── {rite-name}/
│   │   ├── agents/          # Agent definitions (orchestrator + specialists)
│   │   ├── mena/            # Rite-specific legomena (domain knowledge)
│   │   ├── manifest.yaml    # Rite metadata
│   │   ├── orchestrator.yaml # Workflow definition
│   │   └── workflow.yaml    # Agent routing DAG
│   └── shared/              # Cross-rite mena (smell-detection, handoff, templates)
├── knossos/
│   └── templates/           # Materialization templates
│       ├── sections/        # CLAUDE.md region templates (.md.tpl)
│       ├── rules/           # Path-scoped rules
│       └── CLAUDE.md.tpl    # Master inscription template
├── internal/                # Go source for ari binary
│   ├── materialize/         # Pipeline: source → .claude/ projection
│   ├── provenance/          # File ownership + divergence tracking
│   ├── inscription/         # CLAUDE.md region merging
│   ├── hook/                # Hook event handlers
│   │   └── clewcontract/    # Event types + JSONL writer
│   ├── session/             # Session FSM (NONE→ACTIVE→PARKED→ARCHIVED)
│   ├── agent/               # Agent archetype generation
│   └── ...
└── .claude/                 # Projected output (DO NOT EDIT DIRECTLY)
    ├── CLAUDE.md            # Generated inscription
    ├── settings.json        # Hook wiring
    ├── agents/              # Projected agents
    ├── commands/            # Projected dromena
    ├── skills/              # Projected legomena
    └── rules/               # Projected rules
```

### Current Inventory
- **34 dromena** (top-level mena/) + **0 rite-level dromena**
- **13 legomena** (top-level mena/) + **32 rite-level legomena**
- **58 agents** across 11 rites (4-7 per rite)
- **8 rules** (all path-scoped to internal/ packages)
- **7 inscription sections** (CLAUDE.md regions)
- **CLAUDE.md**: 93 lines

### Key Design Patterns
- **Mena model**: Dromena (.dro.md) are TRANSIENT (user-invoked, side effects). Legomena (.lego.md) are PERSISTENT (model-loaded, reference knowledge). This is a context lifecycle distinction, not just routing.
- **INDEX pattern**: Legomena use INDEX.lego.md as entry point with companion files for progressive disclosure. CC loads INDEX first; companions load only if needed.
- **Rite isolation**: Each rite has its own agents, mena, manifest, and workflow. Only one rite is ACTIVE at a time (agents projected from active rite).
- **Shared mena**: `rites/shared/mena/` contains cross-rite legomena (smell-detection, cross-rite-handoff, shared-templates).
- **Materialization pipeline**: `ari sync` reads source (mena/, rites/, knossos/templates/) and projects to .claude/. Provenance tracking prevents divergence.

## Your Mission

Launch **7 context-engineer agents** in parallel, each auditing a specific surface area against CC's context engineering model. Each agent should:

1. **Read all source material** in their scope
2. **Evaluate against CC primitives** — is content in the right primitive for its lifecycle?
3. **Identify context engineering deficiencies** — bloat, misrouted content, missing progressive disclosure, poor skill descriptions, missing context:fork, etc.
4. **Produce a structured findings doc** at `.wip/CE-AUDIT-{domain}.md`
5. **Rate severity**: CRITICAL (context budget killer), HIGH (misrouted content), MEDIUM (suboptimal), LOW (polish)

### Agent 1: `mena-dromena` — Dromena Quality Audit

**Scope**: All 34 `.dro.md` files in `mena/`

**Read**: Every `.dro.md` file. Also read `knossos/templates/sections/commands.md.tpl` to understand how they're surfaced.

**Evaluate**:
- Does every dromena use `context:fork`? (If not, it pollutes the main conversation)
- Are frontmatter fields complete and correct? (`name`, `description`, `scope`)
- Is `allowed-tools` appropriately scoped? (Over-permissive = risk)
- Is `disable-model-invocation: true` set where appropriate? (Heavy operations shouldn't be model-triggered)
- Are dromena truly transient? Or do some embed persistent reference knowledge that should be legomena?
- Is the description precise enough for CC to understand when NOT to invoke?
- Token cost: estimate the context cost of each dromena when invoked

**Output**: `.wip/CE-AUDIT-dromena.md`

### Agent 2: `mena-legomena` — Legomena Quality Audit

**Scope**: All 13 `.lego.md` in `mena/` + all 32 in `rites/*/mena/`

**Read**: Every `.lego.md` INDEX file and their companion files. Also read how CC surfaces them via `.claude/skills/`.

**Evaluate**:
- **Description quality**: Does each skill have precise `Use when:` and `Triggers:` in its description? CC uses these for autonomous loading — vague descriptions mean skills never load when needed or load when irrelevant
- **Progressive disclosure**: Does the INDEX→companion pattern work? Is INDEX small enough to be worth loading? Are companions granular enough?
- **Lifecycle correctness**: Is everything in legomena truly persistent reference knowledge? Or are some skills actually procedural (should be dromena)?
- **Cross-rite legomena** (`rites/shared/mena/`): Are they correctly shared? Do they duplicate content from top-level mena/?
- **Token cost per skill**: Is each skill worth its context cost when loaded? Are there skills that are too large?
- **Naming collisions**: Do any top-level and rite-level legomena have conflicting names?
- **Orphan skills**: Are there legomena that no agent or prompt references?

**Output**: `.wip/CE-AUDIT-legomena.md`

### Agent 3: `rite-agents` — Agent Quality Audit

**Scope**: All 58 agents across `rites/*/agents/*.md`

**Read**: Every agent file. Also read `rites/forge/mena/rite-development/templates/agent-template.md` (the canonical template) and `internal/agent/archetype.go` (the archetype generator).

**Evaluate**:
- **Template fidelity**: Do all agents follow the forge template structure? Which deviate and why?
- **Frontmatter compliance**: Are `name`, `role`, `description`, `type`, `tools`, `model`, `maxTurns` all present and valid?
- **Tool access**: Are tools appropriately scoped per agent type? (orchestrators should be Read-only, specialists get Edit/Write)
- **Domain authority**: Are "You decide" / "You escalate" / "You route to" sections clear and non-overlapping?
- **Handoff criteria**: Does every agent have clear completion criteria?
- **Context cost**: Agent prompts are loaded into subagent context windows. Are they right-sized? Which are bloated?
- **Orchestrator pattern**: Do all orchestrators follow the coach pattern (advise, don't execute)?
- **Cross-rite consistency**: Are shared patterns (Cross-Rite Protocol, escalation criteria) uniform?

**Output**: `.wip/CE-AUDIT-agents.md`

### Agent 4: `rite-structure` — Rite Structural Consistency Audit

**Scope**: `rites/*/` (manifest.yaml, orchestrator.yaml, workflow.yaml, README.md, TODO.md) for all 11 rites + shared

**Read**: All manifest, orchestrator, and workflow files. Also read `rites/forge/mena/rite-development/patterns/rite-composition.md`.

**Evaluate**:
- **Manifest completeness**: Do all rites have consistent manifest.yaml schemas?
- **Orchestrator YAML**: Are all orchestrator.yaml files structurally consistent? Same fields, same patterns?
- **Workflow DAGs**: Do workflow.yaml files correctly define agent routing? Are there unreachable agents?
- **Rite sizing**: Are any rites over-scoped (too many agents, too broad domain) or under-scoped (too few agents for their mission)?
- **Shared mena adequacy**: Is `rites/shared/mena/` the right abstraction? Should more or fewer things be shared?
- **TODO quality**: Are TODO.md files actionable or stale?
- **README consistency**: Do all rites have READMEs with consistent structure?

**Output**: `.wip/CE-AUDIT-rite-structure.md`

### Agent 5: `inscription-rules` — CLAUDE.md + Rules Audit

**Scope**: `.claude/CLAUDE.md`, `knossos/templates/sections/*.md.tpl`, `knossos/templates/rules/*.md`, `.claude/rules/`

**Read**: The generated CLAUDE.md, all section templates, all rules files. Also read `rites/ecosystem/mena/claude-md-architecture/first-principles.md` (the 6 design principles).

**Evaluate**:
- **CLAUDE.md token budget**: At 93 lines, is it too large? Too small? Is every line earning its per-turn cost?
- **Behavioral contract compliance**: Does CLAUDE.md only describe capabilities and workflow (Principle 1)? Or does it contain knowledge that should be skills?
- **Stability test**: Would any content become stale within a month (Principle 6)?
- **Region architecture**: Are the 7 sections the right decomposition? Should any merge or split?
- **Rules coverage**: 8 rules all scoped to `internal/`. Are there missing rules for other paths (`rites/`, `mena/`, `docs/`)?
- **Rules quality**: Are the rules concise enough to be useful without being bloat? Are they actually loaded when relevant?
- **Settings tier awareness**: Does the inscription correctly leverage CC's `skeleton < project < team < user` precedence model?

**Output**: `.wip/CE-AUDIT-inscription.md`

### Agent 6: `pipeline` — Materialization Pipeline Audit

**Scope**: `internal/materialize/`, `internal/provenance/`, `internal/inscription/`, `knossos/templates/CLAUDE.md.tpl`

**Read**: The key Go source files in these packages. Focus on the pipeline flow, not every line of code.

**Evaluate**:
- **Pipeline coherence**: Does the materialize→inscribe→provenance flow make sense? Are there unnecessary steps?
- **Projection fidelity**: Does the pipeline correctly transform source (mena/, rites/) into .claude/ output?
- **Provenance tracking**: Is the ownership model (knossos-owned vs user-owned) working correctly?
- **Idempotency**: Is `ari sync` truly idempotent? (Run twice = same output)
- **Error handling**: What happens when source files are malformed? Does the pipeline degrade gracefully?
- **Performance**: Are there obvious bottlenecks in the pipeline?
- **Missing capabilities**: What should the pipeline do that it doesn't? (e.g., validation, linting, token counting)

**Output**: `.wip/CE-AUDIT-pipeline.md`

### Agent 7: `hooks-events` — Hook + Event System Audit

**Scope**: `internal/hook/`, `internal/hook/clewcontract/`, hook sections in `.claude/settings.json`

**Read**: The hook handler code, event types, BufferedEventWriter, and the settings.json hook configuration.

**Evaluate**:
- **Event coverage**: 20 event types exist. Which CC lifecycle events are NOT covered? What's missing?
- **Hook data flow**: CC sends JSON on stdin. Does the handler parse it correctly for all event types?
- **SessionStart enrichment**: The SessionStart hook injects ephemeral context. Is it injecting the right information? Too much? Too little?
- **Async vs sync**: Which hooks should be async (fire-and-forget) vs sync (blocking)? Are they configured correctly?
- **PreToolUse guards**: The writeguard prevents direct writes to *_CONTEXT.md. Are there other dangerous patterns that should be guarded?
- **BufferedEventWriter**: Is the 5s flush + bounded loss window appropriate? Are events actually useful for anything downstream?
- **Error handling**: Do hooks fail open (allow) as designed? What happens when ari binary is missing from PATH?

**Output**: `.wip/CE-AUDIT-hooks.md`

## Execution Instructions

1. Launch all 7 agents in parallel using the Task tool with `subagent_type: "context-engineer"`
2. Each agent should receive its section above as the prompt (copy the full Agent N section)
3. After all 7 complete, read their output files
4. Synthesize into a single `.wip/CE-AUDIT-SYNTHESIS.md` with:
   - Top 10 highest-impact findings across all domains
   - A "6→9 roadmap" — the specific changes that would most improve context engineering quality
   - Quick wins (< 1 hour each) vs structural changes (multi-session)

## Anti-Patterns to Flag

These are the specific failure modes that keep Knossos at 6/10:

1. **Knowledge in CLAUDE.md** — behavioral contract only, knowledge goes in skills
2. **Vague skill descriptions** — CC can't route to skills it doesn't understand
3. **Missing context:fork on dromena** — heavy commands pollute main context
4. **Bloated agent prompts** — agents that embed reference knowledge instead of referencing skills
5. **Missing progressive disclosure** — monolithic skills instead of INDEX→companion pattern
6. **Rules gaps** — instructions that should be path-scoped but are jammed into CLAUDE.md
7. **Hook under-utilization** — ephemeral context that's hardcoded instead of injected
8. **Token waste** — content that exists in multiple primitives instead of one authoritative source

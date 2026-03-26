---
last_verified: 2026-03-26
---

# Architecture Map

> Subsystem table mapping packages to purpose. For Go source details, see inline godoc.

---

## Subsystems

| Package | Entry Point | Purpose |
|---------|-------------|---------|
| `internal/materialize/` | `ari sync materialize` | Generates channel directory projections from SOURCE |
| `internal/materialize/agent_transform.go` | Called during materialize | Injects CC-OPP capabilities (skills, hooks, memory) into agent prompts |
| `internal/session/` | `ari session *` | Session FSM (ACTIVE/PARKED/ARCHIVED), lifecycle management |
| `internal/hook/` | `ari hook *` | Hook infrastructure — context injection, clew recording, write guards |
| `internal/cmd/hook/agentguard.go` | `ari hook agent-guard` | Runtime tool restriction enforcement for agents |
| `internal/hook/clewcontract/` | PostToolUse hooks | Clew contract types and event schema (`events.jsonl`) |
| `internal/inscription/` | `ari sync inscription` | CLAUDE.md generation from templates with section management |
| `internal/sails/` | `ari sails check` | White Sails confidence computation (WHITE/GRAY/BLACK) |
| `internal/naxos/` | `ari naxos scan` | Orphaned session detection and cleanup |
| `internal/rite/` | `ari rite *` | Rite loading, validation, invoke/swap/release |
| `internal/agent/frontmatter.go` | Materialize pipeline | Agent YAML frontmatter parsing (CC-OPP capabilities) |
| `internal/provenance/` | `ari provenance *` | Provenance tracking and verification |
| `internal/tribute/` | `ari tribute generate` | Session completion reports for stakeholders |
| `internal/lock/` | Session operations | Advisory locking with stale detection |
| `internal/artifact/` | `ari artifact *` | Artifact registry and querying |
| `internal/serve/` | `ari serve` | Clew HTTP webhook server — Slack event processing, reasoning pipeline (knowledge retrieval + trust scoring + Claude generation), health endpoints |
| `internal/procession/` | `ari procession *` | Template-defined station workflows coordinating cross-rite work within a session |
| `internal/registry/` | `ari registry *` | Cross-repo knowledge domain catalog; syncs `.know/` domains from GitHub org repos |
| `internal/org/` | `ari org *` | Organization directory management; shared rites, agents, mena across projects |
| `internal/know/` | `ari knows *` | `.know/` domain inspection, staleness detection, semantic diff |
| `internal/ledge/` | `ari ledge *` | Work product artifact lifecycle (promotable → shelf) |
| `internal/cmd/agent/` | `ari agent *` | Agent validation, listing, summon/dismiss, scaffold, embody |
| `internal/cmd/land/` | `ari land *` | Cross-session knowledge synthesis entry point (delegates to Dionysus agent) |

---

## CLI → Package Mapping

| CLI Family | Primary Package(s) |
|------------|-------------------|
| `session` | `internal/session/`, `internal/lock/` |
| `rite` | `internal/rite/` |
| `sync` | `internal/materialize/`, `internal/inscription/` |
| `hook` | `internal/hook/`, `internal/cmd/hook/` |
| `sails` | `internal/sails/` |
| `naxos` | `internal/naxos/` |
| `tribute` | `internal/tribute/` |
| `artifact` | `internal/artifact/` |
| `provenance` | `internal/provenance/` |
| `agent` | `internal/cmd/agent/`, `internal/agent/` |
| `land` | `internal/cmd/land/` |
| `serve` | `internal/serve/`, `internal/slack/`, `internal/reason/` |
| `procession` | `internal/procession/` |
| `org` | `internal/org/` |
| `registry` | `internal/registry/` |
| `knows` | `internal/know/` |
| `ledge` | `internal/ledge/` |

---

## Clew / Serve Subsystem

The `serve` subsystem implements the Clew initiative service layer — an HTTP server that gives Slack access to cross-project knowledge.

**Purpose**: Process Slack webhook events with a reasoning pipeline that retrieves knowledge domains, scores trust, and generates responses via Claude.

**Entry point**: `ari serve --org <org-name>`

**Key packages**:

| Package | Role |
|---------|------|
| `internal/serve/` | HTTP server, routing, health endpoints (`/health`, `/ready`) |
| `internal/slack/` | Slack signature verification (HMAC-SHA256), event parsing |
| `internal/reason/` | Reasoning pipeline: knowledge retrieval → trust scoring → Claude generation |
| `internal/registry/` | Knowledge domain catalog for cross-repo retrieval |
| `internal/llm/` | Claude API client for generation step |

**Configuration**: Four-tier hierarchy — CLI flags → env vars → org env file (`$XDG_DATA_HOME/knossos/orgs/{org}/serve.env`) → hardcoded defaults.

**Telemetry**: Optional OpenTelemetry tracing via `OTEL_EXPORTER_OTLP_ENDPOINT`; no-op tracer when not set.

---

## Multi-Channel Architecture (ADR-0031)

Knossos uses a multi-channel projection architecture so the same rite definitions generate output for both Claude Code (`.claude/`) and Gemini CLI (`.gemini/`) without forking.

The `internal/materialize/` package implements a `ChannelCompiler` interface. Each channel target (claude, gemini) has its own compiler that handles channel-native format differences (markdown commands vs. TOML, etc.). `ari sync materialize --channel all` dispatches to all compilers in sequence.

**Practical implication**: The `--channel` flag on every `ari` command selects which channel directory to read/write. The default is `all` for write operations. Specify `--channel claude` or `--channel gemini` to target one.

---

## Key Flows

### Materialization (`ari sync materialize`)
1. Load rite manifest (`rites/*/manifest.yaml`)
2. Copy agents from `rites/*/agents/` + `agents/` → `.channel/agents/`
3. Apply agent transforms (inject skills, hooks, memory from frontmatter)
4. Copy mena from `rites/*/mena/` → `.channel/skills/` + `.channel/commands/`
5. Render inscription from `knossos/templates/` → channel context file

### Session Lifecycle
1. `session create` → Clotho domain → ACTIVE state
2. State mutations → Lachesis domain → validates + records events
3. `session wrap` → Atropos domain → quality gates → ARCHIVED state

### Agent-Guard Enforcement
1. The harness fires PreToolUse hook with tool name + agent context (JSON on stdin)
2. `ari hook agent-guard` checks agent frontmatter for `disallowedTools`
3. Returns block/allow decision to the harness runtime

---

**See Also:**
- [agent-capabilities.md](agent-capabilities.md) — CC-OPP capability details
- [../compliance/COMPLIANCE-STATUS.md](../compliance/COMPLIANCE-STATUS.md) — Platform metrics
- [INDEX.md](INDEX.md) — Navigation hub

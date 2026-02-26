---
last_verified: 2026-02-26
---

# Architecture Map

> Subsystem table mapping packages to purpose. For Go source details, see inline godoc.

---

## Subsystems

| Package | Entry Point | Purpose |
|---------|-------------|---------|
| `internal/materialize/` | `ari sync materialize` | Generates `.claude/` projections from SOURCE |
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

---

## Key Flows

### Materialization (`ari sync materialize`)
1. Load rite manifest (`rites/*/manifest.yaml`)
2. Copy agents from `rites/*/agents/` + `agents/` → `.claude/agents/`
3. Apply agent transforms (inject skills, hooks, memory from frontmatter)
4. Copy mena from `rites/*/mena/` → `.claude/skills/` + `.claude/commands/`
5. Render inscription from `knossos/templates/` → `.claude/CLAUDE.md`

### Session Lifecycle
1. `session create` → Clotho domain → ACTIVE state
2. State mutations → Lachesis domain → validates + records events
3. `session wrap` → Atropos domain → quality gates → ARCHIVED state

### Agent-Guard Enforcement
1. CC fires PreToolUse hook with tool name + agent context (JSON on stdin)
2. `ari hook agent-guard` checks agent frontmatter for `disallowedTools`
3. Returns block/allow decision to CC runtime

---

**See Also:**
- [agent-capabilities.md](agent-capabilities.md) — CC-OPP capability details
- [../compliance/COMPLIANCE-STATUS.md](../compliance/COMPLIANCE-STATUS.md) — Platform metrics
- [INDEX.md](INDEX.md) — Navigation hub

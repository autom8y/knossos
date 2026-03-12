---
name: session-context-schema
description: "Session context schema: SESSION_CONTEXT.md fields, hook output format, and query command reference."
---

# Session Context Schema (v2.4)

> Developer documentation for session context internals. Agents do not need to load this doc at runtime -- hook output and query output use self-describing YAML field names that agents parse directly.

**Source of truth**: `internal/session/context.go` (storage), `internal/cmd/hook/context.go` (hook output), `internal/output/output.go` (query output).

## Hook Output Format

The `ari hook context` command fires on `SessionStart` and emits **YAML frontmatter** to stdout. This is the primary channel through which agents receive session state.

Output is delimited by `---` markers. A comment header identifies the source.

### Example: Active Session

```yaml
---
# Session Context (injected by ari hook context)
session_id: session-20260306-122256-4fc1e1cc
harness_session_id: "abc123"
status: ACTIVE
initiative: "Context-Session Alignment Remediation"
active_rite: ecosystem
execution_mode: orchestrated
current_phase: implementation
git_branch: feat/csa-remediation
base_branch: main
complexity: SYSTEM
strands:
  - session_id: session-20260306-130000-aabbccdd
    status: ACTIVE
    frame_ref: .sos/wip/frames/s2-field-widening.md
  - session_id: session-20260306-140000-eeff0011
    status: LANDED
    landed_at: "2026-03-06T15:00:00Z"
available_rites:
  - ecosystem
  - 10x-dev
available_agents:
  - potnia
  - integration-engineer
know_status: "Codebase knowledge: 5 domains (architecture: fresh, ...)"
---
```

### Example: No Session

```yaml
---
# Session Context (injected by ari hook context)
has_session: false
harness_session_id: "abc123"
---
```

### Post-Frontmatter Sections

The hook may append content after the closing `---` delimiter. These are not YAML fields:

- **Throughline Agents**: Active throughline agent IDs (key-value pairs, survive compaction)
- **Recovered State**: Rehydrated `COMPACT_STATE.md` checkpoint content (consumed on read)

## Secondary Channel: `ari session query`

The `ari session query` command provides on-demand session state reads mid-conversation. It returns the same YAML frontmatter format as the hook so agents can parse both identically.

```bash
ari session query                    # Full YAML frontmatter output
ari session query -o json            # JSON output
ari session query --field complexity # Single field value (plain text)
ari session query --field status     # Single field value (plain text)
ari session query --session-id ID    # Explicit session target
```

**Resolution chain** (same as hook): explicit `--session-id` flag > CC session map > smart scan (single active session).

**Differences from hook**: Query output omits `git_branch`, `base_branch`, `available_rites`, `available_agents`, `know_status`, `compact_state`, and `throughline_ids` (environment-dependent fields not relevant to mid-session state pulls). Query is read-only and does not emit lifecycle events.

**Comment header**: `# Session Context (ari session query)` (vs. hook's `# Session Context (injected by ari hook context)`).

### Queryable Fields via `--field`

`session_id`, `status`, `initiative`, `complexity`, `active_rite`, `execution_mode`, `current_phase`, `frayed_from`, `frame_ref`, `park_source`, `claimed_by`.

## SESSION_CONTEXT.md Frontmatter Schema

Fields stored in `SESSION_CONTEXT.md` and parsed by `internal/session/context.go`. All frontmatter is YAML between `---` delimiters.

### Core Identity

| Field | YAML Key | Go Type | Required | Description |
|-------|----------|---------|----------|-------------|
| Schema Version | `schema_version` | `string` | Yes | Current: "2.3" |
| Session ID | `session_id` | `string` | Yes | Format: `session-YYYYMMDD-HHMMSS-XXXXXXXX` |
| Created At | `created_at` | `time.Time` | Yes | ISO 8601 timestamp |
| Initiative | `initiative` | `string` | Yes | Human-readable description |
| Complexity | `complexity` | `string` | Yes | PATCH, MODULE, SYSTEM, INITIATIVE, or MIGRATION |
| Active Rite | `active_rite` | `string` | Yes | Rite name (e.g., "10x-dev") |
| Rite | `rite` | `*string` | No | Nullable rite override (null = cross-cutting) |
| Current Phase | `current_phase` | `string` | Yes | requirements, design, implementation, validation, complete |

### Lifecycle State

| Field | YAML Key | Go Type | Required | Description |
|-------|----------|---------|----------|-------------|
| Status | `status` | `Status` | Yes | ACTIVE, PARKED, or ARCHIVED |
| Parked At | `parked_at` | `*time.Time` | No | When session was parked |
| Parked Reason | `parked_reason` | `string` | No | Why session was parked |
| Park Source | `park_source` | `string` | No | manual, auto, or fray (v2.3+) |
| Archived At | `archived_at` | `*time.Time` | No | When session was archived |
| Resumed At | `resumed_at` | `*time.Time` | No | When session was last resumed |

### Fray / Strand

| Field | YAML Key | Go Type | Required | Description |
|-------|----------|---------|----------|-------------|
| Frayed From | `frayed_from` | `string` | No | Parent session ID (set on child) |
| Fray Point | `fray_point` | `string` | No | Phase at which fork occurred |
| Strands | `strands` | `[]Strand` | No | Child sessions (v2.3: typed struct) |
| Frame Ref | `frame_ref` | `string` | No | Link to `.sos/wip/frames/{slug}.md` (v2.3+) |
| Claimed By | `claimed_by` | `string` | No | CC session ID that claimed this session (v2.3+) |

#### Strand Struct

| Field | YAML Key | Go Type | Required | Description |
|-------|----------|---------|----------|-------------|
| Session ID | `session_id` | `string` | Yes | Child session ID |
| Status | `status` | `string` | Yes | SPAWNED, ACTIVE, LANDED, or ABANDONED |
| Frame Ref | `frame_ref` | `string` | No | Frame link for this strand |
| Landed At | `landed_at` | `string` | No | ISO 8601 timestamp when strand landed |

### Complexity Levels

| Level | Use For |
|-------|---------|
| PATCH | Single-file changes, quick fixes, < 200 LOC |
| MODULE | Multiple files, < 2000 LOC, clear interfaces |
| SYSTEM | Multiple modules, APIs, data persistence |
| INITIATIVE | Multiple services, infrastructure, complex integration |
| MIGRATION | Cross-cutting migrations, large-scale refactors |

## Hook Output Fields (ContextOutput)

The hook's `ContextOutput` struct adds environment-derived fields not stored in `SESSION_CONTEXT.md`. These are computed at hook execution time.

| Field | YAML Key | Source | Description |
|-------|----------|--------|-------------|
| Harness Session ID | `harness_session_id` | Hook env (stdin) | Harness's own session ID |
| Execution Mode | `execution_mode` | Computed | `native`, `orchestrated`, or `cross-cutting` |
| Has Session | `has_session` | Computed | `true` if a session was resolved, `false` otherwise |
| Git Branch | `git_branch` | `git rev-parse` | Current branch name |
| Base Branch | `base_branch` | `git symbolic-ref` | Default remote branch (fallback: `main`) |
| Available Rites | `available_rites` | SourceResolver | List of rite names from 4-tier resolution |
| Available Agents | `available_agents` | Agents dir | List of agent names (`.md` files, extension stripped) |
| Know Status | `know_status` | `.know/` dir | Codebase knowledge freshness summary |
| Compact State | `compact_state` | Session dir | Rehydrated checkpoint (post-frontmatter section, consumed on read) |
| Throughline IDs | `throughline_ids` | Session dir | Active throughline agent IDs (post-frontmatter section) |

All fields use `omitempty` -- absent when empty or zero-valued.

## Body Fields

Fields written by Moirai to the markdown body (after frontmatter). These are NOT parsed by Go -- they exist as prose in the markdown content.

| Field | Written By | Description |
|-------|-----------|-------------|
| `last_agent` | Moirai (Lachesis) | Last agent to work on session |
| `handoff_count` | Moirai (Lachesis) | Total handoffs in session |
| `last_handoff_at` | Moirai (Lachesis) | Timestamp of last handoff |
| `artifacts` | Moirai (Lachesis) | List of produced artifacts with paths and status |
| `blockers` | Moirai (Lachesis) | Current blockers with descriptions |
| `next_steps` | Moirai (Lachesis) | Pending actions |

## State Machine

```
StatusNone   -> [StatusActive]              (create)
StatusActive -> [StatusParked, StatusArchived] (park, wrap)
StatusParked -> [StatusActive, StatusArchived] (resume, wrap)
StatusArchived -> []                          (terminal)
```

## Mutation Authority

- **Frontmatter**: Written by `ari session *` CLI commands. The write guard hook blocks direct agent writes.
- **Body**: Written by Moirai agent via lock-protected Edit operations.
- **Events**: Written by clew hooks to `events.jsonl` (append-only).

## Backward Compatibility

- v2.1/v2.2 sessions parse correctly (missing fields = Go zero values)
- Strands `[]string` (v2.1/v2.2) auto-converts to `[]Strand` on parse via polymorphic `strandList` unmarshaler
- New fields are all `omitempty` -- absent in existing sessions
- `ari session query` returns the same YAML format as hook output, so agents parse both identically

## Removed Fields

These fields appeared in earlier schema documentation but were never implemented in Go:

- `parked_phase` -- Redundant with `current_phase` when `parked_at` is set
- `parked_git_status` -- Never implemented
- `parked_uncommitted_files` -- Never implemented
- `resume_count` -- Derivable from event log
- `quality_gates_passed` -- Replaced by sails system
- `quality_gates_skipped` -- Replaced by sails system

# Knossos Integration Guide

> roster/.claude/ IS Knossos. This repository is the Knossos platform.

## Identity Relationship

The "roster" repository **is** Knossos - an agentic workflow platform for Claude Code. The naming reflects:

- **Knossos**: The platform (this repository, eventually renamed)
- **Ariadne**: The CLI thread (`ari` binary) - named for Ariadne's thread in Greek mythology
- **roster/.claude/**: Source templates for hooks, skills, and agents

### Future Rename

This repository will be renamed from `roster` to `knossos`. Until then:
- `roster/` = `knossos/`
- `.claude/` directory contains the platform configuration
- `ariadne/` contains the CLI implementation

## Ariadne CLI Integration

### Installation

The `ari` binary is built from `ariadne/`:

```bash
cd ariadne
just build  # Builds with CGO_ENABLED=0 for macOS compatibility
```

Or install to GOPATH:

```bash
cd ariadne
just install  # Installs to $GOPATH/bin
```

### Shell-Callable Interface

Hooks and scripts invoke Ariadne via shell command:

```bash
ari <command> [subcommand] [flags]
```

### Command Groups

| Command | Description |
|---------|-------------|
| `ari session` | Manage workflow sessions |
| `ari team` | Switch and manage agent teams |
| `ari hook` | Claude Code hook infrastructure |
| `ari sails` | White Sails quality gates |
| `ari manifest` | Manage configuration manifests |
| `ari sync` | Sync state between locations |
| `ari validate` | Validate artifacts and state |
| `ari worktree` | Git worktree management |
| `ari artifact` | Artifact tracking |
| `ari handoff` | Agent handoff operations |

### Session Commands

```bash
# Create new session
ari session create "feature-name" --complexity MODULE

# Check current session status
ari session status

# List all sessions
ari session list

# Park current session (pause work)
ari session park "reason for parking"

# Resume a parked session
ari session resume <session-id>

# Create seeded session for parallel execution
ari session create "feature-name" --complexity MODULE --seed

# Transition session state
ari session transition <new-state>

# Audit session history
ari session audit
```

### Hook Commands

Hooks run automatically via Claude Code's hook system. Manual invocation:

```bash
# Context injection (SessionStart)
ari hook context

# Clew event tracking (PostToolUse)
ari hook clew

# Auto-park detection (PostToolUse)
ari hook autopark

# Write guard validation (PreToolUse)
ari hook writeguard

# Orchestrator routing (PreToolUse)
ari hook route

# Artifact validation (PostToolUse)
ari hook validate
```

### Handoff Commands

Agent handoffs enable workflow phase transitions with proper event tracking:

```bash
# Prepare handoff (validates readiness, emits task_end)
ari handoff prepare --from architect --to principal-engineer

# Execute handoff (triggers transition, emits task_start)
ari handoff execute --from architect --to principal-engineer

# Query current handoff state
ari handoff status

# View handoff history from events.jsonl
ari handoff history
```

### Sails Commands

```bash
# Check White Sails quality gate
ari sails check

# Check specific session
ari sails check --session-id <session-id>
```

### Global Flags

All commands support these flags:

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: text, json, yaml (default: text) |
| `-v, --verbose` | Enable verbose output (JSON lines to stderr) |
| `-p, --project-dir` | Project root directory (overrides discovery) |
| `-s, --session-id` | Session ID (overrides current) |

## Hook Integration

### From .claude/hooks/

Hooks are shell scripts that invoke `ari`. The hook system uses thin shell wrappers for fast early-exit checks before calling the Go binary.

Example from `.claude/hooks/ari/thread.sh`:

```bash
#!/bin/bash
# thread.sh - Smart dispatch wrapper for thread/artifact tracking
set -euo pipefail

# FAST PATH: Early exit checks (<5ms, no subprocess)
case "$CLAUDE_HOOK_TOOL_NAME" in Edit|Write|Bash) ;; *) exit 0 ;; esac

# Check for active session
SESSION_DIR="${CLAUDE_SESSION_DIR:-.claude/sessions}"
[[ ! -d "$SESSION_DIR" ]] && exit 0

# Feature flag (default: Go enabled)
[[ "${USE_ARI_HOOKS:-1}" != "1" ]] && exit 0

# DISPATCH: Call ari (<100ms total)
ARI="${ARIADNE_BIN:-/path/to/ariadne/ari}"
exec "$ARI" hook thread --output json
```

### Hook Environment Variables

Claude Code provides these environment variables to hooks:

| Variable | Description |
|----------|-------------|
| `CLAUDE_HOOK_TOOL_NAME` | Name of the tool being used (Edit, Write, Bash, etc.) |
| `CLAUDE_HOOK_TOOL_INPUT` | JSON input to the tool |
| `CLAUDE_SESSION_DIR` | Session directory path |

Ariadne-specific:

| Variable | Description |
|----------|-------------|
| `USE_ARI_HOOKS` | Feature flag: 1=enabled (default), 0=disabled |
| `ARIADNE_BIN` | Path to ari binary (for development) |
| `ARIADNE_MSG_WARN` | Cognitive budget warning threshold (default: 250) |
| `ARIADNE_MSG_PARK` | Cognitive budget park suggestion threshold |
| `ARIADNE_BUDGET_DISABLE` | Set to 1 to disable cognitive budget tracking |
| `ARIADNE_SESSION_KEY` | Override session key for budget tracking (testing) |

### Performance Targets

The hook system is designed for minimal latency impact:

- Early exit: <5ms (when hooks disabled or no session)
- Full execution: <100ms (with all processing)
- Maximum timeout: 500ms (safety limit)

## Clew Contract Events

The Clew Contract v2 records significant events to `events.jsonl` in the session directory:

### Event Types

| Event | Description |
|-------|-------------|
| `session_start` | Session created |
| `session_end` | Session wrapped or archived |
| `task_start` | Task work began |
| `task_end` | Task completed |
| `artifact_created` | New artifact produced |
| `error` | Error occurred |
| `tool_call` | Tool invocation recorded |
| `file_change` | File modified |
| `command` | Shell command executed |
| `decision` | Significant decision made |
| `context_switch` | Context change (new file, new task) |
| `sails_generated` | White Sails confidence signal generated |
| `handoff_prepared` | Agent handoff validated and ready |
| `handoff_executed` | Agent handoff completed |

### Artifact Types

| Type | Description |
|------|-------------|
| `prd` | Product Requirements Document |
| `tdd` | Technical Design Document |
| `adr` | Architecture Decision Record |
| `test_plan` | Test Plan |
| `code` | Source Code |
| `white_sails` | White Sails quality gate artifact |

### Event Format

```json
{
  "timestamp": "2026-01-05T14:30:00Z",
  "event_type": "tool_call",
  "tool_name": "Write",
  "file_path": "/path/to/file.go",
  "session_id": "session-20260105-143000-abc12345"
}
```

### sails_generated Event with Evidence

The `sails_generated` event includes evidence paths from WHITE_SAILS.yaml:

```json
{
  "ts": "2026-01-05T14:30:00Z",
  "type": "sails_generated",
  "meta": {
    "color": "white",
    "computed_base": "green",
    "session_id": "session-20260105-143000-abc12345",
    "file_path": ".claude/sessions/.../WHITE_SAILS.yaml",
    "evidence_paths": {
      "tests": "ariadne/internal/..._test.go",
      "build": "ariadne/ari",
      "lint": "ariadne/.golangci.yml"
    }
  }
}
```

## Cognitive Budget

The cognitive budget system tracks tool usage per CLI invocation and warns when approaching context limits.

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ARIADNE_MSG_WARN` | 250 | Emit warning when tool count reaches this threshold |
| `ARIADNE_MSG_PARK` | - | Suggest /park when reaching this threshold |
| `ARIADNE_BUDGET_DISABLE` | 0 | Set to 1 to disable tracking entirely |

### Behavior

- Counter increments on each PostToolUse hook execution
- Scoped per-CLI-invocation (not per-Knossos-session)
- Warnings emitted to stderr (non-blocking)
- State stored in `/tmp/ariadne-msg-count-{session-key}`

### Example Warning

```
[COGNITIVE_BUDGET] Tool count 250 reached warn threshold. Consider /park to preserve context.
```

## State Management

### state-mate Integration

state-mate is the authority for session state mutations. It delegates to ari for state operations:

```bash
# Park via state-mate
Task(moirai, "park session with reason: blocked on review")

# Which internally calls:
ari session park "blocked on review"
```

### SESSION_CONTEXT.md

Session state lives in `.claude/sessions/<session-id>/SESSION_CONTEXT.md`:

```yaml
---
schema_version: "2.1"
session_id: "session-20260105-143000-abc12345"
status: "ACTIVE"  # ACTIVE | PARKED | ARCHIVED
initiative: "feature-name"
complexity: "MODULE"
team: "10x-dev-pack"
created_at: "2026-01-05T14:30:00Z"
---

## Progress

- [x] PRD complete
- [ ] TDD in progress
- [ ] Implementation pending
```

### Session ID Format

Session IDs follow the pattern: `session-YYYYMMDD-HHMMSS-<8-char-hex>`

Example: `session-20260105-143000-abc12345`

## Project Discovery

Ariadne automatically discovers the project root by looking for:

1. `.claude/` directory
2. `CLAUDE.md` file
3. `.git/` directory

The `--project-dir` flag overrides automatic discovery.

## Configuration

### Config File Location

Default: `$XDG_CONFIG_HOME/ariadne/config.yaml`

### Example Config

```yaml
default_output: text
verbose: false
```

## Error Handling

Ariadne uses structured error codes:

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Project not found |
| 3 | Session not found |
| 4 | Validation error |
| 5 | State transition error |

## Related Resources

- [White Sails Guide](white-sails.md) - Confidence signaling system
- [Parallel Sessions Guide](parallel-sessions.md) - Session seeding for parallel execution
- [User Preferences](user-preferences.md) - Configure Claude Code behavior
- [ADR-0001](../decisions/ADR-0001-session-state-machine-redesign.md) - Session FSM design
- [ADR-0005](../decisions/ADR-0005-state-mate-centralized-state-authority.md) - state-mate authority
- [ADR-0010](../decisions/ADR-0010-worktree-session-seeding.md) - Worktree session seeding

# Ariadne CLI

> The clew that navigates the Knossos labyrinth

Ariadne (`ari`) is the command-line interface for the Knossos agentic workflow platform. It provides session management, hook infrastructure, quality gates, and agent handoff operations for Claude Code workflows.

## Installation

### Build from Source

```bash
cd ariadne
just build
```

The binary is built at `./ari`. Add to your PATH or use the full path.

### Install to GOPATH

```bash
cd ariadne
just install  # Installs to $GOPATH/bin
```

## Quick Start

```bash
# Create a new session
ari session create "feature-name" --complexity MODULE

# Check session status
ari session status

# Park session (pause work)
ari session park "reason for parking"

# Resume a parked session
ari session resume <session-id>
```

## Command Reference

### Session Management

| Command | Description |
|---------|-------------|
| `ari session create` | Create new session |
| `ari session status` | Show current session status |
| `ari session list` | List all sessions |
| `ari session park` | Park (pause) current session |
| `ari session resume` | Resume a parked session |
| `ari session transition` | Transition session state |
| `ari session audit` | Show session audit history |

### Session Seeding (`--seed` Flag)

The `--seed` flag enables creating multiple PARKED sessions from a single terminal for parallel execution:

```bash
# Create seeded sessions (creates PARKED sessions without violating single-session constraint)
ari session create "Feature A" --complexity=MODULE --seed
ari session create "Feature B" --complexity=MODULE --seed
ari session create "Feature C" --complexity=PATCH --seed

# Resume in separate terminals
ari session resume session-xxx-feature-a  # Terminal 1
ari session resume session-xxx-feature-b  # Terminal 2
ari session resume session-xxx-feature-c  # Terminal 3
```

#### Seeding Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--seed` | Enable worktree seeding mode | false |
| `--seed-prefix=PATH` | Custom worktree location | `/tmp/roster-seed-` |
| `--seed-keep` | Keep worktree after seeding (debugging) | false |

**How it works**: The `--seed` flag creates an ephemeral git worktree, creates and parks the session in that isolated environment, copies the session back to the main repository, and cleans up the worktree. This bypasses the single-session-per-terminal constraint while maintaining state isolation.

See [ADR-0010](../docs/decisions/ADR-0010-worktree-session-seeding.md) for implementation details and [Parallel Sessions Guide](../docs/guides/parallel-sessions.md) for usage patterns.

### Hook Commands

Hooks integrate with Claude Code's hook system for automatic context injection and event tracking.

| Command | Description |
|---------|-------------|
| `ari hook context` | Context injection (SessionStart) |
| `ari hook clew` | Clew event tracking (PostToolUse) |
| `ari hook autopark` | Auto-park detection (PostToolUse) |
| `ari hook writeguard` | Write guard validation (PreToolUse) |
| `ari hook route` | Orchestrator routing (PreToolUse) |
| `ari hook validate` | Artifact validation (PostToolUse) |

### Quality Gates (Sails)

```bash
# Check White Sails quality gate for current session
ari sails check

# Check specific session
ari sails check --session-id <session-id>
```

### Agent Handoffs

```bash
# Prepare handoff (validates readiness)
ari handoff prepare --from architect --to principal-engineer

# Execute handoff
ari handoff execute --from architect --to principal-engineer

# Query handoff state
ari handoff status

# View handoff history
ari handoff history
```

### Other Commands

| Command | Description |
|---------|-------------|
| `ari team` | Switch and manage agent teams |
| `ari manifest` | Manage configuration manifests |
| `ari sync` | Sync state between locations |
| `ari validate` | Validate artifacts and state |
| `ari worktree` | Git worktree management |
| `ari artifact` | Artifact tracking |

## Global Flags

All commands support these flags:

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: text, json, yaml (default: text) |
| `-v, --verbose` | Enable verbose output |
| `-p, --project-dir` | Project root directory |
| `-s, --session-id` | Override current session ID |

## Build Configuration

### CGO_ENABLED=0 Constraint

The `ari` binary is built with `CGO_ENABLED=0`:

```bash
CGO_ENABLED=0 go build -o ari ./cmd/ari/main.go
```

**Why?** On macOS arm64, CGO-enabled binaries can encounter dyld issues related to LC_UUID mismatches:

1. **LC_UUID Verification**: macOS verifies that shared libraries have matching UUIDs to prevent accidental mixing of library versions
2. **CGO Dependencies**: When CGO is enabled, Go links against system libraries (libc, libSystem) which have platform-specific UUIDs
3. **Cross-Compilation Issues**: Building on one architecture variant can embed UUIDs that don't match the runtime environment

**Consequence**: Pure Go builds (`CGO_ENABLED=0`) avoid system library linking entirely, producing fully static binaries that work across macOS environments without dyld verification failures.

**Tradeoff**: Some Go packages require CGO (e.g., `sqlite3`, `libgit2`). The Ariadne CLI intentionally avoids such dependencies to maintain portability.

See [ADR-0010](../docs/decisions/ADR-0010-worktree-session-seeding.md#7-cgo_enabled0-build-constraint) for the full rationale.

### Build Commands

```bash
just build           # Build binary
just build-verbose   # Build with verbose output
just test            # Run all tests
just test-verbose    # Run tests with verbose output
just lint            # Run linter
just clean           # Clean build artifacts
just install         # Install to $GOPATH/bin
just info            # Show binary info
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `USE_ARI_HOOKS` | Emergency kill switch (set to 0 to disable) | enabled |
| `ARIADNE_BIN` | Path to ari binary | Auto-discovered |
| `ARIADNE_MSG_WARN` | Cognitive budget warning threshold | 250 |
| `ARIADNE_MSG_PARK` | Cognitive budget park threshold | - |
| `ARIADNE_BUDGET_DISABLE` | Disable cognitive budget tracking | 0 |

## Project Structure

```
ariadne/
  cmd/
    ari/           # Main binary entry point
  internal/
    cmd/           # Command implementations
    sails/         # White Sails quality gates
    session/       # Session state management
    ...
  schemas/         # JSON schemas for validation
  testdata/        # Test fixtures
  docs/            # Internal design docs
  justfile         # Build automation
```

## Error Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Project not found |
| 3 | Session not found |
| 4 | Validation error |
| 5 | State transition error |

## Related Documentation

- [Knossos Integration Guide](../docs/guides/knossos-integration.md) - Full CLI reference
- [Parallel Sessions Guide](../docs/guides/parallel-sessions.md) - Session seeding patterns
- [White Sails Guide](../docs/guides/white-sails.md) - Quality gate system
- [ADR-0010](../docs/decisions/ADR-0010-worktree-session-seeding.md) - Worktree session seeding
- [ADR-0006](../docs/decisions/ADR-0006-parallel-session-orchestration.md) - Parallel execution pattern

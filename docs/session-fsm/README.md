# Session State Machine Documentation

This directory contains documentation for the roster session state machine (FSM) system, which provides formally-verified session lifecycle management.

## Contents

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System design, components, and integration points |
| [OPERATIONS.md](./OPERATIONS.md) | Commands, troubleshooting, and recovery procedures |

## Quick Start

### Check Session Status

```bash
./user-hooks/lib/session-manager.sh status
```

### Create a Session

```bash
./user-hooks/lib/session-manager.sh create "Feature: My Initiative" "MODULE"
```

### Park/Resume Session

```bash
# Park
./user-hooks/lib/session-manager.sh mutate park "Reason for parking"

# Resume
./user-hooks/lib/session-manager.sh mutate resume
```

### Complete Session

```bash
./user-hooks/lib/session-manager.sh mutate wrap "true"
```

## Key Concepts

### States

The session FSM has three top-level states:

- **ACTIVE**: Session is in progress, work happening
- **PARKED**: Session is suspended, can be resumed
- **ARCHIVED**: Session is complete, immutable (terminal)

### Single Source of Truth

The `status` field in SESSION_CONTEXT.md is the ONLY authority for session state. No inference from other fields.

### Schema Version

Sessions use v2 schema (`schema_version: "2.0"`). Legacy v1 sessions are auto-migrated on first access.

## Implementation Artifacts

| Artifact | Location | Description |
|----------|----------|-------------|
| Core FSM | `user-hooks/lib/session-fsm.sh` | State machine implementation |
| Migration | `user-hooks/lib/session-migrate.sh` | v1 to v2 migration |
| CLI | `user-hooks/lib/session-manager.sh` | Unified CLI interface |
| TLA+ Spec | `docs/specs/session-fsm.tla` | Formal specification |
| Tests | `tests/session-fsm/` | BATS test suite |

## Related Documentation

| Document | Location |
|----------|----------|
| ADR | `docs/decisions/ADR-0001-session-state-machine-redesign.md` |
| TDD | `docs/design/TDD-session-state-machine.md` |
| TLA+ Spec | `docs/specs/session-fsm.tla` |
| Alloy Spec | `docs/specs/session-permissions.als` |

## Getting Help

- **Troubleshooting**: See [OPERATIONS.md](./OPERATIONS.md#troubleshooting-guide)
- **Error Codes**: See [OPERATIONS.md](./OPERATIONS.md#error-codes-reference)
- **Recovery**: See [OPERATIONS.md](./OPERATIONS.md#recovery-procedures)

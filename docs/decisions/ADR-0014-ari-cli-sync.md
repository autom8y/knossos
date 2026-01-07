# ADR-0014: ari CLI User Commands Synchronization

**Status**: Proposed
**Date**: 2026-01-07
**Context**: Sprint ari-cli-user-commands-sync

## Context

Audit of user-command skills vs `ari` CLI implementation revealed documentation-to-implementation drift. Several flags documented in user-command files (`user-commands/rite-switching/*.md`) are not implemented in the `ari sync materialize` command.

### Documented but Not Implemented

| Flag/Command | Documentation Location | Expected Behavior |
|--------------|----------------------|-------------------|
| `--remove-all` | `user-commands/rite-switching/*.md` | Remove all orphan agents (with backup) |
| `--keep-all` | `user-commands/rite-switching/*.md` | Preserve all orphan agents in project |
| `--promote-all` | `user-commands/rite-switching/*.md` | Move all orphans to user-level (~/.claude/agents/) |
| `--dry-run` | `user-commands/rite-switching/*.md` | Preview changes without applying |
| `--update` | `user-commands/rite-switching/*.md` | Pull latest definitions even if already on rite |
| `ari rite pantheon` | Implicit in `/hygiene` etc. | Display active agents for current rite |

### Current CLI State

```bash
$ ari sync materialize --help
Flags:
      --force         Force regeneration, overwriting local changes
  -h, --help          help for materialize
      --rite string   Rite to materialize (defaults to current ACTIVE_RITE)
```

Only `--force` and `--rite` are implemented.

## Decision

Implement the missing flags and command to achieve full parity between documentation and CLI behavior.

### Implementation Order

1. **Orphan handling flags** (batch - single PR):
   - `--remove-all`: Delete orphan agents, create backup in `.claude/.orphan-backup/`
   - `--keep-all`: Explicitly preserve orphans (default behavior, makes it explicit)
   - `--promote-all`: Copy orphans to `~/.claude/agents/` before removal

2. **Preview flag**:
   - `--dry-run`: Output what would change without modifying filesystem

3. **Update flag**:
   - `--update` / `-u`: Force re-sync even if rite hasn't changed

4. **Pantheon command**:
   - `ari rite pantheon`: List agents in current rite with roles

### Orphan Definition

An "orphan" is an agent file in `.claude/agents/` that:
- Is NOT in the incoming rite's manifest
- Is NOT in `~/.claude/agents/` (user-level)
- Was likely from a previous rite materialization

### Conflict Resolution

When `--remove-all` and `--keep-all` are both specified: ERROR (mutually exclusive)
When neither specified: Default to `--keep-all` (preserve backward compatibility)

## Consequences

### Positive

- Documentation accurately reflects CLI behavior
- Users can safely switch rites without orphan accumulation
- `ari rite pantheon` provides quick visibility into active agents
- Explicit flags give users control over orphan handling

### Negative

- Additional flags increase CLI complexity
- Orphan backup directory may accumulate over time (need cleanup strategy)

### Risks

- Accidental data loss if `--remove-all` used without understanding
  - Mitigation: Always create backup before removal
  - Mitigation: Verbose output showing what was removed

## Implementation Notes

### File Locations

- CLI source: `cmd/ari/sync/materialize.go` (or similar)
- Orphan detection: Use existing manifest comparison logic
- Backup location: `.claude/.orphan-backup/{timestamp}/`

### Testing Requirements

- Integration test: `--remove-all` creates backup and removes orphans
- Integration test: `--promote-all` copies to ~/.claude/agents/
- Integration test: `--dry-run` outputs changes without modifying
- Integration test: Mutually exclusive flag validation
- Integration test: `ari rite pantheon` lists agents correctly

## References

- Sprint: sprint-cli-sync-20260107
- Audit: Path Materialization Audit (2026-01-07)
- Related: ADR-sync-materialization.md

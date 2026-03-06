---
domain: feat/worktree-management
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/worktree/**/*.go"
  - "./internal/cmd/worktree/**/*.go"
  - "./docs/decisions/ADR-0010*.md"
  - "./docs/decisions/ADR-0029*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Git Worktree Management

## Purpose and Design Rationale

Enables parallel CC sessions through isolated git worktree environments with full knossos framework seeding. Works around the single-session-per-terminal constraint by providing filesystem isolation.

**ADR-0010**: Worktree session seeding. **ADR-0029**: Worktree environment contract (three creation paths, rite inheritance, symlink vs real dir decisions).

### Three Creation Paths

1. `ari worktree create` — fully seeded immediately
2. CC `EnterWorktree` + `worktree-seed` hook — seeded via hook
3. `git worktree add` — unseeded, functional after `ari sync`

### Directory Model

`.knossos/` and `.know/` are **symlinks** to main worktree (shared config/knowledge). `.sos/` and `.ledge/` are **real dirs** (worktree-local session state and artifacts).

## Conceptual Model

### Worktree Identity

- **ID**: `wt-YYYYMMDD-HHMMSS-{hex}` (stable, never reused)
- **Name**: user-friendly lookup key
- **Path**: `.worktrees/{id}/` in main project root

### Registry

`.worktrees/metadata.json` — central worktree registry. Per-worktree: `.knossos/.worktree-meta.json`.

### Rite Inheritance

Linked worktrees inherit active rite from main worktree when no `--rite` flag is passed.

## Implementation Map

Domain: `/Users/tomtenuta/Code/knossos/internal/worktree/` (6 files: worktree.go, lifecycle.go, operations.go, metadata.go, git.go, session_integration.go). CLI: `/Users/tomtenuta/Code/knossos/internal/cmd/worktree/` (11 files: 10 subcommands). Hooks: `internal/cmd/hook/worktreeseed.go`, `worktreeremove.go`.

### Manager API

`Create()`, `List()`, `Status()`, `Remove()`, `Cleanup()`, `Switch()`, `Clone()`, `Sync()`, `Export()`, `Import()`.

### Tests

`lifecycle_test.go` (18 tests), `operations_test.go` (~40 tests), `worktree_test.go` (~100 tests).

## Boundaries and Failure Modes

- Does NOT manage git branches (always `--detach HEAD`)
- Does NOT share sessions between worktrees
- `setupWorktreeEcosystem()` errors are swallowed (no error return)
- Hook path (`.knossos/worktrees/`) vs CLI path (`.worktrees/`) divergence — hook worktrees NOT in `metadata.json`
- `session_integration.go` re-implements frontmatter parsing (parallel to `internal/session`)
- Metadata/filesystem divergence recovery is non-fatal (errors discarded)

## Knowledge Gaps

1. Hook worktrees not tracked in `metadata.json` — impact on `ari worktree list` unconfirmed.
2. ADR-0029 P0 tests not found in observed test files.
3. `internal/materialize/worktree.go` not read directly.

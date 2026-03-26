---
domain: feat/worktree-management
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/worktree/**/*.go"
  - "./internal/cmd/worktree/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.87
format_version: "1.0"
---

# Git Worktree Management

## Purpose and Design Rationale

Eliminates file contention between parallel sessions by giving each an isolated filesystem. Git-native isolation (detached HEAD, no branch proliferation). Ecosystem seeding on creation (full ari sync inside worktree). Shared knowledge, isolated state (.knossos/ and .know/ symlinked from main root; .sos/ and .ledge/ created fresh per worktree). ADR-0010 and ADR-0029 referenced but not found on disk.

## Conceptual Model

**Worktree type:** ID (wt-YYYYMMDD-HHMMSS-hex), Name, Path, Branch, Rite, BaseBranch, Complexity, CreatedAt. **Dual registry:** central (.worktrees/metadata.json) + per-worktree (.knossos/.worktree-meta.json). Self-healing via SyncMetadataFromFilesystem. Session integration via worktree_id field in SESSION_CONTEXT.md. **Lifecycle:** Create (git worktree add --detach + setupWorktreeEcosystem) -> Remove (cleanup + git worktree remove) -> Cleanup (age-based bulk remove + git worktree prune).

## Implementation Map

`internal/worktree/` (6 files): worktree.go (types), lifecycle.go (Manager: Create/List/Status/Remove/Cleanup), operations.go (Switch/Clone/Sync/Export/Import + setupWorktreeEcosystem), metadata.go (MetadataManager + PerWorktreeMeta), git.go (GitOperations, 30s timeout), session_integration.go. `internal/cmd/worktree/` (12 files): create, list, status, remove, cleanup, switch, clone, sync, export, import subcommands.

## Boundaries and Failure Modes

Must be invoked from main worktree (nested guard). No branch-per-worktree (detached HEAD). Switch does not change shell working directory. IsWorktree detection is heuristic (path-content check). setupWorktreeEcosystem collision: explicit os.RemoveAll before symlink. Export excludes .git. Pull is fast-forward only. 30s git timeout for all operations. No concurrent Create locking.

## Knowledge Gaps

1. ADR-0010 and ADR-0029 not found on disk
2. internal/assets package not in architecture inventory
3. session.IsValidComplexity not read

# ADR-0016: Sync and Materialization Model

| Field | Value |
|-------|-------|
| **Status** | ACCEPTED |
| **Date** | 2026-01-07 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A |
| **Superseded by** | N/A |

## Context

Knossos requires a configuration materialization model where `.claude/` directories are fully generated from templates rather than checked into repositories. This ADR establishes the patterns for:

1. How templates are rendered into configuration files
2. The UX for initializing and updating configurations
3. How user customizations are preserved during updates
4. The synchronization mechanism between templates and generated files

### Current State

The codebase has partial implementation:

**Existing Infrastructure**:
- `ariadne/internal/inscription/generator.go` uses Go `text/template` with custom functions
- `knossos/templates/` contains 16 template files for CLAUDE.md sections
- `ariadne/internal/cmd/sync/` implements sync commands (status, pull, push, diff, resolve, history, reset)
- `TDD-ariadne-sync.md` specifies three-way merge and conflict resolution

**Gap**: No unified materialization model connecting templates to `.claude/` generation.

### Problem

1. **Initialization Complexity**: Users must understand multiple commands to set up Knossos
2. **Update Fragility**: No standardized way to update templates while preserving customizations
3. **Rite Switching Overhead**: Switching rites requires manual coordination
4. **Conflict Ambiguity**: No clear model for when/how conflicts are detected and resolved

### Requirements (from Upstream SPIKE)

Per `SPIKE-knossos-consolidation-architecture.md` Section 8:
- Content model: Templates in repo; `.claude/` fully generated
- `.claude/` is gitignored, NOT part of repo
- `ari sync` or `ari init` generates `.claude/` from `templates/`
- Single canonical structure (no profiles)
- Must support rite switching (swap-rite functionality)

## Decision

Adopt a **chezmoi-inspired generation model** with a **single idempotent `ari sync` command**.

### Core Principles

1. **Generation over Symlinks**: Templates are rendered to real files in `.claude/`, not symlinked
2. **Single Source of Truth**: `templates/` and `rites/{name}/` are canonical; `.claude/` is derived
3. **Idempotent Operations**: `ari sync` is safe to run repeatedly in any state
4. **Three-Way Merge**: Track base state to detect both local and remote changes
5. **Explicit Conflicts**: Never silently overwrite user modifications

### Command Structure

| Command | Purpose | Modifies Files |
|---------|---------|----------------|
| `ari sync` | Initialize or update configuration (idempotent) | Yes |
| `ari sync status` | Show state without modifying | No |
| `ari sync resolve` | Resolve pending conflicts | Yes |
| `ari sync --force` | Force overwrite local changes | Yes (dangerous) |
| `ari rite switch <name>` | Switch active rite | Yes |

### Materialization Flow

```
┌───────────────────────────────────────────────────────────────┐
│                        TEMPLATE SOURCES                        │
├───────────────────────────────────────────────────────────────┤
│  templates/           │  rites/{active}/   │  rites/shared/   │
│  ├── hooks/           │  ├── agents/       │  └── skills/     │
│  ├── CLAUDE.md.tpl    │  └── skills/       │                  │
│  └── sections/        │                    │                  │
└───────────┬───────────┴─────────┬──────────┴────────┬─────────┘
            │                     │                   │
            └─────────────────────┴───────────────────┘
                                  │
                                  v
                    ┌─────────────────────────┐
                    │      ari sync           │
                    │  (materialization)      │
                    └────────────┬────────────┘
                                 │
                                 v
            ┌────────────────────────────────────────┐
            │              .claude/                   │
            │  ├── agents/       (from rite)         │
            │  ├── skills/       (from rite+shared)  │
            │  ├── hooks/        (from templates)    │
            │  ├── CLAUDE.md     (generated)         │
            │  └── sync/                             │
            │      └── state.json (tracking)         │
            └────────────────────────────────────────┘
```

### Templating Engine

**Go `text/template` with Sprig function library**:

```go
import "github.com/Masterminds/sprig/v3"

func templateFuncs() template.FuncMap {
    funcs := sprig.TxtFuncMap() // 100+ utility functions
    // Add Knossos-specific functions
    funcs["include"] = includePartial
    funcs["agents"] = loadAgentTable
    funcs["term"] = lookupTerminology
    return funcs
}
```

**Rationale**:
- `text/template` already used in `generator.go`
- Sprig is the de facto standard (Helm compatibility)
- No additional binary dependencies
- Extensive string, list, and dict manipulation functions

### Three-Way Merge Model

```
                    BASE STATE
                  (last sync)
                       │
        ┌──────────────┴──────────────┐
        │                             │
        v                             v
   LOCAL STATE                  REMOTE STATE
  (user changes)                 (template)
        │                             │
        └──────────────┬──────────────┘
                       │
                       v
                 MERGE RESULT
```

**Conflict Resolution Matrix**:

| Local vs Base | Remote vs Base | Action |
|---------------|----------------|--------|
| Unchanged | Changed | Auto-update from remote |
| Changed | Unchanged | Preserve local (user customization) |
| Changed | Changed | CONFLICT (requires resolution) |
| N/A (new) | Exists | Auto-add |
| Deleted | Unchanged | Keep deleted |
| Deleted | Changed | CONFLICT (requires resolution) |

### State Tracking

`.knossos/sync/state.json`:

```json
{
  "version": "1.0",
  "initialized_at": "2026-01-07T15:00:00Z",
  "last_sync": "2026-01-07T16:00:00Z",
  "active_rite": "10x-dev",
  "remote": {
    "url": "github.com/autom8y/knossos",
    "ref": "main",
    "commit": "abc123..."
  },
  "files": {
    "agents/architect.md": {
      "checksum": "sha256:...",
      "source": "rites/10x-dev/agents/architect.md",
      "modified_locally": false
    }
  }
}
```

### User Journey

**New User**:
```bash
brew install knossos/tap/ari
cd my-project
ari sync  # Interactive: prompts for rite selection, generates .claude/
```

**Existing User Update**:
```bash
ari sync  # Pulls updates, preserves customizations, reports conflicts
```

**Rite Switch**:
```bash
ari rite switch rnd  # Backs up customizations, switches agents/skills
```

## Consequences

### Positive

- **Single Command UX**: Users only need to remember `ari sync`
- **Safe Updates**: Three-way merge prevents data loss
- **Clear Conflicts**: Explicit conflict detection and resolution
- **Rite Isolation**: Clean switching between workflows
- **Ecosystem Alignment**: Sprig functions match Helm/Kubernetes patterns

### Negative

- **State File Dependency**: `.knossos/sync/state.json` must be maintained
- **Complexity**: Three-way merge is more complex than simple overwrite
- **Merge Conflicts**: Users must resolve conflicts manually when both sides change
- **Learning Curve**: Understanding materialization model takes time

### Neutral

- **No Symlinks**: Claude Code hooks require real files anyway
- **Checksum Overhead**: SHA256 computation adds minimal latency
- **Template Constraints**: Limited to text/template syntax (no complex logic)

## Alternatives Considered

### Alternative 1: Symlink-Based (GNU Stow Pattern)

**Approach**: Symlink `.claude/` contents to canonical locations

**Rejected Because**:
- Claude Code hooks may not work with symlinked directories
- Cannot support encrypted or permission-modified files
- No templating capability (static files only)

### Alternative 2: Separate Init and Update Commands

**Approach**: `ari init` for first-time, `ari update` for subsequent

**Rejected Because**:
- Two commands to learn increases cognitive load
- Error-prone: running wrong command in wrong state
- CLI best practice prefers idempotent single commands

### Alternative 3: Watch-Based Auto-Sync

**Approach**: Automatically sync on file changes (like direnv)

**Rejected Because**:
- Implicit behavior can surprise users
- Conflicts harder to handle automatically
- Resource overhead for file watching

### Alternative 4: gomplate CLI Instead of Library

**Approach**: Shell out to `gomplate` binary for templating

**Rejected Because**:
- Additional binary dependency
- Less control over custom functions
- Already have Go implementation

## Implementation

### Phase 1: Sprig Integration

Update `ariadne/internal/inscription/generator.go`:
- Add `github.com/Masterminds/sprig/v3` dependency
- Merge Sprig functions with existing custom functions
- Update tests for new function availability

### Phase 2: State Tracking

Extend `ariadne/internal/sync/`:
- Define `state.json` schema in `state.go`
- Implement checksum tracking in `tracker.go`
- Add local modification detection

### Phase 3: Materialization Engine

New package `ariadne/internal/materialize/`:
- Template rendering with context
- File generation with checksum tracking
- Rite-aware source resolution

### Phase 4: Command Integration

Update `ariadne/internal/cmd/sync/`:
- Modify `ari sync` to handle initialization
- Add `--init` flag for explicit re-initialization
- Integrate with rite switching

### Rollback Strategy

If materialization model fails:
1. Restore `.claude/` from git (if previously tracked)
2. Remove `.knossos/sync/state.json` to reset state
3. Re-run `ari sync --force` to regenerate

## Related

- **SPIKE-materialization-model.md**: Research findings informing this decision
- **SPIKE-knossos-consolidation-architecture.md**: Parent architecture defining content model
- **TDD-ariadne-sync.md**: Technical design for sync commands
- **ADR-0009-knossos-roster-identity.md**: Knossos platform identity

## Notes

This ADR establishes the foundation for Knossos as a fully generated configuration system. The key insight from research is that chezmoi's copy-based generation model (vs symlinks) enables features impossible otherwise: templating, per-file permissions, and encryption support. While Knossos does not currently require encryption, the generation model leaves this option open.

The single `ari sync` command follows the "crash-only design" principle from CLI best practices: the same command handles all states, and re-running after failure safely recovers.

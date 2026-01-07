{{/* ariadne-cli section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START ariadne-cli -->
## Ariadne CLI

The `ari` binary provides session lifecycle, rite management, and workflow operations.

### Session Management

```bash
ari session create "initiative" COMPLEXITY    # Create new session (PATCH|MODULE|SYSTEM|INITIATIVE|MIGRATION)
ari session status                            # Show current session state
ari session list                              # List all sessions
ari session park --reason "taking break"      # Park session with reason
ari session resume                            # Resume parked session
ari session wrap                              # Complete session with sails
ari session transition <phase>                # Transition to workflow phase
ari session audit                             # Show session audit log
```

### Rite Management

```bash
ari rite list                                 # List available rites
ari rite info <name>                          # Show rite details
ari rite status                               # Show active rite status
ari rite current                              # Show current active rite
ari rite swap <name>                          # Switch to different rite
ari rite invoke <name>                        # Invoke rite entry point
ari rite validate <name>                      # Validate rite manifest
```

### Sync Operations

```bash
ari sync materialize --rite <name>            # Materialize rite to .claude/
ari sync materialize --force                  # Force overwrite existing
ari sync status                               # Show sync status
ari sync diff                                 # Show pending changes
ari sync pull                                 # Pull from source
ari sync push                                 # Push to destination
ari sync history                              # Show sync history
```

### Hook Operations

```bash
ari hook clew                                 # Emit session clew (breadcrumb)
ari hook context                              # Emit full context injection
ari hook validate                             # Validate hook configuration
ari hook route                                # Route to appropriate handler
ari hook writeguard                           # Check write permissions
ari hook autopark                             # Check autopark conditions
```

### Quality Gates

```bash
ari sails check                               # Check White Sails confidence
ari validate artifact <file>                  # Validate PRD/TDD/ADR artifact
ari validate handoff --phase=<phase>          # Validate handoff criteria
ari validate schema <name> <file>             # Validate against schema
```

### Agent Handoffs

```bash
ari handoff prepare --from <agent> --to <agent>   # Prepare handoff package
ari handoff execute --from <agent> --to <agent>   # Execute handoff
ari handoff status                                # Show handoff status
ari handoff history                               # Show handoff history
```

### Manifest Operations

```bash
ari manifest show                             # Show current manifest
ari manifest diff                             # Show manifest differences
ari manifest merge                            # Merge manifest sources
ari manifest validate                         # Validate manifest structure
```

### Inscription (CLAUDE.md)

```bash
ari inscription sync                          # Sync CLAUDE.md with templates
ari inscription sync --dry-run                # Preview changes
ari inscription validate                      # Check manifest and CLAUDE.md
ari inscription diff                          # Show pending changes
ari inscription backups                       # List available backups
ari inscription rollback                      # Restore from backup
```

### Artifact Registry

```bash
ari artifact list                             # List registered artifacts
ari artifact register <path>                  # Register new artifact
ari artifact query <type>                     # Query artifacts by type
ari artifact rebuild                          # Rebuild artifact index
```

### Session Cleanup (Naxos)

```bash
ari naxos scan                                # Scan for orphaned sessions
ari naxos scan --inactive-threshold=12h       # Custom inactivity threshold
ari naxos scan --include-archived             # Include archived sessions
```

### Worktree Management

```bash
ari worktree create <name>                    # Create isolated worktree
ari worktree list                             # List worktrees
ari worktree status                           # Show worktree status
ari worktree switch <name>                    # Switch to worktree
ari worktree sync                             # Sync worktree state
ari worktree remove <name>                    # Remove worktree
ari worktree cleanup                          # Clean up stale worktrees
```

### Tribute Generation

```bash
ari tribute generate                          # Generate session tribute/summary
```

### Cognitive Budget

Tool usage tracking with configurable thresholds:
- `ARIADNE_MSG_WARN=250` - Warning threshold (default)
- `ARIADNE_MSG_PARK` - Park suggestion threshold
- `ARIADNE_BUDGET_DISABLE=1` - Disable tracking

Build: `just build` (from repo root)

Full reference: `docs/guides/ariadne-cli.md`
<!-- KNOSSOS:END ariadne-cli -->

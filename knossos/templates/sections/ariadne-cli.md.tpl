{{/* ariadne-cli section template */}}
{{/* Owner: knossos - Always synced from Knossos templates */}}
<!-- KNOSSOS:START ariadne-cli -->
## Ariadne CLI

The `ari` binary provides session and hook operations:

```bash
# Session management
ari session create "initiative" COMPLEXITY
ari session status
ari session park "reason"

# Hook operations
ari hook thread
ari hook context

# Quality gates
ari sails check

# Agent handoffs
ari handoff prepare --from <agent> --to <agent>
ari handoff execute --from <agent> --to <agent>
ari handoff status
ari handoff history

# CLAUDE.md inscription
ari inscription sync              # Sync CLAUDE.md with templates
ari inscription sync --dry-run    # Preview changes
ari inscription validate          # Check manifest and CLAUDE.md
ari inscription backups           # List available backups
ari inscription rollback          # Restore from backup
```

### Cognitive Budget

Tool usage tracking with configurable thresholds:
- `ARIADNE_MSG_WARN=250` - Warning threshold (default)
- `ARIADNE_MSG_PARK` - Park suggestion threshold
- `ARIADNE_BUDGET_DISABLE=1` - Disable tracking

Build: `cd ariadne && just build`

Full reference: `docs/guides/knossos-integration.md`
<!-- KNOSSOS:END ariadne-cli -->

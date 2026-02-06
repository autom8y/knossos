# CLI Reference

> Complete reference for all `ari` commands.

[Ariadne](../../reference/GLOSSARY.md#ariadne) (`ari`) is the CLI binary providing session management, rite operations, and platform infrastructure.

---

## Quick Reference

| Family | Commands | Description |
|--------|----------|-------------|
| [session](cli-session.md) | 11 | Create, list, park, resume, wrap sessions |
| [rite](cli-rite.md) | 10 | List, invoke, release, swap rites |
| [worktree](cli-worktree.md) | 10 | Parallel sessions with filesystem isolation |
| [sync](cli-sync.md) | 8 | Synchronize .claude/ with remotes |
| [hook](cli-hook.md) | 6 | Claude Code hook infrastructure |
| [handoff](cli-handoff.md) | 4 | Agent handoffs between phases |
| [inscription](cli-inscription.md) | 5 | CLAUDE.md inscription system |
| [artifact](cli-artifact.md) | 4 | Register and query workflow artifacts |
| [validate](cli-validate.md) | 3 | Validate artifacts and handoffs |
| [manifest](cli-manifest.md) | 4 | Show, validate, diff manifests |
| [sails](cli-sails.md) | 1 | White Sails quality gates |
| [naxos](cli-naxos.md) | 1 | Orphaned session cleanup |
| [tribute](cli-tribute.md) | 1 | Session summary generation |
| [completion](cli-completion.md) | 4 | Shell autocompletion |

**Total**: 72 commands across 14 families

---

## Installation

The `ari` binary is built from Go source:

```bash
cd /path/to/roster/ariadne
just build
```

Or install directly:
```bash
go install github.com/your-org/roster/ariadne/cmd/ari@latest
```

---

## Global Flags

All commands support these flags:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `$XDG_CONFIG_HOME/ariadne/config.yaml` | Config file path |
| `-o, --output` | string | `text` | Output format: text, json, yaml |
| `-p, --project-dir` | string | auto-discovered | Project root directory |
| `-s, --session-id` | string | current session | Override session ID |
| `-v, --verbose` | bool | false | Enable verbose output |

---

## Common Workflows

### Start a New Session

```bash
ari session create "Feature implementation" --complexity=MODULE
```

### Check Status

```bash
ari session status
ari rite current
ari sails check
```

### Park and Resume

```bash
ari session park --reason="Taking a break"
# Later...
ari session resume
```

### Complete Work

```bash
ari sails check        # Verify quality
ari session wrap       # Archive session
```

### Parallel Sessions

```bash
ari worktree create "feature-auth" --rite=10x-dev
cd .worktrees/wt-*/
# Work in isolation...
```

### Switch Rites

```bash
ari rite list          # See available rites
ari rite swap docs     # Switch to docs rite
```

---

## Command Families by Use Case

### Session Lifecycle
- `ari session create` — Start work
- `ari session park` — Pause work
- `ari session resume` — Continue work
- `ari session wrap` — Complete work

### Rite Management
- `ari rite list` — Discover rites
- `ari rite swap` — Change rite
- `ari rite invoke` — Borrow components
- `ari rite pantheon` — See agents

### Configuration
- `ari sync materialize` — Generate .claude/
- `ari inscription sync` — Update CLAUDE.md
- `ari manifest show` — View config

### Quality & Validation
- `ari sails check` — Quality gate
- `ari validate artifact` — Artifact validation
- `ari validate handoff` — Handoff validation

### Maintenance
- `ari naxos scan` — Find orphans
- `ari worktree cleanup` — Remove stale worktrees

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `USE_ARI_HOOKS=0` | Emergency kill switch to disable ari hook implementations (default: enabled) |
| `ARIADNE_MSG_WARN` | Message count warning threshold (default: 250) |
| `ARIADNE_MSG_PARK` | Message count park suggestion threshold |
| `ARIADNE_BUDGET_DISABLE=1` | Disable cognitive budget tracking |

---

## See Also

- [Ariadne CLI Guide](../guides/ariadne-cli.md)
- [Knossos Integration Guide](../guides/knossos-integration.md)
- [Ariadne Glossary Entry](../../reference/GLOSSARY.md#ariadne)

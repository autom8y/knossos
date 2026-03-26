---
last_verified: 2026-03-26
---

# CLI Reference

> Complete reference for all `ari` commands.

[Ariadne](../../reference/GLOSSARY.md#ariadne) (`ari`) is the CLI binary providing session management, rite operations, and platform infrastructure.

---

## Quick Reference

| Family | Commands | Description |
|--------|----------|-------------|
| [session](cli-session.md) | 10 | Create, list, park, resume, wrap sessions |
| [rite](cli-rite.md) | 10 | List, invoke, release, swap rites |
| [worktree](cli-worktree.md) | 11 | Parallel sessions with filesystem isolation |
| [sync](cli-sync.md) | 8 | Synchronize channel directory with remotes |
| [hook](cli-hook.md) | 6 | Hook infrastructure + agent-guard |
| [handoff](cli-handoff.md) | 4 | Agent handoffs between phases |
| [inscription](cli-inscription.md) | 5 | Context file inscription system |
| [artifact](cli-artifact.md) | 4 | Register and query workflow artifacts |
| [validate](cli-validate.md) | 3 | Validate artifacts and handoffs |
| [manifest](cli-manifest.md) | 4 | Show, validate, diff manifests |
| [agent](cli-agent.md) | 8 | Validate, scaffold, summon, and manage agents |
| [serve](cli-serve.md) | 1 | Clew HTTP webhook server (Slack + reasoning pipeline) |
| [procession](cli-procession.md) | 6 | Template-defined cross-rite station workflows |
| [land](cli-land.md) | 1 | Cross-session knowledge synthesis via Dionysus |
| [status](cli-status.md) | 1 | Unified project health dashboard |
| [org](cli-org.md) | 4 | Create and manage organization-level resources |
| [registry](cli-registry.md) | 3 | Cross-repo knowledge domain catalog |
| [ask](cli-ask.md) | 1 | Natural language query for commands and workflows |
| [knows](cli-knows.md) | 1 | Inspect `.know/` knowledge domains |
| [ledge](cli-ledge.md) | 2 | List and promote work product artifacts |
| [lint](cli-lint.md) | 1 | Lint agents, dromena, and legomena source files |
| [provenance](cli-provenance.md) | 1 | Channel directory file ownership and checksums |
| [explain](cli-explain.md) | 1 | Look up knossos domain concept definitions |
| [complaint](cli-complaint.md) | 3 | View and manage Cassandra complaint artifacts |
| [init](cli-init.md) | 1 | Scaffold the channel directory for a new project |
| [tour](cli-tour.md) | 1 | Directory tree with file counts |
| [version](cli-version.md) | 1 | Binary version and build metadata |
| [help](cli-help.md) | 1 | Help for any command |
| [sails](cli-sails.md) | 1 | White Sails quality gates |
| [naxos](cli-naxos.md) | 1 | Orphaned session cleanup |
| [tribute](cli-tribute.md) | 1 | Session summary generation |
| [completion](cli-completion.md) | 4 | Shell autocompletion |

**Total**: 32 command families (all documented above)

---

## Installation

The `ari` binary is built from Go source:

```bash
CGO_ENABLED=0 go build ./cmd/ari
```

Install to PATH:
```bash
go install ./cmd/ari
```

---

## Global Flags

All commands support these flags:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `$XDG_CONFIG_HOME/knossos/config.yaml` | Config file path |
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
- `ari sync materialize` — Generate channel directory
- `ari inscription sync` — Update context file
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

- [Worktree Guide](../guides/worktree-guide.md) — Parallel session patterns
- [Ariadne Glossary Entry](../../reference/GLOSSARY.md#ariadne)

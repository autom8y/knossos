# Knossos v0.1.0 Release Notes

**Release Date**: 2026-01-07

## Overview

This is the initial release of **Knossos**, a platform for orchestrating Claude Code workflows through rites, sessions, and the Ariadne CLI (`ari`).

Knossos represents a major consolidation and restructure of the roster project, implementing a clean architecture with proper Go module structure, rite-centric content organization, and automated materialization.

## Highlights

### Ariadne CLI (`ari`)

The Ariadne CLI is the "thread through the labyrinth" - providing:

- **Session Management**: Create, park, resume, and wrap work sessions
- **Rite Operations**: Switch between practice bundles (rites) with agent routing
- **Sync/Materialization**: Generate `.claude/` directory from templates and rites
- **Hook Support**: Context injection for Claude Code workflows
- **Artifact Registry**: Track and query session artifacts

### Architecture Changes

- **Go Module at Repository Root**: Module path `github.com/autom8y/knossos` with standard `cmd/ari` + `internal/` layout
- **Rite-Centric Content Organization**: Each rite has a `manifest.yaml` defining agents, skills, phases, and dependencies
- **Materialization Model**: `.claude/` is now fully generated via `ari sync materialize`, not committed to git
- **12 Rites Available**: 10x-dev, debt-triage, docs, ecosystem, forge, hygiene, intelligence, rnd, security, shared, sre, strategy

## Installation

### Homebrew (macOS/Linux)

```bash
brew install autom8y/tap/ari
```

### Go Install

```bash
go install github.com/autom8y/knossos/cmd/ari@v0.1.0
```

### Binary Download

Download the appropriate archive for your platform from the [Releases page](https://github.com/autom8y/knossos/releases/tag/v0.1.0).

## ADRs Accepted

This release implements the following Architecture Decision Records:

| ADR | Title | Summary |
|-----|-------|---------|
| ADR-0013 | CLI Distribution Strategy | GoReleaser + Homebrew tap |
| ADR-0014 | Go Module Structure | `cmd/ari` + `internal/` at repo root |
| ADR-0015 | Content Organization | Rite-centric structure with manifests |
| ADR-0016 | Sync and Materialization | chezmoi-inspired generation model |
| ADR-0017 | Hook Architecture | Thin shell wrapper + ari subcommand |

## Breaking Changes

This release includes significant breaking changes from the prior roster structure:

1. **Module Path Changed**: `github.com/autom8y/ariadne` → `github.com/autom8y/knossos`
2. **Directory Structure Changed**: Go code moved from `ariadne/` subdirectory to repository root
3. **`.claude/` Now Generated**: Previously committed, now gitignored and generated via `ari sync materialize`
4. **Rite Switching**: `swap-rite.sh` still works but `ari sync materialize --rite <name>` is the new approach

## Quick Start

```bash
# Install ari
brew install autom8y/tap/ari

# Clone and set up knossos
git clone https://github.com/autom8y/knossos.git
cd knossos

# Generate .claude/ for your desired rite
ari sync materialize --rite 10x-dev

# Start a session
ari session create "my-feature" SMALL

# Check session status
ari session status
```

## What's Next

- v0.2.0: Three-way merge for conflict resolution in sync
- v0.3.0: Remote sync from upstream Knossos
- v1.0.0: Stable API and full documentation

## Contributors

- Claude Opus 4.5 (AI pair programmer)

## Links

- [Documentation](https://github.com/autom8y/knossos/tree/main/docs)
- [Ariadne CLI Guide](https://github.com/autom8y/knossos/blob/main/docs/guides/ariadne-cli.md)
- [Knossos Doctrine](https://github.com/autom8y/knossos/blob/main/docs/philosophy/knossos-doctrine.md)

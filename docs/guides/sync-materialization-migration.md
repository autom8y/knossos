# Sync Materialization Migration Guide

## Overview

As of 2026-01-07, Knossos has migrated from checking `.claude/` into git to a **materialization model** where `.claude/` is fully generated from templates and rite manifests.

This document explains:
1. How the new sync system works
2. How to use `ari sync materialize`
3. Migration path from the old model

## What Changed

### Before (Old Model)
- `.claude/` directory was checked into git
- Rite switching used `swap-rite.sh` bash script
- Manual coordination required for updates

### After (New Model)
- `.claude/` is generated from `templates/` and `.knossos/rites/{name}/`
- `.claude/` is gitignored (not tracked)
- Single `ari sync materialize` command regenerates everything
- Idempotent and safe to run multiple times

## Architecture

```
┌───────────────────────────────────────────────────────────────┐
│                        TEMPLATE SOURCES                        │
├───────────────────────────────────────────────────────────────┤
│  .knossos/rites/{active}/      │  rites/shared/     │  templates/      │
│  ├── agents/          │  └── skills/       │  └── hooks/      │
│  ├── skills/          │                    │  (future)        │
│  └── manifest.yaml    │                    │                  │
└───────────┬───────────┴─────────┬──────────┴────────┬─────────┘
            │                     │                   │
            └─────────────────────┴───────────────────┘
                                  │
                                  v
                    ┌─────────────────────────┐
                    │  ari sync materialize   │
                    └────────────┬────────────┘
                                 │
                                 v
            ┌────────────────────────────────────────┐
            │              .claude/                   │
            │  ├── agents/       (from rite)         │
            │  ├── skills/       (from rite+shared)  │
            │  ├── hooks/        (preserved)         │
            │  ├── CLAUDE.md     (generated)         │
            │  └── sync/                             │
            │      └── state.json (tracking)         │
            └────────────────────────────────────────┘
```

## Using `ari sync materialize`

### Basic Usage

```bash
# Generate .claude/ from current ACTIVE_RITE
ari sync materialize

# Generate .claude/ for a specific rite
ari sync materialize --rite ecosystem

# Force regeneration (overwrite local changes)
ari sync materialize --force
```

### What It Does

The materialize command:

1. **Reads** the rite manifest from `.knossos/rites/{name}/manifest.yaml`
2. **Copies** agent files from `.knossos/rites/{name}/agents/` → `.claude/agents/`
3. **Merges** skills from:
   - `.knossos/rites/{name}/skills/`
   - `rites/shared/skills/` (from dependencies)
   - Into `.claude/skills/`
4. **Preserves** hooks in `.claude/hooks/` (until templates/hooks exists)
5. **Generates** CLAUDE.md using the inscription system
6. **Tracks** state in `.knossos/sync/state.json`
7. **Writes** ACTIVE_RITE marker

### Idempotency

Running `ari sync materialize` multiple times is safe:
- It regenerates all managed content
- Preserves user customizations in satellite regions (CLAUDE.md)
- Maintains state tracking

## Migration from Old Model

### For Existing Installations

1. **Backup** (recommended):
   ```bash
   cp -r .claude .claude.backup
   ```

2. **Run materialization**:
   ```bash
   ari sync materialize
   ```

3. **Verify** the generated content:
   ```bash
   ls .claude/agents/
   ls .claude/skills/
   cat .claude/CLAUDE.md
   ```

4. **Confirm** ACTIVE_RITE is correct:
   ```bash
   cat .knossos/ACTIVE_RITE
   ```

### Git Configuration

The `.gitignore` already contains `.claude/`, and `.claude/` has been removed from git tracking in commit:

```bash
git status  # Should show D .claude/* for deletion of tracked files
```

This is expected and correct. The `.claude/` directory still exists on disk (gitignored).

## Rite Manifest Format

Each rite requires a `manifest.yaml`:

```yaml
name: ecosystem
version: "1.0.0"
description: Ecosystem infrastructure lifecycle

agents:
  - name: orchestrator
    role: Coordinates ecosystem infrastructure phases
  - name: ecosystem-analyst
    role: Traces CEM/roster problems to root causes

skills:
  - claude-md-architecture
  - doc-ecosystem
  - ecosystem-ref

dependencies:
  - shared

hooks: []
```

## State Tracking

Materialization tracks state in `.knossos/sync/state.json`:

```json
{
  "schema_version": "1.0",
  "remote": "local:ecosystem",
  "last_sync": "2026-01-07T16:55:37Z",
  "tracked_files": {}
}
```

This enables:
- Conflict detection (future)
- Change tracking
- Remote sync capabilities (future)

## Future Enhancements

Per ADR-0016 (Sync and Materialization Model):

1. **Three-Way Merge**: Track base state to detect local and remote changes
2. **Conflict Resolution**: Explicit conflict handling with `ari sync resolve`
3. **Template Hooks**: Move hooks from .claude/hooks to templates/hooks
4. **Remote Sync**: Pull updates from upstream Knossos
5. **Rite Switching**: Integrated `ari rite switch` command

## Troubleshooting

### Issue: "no ACTIVE_RITE found"

**Solution**: Specify the rite explicitly:
```bash
ari sync materialize --rite ecosystem
```

### Issue: Missing agents or skills

**Cause**: Rite manifest may not list all agents/skills.

**Solution**: Check `.knossos/rites/{name}/manifest.yaml` and ensure all agents are listed.

### Issue: CLAUDE.md sections missing

**Cause**: KNOSSOS_MANIFEST.yaml may be missing or incomplete.

**Solution**: Ensure `.knossos/KNOSSOS_MANIFEST.yaml` exists with region definitions.

### Issue: "failed to load rite manifest"

**Cause**: The rite directory or manifest.yaml is missing.

**Solution**: Verify `.knossos/rites/{name}/manifest.yaml` exists and is valid YAML.

## References

- **ADR-0016**: Sync and Materialization Model
- **ADR-0015**: Content Organization and Structure
- **SPIKE-materialization-model.md**: Research findings
- **TDD-ariadne-sync.md**: Technical design for sync commands

## Implementation Details

The materialization system is implemented in:
- `internal/materialize/materialize.go` - Core materialization engine
- `internal/cmd/sync/materialize.go` - CLI command
- `internal/inscription/` - CLAUDE.md generation
- `internal/sync/` - State tracking

### Sprig Integration

The inscription system now uses Sprig template functions (100+ utility functions):
- `internal/inscription/generator.go` includes Sprig via `sprig.TxtFuncMap()`
- Compatible with Helm/Kubernetes templating patterns

## Summary

The materialization model provides:
- ✅ Single source of truth (templates + rites)
- ✅ Idempotent operations
- ✅ Clean rite switching
- ✅ Version control friendly (only source files tracked)
- ✅ Future-proof for remote sync

Run `ari sync materialize` to regenerate `.claude/` any time.

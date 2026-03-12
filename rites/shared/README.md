# Shared Skills Rite

> Cross-rite primitives shared across all knossos rites

## Purpose

The shared rite provides common skills, hooks, and commands that are universally available regardless of which rite is currently active. This eliminates duplication and ensures consistency across all rites.

## Structure

```
rites/shared/
├── skills/      # Cross-rite skill definitions
├── hooks/       # Shared lifecycle hooks
└── commands/    # Common command definitions
```

**Important**: This rite contains NO agents/ directory and NO ACTIVE_WORKFLOW.yaml. The shared rite is skills-only infrastructure.

## Sync Behavior

`ari sync --rite` automatically syncs shared rite content alongside the active rite:
- Skills, hooks, and commands are synced to the channel skills, hooks, and commands directories
- Shared content is always available regardless of active rite
- Rite-specific content takes precedence over shared content (rite-privileged override)

## Runtime Behavior

At runtime, shared skills are flattened into the channel skills directory with no subdirectory nesting. The skill `rites/shared/skills/foo.md` appears as `{channel_dir}/skills/foo.md`.

## Override Resolution

When both shared and rite-specific content exist with the same name:
1. Rite-specific version wins (rite-privileged)
2. Shared version is ignored during sync
3. No merge or conflict resolution occurs

This ensures rites can override shared primitives with specialized versions.

## Tracking

Shared skill sync state is tracked via a `.shared-skills` marker file in the channel directory, separate from the `.rite-skills` marker used for active rite content.

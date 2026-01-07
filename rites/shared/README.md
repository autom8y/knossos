# Shared Skills Team

> Cross-rite primitives shared across all roster teams

## Purpose

The shared team provides common skills, hooks, and commands that are universally available regardless of which team is currently active. This eliminates duplication and ensures consistency across all rites.

## Structure

```
rites/shared/
├── skills/      # Cross-rite skill definitions
├── hooks/       # Shared lifecycle hooks
└── commands/    # Common command definitions
```

**Important**: This team contains NO agents/ directory and NO ACTIVE_WORKFLOW.yaml. The shared team is skills-only infrastructure.

## Sync Behavior

`swap-rite.sh` automatically syncs shared team content alongside the active rite:
- Skills, hooks, and commands are synced to `.claude/skills/`, `.claude/hooks/`, `.claude/commands/`
- Shared content is always available regardless of active rite
- Team-specific content takes precedence over shared content (rite-privileged override)

## Runtime Behavior

At runtime, shared skills are flattened into `.claude/skills/` with no subdirectory nesting. The skill `rites/shared/skills/foo.md` appears as `.claude/skills/foo.md`.

## Override Resolution

When both shared and rite-specific content exist with the same name:
1. Team-specific version wins (rite-privileged)
2. Shared version is ignored during sync
3. No merge or conflict resolution occurs

This ensures teams can override shared primitives with specialized versions.

## Tracking

Shared skill sync state is tracked via `.claude/.shared-skills` marker file, separate from the `.claude/.rite-skills` marker used for active rite content.

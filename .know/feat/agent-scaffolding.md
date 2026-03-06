---
domain: feat/agent-scaffolding
generated_at: "2026-03-03T21:30:00Z"
expires_after: "14d"
source_scope:
  - "./internal/agent/**/*.go"
  - "./internal/cmd/agent/**/*.go"
  - "./docs/decisions/ADR-0024*.md"
  - "./.know/architecture.md"
generator: theoros
source_hash: "18042fc"
confidence: 0.88
format_version: "1.0"
---

# Agent Scaffolding and Factory

## Purpose and Design Rationale

Replaces freeform agent authoring with structured scaffolding, validation, and platform section updates. Prior: 61+ agents as inconsistent markdown, no schema validation, no type classification, 2-4 hours per agent.

**ADR-0024**: "Smart Authoring, Simple Materialization" — intelligence in CLI commands, materializer copies verbatim. Section ownership (platform/author/derived). Warning-based MCP validation. `additionalProperties: true` in JSON Schema. Two-tier validation (WARN/STRICT).

## Conceptual Model

### Three Archetypes

| Archetype | Model | Tools | Color | MaxTurns |
|-----------|-------|-------|-------|----------|
| `orchestrator` | opus | Read | purple | 40 |
| `specialist` | opus | Bash,Glob,Grep,Read,Edit,Write,TodoWrite,Skill | orange | 150 |
| `reviewer` | opus | +WebFetch,WebSearch | red | 100 |

Types `meta`, `designer`, `analyst`, `engineer` fall back to `specialist` archetype.

### Section Ownership

| Ownership | On Scaffold | On `ari agent update` |
|-----------|-------------|----------------------|
| `platform` | Pre-populated from template | Regenerated |
| `author` | `<!-- TODO: hint -->` | Preserved |
| `derived` | Generated from frontmatter | Regenerated |

### Three Workflows

1. **Authoring**: `ari agent new` → `GetArchetype()` → `ScaffoldAgent()` → write
2. **Validation**: `ari agent validate` → Phase 1 (parse) → Phase 2 (JSON Schema) → Phase 3 (semantic)
3. **Update**: `ari agent update` → `ParseAgentSections()` → `RegeneratePlatformSections()` → `AssembleAgentFile()`

## Implementation Map

16 files in `/Users/tomtenuta/Code/knossos/internal/agent/`. 6 files in `internal/cmd/agent/`. 3 embedded `.md.tpl` archetype templates.

### Key Types

`AgentFrontmatter`, `Archetype`, `SectionOwnership`, `ParsedAgent`, `ValidationMode`, `MemoryField`, `FlexibleStringSlice`.

**Critical boundary**: `internal/materialize` does NOT import `internal/agent`. Materializer copies agent files verbatim.

## Boundaries and Failure Modes

- Does NOT run during `ari sync` (authoring-time only)
- Does NOT enforce strict validation on existing agents (opt-in via `--strict`)
- Type fallback is silent (analyst → specialist without warning)
- `ari agent update` uses `os.WriteFile` not `AtomicWriteFile`
- `ValidationMode` uses `iota` (violates typed string constant convention)

## Knowledge Gaps

1. Archetype template content (`.md.tpl` files) not read.
2. `agent.schema.json` conditional rules not confirmed against current schema.
3. Lint integration uses raw file inspection, not `internal/agent` parsing.

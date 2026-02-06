# ADR-0024: Agent Factory -- Structured Agent Authoring with Schema Validation

| Field | Value |
|-------|-------|
| **Status** | Accepted |
| **Date** | 2026-02-06 |
| **Deciders** | Architecture Team |
| **Supersedes** | N/A (new capability; replaces handcrafted freeform agent workflow) |
| **Superseded by** | N/A |

## Context

Knossos manages 61 agents across 12 rites. Prior to this decision, agents were authored as freeform markdown files with minimal frontmatter (`name`, `description`, `tools` as a comma-separated string, `model`, `color`). There was no schema validation, no type classification, no standard section structure, and no behavioral contracts. This created several problems:

1. **No schema validation.** Agent frontmatter was parsed by the materializer for copying, but never validated. Typos in tool names, invalid model values, and missing required fields were only discovered at runtime when Claude Code rejected the agent.

2. **Inconsistent structure.** Each agent author invented their own section layout. Some agents had 11 sections, others had 5. Platform-wide concerns (behavioral constraints, anti-patterns, acid tests) were copy-pasted between agents and diverged over time.

3. **No type classification.** Orchestrators, specialists, and reviewers are fundamentally different patterns with different tool requirements, section layouts, and behavioral expectations. Without formal types, there was no way to enforce these patterns or scaffold new agents correctly.

4. **No workflow metadata.** Agents had no way to declare their position in a workflow (upstream sources, downstream targets, produced artifacts). Orchestrators had to describe routing in prose rather than structured data.

5. **No MCP integration.** The `tools` field only supported Claude Code built-in tools. Agents that needed MCP server tools had no way to declare the dependency, and rite manifests had no way to carry MCP server configuration to satellites.

6. **Agent creation took 2-4 hours.** Authors started from scratch or copy-pasted from existing agents, manually adapting each section. Platform-owned sections (constraints, anti-patterns) were frequently forgotten or outdated.

### Prior Art

- **SPIKE**: `docs/spikes/SPIKE-agent-factory.md` -- explored the full agent lifecycle, recommended building on existing `internal/materialize` and `internal/validation` infrastructure rather than a standalone system
- **Forge rite agents**: The forge rite's 7 agents (prompt-architect, agent-designer, etc.) represent the manual agent-creation workflow this system replaces

## Decision

Implement the Agent Factory as an extension of the existing validation and materialization infrastructure. The system comprises five components: a YAML frontmatter schema, a two-tier validation engine, an archetype system with section ownership, CLI tooling for agent lifecycle, and MCP server integration.

### 1. Agent Frontmatter Schema

Define `agent.schema.json` (JSON Schema draft 2020-12) in `internal/validation/schemas/`. The schema declares:

- **Required fields**: `name` (kebab-case pattern `^[a-z][a-z0-9-]*$`), `description` (minimum 10 characters)
- **Optional enhanced fields**: `role`, `type`, `tools`, `model`, `color`, `aliases`, `upstream`, `downstream`, `produces`, `contract`, `schema_version`
- **`additionalProperties: true`** for forward compatibility -- new fields can be added to agents without schema changes
- **Conditional rules**: orchestrators must use `opus` model; reviewers must have `contract.must_not` when `type` is set

The Go-side struct is `AgentFrontmatter` in `internal/agent/frontmatter.go`, separate from `MenaFrontmatter` in `internal/materialize/frontmatter.go`. Agent and mena are different content types with different schemas; coupling them would conflate their evolution.

### 2. Two-Tier Validation

Two validation modes serve different lifecycle stages:

| Mode | Required Fields | Enhanced Fields | Archetype Rules | Use Case |
|------|----------------|-----------------|-----------------|----------|
| **WARN** | `name`, `description` (error) | Warnings only | Warnings only | Existing agents, gradual migration |
| **STRICT** | `name`, `description` (error) | `type`, `tools` (error) | Enforced as errors | New agents, CI gates, post-migration |

Validation runs in three phases:
1. **Parse**: Extract YAML frontmatter from markdown, fail on missing delimiters or malformed YAML
2. **Schema**: Validate against `agent.schema.json` via the existing `internal/validation/Validator` infrastructure
3. **Semantic**: Go-level checks beyond JSON Schema -- tool reference validation against known tools, archetype-specific constraints, MCP server cross-references

### 3. Archetypes and Section Ownership

Three archetypes define the standard structure for agent categories:

| Archetype | Purpose | Default Model | Default Tools | Default Color |
|-----------|---------|---------------|---------------|---------------|
| `orchestrator` | Consultative coordinator; routes work, does not execute | `opus` | `Read` | `purple` |
| `specialist` | Domain expert; executes focused work, produces artifacts | `opus` | `Bash`, `Glob`, `Grep`, `Read`, `Edit`, `Write`, `TodoWrite`, `Skill` | `orange` |
| `reviewer` | Quality gate; evaluates work, approve/reject decisions | `opus` | `Bash`, `Glob`, `Grep`, `Read`, `Edit`, `Write`, `WebFetch`, `WebSearch`, `TodoWrite`, `Skill` | `red` |

Each archetype defines an ordered list of sections with ownership designations:

| Ownership | Meaning | On Scaffold | On Update |
|-----------|---------|-------------|-----------|
| **Platform** | Default content from archetype template. Authors should not modify. | Populated from template | Regenerated from template |
| **Author** | Must be filled in by the agent author. Domain-specific content. | `<!-- TODO: hint -->` markers | Preserved exactly as-is |
| **Derived** | Generated from frontmatter data (tools, upstream/downstream). | Generated from frontmatter | Regenerated from frontmatter |

Seven agent type values are supported: `orchestrator`, `specialist`, `reviewer`, `meta`, `designer`, `analyst`, `engineer`. Types without their own archetype template (`meta`, `designer`, `analyst`, `engineer`) fall back to the `specialist` archetype for section layout and defaults.

### 4. CLI Tooling

Four commands under `ari agent`:

| Command | Purpose |
|---------|---------|
| `ari agent new --archetype TYPE --rite NAME --name NAME` | Scaffold new agent from archetype template |
| `ari agent validate [path...] [--strict] [--rite NAME] [--all]` | Validate frontmatter against schema and semantic rules |
| `ari agent list [--rite NAME]` | List agents with type, model, source, description |
| `ari agent update [path...] [--rite NAME] [--all] [--dry-run]` | Regenerate platform/derived sections, preserve author content |

### 5. MCP Server Integration

MCP integration operates at two levels:

**Rite manifest level**: The `RiteManifest` struct carries `MCPServers []MCPServer` with `yaml:"mcp_servers"`. During materialization, `mergeMCPServers()` performs union merge into `settings.local.json` -- rite servers are added/updated, existing satellite servers are preserved.

**Agent tool level**: Agents declare MCP tools via the `mcp:<server>[/<method>]` pattern in the `tools` field. `ValidateAgentMCPReferences()` cross-references these against the rite manifest's declared servers. Undeclared servers produce **warnings, not errors**, because servers may be satellite-provided.

### 6. FlexibleStringSlice

The `tools` field accepts both formats:
```yaml
tools: Bash, Read, Glob              # comma-separated string
tools:                                 # YAML array
  - Bash
  - Read
  - Glob
```

`FlexibleStringSlice` implements `yaml.Unmarshaler` to handle both cases. This preserves backward compatibility with the 61 existing agents that use comma-separated format while supporting the cleaner array format for new agents.

## Key Design Decisions

### "Smart Authoring, Simple Materialization"

Intelligence lives in CLI commands (`ari agent new`, `ari agent validate`, `ari agent update`), not in the materializer. Source agent files are always complete, valid markdown. The materializer continues to copy agents verbatim -- no build-time composition, no template rendering during sync. This means:

- Agents are readable and editable without running any tooling
- The materializer remains simple and fast
- Validation is a separate concern from distribution

**Rationale**: The spike document recommended against YAML-to-markdown generation at materialization time because it couples the critical sync path to template rendering and schema validation. Failures in agent generation should not break `ari sync materialize`.

### Section Ownership Enables Platform Evolution

Platform-owned sections (behavioral constraints, anti-patterns, acid tests) can be updated across all agents via `ari agent update` without touching author content. This solves the copy-paste divergence problem: when platform guidance changes, one template update plus one `ari agent update --all` propagates the change to all agents.

**Rationale**: Without ownership boundaries, any template update risks overwriting hand-authored content. The ownership model makes updates safe by design.

### Warning-Based MCP Validation

MCP server references produce warnings, not errors, when the referenced server is not found in the rite manifest. This is because satellites may provide their own MCP servers not declared in the rite manifest.

**Rationale**: Hard-failing on undeclared MCP servers would make it impossible for satellites to extend agents with project-specific MCP integrations. The warning alerts authors to potential issues while allowing legitimate satellite-provided servers.

### `additionalProperties: true`

The JSON Schema allows unknown fields in agent frontmatter. This enables forward compatibility -- agents can carry fields that the current schema version does not define, and they will be preserved through parse/serialize cycles.

**Rationale**: A strict schema would force all agents to update every time a new field is added. With `additionalProperties: true`, schema evolution is additive and non-breaking.

### Archetype Fallback for Non-Standard Types

Types `meta`, `designer`, `analyst`, and `engineer` are valid type classifications but do not have their own archetype templates. They fall back to the `specialist` archetype for section layout and defaults. This allows fine-grained type classification without requiring a template for every type.

**Rationale**: Creating archetype templates for types that differ only in semantics (not structure) would create maintenance burden with no structural benefit. The specialist template covers the common case; the type field carries the semantic distinction.

## Consequences

### Positive

1. **Agent creation time drops from 2-4 hours to 30 minutes.** `ari agent new` scaffolds a complete agent with platform sections pre-populated. Authors only fill in domain-specific (author-owned) sections.

2. **Schema validation catches errors at authoring time.** Invalid tool names, bad model values, missing required fields, and malformed frontmatter are caught by `ari agent validate` before materialization.

3. **Platform sections stay current across all agents.** `ari agent update --all` regenerates platform-owned and derived sections from templates without touching author content. One template change propagates everywhere.

4. **MCP tools are first-class.** Agents can declare MCP tool dependencies. Rite manifests carry MCP server configuration that is materialized into satellite `settings.local.json`.

5. **Workflow position is machine-readable.** Upstream/downstream declarations and artifact production enable future tooling: dependency graphs, workflow visualization, dead-agent detection.

6. **CI-ready validation.** `ari agent validate --strict` returns exit code 1 on failure, suitable for CI pipeline gates.

7. **Behavioral contracts formalize agent expectations.** `must_use`, `must_produce`, `must_not`, and `max_turns` make agent constraints explicit and validatable.

### Negative

1. **Additional frontmatter complexity.** The enhanced schema has 12+ fields. Mitigated: only `name` and `description` are required; all enhanced fields are optional and added incrementally.

2. **Three archetype templates to maintain.** Template changes require running `ari agent update --all` to propagate. Mitigated: templates change infrequently, and the update command automates propagation.

3. **Section ownership model has a learning curve.** Authors must understand that platform/derived sections are regenerated and should not be customized. Mitigated: `<!-- TODO -->` markers and documentation make the boundary clear.

4. **FlexibleStringSlice adds parsing complexity.** Two input formats for the same field. Mitigated: the type is self-contained (single `UnmarshalYAML` method) and well-tested.

### Neutral

1. **Materialization unchanged.** Agents are still copied verbatim by `materializeAgents()`. No behavioral change to the sync pipeline.

2. **Existing agents continue to work.** WARN-mode validation accepts all existing agents without modification. Migration to enhanced frontmatter is optional and incremental.

3. **Agent source files remain plain markdown.** No build step, no compilation. Files are human-readable and editable in any text editor.

## Implementation

### Components Created

| File | Purpose |
|------|---------|
| `internal/agent/types.go` | `FlexibleStringSlice`, `UpstreamRef`, `DownstreamRef`, `ArtifactDecl`, `BehavioralContract` |
| `internal/agent/frontmatter.go` | `AgentFrontmatter` struct, `ParseAgentFrontmatter()`, `Validate()`, `MCPServers()` |
| `internal/agent/archetype.go` | `Archetype` definitions, `SectionOwnership` enum, `SectionDef`, archetype registry |
| `internal/agent/sections.go` | `ParseAgentSections()`, section parsing state machine, archetype mapping |
| `internal/agent/regenerate.go` | `RegeneratePlatformSections()`, `AssembleAgentFile()`, derived content generation |
| `internal/agent/scaffold.go` | `ScaffoldAgent()`, template rendering with Sprig functions |
| `internal/agent/validate.go` | `AgentValidator`, two-tier validation, archetype-specific rules |
| `internal/agent/mcp_validate.go` | `ValidateAgentMCPReferences()`, MCP server cross-referencing |
| `internal/agent/templates.go` | `//go:embed templates/*.md.tpl` |
| `internal/agent/templates/*.md.tpl` | Archetype templates (orchestrator, specialist, reviewer) |
| `internal/validation/schemas/agent.schema.json` | JSON Schema draft 2020-12 for agent frontmatter |
| `internal/cmd/agent/agent.go` | `ari agent` command group |
| `internal/cmd/agent/validate.go` | `ari agent validate` subcommand |
| `internal/cmd/agent/list.go` | `ari agent list` subcommand |
| `internal/cmd/agent/new.go` | `ari agent new` subcommand |
| `internal/cmd/agent/update.go` | `ari agent update` subcommand |

### Components Modified

| File | Change |
|------|--------|
| `internal/materialize/materialize.go` | Added `MCPServer` struct, `MCPServers` field to `RiteManifest`, `MCPServerNames()` method. Modified `materializeSettings()` to call `mergeMCPServers()` when manifest has MCP servers. |
| `internal/materialize/mcp.go` | New file. `mergeMCPServers()` with union merge semantics, `loadExistingSettings()`, `saveSettings()`. |
| `internal/cmd/root/root.go` | Registered `ari agent` command group. |

### Migration

All 61 existing agents were migrated to enhanced frontmatter in a big-bang execution. Each agent received a `type` classification and validated tools list. Platform sections were regenerated from archetype templates. Author sections were preserved.

## Alternatives Considered

### Alternative 1: YAML Source Files with Build-Time Generation

Agent specs as `.yaml` files in `rites/*/agents/`, with the materializer generating `.md` at sync time (the approach outlined in the spike). This was rejected because it couples the critical sync path to template rendering, makes agent source files unreadable without tooling, and requires maintaining both `.yaml` source and `.md` output formats.

### Alternative 2: Standalone Agent Factory Package

A separate `internal/agentgen/` package as recommended in the spike. This was rejected in favor of `internal/agent/` because the agent package needs tight integration with the validation infrastructure (`internal/validation/`) and the agent lifecycle is a single concern (parse, validate, scaffold, update) that belongs in one package.

### Alternative 3: Strict Schema from Day One

Require all enhanced fields (`type`, `tools`, `upstream`, `downstream`, `contract`) for all agents. This was rejected because it would force a big-bang migration of all 61 agents before any could be validated, making the transition risky. Two-tier validation allows incremental adoption.

### Alternative 4: Hard-Fail on Undeclared MCP Servers

MCP server references without manifest declarations produce errors instead of warnings. This was rejected because it prevents satellites from providing their own MCP servers without forking the rite manifest.

## Related Decisions

- **ADR-0021**: Two-Axis Context Model (unified commands/skills model; Agent Factory follows the same pattern of structured frontmatter with validation)
- **ADR-0023**: Dromena/Legomena Mena Convention (established the `MenaFrontmatter` pattern that `AgentFrontmatter` parallels)
- **ADR-0014**: ari CLI Sync (materialization infrastructure that carries MCP server configuration)
- **ADR-0009**: Knossos-Roster Identity (SOURCE/PROJECTION model for agent distribution)

## References

| Reference | Location |
|-----------|----------|
| Spike document | `docs/spikes/SPIKE-agent-factory.md` |
| Agent schema | `internal/validation/schemas/agent.schema.json` |
| Agent package | `internal/agent/` |
| CLI commands | `internal/cmd/agent/` |
| MCP integration | `internal/materialize/mcp.go` |
| CLI usage docs | `docs/cli/agent-commands.md` |
| Migration guide | `docs/guides/agent-migration-guide.md` |

## Version History

| Date | Author | Change |
|------|--------|--------|
| 2026-02-06 | Claude Code (Documentation Engineer) | Initial acceptance -- documenting Agent Factory implementation |

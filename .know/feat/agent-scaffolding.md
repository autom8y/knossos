---
domain: feat/agent-scaffolding
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/agent/**/*.go"
  - "./internal/cmd/agent/**/*.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.88
format_version: "1.0"
---

# Agent Scaffolding and Factory

## Purpose and Design Rationale

Solves structural authoring: correct frontmatter, sections, tool lists, and behavioral constraints per archetype. Three guarantees: authoring scaffolding (ari agent new from archetype), structural governance (ari agent validate with WARN/STRICT tiers), summonable agent lifecycle (summon/dismiss to ~/.claude/agents/). Three-ownership section model: platform (managed), author (preserved), derived (generated from frontmatter). tier field (standing/rite/summonable) controls materialization lifecycle, not capabilities. knossosOnlyFields stripped at two independent paths (materialize and summon).

## Conceptual Model

**Three archetypes:** orchestrator (Read-only, maxTurns 40, Task disallowed), specialist (broad tools, maxTurns 150), reviewer (Task disallowed, must define contract.must_not). 13 valid type values map to 3 archetype definitions. **Three tiers:** standing (always materialized), rite (per-sync), summonable (on-demand). **Two-tier validation:** Phase 1 (parse), Phase 2 (JSON Schema), Phase 3 (semantic: validateCore + archetype rules). **Section ownership:** platform (re-rendered on update), author (preserved), derived (generated).

## Implementation Map

`internal/agent/` (9 files): frontmatter.go (AgentFrontmatter, validateCore), archetype.go (3 archetypes, SectionDef), scaffold.go (ScaffoldAgent via Go templates), sections.go (ParseAgentSections, exact+prefix heading match), regenerate.go (RegeneratePlatformSections), validate.go (AgentValidator, WARN/STRICT), mcp_validate.go (MCP cross-reference). `internal/cmd/agent/` (8 files): new, validate, update, summon (collision check + knossos field strip + provenance), dismiss (provenance check), roster (3-section: standing/summoned/available), list, embody (delegates to internal/perspective). Materialize: agent_transform.go (7-step transform pipeline).

## Boundaries and Failure Modes

Silent tier mismatch (standing agents without files show no description). BehavioralContract.MaxTurns not surfaced as CC maxTurns. Summon collision checker degrades to noop on uninitialized projects. Non-standard types silently map to specialist. Template placeholder extraction for section content is fragile. Provenance non-fatal for summon/dismiss. ADR-0024 missing from disk.

## Knowledge Gaps

1. JSON Schema content not read
2. Archetype .md.tpl template content not read
3. internal/perspective package not fully traced
4. Skill policy application details not read

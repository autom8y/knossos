---
domain: feat/perspective-system
generated_at: "2026-03-26T19:10:59Z"
expires_after: "14d"
source_scope:
  - "./internal/perspective/**/*.go"
  - "./internal/cmd/agent/embody.go"
  - "./.know/architecture.md"
generator: theoros
source_hash: "b329d719"
confidence: 0.82
format_version: "1.0"
---

# Perspective / Viewpoint System

## Purpose and Design Rationale

The perspective system solves a specific observability problem in knossos: **agents cannot see their own configuration as knossos sees it**. Materialization deliberately strips knossos-only frontmatter fields (behavioral contracts, role, type, schema_version, upstream/downstream routing) before writing to the channel directory. This means neither the agent itself nor a human operator can inspect the full, unstripped configuration just by reading `.claude/agents/`.

The package's package comment articulates this clearly: it "assembles a first-person experiential view of an agent's context by resolving identity, capability, constraint, memory, and provenance layers from source files (not materialized output) to capture knossos-only fields stripped during materialization."

The design is built around three motivations:

1. **Introspection over source truth, not channel output.** The `ParseContext` reads from `rites/{name}/agents/` and rite manifests, not `.claude/`. This is the only way to see the full frontmatter.

2. **Consistency audit capability.** With full configuration visible, cross-layer checks become possible (e.g., a tool appearing in both `tools` and `disallowedTools` -- AUDIT-002). Without the perspective layer this check would require reading multiple source files manually.

3. **Capability simulation.** `RunSimulate` lets operators or agents ask "given a natural language prompt, what can this agent attempt?" before any actual invocation.

The system has no ADRs but its design rationale is embedded in the `SCAR-002` comment in `context.go`: `.knossos/` paths vs `.claude/` paths were a known confusion point.

## Conceptual Model

The perspective system models an agent as nine named layers, numbered L1-L9. Each layer is a distinct facet of the agent's operational reality:

| Layer | Name | Content |
|-------|------|---------|
| L1 | Identity | Name, description, role, type, model, color, schema_version, system prompt body |
| L2 | Perception | Skill awareness: explicit, policy-injected, policy-referenced, on-demand; Skill tool availability |
| L3 | Capability | Tools, MCP tools, hooks; REPLACE semantics vs manifest agent_defaults |
| L4 | Constraint | DisallowedTools, write-guard (3-tier cascade), behavioral contract (must_use, must_produce, must_not) |
| L5 | Memory | Scope (user/project/local), seed file existence + line count, runtime memory accessibility |
| L6 | Position | Workflow phase, predecessor/successor, entry-point/entry-agent flags, back-routes, complexity gates |
| L7 | Surface | Dromena owned, legomena available, artifact types, rite commands |
| L8 | Horizon | Negative space: tools not available, unreachable skills, phases not in, memory blind spots |
| L9 | Provenance | Owner, scope, checksum, last sync time, divergence from materialized file |

**Topological dependency order** in `Assemble()`: L1, L3, L4, L5, L6, L7, L9 are independent (resolved first). L2 depends on L3/L4. L8 is inverse computation over all other layers -- resolved last.

Each layer is wrapped in a `LayerEnvelope` carrying `LayerStatus` (RESOLVED, PARTIAL, OPAQUE, FAILED) and a `Gaps` list.

**Three operational modes:** default (all layers), audit (layers + 11 checks across phases 1-3), simulate (layers + keyword-match against prompt).

**Audit check taxonomy:** Phase 1 (AUDIT-001-006): single-layer consistency. Phase 2 (AUDIT-007-010): cross-layer consistency. Phase 3 (AUDIT-011): reachability.

## Implementation Map

**Entry point:** `ari agent embody <agent-name>` at `internal/cmd/agent/embody.go`.

**Core files:**

| File | Role |
|------|------|
| `internal/perspective/context.go` | `ParseContext` struct, rite resolution, agent source parsing, manifest loading |
| `internal/perspective/assemble.go` | `Assemble()` orchestrator, topological layer resolution |
| `internal/perspective/resolvers.go` | All 9 layer resolver functions (570+ lines) |
| `internal/perspective/audit.go` | `RunAudit()` + 11 audit check functions |
| `internal/perspective/simulate.go` | `RunSimulate()` + keyword-to-tool mapping (34 entries) |
| `internal/perspective/types.go` | All types: PerspectiveDocument, LayerEnvelope, 12 layer-specific data types |
| `internal/perspective/perspective_test.go` | Integration tests with TempDir fixtures |

**Key design decisions:**
- `resolvers.go` replicates `knownChannelTools` map (unexported in agent package) -- deliberate package boundary decision
- `resolvers.go` replays skill policy evaluation in read-only mode, replicating `internal/materialize/skill_policies.go` without importing it
- `context.go` hardcodes `.claude` channel dir (SCAR-002/HA-FS annotation)

## Boundaries and Failure Modes

**Package boundary:** Imports `internal/agent`, `internal/checksum`, `internal/errors`, `internal/frontmatter`, `internal/provenance`. Does NOT import `internal/materialize` or `internal/cmd/*` -- intentionally read-only observer.

**Failure modes:**
- Provenance manifest not found: L9 returns StatusPartial with MISSING gap
- Project-scope memory path: L5 returns PARTIAL + OPAQUE gap (CC path-hashing algorithm not exposed)
- L2 depends on L3/L4: if either fails, L2 is forced to StatusFailed
- Workflow/orchestrator YAML missing: L6 marks InWorkflow=false (may cause AUDIT-008 false positive)
- Rite source directory not found: hard error, no partial document
- `.claude/` hardcoded: Gemini-only projects resolve against wrong directory

**Out-of-scope limitations:**
- No support for embedded (platform) rites -- only checks `rites/` and `.knossos/rites/`
- `ArchetypeSource` field always nil (MVP decision)
- Runtime memory content accessible only for user/local scope

## Knowledge Gaps

1. `resolvers.go` tail (L6/L7/L8 resolver bodies) not fully read
2. Full test suite coverage for audit checks and simulate mode unknown
3. Skill policy replication fidelity vs `internal/materialize/skill_policies.go` not verified
4. Embedded rite fallback not handled
5. MCP server wiring check implementation details not fully read

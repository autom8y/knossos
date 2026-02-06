# Spike: Agent Factory -- From Handcrafted Agents to a Go-Native Generation Pipeline

**Date**: 2026-02-06
**Status**: COMPLETE
**Scope**: Architecture exploration for programmatic agent definition, generation, validation, and distribution across the Knossos platform
**Upstream**: MOONSHOT-agent-template-ecosystem.md (scenario analysis, migration phases)

---

## Executive Summary

Two deep exploration passes mapped the full agent lifecycle from source definition through materialization to satellite projection. The key finding is that Knossos already possesses 80% of the infrastructure needed for an Agent Factory: YAML frontmatter parsing, JSON Schema validation with embedded schemas, template-driven inscription, extension-based routing, and a multi-source materialization pipeline. The critical missing pieces are (1) a formal agent schema enforcing structural contracts, (2) a Go-native generation pipeline replacing the assumed shell-script approach from the MOONSHOT doc, and (3) MCP integration, where the codebase has zero implementation but the settings merge infrastructure is ready to carry `mcpServers` configuration. The recommended path is to build the Agent Factory as an extension of the existing `internal/materialize` and `internal/validation` packages rather than a standalone system.

---

## Question / Context

The MOONSHOT-agent-template-ecosystem document (512 lines, 6-phase roadmap) laid out a 3-year vision for moving from handcrafted agents to a fully automated Agent Factory. That document assumed shell-script-based generation, pre-dated the dromena/legomena mena convention (ADR-0023), and did not account for the Go-native systems that now exist (`internal/materialize`, `internal/validation`, `internal/inscription`). This spike answers:

1. **How would YAML-defined agents work** given the existing frontmatter parsing and materialization system?
2. **What's the behavioral contract model** that enables CI validation of generated agents?
3. **How does distribution to satellites work** beyond the current file-copy materialization?
4. **What's the MCP integration story** for agent-as-service and satellite tool configuration?

---

## Approach

Two-pass codebase exploration:

- **Pass 1**: Mapped the current agent architecture end-to-end. Examined all 58 agents across 12 rites, the rite manifest format, the materialization pipeline (`internal/materialize/materialize.go`), the inscription system (`internal/inscription/`), and the user sync system (`internal/usersync/`). Cataloged the forge rite's 7 agents as the existing manual agent-creation workflow.

- **Pass 2**: Investigated MCP integration readiness, the mena convention, JSON Schema infrastructure, and distribution mechanisms. Confirmed zero MCP integration but identified the settings infrastructure as MCP-ready. Mapped existing validation patterns (`internal/validation/validator.go` with 9 embedded schemas) as the foundation for agent schema validation.

---

## Findings

### 1. YAML-Defined Agents: Schema Design and Generation Pipeline

#### Current State

Agents are 11-section markdown files with YAML frontmatter:

```yaml
---
name: technology-scout
role: "Evaluates emerging technologies for competitive advantage"
description: "Technology horizon specialist who evaluates..."
tools: Glob, Grep, Read, Write, WebSearch, WebFetch, TodoWrite, Skill
model: opus
color: orange
---
```

The frontmatter is parsed by `internal/materialize/frontmatter.go` but only for mena (commands/skills), not for agents. Agent files are copied verbatim by `materializeAgents()` in `materialize.go` -- no parsing, no validation, no transformation.

#### Proposed Agent Schema

Extend the frontmatter convention to carry the full agent specification:

```yaml
---
# Identity (required)
name: technology-scout
role: "Evaluates emerging technologies for competitive advantage"
description: "Technology horizon specialist..."
type: specialist          # orchestrator | specialist | reviewer

# Capabilities (required)
tools:
  - Glob
  - Grep
  - Read
  - Write
  - WebSearch
model: opus               # opus | sonnet | haiku

# Workflow Position (optional, enables generation of Position section)
upstream:
  - source: user
  - source: orchestrator
downstream:
  - agent: integration-researcher
    condition: "recommendation in [Adopt, Trial, Assess]"
produces:
  - artifact: tech-assessment
    format: markdown

# Behavioral Contract (optional, enables CI validation)
contract:
  must_use_tools: [WebSearch]
  must_produce: [tech-assessment]
  must_not: [modify source code, commit changes]
  max_turns: 50

# Composition (optional, enables template-driven generation)
extends: specialist-base
sections:
  domain_authority:
    decides: [maturity rating, risk rating, recommendation verdict]
    escalates: [strategic bets, build vs buy > $50K]
    routes_to: [integration-researcher]
---

# Technology Scout
[... markdown body ...]
```

#### Why This Works With Existing Infrastructure

The `ParseYAMLFrontmatter()` function in `internal/validation/validator.go` already handles extraction. The `MenaFrontmatter` struct in `internal/materialize/frontmatter.go` provides the pattern for a new `AgentFrontmatter` struct. The JSON Schema compiler at `internal/validation/validator.go` already handles embedded schemas via `//go:embed schemas/*.json` -- adding `agent.schema.json` follows the identical pattern used for `session-context.schema.json`, `knossos-manifest.schema.json`, and 7 other schemas.

#### Go-Native Generation Pipeline (Not Shell Scripts)

The MOONSHOT document assumed shell-based generation (`generate-orchestrator.sh`). The codebase has moved past this. The generation pipeline should be a Go package:

```
internal/agentgen/
  agentgen.go       # Core generator: spec + template -> agent markdown
  schema.go         # AgentFrontmatter struct, parsing, validation
  templates.go      # Embedded base templates per archetype
  compose.go        # Section composition from shared fragments
  agentgen_test.go  # Tests including golden-file comparisons
```

This integrates naturally with the existing materialization flow:

```
                        CURRENT FLOW
rites/{name}/agents/*.md  --[copy]--> .claude/agents/*.md

                        FACTORY FLOW
rites/{name}/agents/*.yaml --[agentgen]--> .claude/agents/*.md
                                |
                            validates against
                            agent.schema.json
```

The materializer's `materializeAgents()` function (line 396 of materialize.go) is the integration point. Currently it walks the source directory and copies files verbatim. With the factory, it would detect `.yaml` source files and run them through `agentgen` before writing `.md` to the destination. `.md` source files continue to work as-is (backward compatibility).

#### Templateability by Archetype

| Archetype | Count | Templateable Sections | Unique Sections |
|-----------|-------|----------------------|-----------------|
| Orchestrator | 10 | Identity, Approach (consultation protocol), Tools, Constraints, Anti-patterns, Reference Skills | Routing table, phase definitions |
| Specialist | 35+ | Identity, Approach (generic), Tools, Handoff Criteria, Quality Standards, Constraints | Core Responsibilities, Domain Authority, What You Produce |
| Reviewer | 5+ | Identity, Approach (adversarial), Tools, Handoff Criteria, Quality Standards | Review criteria, pass/fail definitions |

Effective templateability: ~55% of total agent content across all archetypes. The remaining ~45% is domain-specific and must remain in the spec.

---

### 2. Behavioral Contracts: Testing, Validation, CI

#### The Contract Model

Agent behavioral contracts operate at three levels:

**Level 1: Static Schema Validation (build-time)**
- Frontmatter conforms to `agent.schema.json`
- Required sections present (Identity, Core Responsibilities, Domain Authority)
- Tool references resolve to known Claude Code tools
- Model value in allowed set
- Downstream agent references resolve within the rite manifest

Implementation: Extend `internal/validation/Validator` with `ValidateAgentSpec()` method, following the exact pattern of `ValidateSessionContext()`. Add `agent.schema.json` to `internal/validation/schemas/`. Cost: ~200 LOC, 1-2 days.

**Level 2: Structural Contract Validation (CI-time)**
- Generated markdown contains all sections declared in spec
- Tool access section matches frontmatter tool list
- Handoff criteria reference valid downstream agents
- No orphan cross-references (agent mentions non-existent skills/commands)

Implementation: A `ValidateAgentMarkdown()` function that parses the generated markdown and checks structural invariants against the spec. This does not invoke Claude -- it is pure text analysis. Cost: ~400 LOC, 3-5 days.

**Level 3: Behavioral Smoke Tests (optional, expensive)**
- Invoke agent with a synthetic prompt, verify it uses declared tools
- Verify it produces declared artifacts
- Verify it respects `must_not` constraints

Implementation: Requires Claude API calls, costs money, is slow. Recommendation: defer to Phase 3 of the MOONSHOT roadmap. For now, static + structural validation covers the most impactful failure modes.

#### CI Integration

```
ari validate agents [--rite NAME]   # Validates all agents in a rite
ari validate agent PATH             # Validates a single agent spec
```

Hook into `go test` via a test file that globs all agent specs and validates them:

```go
func TestAllAgentSpecs(t *testing.T) {
    specs, _ := filepath.Glob("rites/*/agents/*.yaml")
    validator, _ := validation.NewValidator()
    for _, spec := range specs {
        t.Run(spec, func(t *testing.T) {
            data, _ := os.ReadFile(spec)
            err := validator.ValidateAgentSpec(data)
            if err != nil {
                t.Errorf("invalid agent spec %s: %v", spec, err)
            }
        })
    }
}
```

This means agent validation runs on every `go test ./...` -- no special CI configuration needed.

---

### 3. Distribution to Satellites: Registry, Channels, Overrides

#### Current Distribution Model

```
ROSTER                               SATELLITE
rites/{name}/agents/*.md  --[ari sync materialize]--> .claude/agents/*.md
mena/**/*.{dro,lego}.md   --[ari sync materialize]--> .claude/{commands,skills}/*.md
user-agents/*.md           --[ari sync user agents]--> ~/.claude/agents/*.md
```

This is file-copy with orphan detection. No versioning, no channels, no override composition.

#### What the Factory Adds

**Agent Versioning**: The `version` field in `AgentFrontmatter` enables tracking which version of an agent a satellite is running. The existing `sync/state.json` (already written by `trackState()` in materialize.go) is the natural home for per-agent version tracking.

```json
{
  "version": "1.0",
  "active_rite": "rnd",
  "agents": {
    "technology-scout": {
      "version": "1.2.0",
      "checksum": "sha256:abc123...",
      "generated_at": "2026-02-06T10:00:00Z",
      "source": "rites/rnd/agents/technology-scout.yaml"
    }
  }
}
```

**Override Layers**: Satellites need to customize agents without forking. The pattern mirrors CSS specificity -- more specific layers win:

```
Layer 1 (base):     rites/{name}/agents/agent.yaml       # Platform default
Layer 2 (shared):   rites/shared/agents/agent.yaml        # Cross-rite override
Layer 3 (project):  .knossos/agents/agent.yaml            # Project-level override
Layer 4 (user):     user-agents/agent.yaml                # User-level override
```

Override semantics: YAML deep-merge. A satellite override can add/replace sections without touching the base. The `sourceResolver` in materialize.go already implements 4-tier resolution (project > user > knossos) for rites; extending it to agents is a natural progression.

**Channels**: Not a near-term need. The MOONSHOT doc's stable/beta/edge channels become relevant at 30+ satellites. For now, git branches plus the existing `--source` flag on `ari sync materialize` provide sufficient channel semantics:

```bash
ari sync materialize rnd                        # default (main branch)
ari sync materialize rnd --source ~/roster-dev  # edge (local dev checkout)
```

**Drift Detection**: Compare satellite `state.json` agent checksums against source. Report via `ari sync status`:

```
$ ari sync status
Active rite: rnd (source: knossos)
Agents: 6 materialized
  technology-scout    v1.2.0  OK
  integration-researcher  v1.1.0  DRIFT (local modification)
  prototype-engineer  v1.0.0  STALE (v1.1.0 available)
```

This reuses the checksum infrastructure already in `internal/usersync/manifest.go`.

---

### 4. MCP Integration: Tool Proxying, Agent Services, Satellite Config

#### Current State: Zero MCP Integration

Confirmed across the entire codebase. No MCP server definitions, no protocol handling, no tool proxying. The `mcpServers` key does not appear in any settings file.

#### What Exists That Is MCP-Adjacent

1. **Settings generation**: `materializeSettings()` in materialize.go creates `settings.local.json` with a `hooks` key. The Claude Code settings format supports `mcpServers` at the same level. Adding MCP server configuration is a one-line change to the settings template.

2. **Tool declarations in agent frontmatter**: Agents already declare `tools: [Glob, Grep, Read, ...]`. These are Claude Code built-in tools. MCP tools would appear as `tools: [mcp:github, mcp:jira, Read, Write]` -- extending the tool enum, not changing the architecture.

3. **Hook infrastructure**: Hooks already implement a request/response protocol with JSON I/O (see SPIKE-hook-architecture.md). MCP servers are conceptually similar -- external processes providing tools via JSON-RPC. The hook dispatch pattern (`shell wrapper -> ari subcommand -> JSON response`) maps directly to MCP tool invocation.

#### MCP Integration Path

**Phase A: Satellite MCP Configuration (low effort, high value)**

Allow rite manifests to declare MCP servers that get materialized into satellite settings:

```yaml
# rites/rnd/manifest.yaml
mcp_servers:
  - name: github
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: "${GITHUB_TOKEN}"
  - name: filesystem
    command: npx
    args: ["-y", "@modelcontextprotocol/server-filesystem"]
    args_dynamic: ["${PROJECT_ROOT}"]
```

Materialization merges these into `settings.local.json`:

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}" }
    }
  }
}
```

Implementation: Extend `materializeSettings()` to read MCP declarations from the rite manifest and merge them into the settings JSON. The deep-merge pattern is already established by the inscription system's region merging. Cost: ~150 LOC, 1-2 days.

**Phase B: Agent-Scoped MCP Tools (medium effort)**

Agents declare which MCP tools they need access to. The factory validates that the satellite's MCP configuration provides those tools:

```yaml
# Agent spec
tools:
  - Read
  - Write
  - mcp:github/create_issue
  - mcp:github/search_repositories

# Validation checks:
# 1. "github" MCP server defined in rite manifest
# 2. Tool names match MCP server's advertised tools
```

**Phase C: Ari as MCP Server (long-term, high effort)**

Expose Knossos operations as MCP tools that any MCP client can consume:

```
ari mcp serve --port 3000

Tools exposed:
  knossos/session_create    - Create a new session
  knossos/session_status    - Get session status
  knossos/rite_switch       - Switch active rite
  knossos/agent_validate    - Validate an agent spec
  knossos/sync_status       - Get sync status
```

This is speculative. It becomes relevant if/when MCP becomes the standard protocol for agentic tool access across providers. The MOONSHOT doc rates "MCP spec changes" as High risk -- building a full MCP server before the protocol stabilizes would be premature.

---

## Architecture Vision

### Current State (2026)

```
ROSTER (source of truth)
  rites/
    {name}/
      manifest.yaml          # Rite definition
      agents/*.md            # Handcrafted markdown (no schema)
      mena/*.{dro,lego}.md   # Commands and skills
  mena/                      # User-level mena
  user-agents/*.md           # User-level agents

     |
     | ari sync materialize {rite}     [file copy, no validation]
     | ari sync user agents            [checksum-based sync]
     v

SATELLITE (.claude/)
  agents/*.md                # Copied from rite
  commands/*.md              # Materialized from mena
  skills/*.md                # Materialized from mena
  ACTIVE_RITE                # Current rite marker
  KNOSSOS_MANIFEST.yaml      # Inscription manifest
  CLAUDE.md                  # Generated via inscription
  settings.local.json        # Minimal (hooks only, no MCP)
```

### Target State (2027)

```
ROSTER (source of truth)
  rites/
    {name}/
      manifest.yaml          # Rite definition + MCP server declarations
      agents/*.yaml          # Agent SPECS (YAML, schema-validated)
      agents/*.md            # Escape hatch: handcrafted agents (coexist)
      mena/*.{dro,lego}.md   # Commands and skills (unchanged)
  mena/                      # User-level mena (unchanged)
  user-agents/*.md           # User-level agents (unchanged)
  internal/
    agentgen/                # Go-native agent generation pipeline
    validation/schemas/
      agent.schema.json      # Formal agent schema

     |
     | ari sync materialize {rite}
     |   1. Load rite manifest
     |   2. For each agent:
     |      a. .yaml? -> validate against schema -> generate .md via template
     |      b. .md?   -> copy verbatim (backward compat)
     |   3. Validate generated agents (structural contracts)
     |   4. Materialize MCP servers into settings.local.json
     |   5. Track versions in sync/state.json
     |   6. Inscription generates CLAUDE.md
     v

SATELLITE (.claude/)
  agents/*.md                # Generated from specs OR copied from source
  commands/*.md              # Materialized from mena
  skills/*.md                # Materialized from mena
  ACTIVE_RITE                # Current rite marker
  KNOSSOS_MANIFEST.yaml      # Inscription manifest
  CLAUDE.md                  # Generated via inscription
  settings.local.json        # Hooks + MCP servers
  sync/state.json            # Agent versions, checksums, drift tracking
```

### Key Architectural Difference

The MOONSHOT doc envisioned a standalone "Agent Factory" as a separate system with its own registry, distribution layer, and health monitor. This spike finds that the factory should be **embedded in the existing materialization pipeline** -- not a separate system, but an enhancement of `ari sync materialize`. The registry is `sync/state.json`. The distribution layer is `ari sync`. The health monitor is `ari sync status`. No new binaries, no new services, no new infrastructure.

---

## 2026 Prerequisites

What must be built this year to enable the 2027 target state, ordered by dependency:

### P0: Agent Schema (1-2 weeks)

- Add `agent.schema.json` to `internal/validation/schemas/`
- Create `AgentFrontmatter` struct in `internal/materialize/frontmatter.go` (pattern: `MenaFrontmatter`)
- Add `ValidateAgentSpec()` to `internal/validation/Validator`
- Add `ari validate agent` CLI command
- Add `TestAllAgentSpecs` to `go test ./...`

**Why first**: Everything downstream depends on having a formal schema. Without it, generation produces unvalidated output and contracts have nothing to enforce against.

**Reversibility**: Two-way door. Schema validation is additive. Existing `.md` agents are unaffected.

### P1: Agent Generation Package (2-3 weeks, depends on P0)

- Create `internal/agentgen/` package
- Implement `Generate(spec AgentSpec, template Template) ([]byte, error)`
- Embed base templates for orchestrator, specialist, reviewer archetypes
- Add section composition (shared fragments for Constraints, Anti-patterns, Quality Standards)
- Golden-file tests comparing generated output to handcrafted agents

**Why second**: This is the core value -- going from "2-4 hours to create an agent" to "30 minutes".

**Reversibility**: Two-way door. Generation is optional; `.md` agents continue to work.

### P2: Materialization Integration (1 week, depends on P1)

- Modify `materializeAgents()` to detect `.yaml` source files
- Route `.yaml` through `agentgen` before writing `.md`
- Update `trackState()` to record per-agent version and checksum

**Reversibility**: Two-way door. The `.md` fallback path is preserved.

### P3: MCP Settings Materialization (1 week, independent)

- Extend `RiteManifest` struct with `MCPServers` field
- Modify `materializeSettings()` to merge MCP server declarations
- Add MCP server entries to 1-2 rite manifests as proof of concept

**Reversibility**: Two-way door. MCP configuration is additive to settings.json.

### P4: Drift Detection (1 week, depends on P2)

- Extend `ari sync status` to compare satellite agent checksums against source
- Report DRIFT (local modification) and STALE (newer version available)
- Wire into existing `sync/state.json` infrastructure

**Reversibility**: Two-way door. Read-only status reporting, no mutations.

### Total 2026 Investment

6-8 weeks of focused work, ~1500-2000 LOC of Go. Notably smaller than the MOONSHOT doc's Phase 1-3 estimate of 6-9 weeks because the Go infrastructure already exists -- the MOONSHOT doc assumed building from shell scripts.

---

## Scenario Analysis

The MOONSHOT document defined four scenarios. This spike reassesses them with current codebase knowledge:

### Scenario A: Claude Code Native Agent Templating (Probability: HIGH, 70%)

**What changes since MOONSHOT**: Claude Code has not announced native templating as of 2026-02-06, but the MCP specification continues to expand. The more likely path is MCP-based agent definitions rather than a proprietary template format.

**What survives**: The semantic layer (agent specs as YAML with behavioral contracts) remains valuable regardless of rendering target. If Claude Code introduces its own format, the `agentgen` package becomes a transpiler: `agent.yaml -> claude-native-format` instead of `agent.yaml -> agent.md`.

**What breaks**: Nothing, if we keep the spec/rendering separation clean.

**Preparation**: Ensure the `AgentFrontmatter` struct captures semantics, not formatting. Do not embed markdown-specific assumptions in the schema.

### Scenario B: Multi-LLM Agent Ecosystem (Probability: MEDIUM, 40%)

**What changes since MOONSHOT**: The tool enum (`Glob`, `Grep`, `Read`, etc.) is Claude Code-specific. Other providers have different tool interfaces.

**What survives**: Agent specs with provider-agnostic capability declarations (`needs: web_search, file_read, file_write`) that get mapped to provider-specific tools at generation time.

**Preparation for 2026**: Add `capabilities` as an optional field in the agent schema alongside `tools`. The tools field remains the concrete provider-specific list; capabilities is the abstract list. This is a two-way door -- we can add it later without breaking anything.

### Scenario C: 100x Scale (Probability: MEDIUM-LOW, 30%)

**What changes since MOONSHOT**: The existing 4-tier source resolution in `SourceResolver` (project > user > knossos) already supports multi-source distribution without a registry service. At 100+ satellites, the bottleneck is not distribution (git handles this) but **conflict resolution and drift detection**.

**Preparation for 2026**: P4 (drift detection) is sufficient. A dedicated registry becomes necessary only if git-based distribution fails at scale -- and git handles thousands of consumers routinely.

### Scenario D: Compliance/Regulatory (Probability: LOW, 20%)

**What changes since MOONSHOT**: Agent provenance tracking (`sync/state.json` with checksums, versions, source paths) provides a basic audit trail. The formal agent schema provides a validation checkpoint. Together, these satisfy 80% of a compliance audit without building dedicated infrastructure.

**Preparation for 2026**: P0 (schema) + P2 (materialization integration with version tracking) provide the audit trail foundation for free.

### Risk Matrix

| Risk | Prob. | Impact | 2026 Mitigation |
|------|-------|--------|-----------------|
| Claude native templating obsoletes generation | High | High | Spec-first design; generation is the swappable layer |
| Generated agents inferior to handcrafted | Medium | Medium | Golden-file tests; `.md` escape hatch; gradual migration |
| MCP spec instability breaks integration | Medium | Low | Phase A only (settings merge); no protocol dependency |
| Agent schema too rigid, constrains creativity | Medium | Medium | Schema enforces structure not content; freeform sections allowed |
| Over-engineering before product-market fit | Medium | High | P0-P2 only; defer registry/channels/health until triggered |

---

## Recommendation

### Start Now (regardless of which scenario arrives)

1. **P0: Agent Schema** -- Define `agent.schema.json`, add `ValidateAgentSpec()`. This is the foundation for everything else and costs 1-2 weeks. It is a two-way door.

2. **P3: MCP Settings Materialization** -- Extend `materializeSettings()` to carry MCP server config from rite manifests. This is independent of P0-P2, costs 1 week, and unblocks satellite teams that want MCP tools.

3. **Convert 2-3 agents to YAML spec format** as a design exercise. Start with one of each archetype: one orchestrator (most templateable), one specialist, one reviewer. Use these to validate the schema design before committing to generation.

### Build Next (Q2 2026)

4. **P1: Agent Generation Package** -- Build `internal/agentgen/` once the schema is validated by real specs. Golden-file test against the handcrafted originals.

5. **P2: Materialization Integration** -- Wire generation into `ari sync materialize`. This is the point of no return for the factory approach -- but it is still a two-way door because `.md` agents continue to work.

### Wait For Trigger

| Action | Trigger Signal |
|--------|---------------|
| Agent registry service | >30 active satellites using Knossos |
| Multi-provider rendering | Customer request for non-Claude agents, or GPT-5 launch with competitive agent support |
| Ari as MCP server | MCP 1.0 stable release with wide adoption |
| Behavioral smoke tests | Compliance audit request, or agent regression in production |
| Override composition system | >5 satellite override requests per month |

---

## Open Questions

1. **Schema strictness**: Should the agent schema require all 11 sections, or allow partial specs that get filled with archetype defaults? Partial specs are more practical but harder to validate.

2. **Migration path for 58 existing agents**: Convert all at once (big bang) or convert per-rite as rites are updated? Per-rite is safer but means maintaining two formats indefinitely.

3. **Forge rite relationship**: The forge rite (7 agents: prompt-architect, agent-designer, etc.) currently creates agents manually. Should the forge rite become the Agent Factory's user interface, or should `ari agent create` be a standalone command?

4. **Template language for agent generation**: Go `text/template` (consistent with inscription system) or a simpler string interpolation? `text/template` is more powerful but adds complexity for agent authors.

5. **MCP server lifecycle**: Who starts/stops MCP servers declared in rite manifests? Claude Code manages MCP server processes, but Knossos needs to ensure the configuration is correct before Claude Code launches.

6. **Agent spec backward compatibility**: When the schema evolves (v1 -> v2), do we migrate all specs or support multi-version validation? The JSON Schema infrastructure supports `$schema` version declarations, but multi-version validation adds maintenance burden.

---

## References

| Document | Location | Relationship |
|----------|----------|-------------|
| MOONSHOT-agent-template-ecosystem | `/Users/tomtenuta/Code/roster/docs/MOONSHOT-agent-template-ecosystem.md` | Upstream vision document; this spike refines its implementation path |
| Materialization pipeline | `/Users/tomtenuta/Code/roster/internal/materialize/materialize.go` | Primary integration point for agent generation |
| Validation infrastructure | `/Users/tomtenuta/Code/roster/internal/validation/validator.go` | JSON Schema validation with embedded schemas |
| Mena frontmatter | `/Users/tomtenuta/Code/roster/internal/materialize/frontmatter.go` | Pattern for AgentFrontmatter struct |
| Rite manifest (rnd) | `/Users/tomtenuta/Code/roster/rites/rnd/manifest.yaml` | Example manifest to extend with MCP servers |
| Agent example | `/Users/tomtenuta/Code/roster/rites/rnd/agents/technology-scout.md` | Current agent format to validate schema against |
| SPIKE-materialization-model | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-materialization-model.md` | Upstream spike on materialization patterns |
| SPIKE-hook-architecture | `/Users/tomtenuta/Code/roster/docs/spikes/SPIKE-hook-architecture.md` | Hook contract pattern relevant to MCP integration |

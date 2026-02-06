# ari agent -- Agent Management Commands

**For**: Rite authors, agent developers, CI pipelines
**Requires**: ari CLI (Knossos platform)
**Project context**: Must be run from within a Knossos project root

## Overview

The `ari agent` command group provides tools for managing agent specification files. Agents are markdown files with YAML frontmatter that define specialist, orchestrator, and reviewer roles within rites.

Agent files live in two locations:
- `rites/<rite-name>/agents/*.md` -- rite-scoped agents
- `user-agents/*.md` -- cross-rite agents (meta-agents, shared utilities)

```
ari agent [command]

Available Commands:
  validate    Validate agent specifications
  list        List agents
  new         Scaffold a new agent from an archetype
```

### Global Flags

These flags are available on all `ari agent` subcommands:

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | | Config file path (default: `$XDG_CONFIG_HOME/knossos/config.yaml`) |
| `--output` | `-o` | Output format: `text`, `json`, `yaml` (default: `text`) |
| `--project-dir` | `-p` | Project root directory (overrides auto-discovery) |
| `--session-id` | `-s` | Session ID (overrides current) |
| `--verbose` | `-v` | Enable verbose output (JSON lines to stderr) |

---

## ari agent validate

Validates agent frontmatter against the agent JSON schema and semantic rules.

### Usage

```
ari agent validate [path...] [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--rite` | `-r` | | Validate all agents in the specified rite |
| `--strict` | | `false` | Enable strict validation (requires enhanced fields like `type`, `tools`) |
| `--all` | | `false` | Validate all agents in all rites and user-agents |

When called with no arguments and no flags, the default behavior is to validate all agents (equivalent to `--all`).

### Validation Modes

**Default (warn)**: Validates required fields (`name`, `description`) as errors. Missing optional enhanced fields (`type`, `tools`, `model`) produce warnings instead of errors. Suitable for existing agents that have not yet been fully migrated.

**Strict (`--strict`)**: Requires enhanced fields in addition to the base required fields. `type` and `tools` become required. Archetype-specific rules are enforced as errors (e.g., reviewers must have `contract.must_not`). Use this for post-migration agents and CI gates.

### Validation Pipeline

Validation runs in three phases:

1. **Parse**: Extract YAML frontmatter from the markdown file. Fails if frontmatter delimiters (`---`) are missing or YAML is malformed.
2. **Schema**: Validate the parsed frontmatter against `agent.schema.json`. Checks types, enums, patterns, and conditional rules.
3. **Semantic**: Go-level checks beyond what JSON Schema can express. Validates tool references against the known tool list, checks archetype-specific constraints, and produces archetype warnings.

### Output Format

Each agent file produces one status line:

```
PASS  rites/ecosystem/agents/context-architect.md
WARN  user-agents/moirai.md
  WARN: tools field is empty or missing
FAIL  rites/rnd/agents/broken-agent.md
  ERROR: name: name is required
  ERROR: tools: agent frontmatter: unknown tool "InvalidTool"
         value: InvalidTool
```

Status indicators:
- `PASS` -- No errors and no warnings
- `WARN` -- Valid but with warnings
- `FAIL` -- Validation errors found

The summary follows all individual results:

```
Summary: 12 agents validated
  Valid: 10
  Errors: 2
  Warnings: 3
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All agents passed validation |
| 1 | One or more agents failed validation |

### Examples

Validate all agents across all rites and user-agents:

```bash
ari agent validate
```

Validate only agents in the ecosystem rite:

```bash
ari agent validate --rite ecosystem
```

Validate a specific agent file:

```bash
ari agent validate user-agents/moirai.md
```

Validate with glob pattern (shell-expanded):

```bash
ari agent validate rites/*/agents/*.md
```

Strict validation for CI pipeline:

```bash
ari agent validate --strict --rite ecosystem
# Exit code 1 if any agent is missing required enhanced fields
```

---

## ari agent list

Lists agents with their metadata extracted from frontmatter. Displays a formatted table with agent name, type, model, source, and description.

### Usage

```
ari agent list [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--rite` | `-r` | | List agents from a specific rite only |
| `--all` | | `true` | List all agents (default behavior) |

### Output Format

```
AGENT                          TYPE            MODEL      SOURCE          DESCRIPTION
----------------------------------------------------------------------------------------------------
orchestrator                   orchestrator    opus       rite:ecosystem  Coordinates ecosystem infras...
ecosystem-analyst              specialist      opus       rite:ecosystem  Traces CEM/knossos problems ...
context-architect              specialist      opus       rite:ecosystem  Designs context solutions, s...
moirai                         meta            opus       user            Context mutation meta-agent
```

Column details:
- **AGENT**: The `name` field from frontmatter, or the filename (without `.md`) if name is empty
- **TYPE**: The `type` field from frontmatter, or `-` if not set
- **MODEL**: The `model` field from frontmatter, or `-` if not set
- **SOURCE**: Either `rite:<rite-name>` for rite agents or `user` for user-agents
- **DESCRIPTION**: The `description` field, truncated to 40 characters

Agents that fail frontmatter parsing are silently skipped from the listing.

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (even if no agents found) |
| 1 | Error reading agent directories |

### Examples

List all agents in the project:

```bash
ari agent list
```

List agents in a specific rite:

```bash
ari agent list --rite forge
```

---

## ari agent new

Scaffolds a new agent file from an archetype template. The generated file includes complete frontmatter, platform-owned sections with default content, and author-owned sections marked with `<!-- TODO -->` comments.

### Usage

```
ari agent new [flags]
```

### Required Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--archetype` | `-a` | Archetype to use (`orchestrator`, `reviewer`, `specialist`) |
| `--rite` | `-r` | Target rite (must exist as a directory under `rites/`) |
| `--name` | `-n` | Agent name in kebab-case (e.g., `technology-scout`) |

### Optional Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--description` | `-d` | Agent description. If omitted, defaults to `"<Archetype> agent for the <rite> rite"` |

### Behavior

1. Validates the archetype name against the registry
2. Confirms the target rite directory exists at `rites/<rite>/`
3. Creates `rites/<rite>/agents/` directory if it does not exist
4. Checks that the target file does not already exist (prevents overwriting)
5. Renders the archetype template with provided metadata
6. Validates the generated output passes frontmatter parsing and validation
7. Writes the file to `rites/<rite>/agents/<name>.md`

### Output

On success:

```
Created rites/rnd/agents/technology-scout.md -- fill in author sections marked with TODO
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Agent file created successfully |
| 1 | Error: unknown archetype, rite not found, file already exists, or permission denied |

### Examples

Create a specialist agent:

```bash
ari agent new --archetype specialist --rite rnd --name technology-scout
```

Create a reviewer with a custom description:

```bash
ari agent new \
  --archetype reviewer \
  --rite security \
  --name code-reviewer \
  --description "Reviews code for security vulnerabilities and compliance"
```

Create an orchestrator:

```bash
ari agent new -a orchestrator -r ecosystem -n coordinator
```

---

## Agent Frontmatter Schema Reference

Agent files use YAML frontmatter (delimited by `---`) at the top of the markdown file. The schema is defined in `internal/validation/schemas/agent.schema.json`.

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Agent identifier in kebab-case (e.g., `ecosystem-analyst`). Must match pattern `^[a-z][a-z0-9-]*$`. |
| `description` | `string` | Agent description, minimum 10 characters. Should include use-cases and trigger phrases. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `role` | `string` | Short role summary, 1-200 characters (e.g., "Traces ecosystem issues to root causes"). |
| `type` | `string` | Agent archetype. One of: `orchestrator`, `specialist`, `reviewer`, `meta`, `designer`, `analyst`, `engineer`. |
| `tools` | `string` or `string[]` | Tools available to this agent. Accepts comma-separated string (`"Bash, Read, Glob"`) or YAML array. |
| `model` | `string` | Claude model tier. One of: `opus`, `sonnet`, `haiku`. |
| `color` | `string` | Display color for agent badge. Lowercase alphabetic (e.g., `purple`, `orange`, `red`). |
| `aliases` | `string[]` | Alternative names for invocation. Each must match `^[a-z0-9-]+$`. |
| `upstream` | `UpstreamRef[]` | Agents or sources that feed into this agent. |
| `downstream` | `DownstreamRef[]` | Agents this agent routes work to. |
| `produces` | `ArtifactDecl[]` | Artifacts this agent is expected to produce. |
| `contract` | `BehavioralContract` | Behavioral constraints and requirements. |
| `schema_version` | `string` | Schema version (e.g., `"1.0"`). |

### Nested Types

**UpstreamRef**

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `source` | Yes | `string` | Source agent name or external input description |
| `artifact` | No | `string` | Expected artifact from upstream |

**DownstreamRef**

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `agent` | Yes | `string` | Target agent name |
| `condition` | No | `string` | When to route to this agent |
| `artifact` | No | `string` | Artifact passed to downstream agent |

**ArtifactDecl**

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `artifact` | Yes | `string` | Artifact name or type |
| `format` | No | `string` | Expected format (`markdown`, `yaml`, `json`, etc.) |

**BehavioralContract**

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `must_use` | No | `string[]` | Tools or patterns the agent must use |
| `must_produce` | No | `string[]` | Artifacts the agent must produce |
| `must_not` | No | `string[]` | Actions the agent must not take |
| `max_turns` | No | `integer` | Maximum conversation turns before requiring handoff (minimum: 1) |

### Valid Tool References

Standard Claude Code tools:

```
Bash, Read, Write, Edit, Glob, Grep, Task, TodoWrite, TodoRead,
WebSearch, WebFetch, Skill, NotebookEdit, AskUserQuestion
```

MCP tool references follow the pattern `mcp:<server>` or `mcp:<server>/<method>`:

```yaml
tools:
  - Read
  - Bash
  - mcp:github
  - mcp:github/create_issue
```

MCP tool references produce warnings during validation because server availability depends on the satellite's MCP configuration.

### Conditional Schema Rules

- **Orchestrators** (`type: orchestrator`): The `model` field must be `opus` when set. Validation warns if tools other than `Read` are declared.
- **Reviewers** (`type: reviewer`): In strict mode, `contract.must_not` is required (must have at least one entry). In default mode, a missing `contract.must_not` produces a warning.

### Complete Frontmatter Example

```yaml
---
name: integration-engineer
description: |
  Implements CEM and knossos changes with integration tests.
  Invoke for code changes, test writing, and build verification.
role: Implements infrastructure changes with tests
type: specialist
tools:
  - Bash
  - Glob
  - Grep
  - Read
  - Edit
  - Write
  - TodoWrite
  - Skill
model: opus
color: orange
upstream:
  - source: context-architect
    artifact: design document
downstream:
  - agent: documentation-engineer
    condition: implementation complete with passing tests
    artifact: implementation report
produces:
  - artifact: implementation code
    format: go
  - artifact: integration tests
    format: go
contract:
  must_use:
    - Read
  must_produce:
    - integration tests
  must_not:
    - skip test verification
    - modify files outside scope
  max_turns: 50
schema_version: "1.0"
---
```

---

## Archetype Reference

Archetypes define the default structure, frontmatter values, and section layout for new agents. Three archetypes are available.

### orchestrator

**Purpose**: Consultative coordinator that analyzes context, routes work to specialists, and maintains decision consistency across phases. Orchestrators do not execute work directly.

**Default values**:

| Field | Value |
|-------|-------|
| `model` | `opus` |
| `tools` | `Read` |
| `color` | `purple` |

**Sections** (12 total):

| Section | Ownership | Author Action |
|---------|-----------|---------------|
| Consultation Role | Platform | None (provided) |
| Tool Access | Derived | None (generated) |
| Consultation Protocol | Platform | None (provided) |
| Position in Workflow | Derived | None (generated) |
| Domain Authority | **Author** | Define decisions vs. escalations |
| Phase Routing | **Author** | Define specialist routing conditions |
| Behavioral Constraints | Platform | None (provided) |
| Handling Failures | Platform | None (provided) |
| The Acid Test | Platform | None (provided) |
| Cross-Team Protocol | **Author** | Define cross-rite routing |
| Skills Reference | Derived | None (generated) |
| Anti-Patterns | Platform | None (provided) |

**When to use**: For agents that coordinate multi-phase workflows, route to specialists, and make sequencing decisions without executing work themselves.

### specialist

**Purpose**: Domain expert that executes focused work within a specific discipline. Specialists receive prompts from orchestrators, produce artifacts, and hand off to downstream agents.

**Default values**:

| Field | Value |
|-------|-------|
| `model` | `opus` |
| `tools` | `Bash`, `Glob`, `Grep`, `Read`, `Edit`, `Write`, `TodoWrite`, `Skill` |
| `color` | `orange` |

**Sections** (11 total):

| Section | Ownership | Author Action |
|---------|-----------|---------------|
| Core Responsibilities | **Author** | Define primary and secondary functions |
| Position in Workflow | Derived | None (generated) |
| Domain Authority | **Author** | Define decisions, escalations, and routing |
| Tool Access | Derived | None (generated) |
| What You Produce | **Author** | Define artifacts with format and audience |
| Quality Standards | **Author** | Define quality criteria and verification |
| Handoff Criteria | **Author** | Define completion checklist |
| Behavioral Constraints | Platform | None (provided) |
| The Acid Test | Platform | None (provided) |
| Anti-Patterns | Platform | None (provided) |
| Skills Reference | Derived | None (generated) |

**When to use**: For agents that do the actual work -- writing code, producing documents, running analyses, creating artifacts. Most agents in a rite are specialists.

### reviewer

**Purpose**: Quality gate that reviews work products against domain-specific criteria. Reviewers evaluate, classify findings by severity, and provide clear approve/reject decisions with actionable feedback.

**Default values**:

| Field | Value |
|-------|-------|
| `model` | `opus` |
| `tools` | `Bash`, `Glob`, `Grep`, `Read`, `Edit`, `Write`, `WebFetch`, `WebSearch`, `TodoWrite`, `Skill` |
| `color` | `red` |

**Sections** (10 total):

| Section | Ownership | Author Action |
|---------|-----------|---------------|
| Core Purpose | **Author** | Define what this reviewer catches |
| Position in Workflow | Derived | None (generated) |
| Domain Authority | **Author** | Define decisions, escalations, and routing |
| Quality Standards | **Author** | Define review focus areas and criteria |
| Severity Classification | **Author** | Define severity levels with examples |
| What You Produce | **Author** | Define review artifacts and signoff format |
| Behavioral Constraints | Platform | None (provided) |
| The Acid Test | Platform | None (provided) |
| Anti-Patterns | Platform | None (provided) |
| Skills Reference | Derived | None (generated) |

**When to use**: For terminal or gate agents that approve, reject, or route back work products. Common examples: security reviewers, documentation reviewers, integration test validators.

### Section Ownership Model

Each section in an archetype template has an ownership designation:

| Ownership | Meaning | In Scaffolded File |
|-----------|---------|-------------------|
| **Platform** | Default content provided by the archetype. Authors should not need to modify. | Fully populated |
| **Author** | Must be filled in by the agent author. | Contains `<!-- TODO -->` markers with guidance |
| **Derived** | Generated from frontmatter or context. Future `ari agent update` will regenerate. | Contains template-derived content |

---

## Workflow Examples

### Creating a New Specialist Agent for a Rite

Scenario: You are building a new `rnd` rite and need a technology scout agent.

```bash
# 1. Scaffold the agent
ari agent new \
  --archetype specialist \
  --rite rnd \
  --name technology-scout \
  --description "Researches emerging technologies and produces evaluation reports"

# Expected output:
# Created rites/rnd/agents/technology-scout.md -- fill in author sections marked with TODO

# 2. Edit the generated file -- fill in all sections marked with TODO
#    Focus on: Core Responsibilities, Domain Authority, What You Produce,
#    Quality Standards, Handoff Criteria

# 3. Validate the agent after editing
ari agent validate rites/rnd/agents/technology-scout.md

# Expected output:
# PASS  rites/rnd/agents/technology-scout.md
#
# Summary: 1 agents validated
#   Valid: 1
#   Errors: 0

# 4. Validate with strict mode to confirm all enhanced fields are present
ari agent validate --strict rites/rnd/agents/technology-scout.md
```

### Validating All Agents Before a Release

Scenario: CI pipeline gate that ensures all agents pass strict validation.

```bash
# Run strict validation across all agents
ari agent validate --strict

# In a CI script, check the exit code:
if ! ari agent validate --strict 2>&1; then
  echo "Agent validation failed -- fix errors before release"
  exit 1
fi

# For a specific rite release:
ari agent validate --strict --rite ecosystem
```

### Listing Agents to Understand Rite Composition

Scenario: A new contributor wants to understand which agents exist in the ecosystem rite.

```bash
# List all agents in the ecosystem rite
ari agent list --rite ecosystem

# List all agents across the entire project
ari agent list

# Combine with validate to get a full health picture
ari agent list --rite ecosystem
ari agent validate --rite ecosystem
```

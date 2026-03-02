# SPIKE: Organization-Level Best Practices for Knossos

> Research spike exploring how Knossos should model and support organization-level configuration, policies, and shared context as it scales from single-user to multi-team deployment.

**Date**: 2026-03-01
**Author**: Spike (org-level best practices)
**Prior Art**: `docs/strategy/ROADMAP-distribution-readiness.md`, `docs/decisions/ADR-0025-mena-scope.md`, `docs/decisions/ADR-0021-two-axis-context-model.md`, `docs/MOONSHOT-agent-template-ecosystem.md`

---

## Executive Summary

Knossos currently operates at three configuration tiers: **embedded** (compiled into the `ari` binary), **platform** (`$KNOSSOS_HOME`), and **project** (`.claude/` per-repo). There is no explicit **organization** tier -- the layer between "single developer's global preferences" and "per-project rite configuration." As Knossos moves toward multi-team distribution (Stage 2-3 in the roadmap), the absence of an org tier creates a gap: teams cannot share rites, agents, policies, or conventions across multiple projects without either (a) copying files manually or (b) having every project point at the same `$KNOSSOS_HOME`.

Meanwhile, Claude Code (CC) has shipped a mature four-scope configuration hierarchy (Managed > Project > User > Local) with enterprise-grade enforcement mechanisms (MDM, managed-settings.json, `allowManagedHooksOnly`, `forceLoginOrgUUID`). The industry has also converged on AGENTS.md as a cross-tool standard, with nested file hierarchies for monorepo/org patterns.

**Key finding**: Knossos should not build its own org-level enforcement layer. CC already has one (`managed-settings.json`). Instead, Knossos should focus on what CC does NOT provide: **org-level rite distribution, shared agent libraries, convention templates, and cross-project knowledge sharing** -- and deliver these through an `ari init --org` bootstrap + XDG org directory pattern that slots cleanly into CC's existing hierarchy.

---

## 1. Question and Context

### What are we trying to learn?

1. What does "org-level best practices" mean in the Knossos context?
2. How does Knossos's current configuration hierarchy map to CC's native scopes?
3. Where are the gaps between what CC provides and what teams need at org scale?
4. What should Knossos implement vs. delegate to CC's native mechanisms?
5. What does the industry (AGENTS.md, Cursor Rules, etc.) recommend for org-level patterns?

### What decisions will this inform?

- Whether to add an explicit org tier to Knossos's source resolution chain
- How `ari init` should handle org-level bootstrapping
- Whether org-level CLAUDE.md and rules should be Knossos-managed or CC-native
- The distribution model for shared rites across an organization's projects
- Feature-gating strategy for the enterprise tier (per roadmap Stage 3)

---

## 2. Current State Analysis

### 2.1 Knossos Configuration Tiers (Today)

| Tier | Source | Resolution | Managed By |
|------|--------|-----------|------------|
| **Embedded** | `embed.go` (rites/, mena/, agents/, templates/) | Lowest-priority fallback | `ari` binary |
| **Platform** | `$KNOSSOS_HOME/rites/`, `$KNOSSOS_HOME/config/hooks.yaml` | Mid-priority | Developer checkout |
| **User** | `~/.local/share/knossos/rites/`, `~/.claude/` | Per-developer scope | `ari sync user` |
| **Project** | `./rites/`, `.claude/` | Highest-priority (per-repo) | `ari sync` |

Source resolution for rites: **project > user > knossos platform > embedded** (`internal/materialize/source/resolver.go:52-160`).

### 2.2 Claude Code Configuration Hierarchy (Native)

| Scope | Location | Sharing | Override |
|-------|----------|---------|----------|
| **Managed** | MDM, server, `/Library/Application Support/ClaudeCode/managed-settings.json` | IT-deployed, org-wide | Highest (enforced) |
| **Project** | `.claude/settings.json`, `.claude/CLAUDE.md` | Git-committed, team-shared | Mid |
| **User** | `~/.claude/settings.json`, `~/.claude/CLAUDE.md` | Personal | Low |
| **Local** | `.claude/settings.local.json`, `.claude/CLAUDE.local.md` | Gitignored, personal | Lowest |

Key CC enterprise controls:
- `allowManagedHooksOnly: true` -- org lockdown of hook execution
- `allowManagedPermissionRulesOnly: true` -- org lockdown of tool permissions
- `allowManagedMcpServersOnly: true` -- org lockdown of MCP servers
- `forceLoginOrgUUID` -- automatic org association
- `strictKnownMarketplaces` -- approved plugin sources only
- Array settings **merge** across scopes (concatenate, deduplicate)
- Scalar settings **replace** (higher scope wins)

### 2.3 AGENTS.md Industry Standard

AGENTS.md (stewarded by Linux Foundation's Agentic AI Foundation) provides:
- Nested files for monorepo subproject isolation
- No enforced schema (plain Markdown)
- Closest-file-wins resolution for hierarchical overrides
- 25+ tool compatibility (Cursor, Copilot, Kilo Code, etc.)
- 20,000+ GitHub projects adopted as of August 2025

Org-level pattern: organizations place a canonical AGENTS.md template in a `.github` repo, scaffolding scripts copy it to new projects. "Fallback then override" model.

### 2.4 Cursor Rules Pattern

Cursor uses Modular Markdown (`.mdc`) files with YAML frontmatter for activation patterns:
- `.cursor/rules/00-project-context.mdc`
- `.cursor/rules/01-architecture.mdc`
- Glob-based activation: rules load only when working in matching file paths
- Three levels: Project, Team, User

---

## 3. Gap Analysis

### What CC already provides at org level (do NOT reimplement):

| Capability | CC Mechanism | Knossos Action |
|------------|-------------|----------------|
| Tool permission policies | `managed-settings.json` permissions | Delegate entirely |
| Hook enforcement | `allowManagedHooksOnly` | Delegate; ensure `ari hook` commands are in managed hooks |
| MCP server lockdown | `allowManagedMcpServersOnly` | Delegate entirely |
| Authentication/org binding | `forceLoginOrgUUID` | Delegate entirely |
| CLAUDE.md inheritance | User > Project hierarchy | Align; don't fight the load order |
| Settings merge semantics | Array concat, scalar replace | Use, don't reinvent |

### What CC does NOT provide (Knossos opportunity):

| Capability | Current Gap | Knossos Opportunity |
|------------|-----------|---------------------|
| **Shared rite distribution** | No mechanism to share rite definitions across an org's repos | Org-level rite registry (`~/.local/share/knossos/org/{org-name}/rites/`) |
| **Shared agent libraries** | Agents are per-project only | Org-level agents synced via `ari sync org` |
| **Convention templates** | Each project writes its own standards from scratch | Org-level mena (standards, prompting patterns, doc-artifacts) |
| **Cross-project knowledge** | `.know/` is per-project | Org-level `.know/` for shared architectural patterns |
| **Rite versioning across projects** | No way to pin rite version per-org | Org manifest with version constraints |
| **Org-level rules** | CC rules are per-project only | Shared `.claude/rules/` synced at org level |
| **Bootstrap consistency** | Each `ari init` starts from scratch | `ari init --org autom8y` seeds org conventions |

---

## 4. Recommended Architecture

### 4.1 Org Tier Addition to Source Resolution

Add an explicit **org** tier between user and knossos platform:

```
Resolution order (highest to lowest):
  1. project   ./rites/{rite}/
  2. org       $XDG_DATA_HOME/knossos/orgs/{org-name}/rites/{rite}/
  3. user      $XDG_DATA_HOME/knossos/rites/{rite}/
  4. knossos   $KNOSSOS_HOME/rites/{rite}/
  5. embedded  compiled into ari binary
```

This mirrors CC's own scope layering where "project > org > user > managed" is the conceptual model.

### 4.2 Org Directory Structure

```
~/.local/share/knossos/orgs/
  autom8y/                          # org name (kebab-case)
    org.yaml                        # org manifest
    rites/                          # org-level rite overrides/additions
      custom-review/
        manifest.yaml
        agents/
        mena/
    agents/                         # org-wide shared agents
      compliance-reviewer.md
      security-scanner.md
    mena/                           # org-wide shared mena
      guidance/
        standards/
          code-conventions.md
          api-standards.md
      templates/
        pr-template/INDEX.lego.md
    rules/                          # org-wide rules (synced to projects)
      security-policy.md
      naming-conventions.md
    know/                           # org-wide knowledge base
      architecture.md               # shared architectural patterns
      conventions.md                # cross-project conventions
```

### 4.3 Org Manifest (`org.yaml`)

```yaml
schema_version: "1.0"
name: autom8y
display_name: "Autom8y"
description: "Knossos platform development organization"

# Default rite for new projects bootstrapped with this org
default_rite: 10x-dev

# Rites available to all projects in this org
rites:
  - name: review
    version: ">=1.0.0"
  - name: 10x-dev
    version: ">=1.0.0"
  - name: custom-review
    source: org  # org-local rite

# Shared resources synced to all projects
shared:
  agents: true      # sync org agents to projects
  mena: true        # sync org mena to projects
  rules: true       # sync org rules to projects
  know: false       # opt-in per project

# CC managed-settings.json integration (reference only, not generated by Knossos)
cc_managed_settings:
  hint: "Deploy via MDM or /Library/Application Support/ClaudeCode/managed-settings.json"
```

### 4.4 Bootstrap Flow

```bash
# First time: register org and clone org config
ari org init autom8y --from git@github.com:autom8y/knossos-org-config.git

# Bootstrap new project with org defaults
ari init --org autom8y --rite review

# Sync org-level resources to current project
ari sync --scope org

# List org configuration
ari org show autom8y
```

### 4.5 Integration with CC Managed Settings

Knossos should NOT generate `managed-settings.json`. That is CC's domain. Instead:

1. **Document the mapping**: Provide an `ari org cc-config` command that outputs a recommended `managed-settings.json` based on org.yaml
2. **Hooks alignment**: Org-level hooks defined in Knossos should be documented as what IT should deploy via CC's `managed-settings.json`
3. **CLAUDE.md alignment**: Org-level CLAUDE.md content should live at `~/.claude/CLAUDE.md` (CC user scope) and be synced by `ari sync user` -- this is how CC naturally discovers org guidance

---

## 5. Comparison Matrix

### 5.1 Implementation Approaches

| Approach | Complexity | CC Alignment | Distribution | Risk |
|----------|-----------|-------------|-------------|------|
| **A: Org directory in XDG** (recommended) | Medium | High -- slots into existing hierarchy | Git clone + ari org init | Low -- additive, no breaking changes |
| **B: Org as special rite** | Low | Medium -- reuses existing rite infra | Embedded or registry | Medium -- overloads rite semantics |
| **C: Org as CC managed settings only** | Low | Highest -- pure delegation | MDM/IT deployment | High -- requires IT involvement, no rite/agent sharing |
| **D: Monorepo with nested CLAUDE.md** | Low | High -- CC native hierarchy | Git submodules/worktrees | Medium -- forces monorepo structure |
| **E: Remote rite registry (marketplace)** | High | Medium | HTTP registry, `ari install` | High -- premature for current stage |

### 5.2 Feature Coverage by Approach

| Feature | A (XDG org) | B (special rite) | C (CC managed) | D (monorepo) | E (registry) |
|---------|------------|------------------|-----------------|-------------|--------------|
| Shared rites | Yes | Partial | No | Via monorepo | Yes |
| Shared agents | Yes | Yes | No | Via nested dirs | Yes |
| Shared mena | Yes | Yes | No | Via nested dirs | Yes |
| Shared rules | Yes | No | Yes (CC native) | Via root rules | No |
| Security policies | Via CC | No | Yes | Via CC | No |
| Hook enforcement | Via CC | No | Yes | Via CC | No |
| Offline capability | Yes | Yes | Yes | Yes | No |
| Multi-org support | Yes | Complex | N/A | No | Yes |
| Versioning | Via git | Via manifest | N/A | Via git | Native |

---

## 6. Recommendations

### 6.1 Immediate (Stage 1 -- Internal)

**Do nothing new.** The current single-org model (KNOSSOS_HOME pointing to the repo checkout) is sufficient for internal use. Focus on:

1. **Document the existing hierarchy** in `.know/` and CLAUDE.md: explain how `$KNOSSOS_HOME`, user-level XDG paths, and project-level paths compose
2. **Ensure `ari init` is clean**: currently works for embedded rites; validate it works without KNOSSOS_HOME for external users
3. **Add org metadata to MEMORY.md seeds**: when bootstrapping team members, pre-populate org conventions

### 6.2 Near-term (Stage 2 -- Trusted External)

**Implement Approach A (Org directory in XDG):**

1. Add `ari org init {name}` command -- clones/creates org config directory
2. Add org tier to `SourceResolver.ResolveRite()` -- between user and knossos
3. Add `ari sync --scope org` -- syncs org resources to `~/.claude/` (user scope in CC terms)
4. Ship org template repo (`knossos-org-template`) as canonical starting point
5. Document CC managed-settings.json recommendations per org config

**Estimated effort**: 40-60 hours
- `ari org init/show/list` commands: 8-12 hours
- SourceResolver org tier: 4-6 hours
- Org sync pipeline: 12-16 hours
- Org template repo: 4-6 hours
- Documentation + testing: 12-20 hours

### 6.3 Long-term (Stage 3 -- Broader Availability)

**Build on Approach A toward Approach E:**

1. Remote org registry: `ari org init autom8y --from https://registry.knossos.dev/orgs/autom8y`
2. Rite marketplace integration: `ari rite install review --version 2.0`
3. Org-level CC managed-settings.json generator: `ari org cc-config --output managed-settings.json`
4. Enterprise features: audit trail for org config changes, compliance attestation per ADR

---

## 7. Anti-Patterns to Avoid

| Anti-Pattern | Why It's Tempting | Why It's Wrong |
|-------------|-------------------|----------------|
| **Reimplementing CC managed settings** | Full control over policy enforcement | CC already does this; duplication creates drift |
| **Org as a rite** | Reuses existing infrastructure | Rites have workflows, phases, agents; org config is structural, not procedural |
| **Global mutable org state** | Simple single-source | Multiple orgs become impossible; conflicts with CC's merge semantics |
| **Mandatory org membership** | Enforces consistency | Breaks single-developer and open-source use cases |
| **Org config in project repos** | Easy distribution | Defeats the purpose; every project has its own copy that drifts |

---

## 8. CC-Native Best Practices for Knossos Users (Today)

Even without new Knossos features, teams can implement org-level patterns using CC's existing infrastructure:

### 8.1 CLAUDE.md Hierarchy

```
~/.claude/CLAUDE.md                    # Org conventions (synced by ari sync user)
  ~/Code/project-a/.claude/CLAUDE.md   # Project-specific (generated by ari sync)
    ~/Code/project-a/.claude/CLAUDE.local.md  # Personal overrides (gitignored)
```

### 8.2 Managed Settings for Security

Deploy via MDM or file-based managed settings:

```json
{
  "permissions": {
    "deny": ["Bash(rm -rf *)", "Read(.env)", "Read(.env.*)"]
  }
}
```

> **Note**: Do NOT add a blanket `agent-guard` hook in managed settings without
> `--allow-path` flags. A pathless `ari hook agent-guard` unconditionally denies
> ALL write operations. Agent-guard enforcement is handled correctly by the
> per-agent materialization pipeline via manifest `hook_defaults` and agent
> `write-guard:` frontmatter. See `SPIKE-agent-guard-ledge-path-blockage.md`.

### 8.3 Shared Rules via Git

Maintain a shared rules repo synced to each project's `.claude/rules/`:

```bash
# In CI or developer setup script
cp -r org-config/rules/ .claude/rules/
```

### 8.4 User-Level Mena as Org Conventions

Knossos already syncs mena to `~/.claude/` via `ari sync user`. Content in `mena/guidance/standards/` (code-conventions.md, tech-stack-*.md) already serves as org-level guidance. Ensure these are well-maintained and comprehensive.

---

## 9. Follow-Up Actions

| # | Action | Priority | Depends On | Estimated Effort |
|---|--------|----------|-----------|-----------------|
| 1 | Document current hierarchy in `.know/conventions.md` | P1 | None | 2-3 hours |
| 2 | Validate `ari init` works without KNOSSOS_HOME for clean external bootstrap | P1 | None | 1-2 hours |
| 3 | Write ADR for org tier addition to source resolution | P2 | This spike accepted | 2-3 hours |
| 4 | Implement `ari org init` command skeleton | P2 | ADR accepted | 8-12 hours |
| 5 | Add org tier to SourceResolver | P2 | `ari org init` | 4-6 hours |
| 6 | Create org template repo (knossos-org-template) | P2 | `ari org init` | 4-6 hours |
| 7 | Document CC managed-settings.json recommendations | P1 | None | 2-3 hours |
| 8 | Ship org sync pipeline (`ari sync --scope org`) | P3 | Items 4-5 | 12-16 hours |

---

## 10. References

### Internal

- `internal/materialize/source/resolver.go` -- current 4-tier source resolution
- `internal/config/home.go` -- KNOSSOS_HOME and XDG resolution
- `internal/paths/paths.go` -- user-level directory paths
- `internal/materialize/userscope/sync.go` -- user scope sync pipeline
- `internal/cmd/initialize/init.go` -- `ari init` bootstrap flow
- `docs/strategy/ROADMAP-distribution-readiness.md` -- distribution stages
- `docs/decisions/ADR-0025-mena-scope.md` -- pipeline-targeted mena distribution
- `docs/decisions/ADR-0021-two-axis-context-model.md` -- two-axis context model
- `docs/MOONSHOT-agent-template-ecosystem.md` -- long-term ecosystem vision

### External

- [Claude Code Settings Documentation](https://code.claude.com/docs/en/settings) -- CC configuration hierarchy
- [Claude Code Admin Controls](https://www.anthropic.com/news/claude-code-on-team-and-enterprise) -- enterprise managed settings
- [AGENTS.md Standard](https://agents.md/) -- industry standard for AI agent configuration
- [AGENTS.md GitHub](https://github.com/agentsmd/agents.md) -- specification and examples
- [Cursor Rules Documentation](https://cursor.com/docs/context/rules) -- modular context rules pattern
- [Context Engineering for Developers](https://www.faros.ai/blog/context-engineering-for-developers) -- context engineering patterns
- [Claude Code Configuration Guide (eesel.ai)](https://www.eesel.ai/blog/claude-code-configuration) -- practical CC configuration guide
- [Claude Code Enterprise Security](https://www.mintmcp.com/blog/claude-code-security) -- enterprise security best practices
- [Claude Code Best Practices (Morph)](https://www.morphllm.com/claude-code-best-practices) -- 2026 productivity guide

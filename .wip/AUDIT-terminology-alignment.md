# Terminology Alignment Audit (SL-008)

> Deep adversarial audit of Legacy Terminology Deep Cleanse initiative
> Date: 2026-02-09
> Session: Continuation across 3 sessions

## Initiative Summary

| Metric | Value |
|--------|-------|
| **Starting refs** | 400+ across rites/ + mena/ |
| **Commits** | 9 terminology commits (pre-audit) |
| **Files changed** | 135+ unique files |
| **Pre-audit rites/ refs** | 58 |
| **Pre-audit mena/ refs** | 10 |

## Audit Agents Deployed

| Agent | Surface Area | Status |
|-------|-------------|--------|
| code-smeller #1 | All remaining rites/ + mena/ `team` refs | COMPLETE |
| code-smeller #2 | Go source, docs, templates (pack/roster/skeleton/CEM) | COMPLETE |
| context-engineer | Schema consistency, rite/pantheon precision, context quality | COMPLETE |

## Consolidated Findings

### CRITICAL (3 classes, blocking distribution)

| ID | Finding | Surface | Fix Agent |
|----|---------|---------|-----------|
| C-1 | `validation.sh` uses `source_team`/`target_team` — rejects valid HANDOFFs | 8 field refs | janitor-rites |
| C-2 | 20+ agent prompts say "handoff patterns to other teams" | 21 files across 8 rites | janitor-rites |
| C-3 | JSON schema `rite_name` pattern `^[a-z0-9]+-pack$` rejects ALL valid rite names | 1 schema file | janitor-go |

### HIGH (8 classes)

| ID | Finding | Surface | Fix Agent |
|----|---------|---------|-----------|
| H-1 | 10/11 orchestrators: "outside team's control" (forge is only correct one) | 10 files | janitor-rites |
| H-2 | 3 orchestrators: "affected team(s)" in Cross-Rite Protocol | 3 files | janitor-rites |
| H-3 | sprint-planner.md + code-smeller.md embed wrong HANDOFF field names | 2 files | janitor-rites |
| H-4 | `source_team`/`target_team` in 5 rite TODO files + sre/README.md | 6 files | janitor-rites |
| H-5 | `consultant-pack` phantom rite in handoff prepare code | 2 Go files | janitor-go |
| H-6 | `SchemaTeamManifest` + `validateTeamManifestStructure` in manifest code | 1 Go file | janitor-go |
| H-7 | `team` variable name throughout session create command | 1 Go file | janitor-go |
| H-8 | `Cross-Team Protocol` in agent archetype + template | 2 Go files | janitor-go |

### MEDIUM (4 classes)

| ID | Finding | Surface | Fix Agent |
|----|---------|---------|-----------|
| M-1 | forge rite-composition.md uses "teams" 5 times | 1 file | janitor-rites |
| M-2 | workflow-yaml-schema.md: "team workflow" (2 refs) | 1 file | janitor-rites |
| M-3 | docs/doctrine/rites/forge.md: "agent team" (7 refs) | 1 file | janitor-docs |
| M-4 | handoff-smoke-tests.md: 24 source_team + 12 "Notes for Target Team" | 1 file | janitor-docs |

### KEEP (confirmed legitimate, ~42 refs)

- Human teams: "product team", "engineering team", "executive team", "Backend Team"
- GitHub handles: "@security-team", "@api-team", "@platform-team", "@infra-team"
- CC settings tiers: "skeleton < project < team < user"
- Generic English: "team members", "team discussion", "team expertise", "team skills"
- Pricing: "team plan", "team tier"

### CANARY (need human judgment, 13 refs)

| ID | Finding | Files | Question |
|----|---------|-------|----------|
| CAN-1 | "outside team's control" — generic English vs rite reference | 10 orchestrators | DECIDED: Change to "rite's control" (forge precedent) |
| CAN-2 | "Multi-team coordination" complexity level | complexity-levels.md | Multi-rite or multi-human-team? |
| CAN-3 | "team workflows" in information-architect escalation | docs orchestrator | Rite workflows or human workflows? |
| CAN-4 | "another team" in doc-auditor | doc-auditor.md | Another rite or human team? |
| CAN-5 | `teams/` legacy directory still exists | entire directory | Deletion candidate? |

## Fix Agents Deployed

| Agent | Scope | Status |
|-------|-------|--------|
| janitor-rites | 40+ rite files: agents, orchestrators, TODOs, shared-templates, validation.sh | RUNNING |
| janitor-go | Go source: schema, handoff, manifest, session, archetype, template, tests | RUNNING |
| janitor-docs | docs/: doctrine, testing, design, playbooks, edge-cases | RUNNING |

## Terminology Decision Matrix (Canonical)

| Legacy Term | Replacement | When |
|-------------|------------|------|
| team (agent workflow) | **rite** | Organizational bundle |
| team (agent group) | **pantheon** | Collection of agents in a rite |
| team (human) | **team** (KEEP) | Human collaborators |
| team (pricing) | **team** (KEEP) | Plan/tier name |
| team (GitHub) | **team** (KEEP) | @handle |
| pack | **rite** | Everywhere |
| roster | **manifest** | Agent/resource listing |
| skeleton | **template/scaffold** | Starter structure |
| CEM | **sync/materialize** | Everywhere |
| state-mate | **moirai** | Session lifecycle agent |
| source_team/target_team | **source_rite/target_rite** | HANDOFF schema fields |
| Cross-Team | **Cross-Rite** | Section headers about rites |
| Target Team | **Target Rite** | Routing table headers |
| Notes for Target Team | **Notes for Target Rite** | HANDOFF artifact sections |

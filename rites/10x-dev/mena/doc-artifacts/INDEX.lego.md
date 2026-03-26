---
name: doc-artifacts
description: |
  PRD, TDD, ADR, and Test Plan templates for the 10x development workflow.
  Use when: writing requirements documents, creating technical designs, recording
  architecture decisions, or producing QA test artifacts.
  Triggers: PRD, TDD, ADR, test plan, test case, test summary, requirements document,
  technical design, architecture decision, impact assessment, acceptance criteria.
---

# Development Artifact Templates

> Canonical templates for the four artifact types produced across 10x-dev workflow phases.

## Artifact Types

| Artifact | Owner Agent | Produced In Phase | Key Section |
|----------|-------------|-------------------|-------------|
| **PRD** | Requirements Analyst | Requirements | Impact Assessment (routes workflow) |
| **TDD** | Architect | Architecture | System Design + API contracts |
| **ADR** | Architect | Architecture | Decision + Alternatives Considered |
| **Test Case** | QA Adversary | QA | Steps + Pass/Fail criteria |
| **Test Summary** | QA Adversary | QA | Release recommendation |

## When to Load Each Template

| Template | Load When |
|----------|-----------|
| `prd-template.lego.md` | Starting feature scope, defining acceptance criteria, setting `impact:` flag |
| `tdd-template.lego.md` | Designing system architecture, specifying APIs or data models |
| `adr-template.lego.md` | Recording a discrete architectural decision with alternatives |
| `test-templates.lego.md` | Writing individual test cases OR producing a QA release summary |

## PRD Impact Flag

The PRD `impact:` field controls downstream routing. Set it correctly — it gates specialist activation.

| Value | Meaning | Effect |
|-------|---------|--------|
| `low` | No high-risk categories apply | Standard workflow path |
| `high` | One or more categories apply | Activates impact-specific specialists |

Impact categories: `security`, `performance`, `compliance`, `data-migration`, `breaking-change`.

## Companion Reference

| File | When to Load |
|------|--------------|
| `prd-template.lego.md` | Full PRD template with impact category reference |
| `tdd-template.lego.md` | Full TDD template with architecture diagram and API sections |
| `adr-template.lego.md` | Full ADR template with alternatives format |
| `test-templates.lego.md` | Test case + test summary templates |

## Related Skills

- `10x-workflow` skill — Phase sequencing and agent coordination
- `conventions` skill — Project-wide naming and formatting standards

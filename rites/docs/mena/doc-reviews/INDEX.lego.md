---
name: doc-reviews
description: "Documentation audit, review, and information architecture templates. Use when: auditing doc health, reviewing doc accuracy, planning doc structure, creating content briefs, migrating documentation. Triggers: doc audit, documentation review, information architecture, content brief, staleness, doc migration."
---

# Documentation Reviews & Audits

> Templates for documentation health assessment, accuracy validation, and structural planning.

## Purpose

Provides structured templates for documentation lifecycle management: health audits detecting staleness and redundancy, accuracy reviews validating docs against code, information architecture design, migration planning, and content briefs for new documentation.

## Template Catalog

| Template | Purpose | Agent |
|----------|---------|-------|
| [Audit Report](templates/audit-report.md) | Documentation health: staleness, orphans, redundancy, gaps | doc-auditor |
| [Review Report](templates/review-report.md) | Accuracy validation against actual code behavior | doc-reviewer |
| [Information Architecture](templates/information-architecture.md) | Taxonomy, directory structure, naming conventions | information-architect |
| [Migration Plan](templates/migration-plan.md) | Phased doc restructuring: moves, merges, retirements | information-architect |
| [Content Brief](templates/content-brief.md) | Specification for new documentation to create | doc-lead |

## When to Use Each Template

| Scenario | Template |
|----------|----------|
| Assessing documentation health | Audit Report |
| Validating doc accuracy against code | Review Report |
| Designing documentation structure | Information Architecture |
| Restructuring existing docs | Migration Plan |
| Commissioning new documentation | Content Brief |
| Finding stale or orphaned docs | Audit Report |

## Quality Gates Summary

| Template | Gate Criteria |
|----------|---------------|
| **Audit Report** | All directories scanned, staleness scored, redundancy clusters identified |
| **Review Report** | Issues severity-graded, code references validated, approval status selected |
| **Information Architecture** | Taxonomy complete, directory annotated, naming conventions defined |
| **Migration Plan** | Every doc mapped to action, cross-references listed, retirement verified |
| **Content Brief** | Location follows IA conventions, scope has exclusions, priority rationalized |

## Progressive Disclosure

- [audit-report.md](templates/audit-report.md) - Documentation health audit
- [review-report.md](templates/review-report.md) - Accuracy validation
- [information-architecture.md](templates/information-architecture.md) - Taxonomy and structure
- [migration-plan.md](templates/migration-plan.md) - Phased restructuring
- [content-brief.md](templates/content-brief.md) - New doc specification

---
name: doc-ecosystem
description: "Ecosystem and hygiene templates for CEM sync, migration, compatibility, and code quality workflows. Use when: planning CEM migrations, validating compatibility across satellites, analyzing code smells, designing refactoring sequences, or documenting system-level changes. Triggers: gap analysis, context design, migration runbook, compatibility report, smell report, refactoring plan, CEM sync, satellite migration, code cleanup."
---

# Documentation: Ecosystem & Hygiene

> Templates for ecosystem-level documentation and code hygiene workflows.

## Purpose

This skill provides templates for ecosystem-level documentation: CEM/satellite synchronization, migration planning, compatibility validation, and code hygiene workflows. These templates support cross-repository changes, breaking changes, and systematic code cleanup.

## Core Principles

| Principle | Description |
|-----------|-------------|
| **Ecosystem Awareness** | Changes to CEM affect all satellites. Ensure backward compatibility and clear migration paths. |
| **Hygiene Before Features** | Refactoring and cleanup are first-class work. Templates support smell detection and phased cleanup. |
| **Migration Safety** | Breaking changes require runbooks, rollback procedures, and compatibility matrices. |

## Template Categories

### Ecosystem Change Templates

| Template | Purpose | Agent |
|----------|---------|-------|
| [Gap Analysis](templates/gap-analysis.md) | Issue diagnosis for CEM/satellite problems | ecosystem-analyst |
| [Context Design](templates/context-design.md) | Technical design for ecosystem changes | context-architect |
| [Migration Runbook](templates/migration-runbook.md) | Satellite owner migration guide | documentation-engineer |
| [Compatibility Report](templates/compatibility-report.md) | Cross-satellite validation results | compatibility-tester |

### Code Hygiene Templates

| Template | Purpose | Agent |
|----------|---------|-------|
| [Smell Report](templates/smell-report.md) | Code smell catalog and cleanup priorities | ecosystem-analyst |
| [Refactoring Plan](templates/refactoring-plan.md) | Phased refactoring sequence | context-architect |

## When to Use Each Template

| Scenario | Template |
|----------|----------|
| CEM sync failing | Gap Analysis |
| Planning ecosystem change | Context Design |
| Shipping breaking change | Migration Runbook |
| Validating before release | Compatibility Report |
| Cleaning up codebase | Smell Report + Refactoring Plan |

## Quality Gates Summary

| Template | Gate Criteria |
|----------|---------------|
| **Gap Analysis** | Clear reproduction, root cause identified, success criteria testable |
| **Context Design** | Backward compatibility assessed, migration path documented, tests defined |
| **Migration Runbook** | Step-by-step instructions, rollback tested, compatibility matrix complete |
| **Compatibility Report** | All satellites tested, defects prioritized, recommendation justified |
| **Smell Report** | Evidence-based, severity assigned, cleanup priority established |
| **Refactoring Plan** | Phases sequenced by risk, invariants defined, commit scope clear |

## Progressive Disclosure

### Ecosystem Templates
- [gap-analysis.md](templates/gap-analysis.md) - Issue diagnosis template
- [context-design.md](templates/context-design.md) - Technical design template
- [migration-runbook.md](templates/migration-runbook.md) - Migration guide template
- [compatibility-report.md](templates/compatibility-report.md) - Validation template

### Hygiene Templates
- [smell-report.md](templates/smell-report.md) - Code smell catalog template
- [refactoring-plan.md](templates/refactoring-plan.md) - Phased cleanup template

## Related Skills

- [ecosystem-ref](../ecosystem-ref/SKILL.md) - CEM/skeleton/roster quick reference
- [documentation](../../../../user-skills/documentation/documentation/SKILL.md) - Core artifact templates (PRD, TDD, ADR)
- [standards](../../../../user-skills/documentation/standards/SKILL.md) - Code conventions and repository structure

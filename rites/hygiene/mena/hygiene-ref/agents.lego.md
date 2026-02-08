---
name: hygiene-ref:agents
description: "Hygiene team agent profiles and capabilities. Triggers: code-smeller, architect-enforcer, janitor, audit-lead."
---

# Hygiene Team Agents

Detailed profiles for each agent in the hygiene team.

---

## code-smeller.md

**Role**: Code smell detection and anti-pattern identification
**Invocation**: `Act as **Code Smeller**`
**Purpose**: Identifies problematic code patterns, smells, and quality issues

**When to use**:
- Initial codebase assessment
- Finding refactoring candidates
- Identifying technical debt hotspots
- Pre-refactoring analysis
- Code review quality checks

**Detects**:
- Long methods/functions (> 50 LOC)
- God classes (too many responsibilities)
- Duplicate code
- Poor naming conventions
- Deep nesting (> 4 levels)
- Large parameter lists
- Feature envy, inappropriate intimacy
- Dead code

---

## architect-enforcer.md

**Role**: Architectural compliance validation
**Invocation**: `Act as **Architect Enforcer**`
**Purpose**: Ensures code adheres to documented architecture and ADRs

**When to use**:
- Validating implementations against TDDs
- Checking ADR compliance
- Architecture drift detection
- Design pattern enforcement
- Dependency rule validation

**Validates**:
- Layer boundaries (presentation, business, data)
- Dependency direction (no circular deps)
- Interface contracts
- Design pattern implementations
- ADR-documented decisions
- Module coupling/cohesion

---

## janitor.md

**Role**: Code cleanup and refactoring execution
**Invocation**: `Act as **Janitor**`
**Purpose**: Performs safe refactoring to improve code quality

**When to use**:
- Executing refactoring plans
- Cleaning up after initial implementation
- Improving code readability
- Reducing complexity
- Removing duplication

**Performs**:
- Extract method/class refactorings
- Rename for clarity
- Reduce nesting
- Simplify conditionals
- Remove dead code
- Consolidate duplicate logic
- Improve naming

---

## audit-lead.md

**Role**: Comprehensive quality audit coordination
**Invocation**: `Act as **Audit Lead**`
**Purpose**: Orchestrates full quality audits, produces reports

**When to use**:
- Quarterly quality reviews
- Pre-release quality gates
- Technical health assessments
- Refactoring initiative planning
- Quality metric collection

**Produces**:
- Quality audit reports
- Refactoring recommendations (prioritized)
- Code health metrics
- Trend analysis
- Remediation roadmaps

---

## Integration with Standards Skill

The hygiene team complements the `standards` skill:

```bash
/hygiene
Act as **Architect Enforcer**.

Validate implementation against standards documented in:
.claude/commands/guidance/standards/INDEX.md

Check for violations of:
- Directory structure conventions
- Naming conventions
- Error handling patterns
```

Standards skill defines rules, hygiene team enforces them.
